package sanity

import (
	"fmt"
	"os"
	"path"
	"testing"

	"toyou-csi/pkg/driver"

	"github.com/kubernetes-csi/csi-test/pkg/sanity"
)

var defaultStdParams = map[string]string{
	"pool":     "Pool-1",
	"user":     "toyoucsi",
	"password": "Toyou@123",
}

func TestSanity(t *testing.T) {
	// Set up variables
	tmpDir := "/tmp/csi"
	initiatorsDir := "/etc/iscsi"
	initiatorsFile := fmt.Sprintf("%s/initiatorname.iscsi", initiatorsDir)
	hostNameFile := "/etc/hostname"
	endpoint := fmt.Sprintf("unix://%s/csi.sock", tmpDir)
	mountPath := path.Join(tmpDir, "mount")
	stagePath := path.Join(tmpDir, "stage")
	driverName := "testDriver"
	node := "testNode"

	_, err := os.Stat(tmpDir)
	if err == nil {
		err = os.RemoveAll(tmpDir)
		if err != nil {
			t.Fatalf("Remove sock dir failed %s: %v", tmpDir, err)
		}
	}

	err = os.Mkdir(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create sanity temp working dir %s: %v", tmpDir, err)
	}

	defer func() {
		// Clean up tmp dir
		if err = os.RemoveAll(tmpDir); err != nil {
			t.Fatalf("Failed to clean up sanity temp working dir %s: %v", tmpDir, err)
		}
	}()

	var exist = true
	if _, err := os.Stat(hostNameFile); os.IsNotExist(err) {
		exist = false
	}

	if exist == false {
		f, _ := os.Create(hostNameFile)
		_, _ = f.WriteString("test-host")
	}

	exist = true
	if _, err := os.Stat(initiatorsDir); os.IsNotExist(err) {
		exist = false
	}

	if exist == false {
		err := os.Mkdir(initiatorsDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create sanity temp working dir %s: %v", tmpDir, err)
		}
		f, _ := os.Create(initiatorsFile)
		_, _ = f.WriteString("InitiatorName=iqn.1994-05.com.redhat:373363268f6")
	}

	TydsDriver.InitDiskDriver(diskDriverInput)
	// run your driver
	go driver.Run()

	cfg := &sanity.Config{
		StagingPath:          stagePath,
		TargetPath:           mountPath,
		Address:              endpoint,
		TestVolumeParameters: defaultStdParams,
	}

	// Now call the test suite
	sanity.Test(t, cfg)
}
