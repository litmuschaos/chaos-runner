package utils

import (
	"testing"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetChaosPod(t *testing.T) {
	fakeExperimentImage := "fake-experiment-image"
	experiment := ExperimentDetails{
		Name:               "Fake-Exp-Name",
		Namespace:          "Fake NameSpace",
		JobName:            "fake-jobs-name-12345",
		StatusCheckTimeout: 2,
	}
	tests := map[string]struct {
		chaospod     v1.Pod
		chaospod2    v1.Pod
		isErr        bool
		isSecondTest bool
	}{
		"Test Positive-1": {
			chaospod: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake-chaos-pod",
					Labels: map[string]string{
						"app":      "myapp",
						"job-name": experiment.JobName,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:      "fake-container",
							Image:     fakeExperimentImage,
							Resources: v1.ResourceRequirements{},
						},
					},
				},
			},
			isErr: false,
		},
		"Test Negative-1": {
			chaospod: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake-chaos-pod",
					Labels: map[string]string{
						"app":      "myapp",
						"job-name": "wrong-job-name",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:      "fake-container",
							Image:     fakeExperimentImage,
							Resources: v1.ResourceRequirements{},
						},
					},
				},
			},
			isErr: true,
		},
		"Test Negative-2": {
			chaospod: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake-chaos-pod",
					Labels: map[string]string{
						"app":      "myapp",
						"job-name": experiment.JobName,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:      "fake-container",
							Image:     fakeExperimentImage,
							Resources: v1.ResourceRequirements{},
						},
					},
				},
			},
			chaospod2: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake-chaos-pod-2",
					Labels: map[string]string{
						"app":      "myapp",
						"job-name": experiment.JobName,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:      "fake-container",
							Image:     fakeExperimentImage,
							Resources: v1.ResourceRequirements{},
						},
					},
				},
			},
			isSecondTest: true,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateFakeClient(t)

			_, err := client.KubeClient.CoreV1().Pods(experiment.Namespace).Create(&mock.chaospod)
			if err != nil {
				t.Fatalf("fail to create chaos pod for %v test, err: %v", name, err)
			}
			if mock.isSecondTest {
				_, err = client.KubeClient.CoreV1().Pods(experiment.Namespace).Create(&mock.chaospod2)
				if err != nil {
					t.Fatalf("fail to create chaos pod 2 for %v test, err: %v", name, err)
				}
			}

			_, err = GetChaosPod(&experiment, client)
			if err != nil && !mock.isErr && !mock.isSecondTest {
				t.Fatalf("%v test failed, fail to get the chaos pod, err: %v", name, err)
			} else if err == nil && mock.isErr && mock.isSecondTest {
				t.Fatalf("%v test failed, the err should not be nil", name)
			}
		})
	}
}

func TestGetChaosContainerStatus(t *testing.T) {
	fakeExperimentImage := "fake-experiment-image"
	experiment := ExperimentDetails{
		Name:               "Fake-Exp-Name",
		Namespace:          "Fake NameSpace",
		JobName:            "fake-jobs-name-12345",
		StatusCheckTimeout: 2,
	}
	tests := map[string]struct {
		chaospod v1.Pod
		isErr    bool
	}{
		"Test Positive-1": {
			chaospod: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake-chaos-pod",
					Labels: map[string]string{
						"app":      "myapp",
						"job-name": experiment.JobName,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:      experiment.JobName,
							Image:     fakeExperimentImage,
							Resources: v1.ResourceRequirements{},
						},
					},
				},
			},
			isErr: false,
		},
		"Test Negative-1": {
			chaospod: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake-chaos-pod",
					Labels: map[string]string{
						"app":      "myapp",
						"job-name": experiment.JobName,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:      "wrong-container-name",
							Image:     fakeExperimentImage,
							Resources: v1.ResourceRequirements{},
						},
					},
				},
			},
			isErr: true,
		},
		"Test Negative-2": {
			chaospod: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake-chaos-pod",
					Labels: map[string]string{
						"app":      "myapp",
						"job-name": experiment.JobName,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:      experiment.JobName,
							Image:     fakeExperimentImage,
							Resources: v1.ResourceRequirements{},
						},
					},
				},
				Status: v1.PodStatus{
					Phase: v1.PodPending,
				},
			},
			isErr: true,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateFakeClient(t)

			var ns string
			if !mock.isErr {
				ns = experiment.Namespace
			} else {
				ns = "wrong-ns"
			}
			_, err := client.KubeClient.CoreV1().Pods(ns).Create(&mock.chaospod)
			if err != nil {
				t.Fatalf("fail to create chaos pod for %v test, err: %v", name, err)
			}

			_, err = GetChaosContainerStatus(&experiment, client)
			if err != nil && !mock.isErr {
				t.Fatalf("%v test failed, fail to get the chaos pod, err: %v", name, err)
			} else if err == nil && mock.isErr {
				t.Fatalf("%v test failed, the err should not be nil", name)
			}
		})
	}
}

func TestWatchChaosContainerForCompletion(t *testing.T) {
	fakeExperimentImage := "fake-experiment-image"
	fakeNamespace := "Fake NameSpace"
	experiment := ExperimentDetails{
		Name:               "Fake-Exp-Name",
		Namespace:          fakeNamespace,
		JobName:            "fake-jobs-name-12345",
		StatusCheckTimeout: 2,
	}
	engineDetails := EngineDetails{
		Name:            "Fake Engine",
		EngineNamespace: fakeNamespace,
	}
	tests := map[string]struct {
		chaosengine *v1alpha1.ChaosEngine
		chaospod    v1.Pod
		isErr       bool
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
								Components: v1alpha1.ExperimentComponents{},
							},
						},
					},
				},
				Status: v1alpha1.ChaosEngineStatus{
					EngineStatus: v1alpha1.EngineStatusCompleted,
					Experiments: []v1alpha1.ExperimentStatuses{
						{
							Name: experiment.Name,
						},
					},
				},
			},
			chaospod: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake-chaos-pod",
					Labels: map[string]string{
						"app":      "myapp",
						"job-name": experiment.JobName,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:      experiment.JobName,
							Image:     fakeExperimentImage,
							Resources: v1.ResourceRequirements{},
						},
					},
				},
				Status: v1.PodStatus{
					Phase: v1.PodSucceeded,
					ContainerStatuses: []v1.ContainerStatus{
						{
							Name: experiment.JobName,
							State: v1.ContainerState{
								Terminated: &v1.ContainerStateTerminated{
									Reason: "Completed",
								},
							},
						},
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
								Components: v1alpha1.ExperimentComponents{},
							},
						},
					},
				},
			},
			chaospod: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake-chaos-pod",
					Labels: map[string]string{
						"app":      "myapp",
						"job-name": experiment.JobName,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:      experiment.JobName,
							Image:     fakeExperimentImage,
							Resources: v1.ResourceRequirements{},
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

			_, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(fakeNamespace).Create(mock.chaosengine)
			if err != nil {
				t.Fatalf("engine not created for %v test, err: %v", name, err)
			}

			_, err = client.KubeClient.CoreV1().Pods(fakeNamespace).Create(&mock.chaospod)
			if err != nil {
				t.Fatalf("fail to create chaos pod for %v test, err: %v", name, err)
			}

			err = engineDetails.WatchChaosContainerForCompletion(&experiment, client)
			if err != nil && !mock.isErr {
				t.Fatalf("%v failed, err: %v", name, err)
			} else if err == nil && mock.isErr {
				t.Fatalf("%v failed, the err should be nil", name)
			}
		})
	}
}
