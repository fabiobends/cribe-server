package main

import (
	"os"
	"testing"

	"cribeapp.com/cribe-server/internal/utils"
)

func TestMain(m *testing.M) {
	_ = utils.CleanDatabase()

	code := m.Run()

	_ = utils.CleanDatabase()

	os.Exit(code)
}
