package main

import (
	"log"
	"net/http"
	"os"

	"paldab/home-agent-operator/config"
	databasemanager "paldab/home-agent-operator/controllers/databaseManager"
	mediaserver "paldab/home-agent-operator/controllers/mediaServer"

	longhornv1beta2 "github.com/longhorn/longhorn-manager/k8s/pkg/apis/longhorn/v1beta2"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
)

var router *http.ServeMux = http.NewServeMux()

func init() {
	config.GetConfig()
	config.SetupLogger()
	longhornv1beta2.AddToScheme(scheme.Scheme)
}

func main() {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{Scheme: scheme.Scheme})
	if err != nil {
		log.Fatalf("failed to construct manager: %w", err)
	}

	// Handles API on different routine
	go setupAPI()

	setupMediaServerController(mgr)

	setupDatabaseScalerController(mgr)

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Fatalf("failed to start controller. error: %v", err)
	}
}

func setupAPI() {
	zap.L().Info("started operator listening on port 8000")
	if err := http.ListenAndServe(":8000", router); err != nil {
		zap.L().Error("could not start API", zap.Error(err))
		os.Exit(1)
	}
}

func setupMediaServerController(mgr ctrl.Manager) {
	if config.MEDIASERVER_NAMESPACE == "" {
		log.Fatal("could not start operator. missing MEDIASERVER_NAMESPACE env variable")
	}

	mediaServerController := mediaserver.MediaServerController{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}

	mediaServerController.RegisterApiEndpoints(router)

	if err := mediaServerController.RegisterController(mgr); err != nil {
		log.Fatalf("failed to add controller to controller manager. error: %v", err)
	}
}

func setupDatabaseScalerController(mgr ctrl.Manager) {
	databaseScalerControler := databasemanager.DatabaseManagerController{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}

	if err := databaseScalerControler.RegisterController(mgr); err != nil {
		log.Fatalf("failed to add controller to controller manager. error: %v", err)
	}
}
