/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package v2

import (
	"hccl-controller/pkg/ring-controller/ranktable/v1"
)

// RankTable : ranktable of v2
type RankTable struct {
	v1.RankTableStatus
	ServerList  []*Server `json:"server_list"`  // hccl_json server list
	ServerCount string    `json:"server_count"` // hccl_json server count
	Version     string    `json:"version"`      // hccl_json version
}

// Server to hccl
type Server struct {
	DeviceList []*Device `json:"device"`    // device list in each server
	ServerID   string    `json:"server_id"` // server id, represented by ip address
	PodID      string    `json:"-"`         // pod id, equal to the last integer of pod name
}

// Device to hccl
type Device struct {
	v1.Device
	RankID string `json:"rank_id"` // rank id
}
