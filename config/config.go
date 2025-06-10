package config

import (
	"log"
	"os"
	"time"

	"go.uber.org/zap"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	API_PORT          = 8080
	ERR_RETRY_TIMEOUT = time.Minute * 5

	// global labels
	MANAGED_BY_OPERATOR_LABEL    = "operator.paldab.io/managed-by"
	DATABASE_INSTANCE_NAME_LABEL = "operator.paldab.io/database-name"
	POD_NEEDS_DATABASE_LABEL     = "operator.paldab.io/database"
)

var (
	MEDIASERVER_NAMESPACE string
)

func GetConfig() {
	MEDIASERVER_NAMESPACE = os.Getenv("MEDIASERVER_NAMESPACE")
}

func SetupLogger() {
	cfg := zap.NewDevelopmentConfig()

	logger, err := cfg.Build()
	if err != nil {
		log.Fatal(err)
	}

	ctrl.SetLogger(ctrlzap.New(ctrlzap.UseDevMode(true)))
	zap.ReplaceGlobals(logger)
}
