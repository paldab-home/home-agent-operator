package config

import (
	"os"
	"time"
)

const (
	API_PORT          = 8080
	ERR_RETRY_TIMEOUT = time.Minute * 5
)

var (
	MEDIASERVER_NAMESPACE string
)

func GetConfig() {
	MEDIASERVER_NAMESPACE = os.Getenv("MEDIASERVER_NAMESPACE")
}
