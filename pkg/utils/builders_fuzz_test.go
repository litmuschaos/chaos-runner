package utils

import (
	fuzz "github.com/AdaLogics/go-fuzz-headers"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func FuzzBuildContainerSpec(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		fuzzConsumer := fuzz.NewConsumer(data)
		targetStruct := &struct {
			ExpDetails *ExperimentDetails
			EnvVars    []corev1.EnvVar
		}{}

		err := fuzzConsumer.GenerateStruct(targetStruct)
		if err != nil {
			return
		}
		containerSpec, err := buildContainerSpec(targetStruct.ExpDetails, targetStruct.EnvVars)
		if err != nil {
			return
		}

		container, err := containerSpec.Build()
		if err != nil {
			return
		}

		require.Equal(t, targetStruct.ExpDetails.JobName, container.Name)
		require.Equal(t, targetStruct.ExpDetails.ExpImage, container.Image)
		require.Equal(t, targetStruct.ExpDetails.ExpCommand, container.Command)
		require.Equal(t, targetStruct.ExpDetails.ExpArgs, container.Args)
		require.Equal(t, targetStruct.ExpDetails.ExpImagePullPolicy, container.ImagePullPolicy)
		require.Equal(t, targetStruct.EnvVars, container.Env)
	})
}

func FuzzGetEnvFromMap(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		fuzzConsumer := fuzz.NewConsumer(data)
		targetStruct := &struct {
			m map[string]corev1.EnvVar
		}{}
		err := fuzzConsumer.GenerateStruct(targetStruct)
		if err != nil {
			return
		}
		envs := getEnvFromMap(targetStruct.m)
		var envCount = len(envs)
		require.Equal(t, envCount, len(targetStruct.m)+1)
	})
}

func FuzzSetSidecarSecrets(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		fuzzConsumer := fuzz.NewConsumer(data)
		targetStruct := &struct {
			experiment *ExperimentDetails
		}{}

		err := fuzzConsumer.GenerateStruct(targetStruct)
		if err != nil {
			return
		}
		if targetStruct.experiment != nil {
			secrets := setSidecarSecrets(targetStruct.experiment)
			require.GreaterOrEqual(t, len(secrets), 1)

			for _, sidecar := range targetStruct.experiment.SideCars {
				for _, secret := range sidecar.Secrets {
					for _, s := range secrets {
						require.Equal(t, s.Name, secret.Name)
					}
				}
			}
		}
	})
}
