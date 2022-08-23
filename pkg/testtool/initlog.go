/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
 */

// Package testtool for init logger for llt
package testtool

import (
	"context"
	"fmt"

	"huawei.com/mindx/common/hwlog"
)

// this is for llt
func init() {
	config := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitRunLogger(&config, context.Background()); err != nil {
		fmt.Printf("%v", err)
	}
}
