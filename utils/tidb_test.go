package utils

import (
	"testing"
)

func TestStartTiDB(t *testing.T) {
	s, err := StartLocalTiDBServer("")
	if err != nil {
		panic(err)
	}
	if !PingLocalTiDB(s.DSN()) {
		panic("failed to ping TiDB")
	}
	s.Release()
	if PingLocalTiDB(s.DSN()) {
		panic("TiDB should be killed")
	}
}
