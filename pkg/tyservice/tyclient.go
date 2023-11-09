package tyservice

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type TydsClient struct {
	Username        string
	Password        string
	BaseURL         string
	SnapshotCount   int
	Token           string
	TokenExpiration int64
	IP              string
}

func NewTydsClient(hostname string, port int, username string, password string) *TydsClient {
	client := &TydsClient{
		Username:      username,
		Password:      base64.StdEncoding.EncodeToString([]byte(password)),
		BaseURL:       fmt.Sprintf("https://%s:%d/api", hostname, port),
		SnapshotCount: 999,
		Token:         "",
		IP:            getLocalIP(),
	}
	return client
}

func (c *TydsClient) GetToken() string {
	if c.Token != "" && time.Now().Unix() < c.TokenExpiration {
		// Token is not expired, directly return the existing Token
		return c.Token
	}

	// Token has expired or has not been obtained before,
	// retrieving the Token again
	c.Token = c.Login()
	// expire time set to 5 hours, less than actual 6.5 hours
	c.TokenExpiration = time.Now().Unix() + 300*60
	return c.Token
}

func (c *TydsClient) SendHTTPAPI(url string, params interface{}, method string) (interface{}, error) {
	jsonParams, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	fullURL := fmt.Sprintf("%s/%s", c.BaseURL, url)
	request, err := http.NewRequest(method, fullURL, bytes.NewReader(jsonParams))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", c.GetToken())
	request.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(response.Body)

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var jsonResponse map[string]interface{}
	err = json.Unmarshal(responseData, &jsonResponse)
	if err != nil {
		return nil, err
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

func (c *TydsClient) Login() string {
	params := map[string]interface{}{
		"REMOTE_ADDR": c.IP,
		"username":    c.Username,
		"password":    c.Password,
	}
	jsonParams, _ := json.Marshal(params)

	url := fmt.Sprintf("%s/auth/login/", c.BaseURL)
	responseData, err := c.doRequest("POST", url, jsonParams)
	if err != nil {
		// Handle login request failure
		panic(err)
	}

	var jsonResponse map[string]interface{}
	err = json.Unmarshal(responseData, &jsonResponse)
	if err != nil {
		// Handle response parsing failure
		panic(err)
	}

	token, ok := jsonResponse["token"].(string)
	if !ok {
		// Handle token not found in the response
		panic("Authentication token not found")
	}

	return token
}

func (c *TydsClient) doRequest(method string, url string, data []byte) ([]byte, error) {
	request, err := http.NewRequest(method, url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(response.Body)

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return responseData, nil
}

func getLocalIP() string {
	// Implement your logic to retrieve the local IP address here
	return ""
}

//
// func main() {
// 	client := NewTydsClient("hostname", 1234, "username", "password")
// 	// Use the client to make requests
// }

func (c *TydsClient) GetPools() []interface{} {
	url := "pool/pool/"
	response, err := c.SendHTTPAPI(url, nil, "GET")
	if err != nil {
		// Handle API request failure
		panic(err)
	}

	poolList, ok := response.(map[string]interface{})["poolList"].([]interface{})
	if !ok {
		// Handle poolList not found in the response
		panic("poolList not found in the response")
	}

	return poolList
}

func (c *TydsClient) GetVolumes() []interface{} {
	url := "block/block"
	response, err := c.SendHTTPAPI(url, nil, "GET")
	if err != nil {
		// Handle API request failure
		panic(err)
	}

	volList, ok := response.(map[string]interface{})["blockList"].([]interface{})
	if !ok {
		// Handle blockList not found in the response
		panic("blockList not found in the response")
	}

	return volList
}

func (c *TydsClient) CreateVolume(volName string, size int, poolName string, stripeSize int) interface{} {
	url := "block/block/"
	params := map[string]interface{}{
		"blockName": volName,
		"sizeMB":    size,
		"poolName":  poolName,
		"stripSize": stripeSize,
	}
	response, err := c.SendHTTPAPI(url, params, "POST")
	if err != nil {
		// Handle API request failure
		panic(err)
	}

	return response
}

func (c *TydsClient) DeleteVolume(volID string) {
	url := fmt.Sprintf("block/block/forcedelete/?id=%s", volID)
	_, err := c.SendHTTPAPI(url, nil, "DELETE")
	if err != nil {
		// Handle API request failure
		panic(err)
	}
}

func (c *TydsClient) ExtendVolume(volName string, poolName string, sizeMB int) {
	url := fmt.Sprintf("block/block/%s/", volName)
	params := map[string]interface{}{
		"blockName": volName,
		"sizeMB":    sizeMB,
		"poolName":  poolName,
	}
	_, err := c.SendHTTPAPI(url, params, "PUT")
	if err != nil {
		// Handle API request failure
		panic(err)
	}
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

func (c *TydsClient) CreateInitiatorGroup(groupName string, client string) error {
	url := "iscsi/client-group/"
	params := map[string]interface{}{
		"group_name": groupName,
		"client":     client,
		"chap_auth":  0,
		"mode":       "ISCSI",
	}
	_, err := c.SendHTTPAPI(url, params, "POST")
	return err
}

func (c *TydsClient) DeleteInitiatorGroup(groupID string) error {
	url := fmt.Sprintf("iscsi/client-group/?group_id=%s", groupID)
	_, err := c.SendHTTPAPI(url, nil, "DELETE")
	return err
}

func (c *TydsClient) GetInitiatorList() (interface{}, error) {
	url := "iscsi/client-group/"
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
	targetListInterface, ok := res.(map[string]interface{})["targetList"].([]interface{})
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
