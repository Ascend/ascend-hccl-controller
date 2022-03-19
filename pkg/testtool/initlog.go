/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
 */

// Package testtool for init logger for llt
package testtool

import (
	"fmt"
	
	"huawei.com/npu-exporter/hwlog"
)

// this is for llt
func init() {
	config := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	stopCh := make(chan struct{})
	if err := hwlog.InitRunLogger(&config, stopCh); err != nil {
		fmt.Printf("%v", err)
	}
}
