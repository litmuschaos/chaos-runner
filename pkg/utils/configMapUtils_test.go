package utils

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	litmuschaosv1alpha1 "github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
)

func TestPatchConfigMaps(t *testing.T) {
	fakeConfigMap := "fake configmap"
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
		chaosengine     *litmuschaosv1alpha1.ChaosEngine
		chaosexperiment *litmuschaosv1alpha1.ChaosExperiment
		configmap       v1.ConfigMap
		isErr           bool
	}{
		"Test Positive-1": {
			chaosengine: &litmuschaosv1alpha1.ChaosEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      engineDetails.Name,
					Namespace: engineDetails.EngineNamespace,
				},
				Spec: litmuschaosv1alpha1.ChaosEngineSpec{
					Experiments: []litmuschaosv1alpha1.ExperimentList{
						{
							Name: experiment.Name,
							Spec: v1alpha1.ExperimentAttributes{
								Components: v1alpha1.ExperimentComponents{
									ConfigMaps: []v1alpha1.ConfigMap{
										{
											Name:      fakeConfigMap,
											MountPath: "fake mountpath",
										},
									},
								},
							},
						},
					},
				},
			},
			chaosexperiment: &litmuschaosv1alpha1.ChaosExperiment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      experiment.Name,
					Namespace: experiment.Namespace,
				},
				Spec: litmuschaosv1alpha1.ChaosExperimentSpec{
					Definition: litmuschaosv1alpha1.ExperimentDef{
						Image: fakeExperimentImage,
					},
				},
			},
			configmap: v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fakeConfigMap,
					Namespace: experiment.Namespace,
				},
				Data: map[string]string{
					"my-fake-key": "myfake-val",
				}},
			isErr: false,
		},
		"Test Negative-1": {
			chaosengine: &litmuschaosv1alpha1.ChaosEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      engineDetails.Name,
					Namespace: engineDetails.EngineNamespace,
				},
				Spec: litmuschaosv1alpha1.ChaosEngineSpec{
					Experiments: []litmuschaosv1alpha1.ExperimentList{
						{
							Name: experiment.Name,
							Spec: v1alpha1.ExperimentAttributes{
								Components: v1alpha1.ExperimentComponents{
									ConfigMaps: []v1alpha1.ConfigMap{
										{
											Name:      fakeConfigMap,
											MountPath: "fake mountpath",
										},
									},
								},
							},
						},
					},
				},
			},
			chaosexperiment: &litmuschaosv1alpha1.ChaosExperiment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      experiment.Name,
					Namespace: experiment.Namespace,
				},
				Spec: litmuschaosv1alpha1.ChaosExperimentSpec{
					Definition: litmuschaosv1alpha1.ExperimentDef{
						Image: fakeExperimentImage,
					},
				},
			},
			isErr: true,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateFakeClient(t)

			_, err := client.KubeClient.CoreV1().ConfigMaps(experiment.Namespace).Create(&mock.configmap)
			if err != nil {
				t.Fatalf("configmap not created for %v test, err: %v", err, name)
			}
			_, err = client.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(mock.chaosexperiment.Namespace).Create(mock.chaosexperiment)
			if err != nil {
				t.Fatalf("experiment not created for %v test, err: %v", err, name)
			}
			_, err = client.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(mock.chaosengine.Namespace).Create(mock.chaosengine)
			if err != nil {
				t.Fatalf("engine not created for %v test, err: %v", err, name)
			}
			err = experiment.PatchConfigMaps(client, engineDetails)
			if !mock.isErr && err != nil {
				t.Fatalf("fail to patch the configmap, err: %v", err)
			}
			if mock.isErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}

			if !mock.isErr {
				actualResult := len(experiment.ConfigMaps)
				expectedResult := 1
				if actualResult != expectedResult {
					t.Fatalf("Test %q failed: expected length of configmap is %v but the actual lenght is %v", name, expectedResult, actualResult)
				}
			}
		})
	}
}
