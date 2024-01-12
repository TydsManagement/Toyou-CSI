package driver

// 创建一个虚拟的 TydsManager，用于模拟 CreateVolume 函数中的 TydsManager
type FakeTydsManager struct {
}

func (m *FakeTydsManager) CreateVolume(name string, size int) (string, error) {
	// 返回虚拟的卷 ID
	return "fake-volume-id", nil
}

func (m *FakeTydsManager) FindSnapshot(snapId string) (*interface{}, error) {
	// TODO implement me
	panic("implement me")
}

func (m *FakeTydsManager) FindSnapshotByName(snapName string) (*interface{}, error) {
	// TODO implement me
	panic("implement me")
}

func (m *FakeTydsManager) CreateSnapshot(snapName string, volID string) error {
	// TODO implement me
	panic("implement me")
}

func (m *FakeTydsManager) DeleteSnapshot(snapID string) error {
	// TODO implement me
	panic("implement me")
}

func (m *FakeTydsManager) CreateVolumeFromSnapshot(volName string, snapId string) (volId string, err error) {
	// TODO implement me
	panic("implement me")
}

func (m *FakeTydsManager) FindVolume(volId string) (interface{}, error) {
	// TODO implement me
	panic("implement me")
}

func (m *FakeTydsManager) FindVolumeByName(volName string) map[string]interface{} {
	// TODO implement me
	panic("implement me")
}

func (m *FakeTydsManager) DeleteVolume(volId string) (err error) {
	// TODO implement me
	panic("implement me")
}

func (m *FakeTydsManager) ListVolumes() []map[string]interface{} {
	// TODO implement me
	panic("implement me")
}

func (m *FakeTydsManager) AttachVolume(volId string, instanceId string) (err error) {
	// TODO implement me
	panic("implement me")
}

func (m *FakeTydsManager) DetachVolume(volId string, instanceId string) (err error) {
	// TODO implement me
	panic("implement me")
}

func (m *FakeTydsManager) ResizeVolume(volId string, requestSize int64) (err error) {
	// TODO implement me
	panic("implement me")
}
