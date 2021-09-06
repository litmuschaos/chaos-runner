package utils

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
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
		chaosengine     *v1alpha1.ChaosEngine
		chaosexperiment *v1alpha1.ChaosExperiment
		configmap       v1.ConfigMap
		isErr           bool
	}{
		"Test Positive-1": {
			chaosengine: &v1alpha1.ChaosEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      engineDetails.Name,
					Namespace: engineDetails.EngineNamespace,
				},
				Spec: v1alpha1.ChaosEngineSpec{
					Experiments: []v1alpha1.ExperimentList{
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
			chaosexperiment: &v1alpha1.ChaosExperiment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      experiment.Name,
					Namespace: experiment.Namespace,
				},
				Spec: v1alpha1.ChaosExperimentSpec{
					Definition: v1alpha1.ExperimentDef{
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
			chaosengine: &v1alpha1.ChaosEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      engineDetails.Name,
					Namespace: engineDetails.EngineNamespace,
				},
				Spec: v1alpha1.ChaosEngineSpec{
					Experiments: []v1alpha1.ExperimentList{
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
			chaosexperiment: &v1alpha1.ChaosExperiment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      experiment.Name,
					Namespace: experiment.Namespace,
				},
				Spec: v1alpha1.ChaosExperimentSpec{
					Definition: v1alpha1.ExperimentDef{
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
				t.Fatalf("configmap not created for %v test, err: %v", name, err)
			}
			_, err = client.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(mock.chaosexperiment.Namespace).Create(mock.chaosexperiment)
			if err != nil {
				t.Fatalf("experiment not created for %v test, err: %v", name, err)
			}
			_, err = client.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(mock.chaosengine.Namespace).Create(mock.chaosengine)
			if err != nil {
				t.Fatalf("engine not created for %v test, err: %v", name, err)
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
					t.Fatalf("Test %q failed: expected length of configmap is %v but the actual length is %v", name, expectedResult, actualResult)
				}
			}
		})
	}
}

func TestValidateConfigMaps(t *testing.T) {
	fakeConfigMapName := "fake configmap"
	fakeNamespace := "fake-namespace"

	tests := map[string]struct {
		configmap  v1.ConfigMap
		experiment ExperimentDetails
		isErr      bool
	}{
		"Test Positive-1": {
			configmap: v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fakeConfigMapName,
					Namespace: fakeNamespace,
				},
				Data: map[string]string{
					"my-fake-key": "myfake-val",
				}},
			experiment: ExperimentDetails{
				Name:               "Fake-Exp-Name",
				Namespace:          fakeNamespace,
				JobName:            "fake-job-name",
				StatusCheckTimeout: 10,
				ConfigMaps: []v1alpha1.ConfigMap{
					{
						Name:      fakeConfigMapName,
						MountPath: "fake mountpath",
					},
				},
			},
			isErr: false,
		},
		"Test Negative-1": {
			configmap: v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fakeConfigMapName,
					Namespace: fakeNamespace,
				},
				Data: map[string]string{
					"my-fake-key": "myfake-val",
				}},
			experiment: ExperimentDetails{
				Name:               "Fake-Exp-Name",
				Namespace:          fakeNamespace,
				JobName:            "fake-job-name",
				StatusCheckTimeout: 10,
				ConfigMaps: []v1alpha1.ConfigMap{
					{
						Name: fakeConfigMapName,
					},
				},
			},
			isErr: true,
		},
		"Test Negative-2": {
			configmap: v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fakeConfigMapName,
					Namespace: fakeNamespace,
				},
				Data: map[string]string{
					"my-fake-key": "myfake-val",
				}},
			experiment: ExperimentDetails{
				Name:               "Fake-Exp-Name",
				Namespace:          fakeNamespace,
				JobName:            "fake-job-name",
				StatusCheckTimeout: 10,
				ConfigMaps: []v1alpha1.ConfigMap{
					{
						MountPath: "fake mountpath",
					},
				},
			},
			isErr: true,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateFakeClient(t)

			_, err := client.KubeClient.CoreV1().ConfigMaps(fakeNamespace).Create(&mock.configmap)
			if err != nil {
				t.Fatalf("configmap not created for %v test, err: %v", name, err)
			}

			err = mock.experiment.ValidateConfigMaps(client)
			if (!mock.isErr && err != nil) || (mock.isErr && err == nil) {
				t.Fatalf("Validation for presence of configmap failed for %v test, err: %v", name, err)
			}

		})
	}
}

func TestValidatePresenceOfConfigMapResourceInCluster(t *testing.T) {
	fakeConfigMap := "fake configmap"
	experiment := ExperimentDetails{
		Name:               "Fake-Exp-Name",
		Namespace:          "Fake NameSpace",
		JobName:            "fake-job-name",
		StatusCheckTimeout: 10,
	}

	tests := map[string]struct {
		configmap v1.ConfigMap
		isErr     bool
	}{
		"Test Positive-1": {
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
			isErr: true,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateFakeClient(t)

			if !mock.isErr {
				_, err := client.KubeClient.CoreV1().ConfigMaps(experiment.Namespace).Create(&mock.configmap)
				if err != nil {
					t.Fatalf("configmap not created for %v test, err: %v", name, err)
				}
			}

			err := client.ValidatePresenceOfConfigMapResourceInCluster(fakeConfigMap, experiment.Namespace)
			if (!mock.isErr && err != nil) || (mock.isErr && err == nil) {
				t.Fatalf("Validation for presence of configmap failed for %v test, err: %v", name, err)
			}

		})
	}
}

