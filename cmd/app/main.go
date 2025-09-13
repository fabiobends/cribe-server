package main

import (
	router "cribeapp.com/cribe-server/internal/core"
	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/utils"
)

func main() {
	// Initialize logger for main application
	appLogger := logger.NewCoreLogger("Application")

	port := utils.GetPort()

	appLogger.Info("Starting Cribe Server", map[string]interface{}{
		"port": port,
	})

	err := router.Handler(port)

	if err != nil {
		appLogger.Error("Server failed to start", map[string]interface{}{
			"error": err.Error(),
			"port":  port,
		})
		panic(err)
	}
}
