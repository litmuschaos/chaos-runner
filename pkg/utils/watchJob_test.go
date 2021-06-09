package utils

import (
	"fmt"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	litmusFakeClientset "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned/fake"
)

func TestPatchChaosEngineStatus(t *testing.T) {
	fakeServiceAcc := "Fake Service Account"
	fakeAppLabel := "Fake Label"
	fakeAppKind := "Fake Kind"
	fakeAnnotationCheck := "Fake Annotation Check"
	expStatus := ExperimentStatus{
		Name:   "Fake exp Name",
		Status: v1alpha1.ExperimentStatusRunning,
	}
	engineDetails := EngineDetails{
		Name:            "Fake Engine",
		EngineNamespace: "Fake NameSpace",
	}

	tests := map[string]struct {
		chaosengine *v1alpha1.ChaosEngine
		isErr       bool
	}{
		"Test Positive-1": {
			chaosengine: &v1alpha1.ChaosEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      engineDetails.Name,
					Namespace: engineDetails.EngineNamespace,
				},
				Spec: v1alpha1.ChaosEngineSpec{
					ChaosServiceAccount: fakeServiceAcc,
					AnnotationCheck:     fakeAnnotationCheck,
					Appinfo: v1alpha1.ApplicationParams{
						Applabel: fakeAppLabel,
						Appns:    engineDetails.EngineNamespace,
						AppKind:  fakeAppKind,
					},
				},
				Status: v1alpha1.ChaosEngineStatus{
					EngineStatus: v1alpha1.EngineStatusCompleted,
					Experiments: []v1alpha1.ExperimentStatuses{
						{
							Name: expStatus.Name,
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
					ChaosServiceAccount: fakeServiceAcc,
					AnnotationCheck:     fakeAnnotationCheck,
					Appinfo: v1alpha1.ApplicationParams{
						Applabel: fakeAppLabel,
						Appns:    engineDetails.EngineNamespace,
						AppKind:  fakeAppKind,
					},
				},
				Status: v1alpha1.ChaosEngineStatus{
					EngineStatus: v1alpha1.EngineStatusCompleted,
				},
			},
			isErr: true,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateFakeClient(t)

			_, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(mock.chaosengine.Namespace).Create(mock.chaosengine)
			if err != nil {
				t.Fatalf("engine not created for %v test, err: %v", name, err)
			}

			err = expStatus.PatchChaosEngineStatus(engineDetails, client)
			if !mock.isErr && err != nil {
				t.Fatalf("fail to patch the engine status for %v test, err: %v", name, err)
			}
			if mock.isErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}

			chaosEngine, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(mock.chaosengine.Namespace).Get(engineDetails.Name, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("fail to get chaos engine after status patch, err: %v", err)
			}

			if !mock.isErr {
				actualResult := chaosEngine.Status.Experiments[0].Status
				expectedResult := expStatus.Status
				println(actualResult)
				if expectedResult != actualResult {
					t.Fatalf("Test %q failed: expected result is %v, got %v", name, expectedResult, actualResult)
				}
			}
			if mock.isErr && len(chaosEngine.Status.Experiments) != 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
		})
	}
}

