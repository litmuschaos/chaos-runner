package utils

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
)

func TestPatchSecrets(t *testing.T) {
	fakeSecretName := "fake secret"
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
		secret          v1.Secret
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
									Secrets: []v1alpha1.Secret{
										{
											Name:      fakeSecretName,
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
			secret: v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fakeSecretName,
					Namespace: experiment.Namespace,
				},
				StringData: map[string]string{
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
									Secrets: []v1alpha1.Secret{
										{
											Name:      fakeSecretName,
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

			_, err := client.KubeClient.CoreV1().Secrets(experiment.Namespace).Create(&mock.secret)
			if err != nil {
				t.Fatalf("secret not created for %v test, err: %v", name, err)
			}
			_, err = client.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(mock.chaosexperiment.Namespace).Create(mock.chaosexperiment)
			if err != nil {
				t.Fatalf("experiment not created for %v test, err: %v", name, err)
			}
			_, err = client.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(mock.chaosengine.Namespace).Create(mock.chaosengine)
			if err != nil {
				t.Fatalf("engine not created for %v test, err: %v", name, err)
			}
			err = experiment.PatchSecrets(client, engineDetails)
			if !mock.isErr && err != nil {
				t.Fatalf("fail to patch the secret, err: %v", err)
			}
			if mock.isErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}

			if !mock.isErr {
				actualResult := len(experiment.Secrets)
				expectedResult := 1
				if actualResult != expectedResult {
					t.Fatalf("Test %q failed: expected length of secret is %v but the actual length is %v", name, expectedResult, actualResult)
				}
			}
		})
	}
}

func TestSetSecrets(t *testing.T) {
	fakeSecretName := "fake secret"
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
									Secrets: []v1alpha1.Secret{
										{
											Name:      fakeSecretName,
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
									Secrets: []v1alpha1.Secret{
										{
											Name:      fakeSecretName,
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

			err = experiment.SetSecrets(client, engineDetails)
			if (!mock.isErr && err != nil) || (mock.isErr && err == nil) {
				t.Fatalf("%v Test Failed, err: %v", name, err)
			}

			actualResult := experiment.Secrets
			expectedResult := mock.chaosengine.Spec.Experiments[0].Spec.Components.Secrets

			if !reflect.DeepEqual(expectedResult, actualResult) && !mock.isErr {
				t.Fatalf("%v Test Failed the expectedResult '%v' is not equal to actual result '%v'", name, expectedResult, actualResult)
			}

		})
	}
}

func TestValidateSecrets(t *testing.T) {
	fakeSecretName := "fake secret"
	fakeNamespace := "fake-namespace"

	tests := map[string]struct {
		secret     v1.Secret
		experiment ExperimentDetails
		isErr      bool
	}{
		"Test Positive-1": {
			secret: v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fakeSecretName,
					Namespace: fakeNamespace,
				},
				StringData: map[string]string{
					"my-fake-key": "myfake-val",
				}},
			experiment: ExperimentDetails{
				Name:               "Fake-Exp-Name",
				Namespace:          fakeNamespace,
				JobName:            "fake-job-name",
				StatusCheckTimeout: 10,
				Secrets: []v1alpha1.Secret{
					{
						Name:      fakeSecretName,
						MountPath: "fake mountpath",
					},
				},
			},
			isErr: false,
		},
		"Test Negative-1": {
			secret: v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fakeSecretName,
					Namespace: fakeNamespace,
				},
				StringData: map[string]string{
					"my-fake-key": "myfake-val",
				}},
			experiment: ExperimentDetails{
				Name:               "Fake-Exp-Name",
				Namespace:          fakeNamespace,
				JobName:            "fake-job-name",
				StatusCheckTimeout: 10,
				Secrets: []v1alpha1.Secret{
					{
						Name: fakeSecretName,
					},
				},
			},
			isErr: true,
		},
		"Test Negative-2": {
			secret: v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fakeSecretName,
					Namespace: fakeNamespace,
				},
				StringData: map[string]string{
					"my-fake-key": "myfake-val",
				}},
			experiment: ExperimentDetails{
				Name:               "Fake-Exp-Name",
				Namespace:          fakeNamespace,
				JobName:            "fake-job-name",
				StatusCheckTimeout: 10,
				Secrets: []v1alpha1.Secret{
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

			_, err := client.KubeClient.CoreV1().Secrets(fakeNamespace).Create(&mock.secret)
			if err != nil {
				t.Fatalf("secret not created for %v test, err: %v", name, err)
			}

			err = mock.experiment.ValidateSecrets(client)
			if (!mock.isErr && err != nil) || (mock.isErr && err == nil) {
				t.Fatalf("Validation for presence of secret failed for %v test, err: %v", name, err)
			}

		})
	}
}

func TestGetSecretsFromChaosExperiment(t *testing.T) {
	fakeSecretName := "fake secret"
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
						Secrets: []v1alpha1.Secret{
							{
								Name:      fakeSecretName,
								MountPath: "fake mountpath",
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
				t.Fatalf("experiment not created for %v test, err: %v", name, err)
			}

			experimentSecrets, err := experiment.getSecretsFromChaosExperiment(client)
			if err != nil && !mock.isErr {
				t.Fatalf("%v Test Failed, err: %v", name, err)
			}
			if !mock.isErr {
				if experimentSecrets[0].Name != fakeSecretName {
					t.Fatalf("Test %q failed to get the secret from experiment: ", name)
				}
			}

		})
	}
}

func TestGetOverridingSecretsFromChaosEngine(t *testing.T) {
	fakeSecretName := "fake-experiment-image"
	tests := map[string]struct {
		experiment    ExperimentDetails
		engineSecrets []v1alpha1.Secret
		isErr         bool
	}{
		"Test Positive-1": {

			experiment: ExperimentDetails{
				Name:               "Fake-Exp-Name",
				Namespace:          "Fake NameSpace",
				JobName:            "fake-job-name",
				StatusCheckTimeout: 10,
			},

			engineSecrets: []v1alpha1.Secret{
				{
					Name:      fakeSecretName,
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

			engineSecrets: []v1alpha1.Secret{},
			isErr:         false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			var err error
			mock.experiment.getOverridingSecretsFromChaosEngine(mock.engineSecrets, mock.engineSecrets)
			if err != nil {
				t.Fatalf("%v Test Failed, err: %v", name, err)
			}

			actualResult := mock.engineSecrets
			expectedResult := mock.experiment.Secrets
			if !reflect.DeepEqual(expectedResult, actualResult) && !mock.isErr {
				t.Fatalf("Test %q failed: expected secret is %v but the we get is '%v' from the experiment", name, expectedResult, actualResult)
			} else if !reflect.DeepEqual(expectedResult, actualResult) && mock.isErr {
				t.Fatalf("Test %q failed: expected secret is %v and the we get is '%v' from the experiment", name, expectedResult, actualResult)
			}
		})
	}
}
