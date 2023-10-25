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

package rpcserver

import (
	"toyou_csi/pkg/common"
	"toyou_csi/pkg/disk/driver"
)

type ControllerServer struct {
	driver *driver.DiskDriver
	locks  *common.ResourceLocks
}

func NewControllerServer(d *driver.DiskDriver) *ControllerServer {
	return &ControllerServer{
		driver: d,
		locks:  common.NewResourceLocks(),
	}
}
