/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
 */

// Package signals package
package signals

const (
	stopChCapacity = 100
)

// SetupSignalHandler registered for SIGTERM and SIGINT. A stop channel is returned
// which is closed on one of these signals. If a second signal is caught, the program
// is terminated with exit code 1.
func SetupSignalHandler() chan struct{} {
	stop := make(chan struct{}, stopChCapacity)
	return stop
}
