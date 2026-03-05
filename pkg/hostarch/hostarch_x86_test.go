// Copyright 2026 The gVisor Authors.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

//go:build amd64 || 386
// +build amd64 386

package hostarch

import (
	"testing"
)

func TestEndianString(t *testing.T) {
	if got := EndianString(); got != "little" {
		t.Errorf("got %s, want little", got)
	}
}
