package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"k8s.io/klog/v2"
)

type TydsClient struct {
	Username        string
	Password        string
	BaseURL         string
	SnapshotCount   int
	Token           string
	TokenExpiration int64
	IP              string
	Port            int
}

func NewTydsClient(hostip string, port int, username string, password string) *TydsClient {
	klog.Infof("Get client args: hostip=%s, port=%d, username=%s, password=%s", hostip, port, username, password)
	client := &TydsClient{
		Username:      username,
		Password:      base64.StdEncoding.EncodeToString([]byte(password)),
		BaseURL:       fmt.Sprintf("http://%s:%d/api", hostip, port),
		SnapshotCount: 999,
		Token:         "",
		IP:            getLocalIP(),
	}
	return client
}

func (c *TydsClient) GetToken() (string, error) {
	if c.Token != "" && time.Now().Unix() < c.TokenExpiration {
		// Token is not expired, directly return the existing Token
		return c.Token, nil
	}

	// Token has expired or has not been obtained before,
	// retrieving the Token again
	token, err := c.Login()
	if err != nil {
		klog.Errorf("Failed to obtain token: %v", err)
		return "", err
	}

	// expire time set to 5 hours, less than actual 6.5 hours
	c.TokenExpiration = time.Now().Unix() + 300*60
	c.Token = token
	return c.Token, nil
}

func (c *TydsClient) SendHTTPAPI(url string, params interface{}, method string) (interface{}, error) {
	jsonParams, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON params: %v", err)
	}

	fullURL := fmt.Sprintf("%s/%s", c.BaseURL, url)
	request, err := http.NewRequest(method, fullURL, bytes.NewReader(jsonParams))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}
	token, err := c.GetToken()
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", token)
	request.Header.Set("Content-Type", "application/json")

	klog.Infof("Sending %s request to: %s with body: %s", method, fullURL, string(jsonParams))

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %v", err)
	}

	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			err = fmt.Errorf("failed to close response body: %v", closeErr)
		}
	}()

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	klog.Infof("Received response: %s", string(responseData))

	var jsonResponse map[string]interface{}
	err = json.Unmarshal(responseData, &jsonResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", response.StatusCode)
	}

	errorCode, ok := jsonResponse["code"].(string)
	if !ok || errorCode != "0000" {
		return nil, fmt.Errorf("API request failed with error: %v", jsonResponse)
	}

	return jsonResponse["data"], nil
}

type LoginResponse struct {
	Code string `json:"code"`
	Data struct {
		UserID     int    `json:"userId"`
		Token      string `json:"token"`
		RoleNameEn string `json:"role_name_en"`
		RoleNameZh string `json:"role_name_zh"`
	} `json:"data"`
	Message string `json:"message"`
}

func (c *TydsClient) Login() (string, error) {
	// 构建请求参数
	params := map[string]interface{}{
		"REMOTE_ADDR": c.IP,
		"username":    c.Username,
		"password":    c.Password,
	}
	jsonParams, err := json.Marshal(params)
	if err != nil {
		klog.Errorf("Failed to marshal JSON params: %v", err)
		return "", err
	}

	// 发起登录请求
	url := fmt.Sprintf("%s/auth/login/", c.BaseURL)
	responseData, err := c.doRequest("POST", url, jsonParams)
	if err != nil {
		klog.Errorf("Failed to make login request: %v", err)
		return "", err
	}

	// 将响应内容添加到日志中
	klog.Infof("Login response: %s", string(responseData))

	// 解析登录响应
	var jsonResponse LoginResponse
	err = json.Unmarshal(responseData, &jsonResponse)
	if err != nil {
		klog.Errorf("Failed to parse login response: %v", err)
		return "", err
	}

	// 提取 token
	token := jsonResponse.Data.Token
	if token == "" {
		err := fmt.Errorf("authentication token not found in response")
		klog.Error(err)
		return "", err
	}

	return token, nil
}

