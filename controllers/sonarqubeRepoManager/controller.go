package sonarquberepomanager

import (
	"context"
	"paldab/home-agent-operator/config"
	"time"

	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var trackedRepositories map[string]*SonarScanObject

// assume multiple orgs and repos can enter
func (r *SonarqubeRepoManager) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var job batchv1.Job

	if err := r.Client.Get(ctx, req.NamespacedName, &job); err != nil {
		return ctrl.Result{}, nil
	}

	// Maybe do smth

	return ctrl.Result{}, nil
}

func (r *SonarqubeRepoManager) SetupWithManager(mgr ctrl.Manager) error {
	// go r.StartProcessingRepoChannel()

	return ctrl.NewControllerManagedBy(mgr).
		Named("SonarqubeManager").
		Watches(&batchv1.Job{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, o client.Object) []reconcile.Request {
			job := o.(*batchv1.Job)
			labels := job.GetLabels()

			if job.GetNamespace() != config.NAMESPACE {
				return nil
			}

			if labels["app.kubernetes.io/component"] != "sonarqube-job" && labels["app.kubernetes.io/created-by"] != config.APP_NAME {
				return nil
			}

			return []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Name:      job.GetName(),
						Namespace: job.GetNamespace(),
					},
				},
			}
		})).
		Complete(r)
}

func (r *SonarqubeRepoManager) StartProcessingRepoChannel() {
	if config.SONARQUBE_HOST_URL == "" || config.SONARQUBE_TOKEN == "" {
		zap.L().Error("Missing either SONARQUBE_HOST_URL or SONARQUBE_TOKEN environment variable. Cannot start Sonarqube controller")
		return
	}

	zap.L().Info("Starting SonarqubeRepoManager", zap.String("sonarqubeHost", config.SONARQUBE_HOST_URL))

	if trackedRepositories == nil {
		trackedRepositories = make(map[string]*SonarScanObject)
	}

	for repoChannelObject := range r.RepoChannel {
		repoName := repoChannelObject.Repository.GetName()
		val, exists := trackedRepositories[repoName]

		if exists {
			repoHasChanged := val.CommitSHA != repoChannelObject.LatestCommitSha
			if !repoHasChanged {
				continue
			}
		}

		newObj := SonarScanObject{
			CommitSHA: repoChannelObject.LatestCommitSha,
			LastScan:  time.Now(),
		}

		repoUrl := repoChannelObject.Repository.GetCloneURL()

		zap.L().Info("scanning using sonarqube", zap.String("repo", repoName), zap.String("url", repoUrl))

		err := sonarScan(repoName, repoUrl, r.Client)

		if err != nil {
			zap.L().Error("failed to create job", zap.Error(err))
			continue
		}

		// On success save in memory
		trackedRepositories[repoName] = &newObj
	}
}
