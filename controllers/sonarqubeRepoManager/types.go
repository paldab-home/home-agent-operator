package sonarquberepomanager

import (
	reposerver "paldab/home-agent-operator/repoServer"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SonarqubeRepoManager struct {
	Client      client.Client
	Scheme      *runtime.Scheme
	RepoChannel <-chan reposerver.RepoChannelObject
}

// idea, map[reponame]SonarScanObject
type SonarScanObject struct {
	CommitSHA string
	LastScan  time.Time
}