func (c *TydsClient) doRequest(method string, url string, data []byte) ([]byte, error) {
	request, err := http.NewRequest(method, url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")

	// Print request information
	klog.Infof("Request: Method=%s, URL=%s, Body=%s", method, url, string(data))

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			// Handle error if closing the response body fails
			klog.Errorf("Failed to close response body: %v", err)
		}
	}()
	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// Print response information
	klog.Infof("Response: Status=%s, Body=%s", response.Status, string(responseData))

	return responseData, nil
}

func getLocalIP() string {
	// 获取所有网络接口的信息
	interfaces, err := net.Interfaces()
	if err != nil {
		// 处理错误，根据你的需求返回适当的默认值或错误信息
		return ""
	}

	// 遍历所有网络接口，查找非回环接口的第一个非零 IP 地址
	for _, iface := range interfaces {
		// 排除回环接口和无效接口
		if iface.Flags&net.FlagLoopback == 0 && iface.Flags&net.FlagUp != 0 {
			addrs, err := iface.Addrs()
			if err != nil {
				// 处理错误，根据你的需求返回适当的默认值或错误信息
				return ""
			}

			// 遍历接口地址，查找第一个非回环 IPv4 或 IPv6 地址
			for _, addr := range addrs {
				ipNet, ok := addr.(*net.IPNet)
				if ok && !ipNet.IP.IsLoopback() {
					ip := ipNet.IP
					// 过滤 IPv6 链路本地地址
					if !ip.IsLinkLocalUnicast() {
						// 返回第一个非回环 IPv4 或 IPv6 地址的字符串表示
						return ip.String()
					}
				}
			}
		}
	}

	// 未找到合适的 IP 地址，根据你的需求返回适当的默认值或错误信息
	return ""
}

//
// func main() {
// 	service := NewTydsClient("hostname", 1234, "username", "password")
// 	// Use the service to make requests
// }

// GetPools 获取所有存储池的列表。
func (c *TydsClient) GetPools() ([]map[string]interface{}, error) {
	url := "pool/pool/"
	response, err := c.SendHTTPAPI(url, nil, "GET")
	if err != nil {
		// 返回错误而不是 panic
		return nil, fmt.Errorf("failed to get pools: %v", err)
	}

	poolList, ok := response.(map[string]interface{})["poolList"].([]interface{})
	if !ok {
		// 返回错误而不是 panic
		return nil, fmt.Errorf("poolList not found in the response")
	}

	// 将 poolList 转换为 []map[string]interface{}
	var pools []map[string]interface{}
	for _, item := range poolList {
		pool, ok := item.(map[string]interface{})
		if !ok {
			// 如果某个元素无法转换，跳过该元素
			continue
		}
		pools = append(pools, pool)
	}

	return pools, nil
}

func (c *TydsClient) GetVolume(volID string) (interface{}, error) {
	url := fmt.Sprintf("block/blocks/%s", volID)
	response, err := c.SendHTTPAPI(url, nil, "GET")
	if err != nil {
		// Handle API request failure
		return nil, err
	}

	volume, ok := response.(map[string]interface{})
	if !ok {
		// Handle invalid response format
		return nil, fmt.Errorf("unexpected response format")
	}

	return volume, nil
}

func (c *TydsClient) GetVolumes() []map[string]interface{} {
	url := "block/blocks"
	response, err := c.SendHTTPAPI(url, nil, "GET")
	if err != nil {
		// 处理 API 请求失败
		panic(err)
	}

	volList, ok := response.(map[string]interface{})["blockList"].([]interface{})
	if !ok {
		// 处理响应中找不到 blockList
		panic("blockList not found in the response")
	}

	volumes := make([]map[string]interface{}, len(volList))
	for i, vol := range volList {
		volumes[i] = vol.(map[string]interface{})
	}

	return volumes
}

