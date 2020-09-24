/*
Copyright 2017 The Beethoven Authors.

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

// Package waitcycle to hccl
package waitcycle

import (
	"errors"
	"k8s.io/klog"
	"math/rand"
	"time"
)

const (
	// DefaultTimeout to ge timeout
	DefaultTimeout = 30
	loggerTypeFour = 4
)

// NeverStop may be passed to Until to make it never stop.
var NeverStop <-chan struct{} = make(chan struct{})

// Func to get func
type Func func() (bool, error)

var (
	// ErrTimeout to timeout
	ErrTimeout = errors.New("timeout")
	// ErrAfterRetry to retry
	ErrAfterRetry = errors.New("exceed retry counts")
)

// Wait to get signal
func Wait(waitTimeout time.Duration, fn Func, checkoutInterval time.Duration) error {
	interrupt := false
	var timeout <-chan time.Time
	if waitTimeout > time.Duration(0) {
		timeout = time.After(waitTimeout)
	} else {
		timeout = time.After(DefaultTimeout)
	}

	go func() {
		select {
		case _, ok := <-timeout:
			if ok {
				interrupt = true
			}
		}
	}()

	for !interrupt {
		ok, err := fn()
		if ok {
			return err
		}
		time.Sleep(checkoutInterval)
	}
	return ErrTimeout
}

// WaitForCount for wait count
func WaitForCount(retryCount int, fn Func, checkoutInterval time.Duration) error {
	for count := 0; count < retryCount; count++ {
		ok, err := fn()
		if ok {
			return err
		}
		time.Sleep(checkoutInterval)
	}
	return ErrAfterRetry
}

// Until loops until stop channel is closed, running f every period.
//
// Until is syntactic sugar on top of JitterUntil with zero jitter factor and
// with sliding = true (which means the timer for period starts after the f
// completes).
func Until(f func(), period time.Duration, stopCh <-chan struct{}) {
	JitterUntil(f, period, 0.0, true, stopCh)
}

// JitterUntil loops until stop channel is closed, running f every period.
//
// If jitterFactor is positive, the period is jittered before every run of f.
// If jitterFactor is not positive, the period is unchanged and not jitterd.
//
// If slidingis true, the period is computed after f runs. If it is false then
// period includes the runtime for f.
//
// Close stopCh to stop. f may not be invoked if stop channel is already
// closed. Pass NeverStop to if you don't want it stop.
func JitterUntil(f func(), period time.Duration, jitterFactor float64, sliding bool, stopCh <-chan struct{}) {
	for {

		select {
		case c, ok := <-stopCh:
			if ok {
				klog.V(loggerTypeFour).Info(c)
			}
			return
		default:
		}

		jitteredPeriod := period
		if jitterFactor > 0.0 {
			jitteredPeriod = Jitter(period, jitterFactor)
		}

		var t *time.Timer
		if !sliding {
			t = time.NewTimer(jitteredPeriod)
		}

		func() {
			f()
		}()

		if sliding {
			t = time.NewTimer(jitteredPeriod)
		}

		// NOTE: b/c there is no priority selection in golang
		// it is possible for this to race, meaning we could
		// trigger t.C and stopCh, and t.C select falls through.
		// In order to mitigate we re-check stopCh at the beginning
		// of every loop to prevent extra executions of f().
		select {
		case c, ok := <-stopCh:
			if ok {
				klog.V(loggerTypeFour).Info(c)
			}
			return
		case <-t.C:
		}
	}
}

// Jitter returns a time.Duration between duration and duration + maxFactor *
// duration.
//
// This allows clients to avoid converging on periodic behavior. If maxFactor
// is 0.0, a suggested default value will be chosen.
func Jitter(duration time.Duration, maxFactor float64) time.Duration {
	if maxFactor <= 0.0 {
		maxFactor = 1.0
	}
	wait := duration + time.Duration(rand.Float64()*maxFactor*float64(duration))
	return wait
}
