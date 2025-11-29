//go:build !test

package main

import (
	"os"

	router "cribeapp.com/cribe-server/internal/core"
	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/utils"
)

func main() {
	// Initialize logger for main application
	log := logger.NewCoreLogger("Application")

	port := utils.GetPort()

	log.Info("Starting Cribe Server", map[string]any{
		"port": port,
	})

	err := router.Handler(port)

	if err != nil {
		log.Error("Server failed to start", map[string]any{
			"error": err.Error(),
			"port":  port,
		})
		os.Exit(1)
	}
}
