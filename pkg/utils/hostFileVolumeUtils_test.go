package utils

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
)

func TestPatchHostFileVolumes(t *testing.T) {
	fakehostpathname := "fake-hostpath-name"
	fakeExperimentImage := "fake-experiment-image"
	experiment := ExperimentDetails{
		Name:               "Fake-Exp-Name",
		Namespace:          "Fake NameSpace",
		JobName:            "fake-job-name",
		StatusCheckTimeout: 10,
	}
	engineDetails := EngineDetails{
		Name:            "Fake Engine",
		EngineNamespace: "Fake NameSpace",
	}

	tests := map[string]struct {
		chaosengine     *v1alpha1.ChaosEngine
		chaosexperiment *v1alpha1.ChaosExperiment
		configmap       v1.ConfigMap
		isErr           bool
	}{
		"Test Positive-1": {
			chaosexperiment: &v1alpha1.ChaosExperiment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      experiment.Name,
					Namespace: experiment.Namespace,
				},
				Spec: v1alpha1.ChaosExperimentSpec{
					Definition: v1alpha1.ExperimentDef{
						Image: fakeExperimentImage,
						HostFileVolumes: []v1alpha1.HostFile{
							{
								Name:      fakehostpathname,
								MountPath: "fake-mount-path",
								NodePath:  "fake-node-path",
							},
						},
					},
				},
			},
			isErr: false,
		},
		"Test Negative-1": {
			chaosexperiment: &v1alpha1.ChaosExperiment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      experiment.Name,
					Namespace: experiment.Namespace,
				},
				Spec: v1alpha1.ChaosExperimentSpec{
					Definition: v1alpha1.ExperimentDef{
						Image: fakeExperimentImage,
						HostFileVolumes: []v1alpha1.HostFile{
							{
								Name:      "",
								MountPath: "",
								NodePath:  "",
							},
						},
					},
				},
			},
			isErr: true,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateFakeClient(t)

			_, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(mock.chaosexperiment.Namespace).Create(mock.chaosexperiment)
			if err != nil {
				t.Fatalf("experiment not created for %v test, err: %v", err, name)
			}
			err = experiment.PatchHostFileVolumes(client, engineDetails)
			if !mock.isErr && err != nil {
				t.Fatalf("fail to patch the host file volume, err: %v", err)
			}
			if mock.isErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.isErr {
				actualResult := len(experiment.HostFileVolumes)
				expectedResult := 1
				if actualResult != expectedResult {
					t.Fatalf("Test %q failed: expected length of configmap is %v but the actual length is %v", name, expectedResult, actualResult)
				}
			}
		})
	}
}