func TestSetConfigMaps(t *testing.T) {
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
		chaosexperiment *v1alpha1.ChaosExperiment
		chaosengine     *v1alpha1.ChaosEngine
		isErr           bool
	}{
		"Test Positive-1": {
			chaosengine: &v1alpha1.ChaosEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      engineDetails.Name,
					Namespace: engineDetails.EngineNamespace,
				},
				Spec: v1alpha1.ChaosEngineSpec{
					Experiments: []v1alpha1.ExperimentList{
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
			chaosexperiment: &v1alpha1.ChaosExperiment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      experiment.Name,
					Namespace: experiment.Namespace,
				},
				Spec: v1alpha1.ChaosExperimentSpec{
					Definition: v1alpha1.ExperimentDef{
						Image: fakeExperimentImage,
					},
				},
			},
			isErr: false,
		},
		"Test Negative-1": {
			chaosengine: &v1alpha1.ChaosEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      engineDetails.Name,
					Namespace: engineDetails.EngineNamespace,
				},
				Spec: v1alpha1.ChaosEngineSpec{
					Experiments: []v1alpha1.ExperimentList{
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
			isErr: true,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateFakeClient(t)

			if !mock.isErr {
				_, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(mock.chaosexperiment.Namespace).Create(mock.chaosexperiment)
				if err != nil {
					t.Fatalf("experiment not created for %v test, err: %v", name, err)
				}
			}

			_, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(mock.chaosengine.Namespace).Create(mock.chaosengine)
			if err != nil {
				t.Fatalf("engine not created for %v test, err: %v", name, err)
			}

			err = experiment.SetConfigMaps(client, engineDetails)
			if (!mock.isErr && err != nil) || (mock.isErr && err == nil) {
				t.Fatalf("%v Test Failed, err: %v", name, err)
			}

			actualResult := experiment.ConfigMaps
			expectedResult := mock.chaosengine.Spec.Experiments[0].Spec.Components.ConfigMaps

			if !reflect.DeepEqual(expectedResult, actualResult) && !mock.isErr {
				t.Fatalf("%v Test Failed the expectedResult '%v' is not equal to actual result '%v'", name, expectedResult, actualResult)
			}

		})
	}
}

func TestGetConfigMapsFromChaosExperiment(t *testing.T) {
	fakeExperimentImage := "fake-experiment-image"
	experiment := ExperimentDetails{
		Name:               "Fake-Exp-Name",
		Namespace:          "Fake NameSpace",
		JobName:            "fake-job-name",
		StatusCheckTimeout: 10,
	}

	tests := map[string]struct {
		chaosexperiment *v1alpha1.ChaosExperiment
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
					},
				},
			},
			isErr: false,
		},
		"Test Negative-1": {
			isErr: true,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateFakeClient(t)

			if !mock.isErr {
				_, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(mock.chaosexperiment.Namespace).Create(mock.chaosexperiment)
				if err != nil {
					t.Fatalf("experiment not created for %v test, err: %v", name, err)
				}
			}

			experimentConfigMaps, err := experiment.getConfigMapsFromChaosExperiment(client)
			if (!mock.isErr && err != nil) || (mock.isErr && err == nil) {
				t.Fatalf("%v Test Failed, err: %v", name, err)
			}

			if !mock.isErr {
				if experimentConfigMaps != nil {
					t.Fatalf("Test %q failed to get the config map from experiment: ", name)
				}
			}

		})
	}
}

func TestGetOverridingConfigMapsFromChaosEngine(t *testing.T) {
	fakeConfigMapName := "fake-configmap"
	tests := map[string]struct {
		experiment       ExperimentDetails
		engineConfigMaps []v1alpha1.ConfigMap
		isErr            bool
	}{
		"Test Positive-1": {

			experiment: ExperimentDetails{
				Name:               "Fake-Exp-Name",
				Namespace:          "Fake NameSpace",
				JobName:            "fake-job-name",
				StatusCheckTimeout: 10,
			},

			engineConfigMaps: []v1alpha1.ConfigMap{
				{
					Name:      fakeConfigMapName,
					MountPath: "fake-mount-path",
				},
			},
			isErr: false,
		},
		"Test Negative-1": {

			experiment: ExperimentDetails{
				Name:               "Fake-Exp-Name",
				Namespace:          "Fake NameSpace",
				JobName:            "fake-job-name",
				StatusCheckTimeout: 10,
			},

			engineConfigMaps: []v1alpha1.ConfigMap{},
			isErr:            false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			var err error
			mock.experiment.getOverridingConfigMapsFromChaosEngine(mock.engineConfigMaps, mock.engineConfigMaps)
			if err != nil {
				t.Fatalf("%v Test Failed, err: %v", name, err)
			}

			actualResult := mock.engineConfigMaps
			expectedResult := mock.experiment.ConfigMaps
			if !reflect.DeepEqual(expectedResult, actualResult) {
				t.Fatalf("Test %q failed: expected configmap is %v but the we get is '%v' from the experiment", name, expectedResult, actualResult)
			} else if !reflect.DeepEqual(expectedResult, actualResult) && mock.isErr {
				t.Fatalf("Test %q failed: expected configmap is %v and the we get is '%v' from the experiment", name, expectedResult, actualResult)
			}
		})
	}
}
