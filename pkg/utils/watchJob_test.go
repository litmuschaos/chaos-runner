package utils

import (
	"fmt"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	litmuschaosv1alpha1 "github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	litmusFakeClientset "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned/fake"
)

func TestPatchChaosEngineStatus(t *testing.T) {
	fakeServiceAcc := "Fake Service Account"
	fakeAppLabel := "Fake Label"
	fakeAppKind := "Fake Kind"
	fakeAnnotationCheck := "Fake Annotation Check"
	var expStatus ExperimentStatus
	expStatus.Name = "Fake exp Name"
	expStatus.Status = v1alpha1.ExperimentStatusRunning
	var engineDetails EngineDetails
	engineDetails.Name = "Fake Engine"
	engineDetails.EngineNamespace = "Fake NameSpace"

	tests := map[string]struct {
		instance *litmuschaosv1alpha1.ChaosEngine
		isErr    bool
	}{
		"Test Positive-1": {
			instance: &litmuschaosv1alpha1.ChaosEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      engineDetails.Name,
					Namespace: engineDetails.EngineNamespace,
				},
				Spec: litmuschaosv1alpha1.ChaosEngineSpec{
					ChaosServiceAccount: fakeServiceAcc,
					AnnotationCheck:     fakeAnnotationCheck,
					Appinfo: litmuschaosv1alpha1.ApplicationParams{
						Applabel: fakeAppLabel,
						Appns:    engineDetails.EngineNamespace,
						AppKind:  fakeAppKind,
					},
				},
				Status: litmuschaosv1alpha1.ChaosEngineStatus{
					EngineStatus: litmuschaosv1alpha1.EngineStatusCompleted,
					Experiments: []litmuschaosv1alpha1.ExperimentStatuses{
						{
							Name: expStatus.Name,
						},
					},
				},
			},
			isErr: false,
		},
		"Test Negative-1": {
			instance: &litmuschaosv1alpha1.ChaosEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      engineDetails.Name,
					Namespace: engineDetails.EngineNamespace,
				},
				Spec: litmuschaosv1alpha1.ChaosEngineSpec{
					ChaosServiceAccount: fakeServiceAcc,
					AnnotationCheck:     fakeAnnotationCheck,
					Appinfo: litmuschaosv1alpha1.ApplicationParams{
						Applabel: fakeAppLabel,
						Appns:    engineDetails.EngineNamespace,
						AppKind:  fakeAppKind,
					},
				},
				Status: litmuschaosv1alpha1.ChaosEngineStatus{
					EngineStatus: litmuschaosv1alpha1.EngineStatusCompleted,
				},
			},
			isErr: true,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateFakeClient(t)

			_, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(mock.instance.Namespace).Create(mock.instance)
			if err != nil {
				t.Fatalf("engine not created, err: %v", err)
			}

			err = expStatus.PatchChaosEngineStatus(engineDetails, client)
			if !mock.isErr && err != nil {
				t.Fatalf("fail to patch the engine status, err: %v", err)
			}
			if mock.isErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}

			chaosEngine, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(mock.instance.Namespace).Get(engineDetails.Name, metav1.GetOptions{})
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
	var expStatus ExperimentStatus
	expStatus.Name = "Fake-Exp-Name"
	expStatus.Status = v1alpha1.ExperimentStatusRunning
	var experiment ExperimentDetails
	experiment.Name = "Fake-Exp-Name"
	experiment.Namespace = "Fake NameSpace"
	experiment.JobName = "fake-job-name"
	experiment.StatusCheckTimeout = 10
	var engineDetails EngineDetails
	engineDetails.Name = "Fake-Engine"
	engineDetails.EngineNamespace = "Fake NameSpace"

	tests := map[string]struct {
		instance    *litmuschaosv1alpha1.ChaosEngine
		chaosresult *litmuschaosv1alpha1.ChaosResult
		chaospod    v1.Pod
		isErr       bool
	}{
		"Test Positive-1": {
			instance: &litmuschaosv1alpha1.ChaosEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      engineDetails.Name,
					Namespace: engineDetails.EngineNamespace,
				},
				Spec: litmuschaosv1alpha1.ChaosEngineSpec{
					ChaosServiceAccount: fakeServiceAcc,
					AnnotationCheck:     fakeAnnotationCheck,
					Appinfo: litmuschaosv1alpha1.ApplicationParams{
						Applabel: fakeAppLabel,
						Appns:    engineDetails.EngineNamespace,
						AppKind:  fakeAppKind,
					},
					JobCleanUpPolicy: "retain",
				},
				Status: litmuschaosv1alpha1.ChaosEngineStatus{
					EngineStatus: litmuschaosv1alpha1.EngineStatusCompleted,
					Experiments: []litmuschaosv1alpha1.ExperimentStatuses{
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
			chaosresult: &litmuschaosv1alpha1.ChaosResult{
				ObjectMeta: metav1.ObjectMeta{
					Name:      engineDetails.Name + "-" + expStatus.Name,
					Namespace: engineDetails.EngineNamespace,
				},
				Spec: litmuschaosv1alpha1.ChaosResultSpec{
					EngineName:     engineDetails.Name,
					ExperimentName: expStatus.Name,
				},
				Status: litmuschaosv1alpha1.ChaosResultStatus{
					ExperimentStatus: v1alpha1.TestStatus{
						Phase:   "Completed",
						Verdict: "Pass",
					},
				},
			},
			isErr: false,
		},
		"Test Negative-1": {
			instance: &litmuschaosv1alpha1.ChaosEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      engineDetails.Name,
					Namespace: engineDetails.EngineNamespace,
				},
				Spec: litmuschaosv1alpha1.ChaosEngineSpec{
					ChaosServiceAccount: fakeServiceAcc,
					AnnotationCheck:     fakeAnnotationCheck,
					Appinfo: litmuschaosv1alpha1.ApplicationParams{
						Applabel: fakeAppLabel,
						Appns:    engineDetails.EngineNamespace,
						AppKind:  fakeAppKind,
					},
					JobCleanUpPolicy: "retain",
				},
				Status: litmuschaosv1alpha1.ChaosEngineStatus{
					EngineStatus: litmuschaosv1alpha1.EngineStatusCompleted,
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
			chaosresult: &litmuschaosv1alpha1.ChaosResult{
				ObjectMeta: metav1.ObjectMeta{
					Name:      engineDetails.Name + "-" + expStatus.Name,
					Namespace: engineDetails.EngineNamespace,
				},
				Spec: litmuschaosv1alpha1.ChaosResultSpec{
					EngineName:     engineDetails.Name,
					ExperimentName: expStatus.Name,
				},
				Status: litmuschaosv1alpha1.ChaosResultStatus{
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

			_, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(mock.instance.Namespace).Create(mock.instance)
			if err != nil {
				t.Fatalf("engine not created, err: %v", err)
			}
			_, err = client.KubeClient.CoreV1().Pods(engineDetails.EngineNamespace).Create(&mock.chaospod)
			if err != nil {
				fmt.Printf("fail to create chaos pod, err: %v", err)
			}
			_, err = client.LitmusClient.LitmuschaosV1alpha1().ChaosResults(mock.chaosresult.Namespace).Create(mock.chaosresult)
			if err != nil {
				t.Fatalf("chaosresult not created, err: %v", err)
			}
			err = engineDetails.UpdateEngineWithResult(&experiment, client)
			if !mock.isErr && err != nil {
				t.Fatalf("fail to update chaos engine with result, err: %v", err)
			}
			if mock.isErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			chaosEngine, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(mock.instance.Namespace).Get(engineDetails.Name, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("fail to get chaos engine after status patch, err: %v", err)
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