func TestUpdateEngineWithResult(t *testing.T) {
	fakeServiceAcc := "Fake Service Account"
	fakeAppLabel := "Fake Label"
	fakeAppKind := "Fake Kind"
	fakeAnnotationCheck := "Fake Annotation Check"
	expStatus := ExperimentStatus{
		Name:   "Fake-Exp-Name",
		Status: v1alpha1.ExperimentStatusRunning,
	}
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
		chaosresult *v1alpha1.ChaosResult
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
					ChaosServiceAccount: fakeServiceAcc,
					AnnotationCheck:     fakeAnnotationCheck,
					Appinfo: v1alpha1.ApplicationParams{
						Applabel: fakeAppLabel,
						Appns:    engineDetails.EngineNamespace,
						AppKind:  fakeAppKind,
					},
					JobCleanUpPolicy: "retain",
				},
				Status: v1alpha1.ChaosEngineStatus{
					EngineStatus: v1alpha1.EngineStatusCompleted,
					Experiments: []v1alpha1.ExperimentStatuses{
						{
							Name:   expStatus.Name,
							Status: expStatus.Status,
						},
					},
				},
			},
			chaospod: v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      experiment.JobName,
					Namespace: experiment.Namespace,
					Labels: map[string]string{
						"job-name": experiment.JobName,
					},
				},
			},
			chaosresult: &v1alpha1.ChaosResult{
				ObjectMeta: metav1.ObjectMeta{
					Name:      engineDetails.Name + "-" + expStatus.Name,
					Namespace: engineDetails.EngineNamespace,
				},
				Spec: v1alpha1.ChaosResultSpec{
					EngineName:     engineDetails.Name,
					ExperimentName: expStatus.Name,
				},
				Status: v1alpha1.ChaosResultStatus{
					ExperimentStatus: v1alpha1.TestStatus{
						Phase:   "Completed",
						Verdict: "Pass",
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
					ChaosServiceAccount: fakeServiceAcc,
					AnnotationCheck:     fakeAnnotationCheck,
					Appinfo: v1alpha1.ApplicationParams{
						Applabel: fakeAppLabel,
						Appns:    engineDetails.EngineNamespace,
						AppKind:  fakeAppKind,
					},
					JobCleanUpPolicy: "retain",
				},
				Status: v1alpha1.ChaosEngineStatus{
					EngineStatus: v1alpha1.EngineStatusCompleted,
				},
			},
			chaospod: v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      experiment.JobName,
					Namespace: experiment.Namespace,
					Labels: map[string]string{
						"job-name": experiment.JobName,
					},
				},
			},
			chaosresult: &v1alpha1.ChaosResult{
				ObjectMeta: metav1.ObjectMeta{
					Name:      engineDetails.Name + "-" + expStatus.Name,
					Namespace: engineDetails.EngineNamespace,
				},
				Spec: v1alpha1.ChaosResultSpec{
					EngineName:     engineDetails.Name,
					ExperimentName: expStatus.Name,
				},
				Status: v1alpha1.ChaosResultStatus{
					ExperimentStatus: v1alpha1.TestStatus{
						Phase:   "Completed",
						Verdict: "Pass",
					},
				},
			},
			isErr: true,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateFakeClient(t)

			_, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(mock.chaosengine.Namespace).Create(mock.chaosengine)
			if err != nil {
				t.Fatalf("engine not created for %v test, err: %v", name, err)
			}
			_, err = client.KubeClient.CoreV1().Pods(engineDetails.EngineNamespace).Create(&mock.chaospod)
			if err != nil {
				fmt.Printf("fail to create chaos pod for %v test, err: %v", name, err)
			}
			_, err = client.LitmusClient.LitmuschaosV1alpha1().ChaosResults(mock.chaosresult.Namespace).Create(mock.chaosresult)
			if err != nil {
				t.Fatalf("chaosresult not created for %v test, err: %v", name, err)
			}
			err = engineDetails.UpdateEngineWithResult(&experiment, client)
			if !mock.isErr && err != nil {
				t.Fatalf("fail to update chaos engine with result for %v test, err: %v", name, err)
			}
			if mock.isErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			chaosEngine, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(mock.chaosengine.Namespace).Get(engineDetails.Name, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("fail to get chaosengine after status patch for %v test, err: %v", name, err)
			}
			if !mock.isErr {
				actualResult := chaosEngine.Status.Experiments[0].Status
				expectedResult := v1alpha1.ExperimentStatusCompleted
				println(actualResult)
				if expectedResult != actualResult {
					t.Fatalf("Test %q failed: expected result is %v, got %v", name, expectedResult, actualResult)
				}
			}
			if mock.isErr && len(chaosEngine.Status.Experiments) != 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
		})
	}
}

