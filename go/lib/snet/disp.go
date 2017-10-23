// Copyright 2017 ETH Zurich
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package snet

import (
	log "github.com/inconshreveable/log15"

	"github.com/netsec-ethz/scion/go/lib/common"
	liblog "github.com/netsec-ethz/scion/go/lib/log"
)

type DispPkt struct {
	Raw  common.RawBytes
	Addr *Addr
}

type DispatchFunc func(*DispPkt)

// PktDispatcher listens on c, and calls f for every packet read.
// N.B. the DispPkt passed to f is reused, so applications should make a copy if
// this is a problem.
func PktDispatcher(c *Conn, f DispatchFunc) {
	defer liblog.LogPanicAndExit()
	var err error
	var n int
	dp := &DispPkt{Raw: make(common.RawBytes, common.MaxMTU)}
	for {
		dp.Raw = dp.Raw[:cap(dp.Raw)]
		n, dp.Addr, err = c.ReadFromSCION(dp.Raw)
		if err != nil {
			log.Error("PktDispatcher: Error reading from connection", "err", err)
			break
		}
		dp.Raw = dp.Raw[:n]
		f(dp)
	}
}

func DispLogger(dp *DispPkt) {
	log.Debug("DispLogger", "src", dp.Addr, "raw", dp.Raw)
}
