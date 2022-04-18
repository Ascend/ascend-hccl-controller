/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
 */

package v2

import (
	ranktablev1 "hccl-controller/pkg/ring-controller/ranktable/v1"
)

// RankTable : ranktable of v2
type RankTable struct {
	ranktablev1.RankTableStatus
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
	ranktablev1.Device
	RankID string `json:"rank_id"` // rank id
}
