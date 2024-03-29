/*
Copyright (C) 2023 Toyou, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this work except in compliance with the License.
You may obtain a copy of the License in the LICENSE file, or at:

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package driver

import (
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"k8s.io/klog"
)

type ToyouDriver struct {
	name          string
	version       string
	nodeId        string
	maxVolume     int64
	volumeCap     []*csi.VolumeCapability_AccessMode
	controllerCap []*csi.ControllerServiceCapability
	nodeCap       []*csi.NodeServiceCapability
	pluginCap     []*csi.PluginCapability

	IdS *IdentityServer
	NS  *NodeServer
	CS  *ControllerServer

	Driver *csicommon.CSIDriver
}

type InitDiskDriverInput struct {
	Name          string
	Version       string
	NodeId        string
	MaxVolume     int64
	VolumeCap     []csi.VolumeCapability_AccessMode_Mode
	ControllerCap []csi.ControllerServiceCapability_RPC_Type
	NodeCap       []csi.NodeServiceCapability_RPC_Type
	PluginCap     []*csi.PluginCapability
}

// NewToyouDriver GetDiskDriver
func NewToyouDriver() *ToyouDriver {
	return &ToyouDriver{
		name:          "csi.toyou.com",
		version:       "1.0.0",
		volumeCap:     make([]*csi.VolumeCapability_AccessMode, 0),
		controllerCap: make([]*csi.ControllerServiceCapability, 0),
		nodeCap:       make([]*csi.NodeServiceCapability, 0),
		pluginCap:     make([]*csi.PluginCapability, 0),
	}
}

func (d *ToyouDriver) InitDiskDriver(input *InitDiskDriverInput) {
	if input == nil {
		return
	}
	d.name = input.Name
	d.version = input.Version
	// Setup Node Id
	d.nodeId = input.NodeId
	// Setup max volume
	d.maxVolume = input.MaxVolume
	// Setup cap
	d.addVolumeCapabilityAccessModes(input.VolumeCap)
	d.addControllerServiceCapabilities(input.ControllerCap)
	d.addNodeServiceCapabilities(input.NodeCap)
	d.addPluginCapabilities(input.PluginCap)
}

func (d *ToyouDriver) addVolumeCapabilityAccessModes(vc []csi.VolumeCapability_AccessMode_Mode) {
	var vca []*csi.VolumeCapability_AccessMode
	for _, c := range vc {
		klog.V(4).Infof("Enabling volume access mode: %v", c.String())
		vca = append(vca, NewVolumeCapabilityAccessMode(c))
	}
	d.volumeCap = vca
}

func (d *ToyouDriver) addControllerServiceCapabilities(cl []csi.ControllerServiceCapability_RPC_Type) {
	var csc []*csi.ControllerServiceCapability
	for _, c := range cl {
		klog.V(4).Infof("Enabling controller service capability: %v", c.String())
		csc = append(csc, NewControllerServiceCapability(c))
	}
	d.controllerCap = csc
}

func (d *ToyouDriver) addNodeServiceCapabilities(nl []csi.NodeServiceCapability_RPC_Type) {
	var nsc []*csi.NodeServiceCapability
	for _, n := range nl {
		klog.V(4).Infof("Enabling node service capability: %v", n.String())
		nsc = append(nsc, NewNodeServiceCapability(n))
	}
	d.nodeCap = nsc
}

func (d *ToyouDriver) addPluginCapabilities(cap []*csi.PluginCapability) {
	d.pluginCap = cap
}

func (d *ToyouDriver) ValidateControllerServiceRequest(c csi.ControllerServiceCapability_RPC_Type) bool {
	if c == csi.ControllerServiceCapability_RPC_UNKNOWN {
		return true
	}

	for _, controllerServiceCapability := range d.controllerCap {
		if c == controllerServiceCapability.GetRpc().Type {
			return true
		}
	}
	return false
}

func (d *ToyouDriver) ValidateNodeServiceRequest(c csi.NodeServiceCapability_RPC_Type) bool {
	if c == csi.NodeServiceCapability_RPC_UNKNOWN {
		return true
	}
	for _, nodeServiceCapability := range d.nodeCap {
		if c == nodeServiceCapability.GetRpc().Type {
			return true
		}
	}
	return false

}

func (d *ToyouDriver) ValidateVolumeCapability(cap *csi.VolumeCapability) bool {
	if !d.ValidateVolumeAccessMode(cap.GetAccessMode().GetMode()) {
		return false
	}
	return true
}

func (d *ToyouDriver) ValidateVolumeCapabilities(caps []*csi.VolumeCapability) bool {
	for _, volumeCapability := range caps {
		if !d.ValidateVolumeAccessMode(volumeCapability.GetAccessMode().GetMode()) {
			return false
		}
	}
	return true
}

func (d *ToyouDriver) ValidateVolumeAccessMode(c csi.VolumeCapability_AccessMode_Mode) bool {
	for _, mode := range d.volumeCap {
		if c == mode.GetMode() {
			return true
		}
	}
	return false
}

func (d *ToyouDriver) ValidatePluginCapabilityService(cap csi.PluginCapability_Service_Type) bool {
	for _, v := range d.GetPluginCapability() {
		if v.GetService() != nil && v.GetService().GetType() == cap {
			return true
		}
	}
	return false
}

func (d *ToyouDriver) GetName() string {
	return d.name
}

func (d *ToyouDriver) GetVersion() string {
	return d.version
}

func (d *ToyouDriver) GetInstanceId() string {
	return d.nodeId
}

func (d *ToyouDriver) GetMaxVolumePerNode() int64 {
	return d.maxVolume
}

func (d *ToyouDriver) GetControllerCapability() []*csi.ControllerServiceCapability {
	return d.controllerCap
}

func (d *ToyouDriver) GetNodeCapability() []*csi.NodeServiceCapability {
	return d.nodeCap
}

func (d *ToyouDriver) GetPluginCapability() []*csi.PluginCapability {
	return d.pluginCap
}

func (d *ToyouDriver) GetVolumeCapability() []*csi.VolumeCapability_AccessMode {
	return d.volumeCap
}

func (d *ToyouDriver) GetTopologyZoneKey() string {
	return fmt.Sprintf("topology.%s/zone", d.GetName())
}

func (d *ToyouDriver) GetTopologyInstanceTypeKey() string {
	return fmt.Sprintf("topology.%s/instance-type", d.GetName())
}
