package sonarquberepomanager

import (
	"context"
	"fmt"
	"paldab/home-agent-operator/config"

	batchv1 "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func sonarScan(repoName, repoUrl string, k8sClient client.Client) error {
	var (
		jobName                    = fmt.Sprintf("sonar-scan-%s", repoName)
		backoffLimit         int32 = 3
		repoDestinationPath        = "/repo"
		jobTimeToLiveSeconds       = int32(1800)
	)

	job := batchv1.Job{
		ObjectMeta: v1.ObjectMeta{
			Name:      jobName,
			Namespace: config.NAMESPACE,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": config.APP_NAME,
				"app.kubernetes.io/created-by": config.APP_NAME,
				"app.kubernetes.io/component":  "sonarqube-job",
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            &backoffLimit,
			TTLSecondsAfterFinished: &jobTimeToLiveSeconds,
			Template: core.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Name: jobName,
				},
				Spec: core.PodSpec{
					RestartPolicy: core.RestartPolicyNever,
					InitContainers: []core.Container{
						newGitCloneContainer(repoUrl, repoDestinationPath),
					},
					Containers: []core.Container{
						newSonarqubeScanContainer(repoName, repoDestinationPath, config.SONARQUBE_HOST_URL, config.SONARQUBE_TOKEN),
					},
					Volumes: []core.Volume{
						{
							Name: "repo",
							VolumeSource: core.VolumeSource{
								EmptyDir: &core.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	err := k8sClient.Create(context.Background(), &job)

	if err != nil {
		return err
	}

	return nil
}
