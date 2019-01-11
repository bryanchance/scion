// Copyright 2019 Anapaya Systems

package xtest

import (
	"os"

	"github.com/scionproto/scion/go/lib/util"
)

// PostgresHost returns the postgres hostname to use for testing.
// In case of an error determining the host this function panicks.
func PostgresHost() string {
	inDocker, err := util.RunsInDocker()
	if err != nil {
		panic(err)
	}
	if inDocker {
		dbHost, ok := os.LookupEnv("DOCKER0")
		if !ok {
			panic("Expected DOCKER0 env variable")
		}
		return dbHost
	}
	return "localhost"
}