func (c *TydsClient) CreateVolume(volName string, size int, poolName string, stripeSize string) (string, error) {
	url := "block/blocks/"
	params := map[string]interface{}{
		"blockName": volName,
		"sizeMB":    size,
		"poolName":  poolName,
		"stripSize": stripeSize,
	}
	_, err := c.SendHTTPAPI(url, params, "POST")
	if err != nil {
		return "", fmt.Errorf("failed to create volume: %v", err)
	}
	return volName, nil
}

func (c *TydsClient) DeleteVolume(volID int) error {
	params := map[string]interface{}{
		"id": []int{volID},
	}

	url := "block/recycle/"
	_, err := c.SendHTTPAPI(url, params, "POST")
	if err != nil {
		// 发生错误时返回错误信息
		return fmt.Errorf("failed to delete volume: %v", err)
	}
	// 没有错误时返回 nil
	return nil
}

// ExtendVolume 在给定的存储池中扩展指定的卷。
func (c *TydsClient) ExtendVolume(volName string, poolName string, sizeMB int64) error {
	url := fmt.Sprintf("block/blocks/%s/", volName)
	params := map[string]interface{}{
		"blockName": volName,
		"sizeMB":    sizeMB,
		"poolName":  poolName,
	}

	_, err := c.SendHTTPAPI(url, params, "PUT")
	if err != nil {
		// 返回错误而不是直接触发 panic
		return fmt.Errorf("failed to extend volume: %v", err)
	}

	// 如果一切顺利，返回 nil 表示没有错误发生
	return nil
}

func (c *TydsClient) CreateCloneVolume(poolName, blockName, blockID, targetPoolName, targetPoolID, targetBlockName string) {
	params := map[string]interface{}{
		"poolName":           poolName,
		"blockName":          blockName,
		"blockId":            blockID,
		"copyType":           0,
		"metapoolName":       "NULL",
		"targetMetapoolName": "NULL",
		"targetPoolName":     targetPoolName,
		"targetPoolId":       targetPoolID,
		"targetBlockName":    targetBlockName,
	}
	url := "block/block/copy/"
	_, err := c.SendHTTPAPI(url, params, "POST")
	if err != nil {
		// Handle API request failure
		panic(err)
	}
}

func (c *TydsClient) GetSnapshot(volumeID string) []interface{} {
	url := "block/snapshot?pageNumber=1"
	if volumeID != "" {
		url += fmt.Sprintf("&blockId=%s", volumeID)
	}
	url += "&pageSize=%d"

	response, err := c.SendHTTPAPI(fmt.Sprintf(url, c.SnapshotCount), nil, "GET")
	if err != nil {
		// Handle API request failure
		panic(err)
	}

	total, ok := response.(map[string]interface{})["total"].(float64)
	if !ok {
		// Handle total not found in the response
		panic("total not found in the response")
	}

	if c.SnapshotCount < int(total) {
		c.SnapshotCount = int(total)
		response, err = c.SendHTTPAPI(fmt.Sprintf(url, c.SnapshotCount), nil, "GET")
		if err != nil {
			// Handle API request failure
			panic(err)
		}
	}

	snapshotList, ok := response.(map[string]interface{})["snapShotList"].([]interface{})
	if !ok {
		// Handle snapShotList not found in the response
		panic("snapShotList not found in the response")
	}

	return snapshotList
}
func (c *TydsClient) CreateSnapshot(name string, volumeID string, comment string) error {
	url := "block/snapshot/"
	params := map[string]interface{}{
		"sourceBlock":  volumeID,
		"snapShotName": name,
		"remark":       comment,
	}
	_, err := c.SendHTTPAPI(url, params, "POST")
	return err
}

func (c *TydsClient) DeleteSnapshot(snapshotID string) error {
	url := fmt.Sprintf("block/snapshot/%s/", snapshotID)
	_, err := c.SendHTTPAPI(url, nil, "DELETE")
	return err
}

func (c *TydsClient) CreateVolumeFromSnapshot(volumeName string, poolName string, snapshotName string, sourceVolumeName string, sourcePoolName string) error {
	url := "block/clone/"
	params := map[string]interface{}{
		"cloneBlockName":     volumeName,
		"targetPoolName":     poolName,
		"snapName":           snapshotName,
		"blockName":          sourceVolumeName,
		"poolName":           sourcePoolName,
		"targetMetapoolName": "NULL",
	}
	_, err := c.SendHTTPAPI(url, params, "POST")
	return err
}

