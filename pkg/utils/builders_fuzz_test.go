package utils

import (
	fuzz "github.com/AdaLogics/go-fuzz-headers"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

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
