package utils

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	"github.com/litmuschaos/chaos-runner/pkg/log"
)

func TestSetExpImageFromEngine(t *testing.T) {
	fakeNewExperimentImage := "fake-new-experiment-image"
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
		chaosengine *v1alpha1.ChaosEngine
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
									ExperimentImage: fakeNewExperimentImage,
								},
							},
						},
					},
				},
			},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateFakeClient(t)

			_, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(mock.chaosengine.Namespace).Create(mock.chaosengine)
			if err != nil {
				t.Fatalf("engine not created, err: %v", err)
			}
			expDetails := experiment.SetExpImageFromEngine(mock.chaosengine)

			actualResult := expDetails.ExpImage
			expectedResult := fakeNewExperimentImage
			log.Infof("Actual output is: %v", actualResult)
			if actualResult != expectedResult {
				t.Fatalf("Test %q failed: expectedOutput is: %v but the actualOutput is: %v", name, expectedResult, actualResult)
			}
		})
	}
}

func TestSetExpAnnotationFromEngine(t *testing.T) {
	fakeNewExperimentAnnotation := map[string]string{"my-fake-key": "myfake-val"}
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
		chaosengine *v1alpha1.ChaosEngine
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
									ExperimentAnnotations: fakeNewExperimentAnnotation,
								},
							},
						},
					},
				},
			},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {

			expDetails := experiment.SetExpAnnotationFromEngine(mock.chaosengine)

			actualResult := expDetails.Annotations
			expectedResult := fakeNewExperimentAnnotation
			log.Infof("Actual output is: %v", actualResult)
			if !reflect.DeepEqual(expectedResult, actualResult) {
				t.Fatalf("Test %q failed: expectedOutput is: %v but the actualOutput is: %v", name, expectedResult, actualResult)
			}
		})
	}
}

func TestSetResourceRequirementsFromEngine(t *testing.T) {
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
		chaosengine *v1alpha1.ChaosEngine
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
									Resources: v1.ResourceRequirements{
										Limits: v1.ResourceList{
											"cpu":    resource.MustParse("500"),
											"memory": resource.MustParse("100"),
										},
										Requests: v1.ResourceList{
											"cpu":    resource.MustParse("10"),
											"memory": resource.MustParse("5"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {

			expDetails := experiment.SetResourceRequirementsFromEngine(mock.chaosengine)
			actualResult := expDetails.ResourceRequirements
			expectedResult := mock.chaosengine.Spec.Experiments[0].Spec.Components.Resources
			if !reflect.DeepEqual(expectedResult, actualResult) {
				t.Fatalf("Test %q failed: expectedOutput is: '%v' but the actualOutput is: '%v'", name, expectedResult, actualResult)
			}
		})
	}
}
func TestSetImagePullSecretsFromEngine(t *testing.T) {
	fakeImagePullSecret := "Fake Image Pull Secret"
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
		chaosengine *v1alpha1.ChaosEngine
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
									ExperimentImagePullSecrets: []v1.LocalObjectReference{
										{
											Name: fakeImagePullSecret,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {

			expDetails := experiment.SetImagePullSecretsFromEngine(mock.chaosengine)
			actualResult := expDetails.ImagePullSecrets[0].Name
			expectedResult := fakeImagePullSecret
			if expectedResult != actualResult {
				t.Fatalf("Test %q failed: expectedOutput is: '%v' but the actualOutput is: '%v'", name, expectedResult, actualResult)
			}
		})
	}
}

func TestSetExpNodeSelectorFromEngine(t *testing.T) {
	fakeNodeSelector := make(map[string]string)
	fakeNodeSelector["my-fake-key"] = "my-fake-val"
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
		chaosengine *v1alpha1.ChaosEngine
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
									NodeSelector: fakeNodeSelector,
								},
							},
						},
					},
				},
			},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			expDetails := experiment.SetExpNodeSelectorFromEngine(mock.chaosengine)
			actualResult := expDetails.NodeSelector
			expectedResult := fakeNodeSelector
			if !reflect.DeepEqual(expectedResult, actualResult) {
				t.Fatalf("Test %q failed: expectedOutput is: '%v' but the actualOutput is: '%v'", name, expectedResult, actualResult)
			}
		})
	}
}

func TestSetTolerationsFromEngine(t *testing.T) {
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
		chaosengine *v1alpha1.ChaosEngine
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
									Tolerations: []v1.Toleration{
										{
											Key:      "fake-key",
											Operator: v1.TolerationOperator("Exists"),
											Effect:   v1.TaintEffect("NoSchedule"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {

			expDetails := experiment.SetTolerationsFromEngine(mock.chaosengine)
			actualResult := expDetails.Tolerations
			expectedResult := mock.chaosengine.Spec.Experiments[0].Spec.Components.Tolerations
			if !reflect.DeepEqual(expectedResult, actualResult) {
				t.Fatalf("Test %q failed: expectedOutput is: '%v' but the actualOutput is: '%v'", name, expectedResult, actualResult)
			}
		})
	}
}

func TestInstanceAttributeValuesFromChaosEngine(t *testing.T) {
	fakeNewExperimentAnnotation := map[string]string{"my-fake-key": "myfake-val"}
	fakeNodeSelector := make(map[string]string)
	fakeImagePullSecret := "Fake Image Pull Secret"
	fakeNewExperimentImage := "fake-new-experiment-image"
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
		chaosengine *v1alpha1.ChaosEngine
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
									NodeSelector:          fakeNodeSelector,
									ExperimentAnnotations: fakeNewExperimentAnnotation,
									ExperimentImage:       fakeNewExperimentImage,
									Resources: v1.ResourceRequirements{
										Limits: v1.ResourceList{
											"cpu":    resource.MustParse("500"),
											"memory": resource.MustParse("100"),
										},
										Requests: v1.ResourceList{
											"cpu":    resource.MustParse("10"),
											"memory": resource.MustParse("5"),
										},
									},
									Tolerations: []v1.Toleration{
										{
											Key:      "fake-key",
											Operator: v1.TolerationOperator("Exists"),
											Effect:   v1.TaintEffect("NoSchedule"),
										},
									},
									ExperimentImagePullSecrets: []v1.LocalObjectReference{
										{
											Name: fakeImagePullSecret,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateFakeClient(t)

			_, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(mock.chaosengine.Namespace).Create(mock.chaosengine)
			if err != nil {
				t.Fatalf("engine not created, err: %v", err)
			}
			if err = experiment.SetInstanceAttributeValuesFromChaosEngine(&engineDetails, client); err != nil {
				t.Fatalf("%v test failed, err: %v", name, err)
			}
		})
	}
}