func TestDeleteJobAccordingToJobCleanUpPolicy(t *testing.T) {
	fakeServiceAcc := "Fake Service Account"
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
		expjob      batchv1.Job
		isErr       bool
		retain      bool
	}{
		"Test Positive-1": {
			chaosengine: &v1alpha1.ChaosEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      engineDetails.Name,
					Namespace: engineDetails.EngineNamespace,
				},
				Spec: v1alpha1.ChaosEngineSpec{
					ChaosServiceAccount: fakeServiceAcc,
					JobCleanUpPolicy:    v1alpha1.CleanUpPolicyDelete,
				},
				Status: v1alpha1.ChaosEngineStatus{
					EngineStatus: v1alpha1.EngineStatusCompleted,
				},
			},
			expjob: batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      experiment.JobName,
					Namespace: experiment.Namespace,
					Labels: map[string]string{
						"job-name": experiment.JobName,
					},
				},
			},
			isErr: true,
		},
		"Test Positive-2": {
			chaosengine: &v1alpha1.ChaosEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      engineDetails.Name,
					Namespace: engineDetails.EngineNamespace,
				},
				Spec: v1alpha1.ChaosEngineSpec{
					ChaosServiceAccount: fakeServiceAcc,
					JobCleanUpPolicy:    v1alpha1.CleanUpPolicyRetain,
				},
				Status: v1alpha1.ChaosEngineStatus{
					EngineStatus: v1alpha1.EngineStatusCompleted,
				},
			},
			expjob: batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      experiment.JobName,
					Namespace: experiment.Namespace,
					Labels: map[string]string{
						"job-name": experiment.JobName,
					},
				},
			},
			isErr: true,
		},
		"Test Negative-1": {
			chaosengine: &v1alpha1.ChaosEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      engineDetails.Name,
					Namespace: engineDetails.EngineNamespace,
				},
				Spec: v1alpha1.ChaosEngineSpec{
					ChaosServiceAccount: fakeServiceAcc,
				},
				Status: v1alpha1.ChaosEngineStatus{
					EngineStatus: v1alpha1.EngineStatusCompleted,
				},
			},
			expjob: batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      experiment.JobName,
					Namespace: experiment.Namespace,
					Labels: map[string]string{
						"job-name": experiment.JobName,
					},
				},
			},
			isErr: false,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateFakeClient(t)

			_, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(mock.chaosengine.Namespace).Create(mock.chaosengine)
			if err != nil {
				t.Fatalf("engine not created for %v test, err: %v", name, err)
			}
			_, err = client.KubeClient.BatchV1().Jobs(engineDetails.EngineNamespace).Create(&mock.expjob)
			if err != nil {
				t.Fatalf("fail to create exp job pod for %v test, err: %v", name, err)
			}
			cleanupPolicy, err := engineDetails.DeleteJobAccordingToJobCleanUpPolicy(&experiment, client)
			if err != nil {
				t.Fatalf("fail to create exp job for %v test, err: %v", name, err)
			}

			jobList, err := client.KubeClient.BatchV1().Jobs(engineDetails.EngineNamespace).List(metav1.ListOptions{LabelSelector: "job-name=" + experiment.JobName})
			if !mock.isErr && err != nil && len(jobList.Items) != 0 {
				t.Fatalf("[%v] test failed experiment job is not deleted when the job cleanup policy is %v , err: %v", name, err, cleanupPolicy)
			}
			if mock.isErr && err != nil && len(jobList.Items) == 0 {
				t.Fatalf("[%v] test failed experiment job is not retained when the job cleanup policy is %v , err: %v", name, err, cleanupPolicy)
			}
		})
	}
}

func CreateFakeClient(t *testing.T) ClientSets {

	clients := ClientSets{}
	clients.SetFakeClient()
	return clients
}

// SetFakeClient initilizes the fake required clientsets
func (clients *ClientSets) SetFakeClient() {

	// Load kubernetes client set by preloading with k8s objects.
	clients.KubeClient = fake.NewSimpleClientset([]runtime.Object{}...)

	// Load litmus client set by preloading with litmus objects.
	clients.LitmusClient = litmusFakeClientset.NewSimpleClientset([]runtime.Object{}...)
}
