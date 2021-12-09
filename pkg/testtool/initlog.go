/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
 */

// Package testtool for init logger for llt
package testtool

import "huawei.com/npu-exporter/hwlog"

// this is for llt
func init() {
	config := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	stopCh := make(chan struct{})
	hwlog.InitRunLogger(&config, stopCh)
}
