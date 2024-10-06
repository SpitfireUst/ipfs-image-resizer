package main

import (
	"time"

	"ipfsimageresizer/pkg/server"
)

const (
	CacheExpiration      = 12 * time.Hour
	CacheCleanupInterval = 24 * time.Hour
	Port                 = 9191
)

func main() {
	server := server.New(CacheExpiration, CacheCleanupInterval)
	server.Start(Port)
}
