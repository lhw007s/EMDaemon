package main

import (
	"testing"
)

func TestWriteMsg(*testing.T) {
	gGlobalScope.Init()
	GLogManager.WriteLog(DEBUG, "testMsg")
}