func (c *TydsClient) GetCloneProgress(volumeID string, volumeName string) (interface{}, error) {
	url := "block/clone/progress/"
	params := map[string]interface{}{
		"blockId":   volumeID,
		"blockName": volumeName,
	}
	return c.SendHTTPAPI(url, params, "POST")
}

func (c *TydsClient) GetCopyProgress(blockID string, blockName string, targetBlockName string) (interface{}, error) {
	url := "block/block/copyprogress/"
	params := map[string]interface{}{
		"blockId":         blockID,
		"blockName":       blockName,
		"targetBlockName": targetBlockName,
	}
	return c.SendHTTPAPI(url, params, "GET")
}

func (c *TydsClient) CreateInitiatorGroup(groupName string, client []map[string]string) error {
	url := "iscsi/service-group/"
	params := map[string]interface{}{
		"group_name": groupName,
		"service":    client,
		"chap_auth":  0,
		"mode":       "ISCSI",
	}
	_, err := c.SendHTTPAPI(url, params, "POST")
	return err
}

func (c *TydsClient) DeleteInitiatorGroup(groupID string) error {
	url := fmt.Sprintf("iscsi/service-group/?group_id=%s", groupID)
	_, err := c.SendHTTPAPI(url, nil, "DELETE")
	return err
}

func (c *TydsClient) GetInitiatorList() (interface{}, error) {
	url := "iscsi/service-group/"
	return c.SendHTTPAPI(url, nil, "GET")
}

func (c *TydsClient) GetTarget() (interface{}, error) {
	url := "/host/host/"
	return c.SendHTTPAPI(url, nil, "GET")
}

func (c *TydsClient) CreateTarget(groupName string, targetList []string, volsInfo interface{}) (interface{}, error) {
	url := "iscsi/target/"
	params := map[string]interface{}{
		"group_name":  groupName,
		"chap_auth":   0,
		"write_cache": 1,
		"hostName":    strings.Join(targetList, ","),
		"block":       volsInfo,
	}
	return c.SendHTTPAPI(url, params, "POST")
}

func (c *TydsClient) DeleteTarget(targetName string) (interface{}, error) {
	url := fmt.Sprintf("iscsi/target/?targetIqn=%s", targetName)
	return c.SendHTTPAPI(url, nil, "DELETE")
}

func (c *TydsClient) ModifyTarget(targetName string, targetList []string, volInfo interface{}) (interface{}, error) {
	url := "iscsi/target/"
	params := map[string]interface{}{
		"targetIqn": targetName,
		"chap_auth": 0,
		"hostName":  targetList,
		"block":     volInfo,
	}
	return c.SendHTTPAPI(url, params, "PUT")
}

func (c *TydsClient) GetInitiatorTargetConnections() ([]interface{}, error) {
	url := "iscsi/target/"
	res, err := c.SendHTTPAPI(url, nil, "GET")
	if err != nil {
		return nil, err
	}
	targetListInterface, ok := res.(map[string]interface{})["target_list"].([]interface{})
	if !ok {
		return nil, errors.New("target_list not found or has invalid type")
	}
	targetList := make([]interface{}, len(targetListInterface))
	copy(targetList, targetListInterface)
	return targetList, nil
}

func (c *TydsClient) GenerateConfig(targetName string) error {
	url := "iscsi/target-config/"
	params := map[string]interface{}{
		"targetName": targetName,
	}
	_, err := c.SendHTTPAPI(url, params, "POST")
	return err
}

func (c *TydsClient) RestartService(hostName string) error {
	url := "iscsi/service/restart/"
	params := map[string]interface{}{
		"hostName": hostName,
	}
	_, err := c.SendHTTPAPI(url, params, "POST")
	return err
}
