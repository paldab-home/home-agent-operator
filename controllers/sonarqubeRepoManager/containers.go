package sonarquberepomanager

import (
	"fmt"

	core "k8s.io/api/core/v1"
)

func newGitCloneContainer(repoUrl, repoDestinationPath string) core.Container {
	return core.Container{
		Name:  "clone-repo",
		Image: "alpine/git",
		Env: []core.EnvVar{
			{
				Name:  "CLONE_URL",
				Value: repoUrl,
			},
			{
				Name: "GITHUB_TOKEN",
				ValueFrom: &core.EnvVarSource{
					SecretKeyRef: &core.SecretKeySelector{
						Key: "token",
						LocalObjectReference: core.LocalObjectReference{
							Name: "github-token",
						},
					},
				},
			},
		},
		Command: []string{
			"/bin/sh",
			"-c",
			`TOKENIZED_URL=$(echo "$CLONE_URL" | sed "s#https://#https://${GITHUB_TOKEN}@#")
			git clone "$TOKENIZED_URL" /repo`,
		},
		VolumeMounts: []core.VolumeMount{
			{
				Name:      "repo",
				MountPath: repoDestinationPath,
			},
		},
	}
}

func newSonarqubeScanContainer(repoName, repoDestinationPath, sonarHostUrl, sonarToken string) core.Container {
	var (
		containerName  = "sonarscan"
		containerImage = "sonarsource/sonar-scanner-cli"
	)

	return core.Container{
		Name:       containerName,
		Image:      containerImage,
		WorkingDir: repoDestinationPath,
		Command: []string{
			"/bin/sh",
			"-c",
			fmt.Sprintf(`sonar-scanner \
			  -Dsonar.projectKey="%s" \
	          -Dsonar.sources="." \
			  -Dsonar.host.url=${SONAR_HOST_URL} \
			  -Dsonar.token=${SONAR_TOKEN}`,
				repoName),
		},

		Env: []core.EnvVar{
			{
				Name:  "SONAR_HOST_URL",
				Value: sonarHostUrl,
			},
			{
				Name:  "SONAR_TOKEN",
				Value: sonarToken,
			},
		},

		ImagePullPolicy: "IfNotPresent",
		VolumeMounts: []core.VolumeMount{
			{
				Name:      "repo",
				MountPath: repoDestinationPath,
			},
		},
	}
}
