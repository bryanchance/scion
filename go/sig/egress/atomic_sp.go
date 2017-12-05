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

package egress

import (
	"sync/atomic"

	"github.com/netsec-ethz/scion/go/lib/pathmgr"
)

// AtomicSP contains a pointer to a SyncPaths object; the pointer itself can be
// changed atomically via method UpdateSP. Method GetSPD returns the paths
// within the current SyncPaths object.
type AtomicSP struct {
	// v contains *pathmgr.SyncPaths
	v atomic.Value
}

func NewAtomicSP() *AtomicSP {
	return &AtomicSP{}
}

func (a *AtomicSP) UpdateSP(sp *pathmgr.SyncPaths) {
	a.v.Store(sp)
}

func (a *AtomicSP) GetSPD() *pathmgr.SyncPathsData {
	sp := a.v.Load().(*pathmgr.SyncPaths)
	return sp.Load()
}
