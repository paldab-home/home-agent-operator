package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"paldab/home-agent-operator/config"
	databasemanager "paldab/home-agent-operator/controllers/databaseManager"
	mediaserver "paldab/home-agent-operator/controllers/mediaServer"
	sonarquberepomanager "paldab/home-agent-operator/controllers/sonarqubeRepoManager"
	reposerver "paldab/home-agent-operator/repoServer"

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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{Scheme: scheme.Scheme})
	if err != nil {
		log.Fatalf("failed to construct manager: %v", err)
	}

	// Handles API on different routine
	go setupAPI()

	repoChannel := make(chan reposerver.RepoChannelObject, 15)
	go setupRepoServer(ctx, repoChannel)

	setupMediaServerController(mgr)

	setupDatabaseScalerController(mgr)

	setupSonarqubeManager(mgr, repoChannel)

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

func setupRepoServer(ctx context.Context, repoChannel chan reposerver.RepoChannelObject) {
	var repoServerRefreshInterval = 30
	interval, err := config.ConvertEnvToInt(config.REPOSERVER_REFRESH_INTERVAL)

	if err != nil {
		zap.L().Warn("using default refresh interval value", zap.Int("refreshInterval", repoServerRefreshInterval), zap.Error(err))
	} else {
		repoServerRefreshInterval = interval
	}

	refreshIntervalMin := time.Duration(time.Minute * time.Duration(repoServerRefreshInterval))
	repoServer := reposerver.NewRepoServer(config.GITHUB_TOKEN, refreshIntervalMin, repoChannel)

	if err := repoServer.StartServer(ctx); err != nil {
		zap.L().Error("could not start Repo Server", zap.Error(err))
	}
}

func setupMediaServerController(mgr ctrl.Manager) {
	if config.MEDIASERVER_NAMESPACE == "" {
		zap.L().Error("could not start operator. missing MEDIASERVER_NAMESPACE env variable")
		return
	}

	mediaServerController := mediaserver.MediaServerController{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}

	mediaServerController.RegisterApiEndpoints(router)

	if err := mediaServerController.SetupWithManager(mgr); err != nil {
		log.Fatalf("failed to add controller to controller manager. error: %v", err)
	}
}

func setupDatabaseScalerController(mgr ctrl.Manager) {
	databaseScalerControler := databasemanager.DatabaseManagerController{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}

	if err := databaseScalerControler.SetupWithManager(mgr); err != nil {
		log.Fatalf("failed to add controller to controller manager. error: %v", err)
	}
}

func setupSonarqubeManager(mgr ctrl.Manager, channel <-chan reposerver.RepoChannelObject) {
	sonarqubeManagerController := sonarquberepomanager.SonarqubeRepoManager{
		Client:      mgr.GetClient(),
		Scheme:      mgr.GetScheme(),
		RepoChannel: channel,
	}

	go sonarqubeManagerController.StartProcessingRepoChannel()

	// if err := sonarqubeManagerController.SetupWithManager(mgr); err != nil {
	// 	log.Fatalf("failed to add controller to controller manager. error: %v", err)
	// }
}
