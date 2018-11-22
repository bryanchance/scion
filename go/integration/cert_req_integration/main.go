// Copyright 2018 ETH Zurich
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

package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/scionproto/scion/go/lib/integration"
	"github.com/scionproto/scion/go/lib/log"
)

var (
	name     = "certreq_integration"
	cmd      = "./bin/cert_req"
	attempts = flag.Int("attempts", 2, "Number of attempts before giving up.")
)

func main() {
	os.Exit(realMain())
}

func realMain() int {
	if err := integration.Init(name); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init: %s\n", err)
		return 1
	}
	defer log.LogPanicAndExit()
	defer log.Flush()
	clientArgs := []string{"-log.console", "debug", "-attempts", strconv.Itoa(*attempts),
		"-local", integration.SrcAddrPattern, "-remoteIA", integration.DstIAReplace}
	in := integration.NewBinaryIntegration(name, cmd, clientArgs, []string{}, integration.StdLog)
	if err := runTests(in, integration.IAPairs()); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run tests: %s\n", err)
		return 1
	}
	return 0
}

// RunTests runs the client for each IAPair.
// In case of an error the function is terminated immediately.
func runTests(in integration.Integration, pairs []integration.IAPair) error {
	return integration.ExecuteTimed(in.Name(), func() error {
		// Start the clients for srcDest pair
		for i, conn := range integration.IAPairs() {
			log.Info(fmt.Sprintf("Test %v: %v -> %v (%v/%v)",
				in.Name(), conn.Src.IA, conn.Dst.IA, i+1, len(integration.IAPairs())))
			t := integration.DefaultRunTimeout + integration.RetryTimeout*time.Duration(*attempts)
			if err := integration.RunClient(in, conn, t); err != nil {
				log.Error("Error during client execution", "err", err)
				return err
			}
		}
		return nil
	})
}
