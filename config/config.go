package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	API_PORT          = 8080
	ERR_RETRY_TIMEOUT = time.Minute * 5
	APP_NAME          = "home-agent-operator"

	// database labels
	DATABASE_INSTANCE_NAME_LABEL = "operator.paldab.io/database-instance-name"
	POD_NEEDS_DATABASE_LABEL     = "operator.paldab.io/database"
)

var (
	// global
	ENV       string
	NAMESPACE string

	// Mediaserver
	MEDIASERVER_NAMESPACE string

	// Reposerver
	GITHUB_TOKEN                string
	GITHUB_ORG                  string
	REPOSERVER_REFRESH_INTERVAL string

	// sonarqube
	SONARQUBE_TOKEN    string
	SONARQUBE_HOST_URL string
)

func GetConfig() {
	ENV = os.Getenv("ENV")
	NAMESPACE = os.Getenv("NAMESPACE")

	MEDIASERVER_NAMESPACE = os.Getenv("MEDIASERVER_NAMESPACE")

	GITHUB_TOKEN = os.Getenv("GITHUB_TOKEN")
	GITHUB_ORG = os.Getenv("GITHUB_ORG")
	REPOSERVER_REFRESH_INTERVAL = os.Getenv("REPOSERVER_REFRESH_INTERVAL")

	SONARQUBE_TOKEN = os.Getenv("SONARQUBE_TOKEN")
	SONARQUBE_HOST_URL = os.Getenv("SONARQUBE_HOST_URL")
}

func SetupLogger() {
	if ENV == "dev" {
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		logger, err := cfg.Build()
		if err != nil {
			log.Fatal(err)
		}

		ctrl.SetLogger(ctrlzap.New(ctrlzap.UseDevMode(true)))
		zap.ReplaceGlobals(logger)
	} else {
		cfg := zap.NewProductionConfig()
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg.DisableCaller = true

		logger, err := cfg.Build()
		if err != nil {
			log.Fatal(err)
		}

		ctrl.SetLogger(ctrlzap.New(ctrlzap.UseDevMode(false)))
		zap.ReplaceGlobals(logger)
	}
}

func ConvertEnvToInt(env string) (int, error) {
	if env == "" {
		return 0, fmt.Errorf("env is an empty string")
	}

	number, err := strconv.Atoi(env)
	if err != nil {
		return 0, fmt.Errorf("could not convert env to number. Using default interval")
	}

	return number, nil
}
