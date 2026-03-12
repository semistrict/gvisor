// Copyright 2021 The gVisor Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package transport

import (
	"context"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"gvisor.dev/gvisor/pkg/log"
)

// beforeSave is invoked by stateify.
func (e *connectionlessEndpoint) beforeSave() {
	frames := runtime.CallersFrames(e.closerStack[:e.closerStackLen])
	var b strings.Builder
	for {
		frame, more := frames.Next()
		fmt.Fprintf(&b, "%s\n\t%s:%d pc=%#x\n", frame.Function, frame.File, frame.Line, frame.PC)
		if !more {
			break
		}
	}
	e.closerStackStr = b.String()
}

// afterLoad is invoked by stateify.
func (e *connectionlessEndpoint) afterLoad(context.Context) {
	e.ops.InitHandler(e, &stackHandler{}, getSendBufferLimits, getReceiveBufferLimits)
	if len(e.closerStackStr) == 0 {
		return
	}

	var (
		closerStack    [32]uintptr
		closerStackLen int
	)
	pcRegex := regexp.MustCompile(`pc=0x([0-9a-fA-F]+)`)
	lines := strings.Split(e.closerStackStr, "\n")
	for _, line := range lines {
		matches := pcRegex.FindStringSubmatch(line)
		if len(matches) == 2 {
			hexPC := matches[1]
			pc, err := strconv.ParseUint(hexPC, 16, 64)
			if err != nil {
				log.Debugf("failed to parse hex PC %q: %v", hexPC, err)
				break
			}

			if closerStackLen < len(closerStack) {
				closerStack[closerStackLen] = uintptr(pc)
				closerStackLen++
			} else {
				log.Debugf("symbolized stack contains more than 32 frames; truncating to 32.")
				break
			}
		}
	}
	e.closerStack = closerStack
	e.closerStackLen = closerStackLen
}
