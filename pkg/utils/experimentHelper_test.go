package utils

import (
	"reflect"
	"testing"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateExperimentList(t *testing.T) {

	tests := map[string]struct {
		engineDetails EngineDetails
		isErr         bool
	}{
		"Test Positive-1": {
			engineDetails: EngineDetails{
				Name:            "Fake Engine",
				EngineNamespace: "Fake NameSpace",
				Experiments: []string{
					"fake-exp-1",
					"fake-exp-2",
				},
			},
			isErr: false,
		},
		"Test Negative-1": {
			engineDetails: EngineDetails{
				Name:            "Fake Engine",
				EngineNamespace: "Fake NameSpace",
			},
			isErr: true,
		},
	}

	for name, moke := range tests {
		t.Run(name, func(t *testing.T) {

			ExpList := moke.engineDetails.CreateExperimentList()
			if len(ExpList) == 0 && !moke.isErr {
				t.Fatalf("%v test failed as the experiment list is still empty", name)
			} else if len(ExpList) != 0 && moke.isErr {
				t.Fatalf("%v test failed as the experiment list is non empty for non empty experiment details on engine", name)
			}

			for i := range ExpList {
				if ExpList[i].Name != moke.engineDetails.Experiments[i] && !moke.isErr {
					t.Fatalf("The expected experimentName is %v but got %v", ExpList[i].Name, moke.engineDetails.Experiments[i])
				}
			}
		})
	}
}

func TestNewExperimentDetails(t *testing.T) {

	tests := map[string]struct {
		engineDetails EngineDetails
	}{
		"Test Positive-1": {
			engineDetails: EngineDetails{
				Name:            "Fake Engine",
				EngineNamespace: "Fake NameSpace",
				Experiments: []string{
					"fake-exp-1",
				},
			},
		},
	}

	for name, moke := range tests {
		t.Run(name, func(t *testing.T) {

			ExpDetails := moke.engineDetails.NewExperimentDetails(0)

			if ExpDetails.Name != moke.engineDetails.Experiments[0] {
				t.Fatalf("%v test failed to create experiment details", name)
			}
		})
	}
}

func TestSetDefaultEnvFromChaosExperiment(t *testing.T) {
	fakeExperimentImage := "fake-experiment-image"
	fakeTotalChaosDuration := "20"
	experiment := ExperimentDetails{
		Name:               "Fake-Exp-Name",
		Namespace:          "Fake NameSpace",
		JobName:            "fake-job-name",
		StatusCheckTimeout: 10,
		envMap:             map[string]v1.EnvVar{},
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
						ENVList: []v1.EnvVar{
							{
								Name:  "TOTAL_CHAOS_DURATION",
								Value: fakeTotalChaosDuration,
							},
							{
								Name:  "CHAOS_INTERVAL",
								Value: "10",
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
						Image:   fakeExperimentImage,
						ENVList: []v1.EnvVar{},
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

			err = experiment.SetDefaultEnvFromChaosExperiment(client)
			if err != nil {
				t.Fatalf("%v Test Failed, err: %v", name, err)
			}
			expectedResult := experiment.envMap["TOTAL_CHAOS_DURATION"].Value
			actualResult := fakeTotalChaosDuration

			if expectedResult != actualResult && !mock.isErr {
				t.Fatalf("Test %q failed to set the default env from experiment", name)
			}
		})
	}
}

func TestSetDefaultAttributeValuesFromChaosExperiment(t *testing.T) {
	fakeExperimentImage := "fake-experiment-image"
	fakeExperimentImagePullPolicy := "fake-exp-pull-policy"
	fakeTotalChaosDuration := "20"
	experiment := ExperimentDetails{
		Name:               "Fake-Exp-Name",
		Namespace:          "Fake NameSpace",
		JobName:            "fake-job-name",
		StatusCheckTimeout: 10,
		envMap:             map[string]v1.EnvVar{},
		ExpLabels:          map[string]string{},
		ExpArgs:            []string{},
		ExpCommand:         []string{},
		ConfigMaps:         []v1alpha1.ConfigMap{},
		Secrets:            []v1alpha1.Secret{},
		ExpImage:           "",
		ExpImagePullPolicy: "",
		SecurityContext:    v1alpha1.SecurityContext{},
		HostPID:            false,
	}

	engineDetails := EngineDetails{
		Name:            "Fake Engine",
		EngineNamespace: "Fake NameSpace",
		UID:             "fake-chaosuid",
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
						Image:           fakeExperimentImage,
						ImagePullPolicy: v1.PullPolicy(fakeExperimentImagePullPolicy),
						Labels: map[string]string{
							"fake-label-key": "fake-label-value",
							"chaosUID":       "",
						},
						ENVList: []v1.EnvVar{
							{
								Name:  "TOTAL_CHAOS_DURATION",
								Value: fakeTotalChaosDuration,
							},
							{
								Name:  "CHAOS_INTERVAL",
								Value: "10",
							},
						},
						HostPID: false,
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
						ImagePullPolicy: v1.PullPolicy(fakeExperimentImagePullPolicy),
						Labels: map[string]string{
							"fake-label-key": "fake-label-value",
							"chaosUID":       "",
						},
						ENVList: []v1.EnvVar{
							{
								Name:  "TOTAL_CHAOS_DURATION",
								Value: fakeTotalChaosDuration,
							},
							{
								Name:  "CHAOS_INTERVAL",
								Value: "10",
							},
						},
						HostPID: false,
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

			err = experiment.SetDefaultAttributeValuesFromChaosExperiment(client, &engineDetails)
			if err != nil {
				t.Fatalf("%v Test Failed, err: %v", name, err)
			}

			if !mock.isErr {
				expectedResult := experiment.ExpImage
				actualResult := fakeExperimentImage
				if expectedResult != actualResult {
					t.Fatalf("Test %q failed to set the default env from experiment", name)
				}
			}
		})
	}
}

func TestSetValueFromChaosResources(t *testing.T) {
	fakeExperimentImage := "fake-exp-image"
	fakeExperimentImagePullPolicy := v1.PullPolicy("fake-exp-image-pull-policy")
	fakeExperimentArgs := "fake-exp-args"
	fakeHostPID := false
	experiment := ExperimentDetails{
		Name:               "Fake-Exp-Name",
		Namespace:          "Fake NameSpace",
		JobName:            "fake-job-name",
		StatusCheckTimeout: 10,
		envMap:             map[string]v1.EnvVar{},
		ExpLabels:          map[string]string{},
		ExpArgs:            []string{},
		ConfigMaps:         []v1alpha1.ConfigMap{},
		Secrets:            []v1alpha1.Secret{},
		ExpImage:           "",
		ExpImagePullPolicy: "",
		SecurityContext:    v1alpha1.SecurityContext{},
		HostPID:            false,
	}
	engineDetails := EngineDetails{
		Name:            "Fake Engine",
		EngineNamespace: "Fake NameSpace",
		UID:             "fake-chaosuid",
	}
	expStatus := ExperimentStatus{
		Name:   "Fake-Exp-Name",
		Status: v1alpha1.ExperimentStatusRunning,
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
				Spec: v1alpha1.ChaosEngineSpec{},
				Status: v1alpha1.ChaosEngineStatus{
					EngineStatus: v1alpha1.EngineStatusInitialized,
					Experiments: []v1alpha1.ExperimentStatuses{
						{
							Name: expStatus.Name,
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
						Image:           fakeExperimentImage,
						ImagePullPolicy: fakeExperimentImagePullPolicy,
						Labels: map[string]string{
							"fake-label-key": "fake-label-value",
						},
						Args: []string{
							fakeExperimentArgs,
						},
						SecurityContext: v1alpha1.SecurityContext{
							PodSecurityContext: v1.PodSecurityContext{
								RunAsUser: ptrint64(1000),
							},
						},
						HostPID: fakeHostPID,
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
						Image:           fakeExperimentImage,
						ImagePullPolicy: fakeExperimentImagePullPolicy,
						Labels: map[string]string{
							"fake-label-key": "fake-label-value",
						},
						Args: []string{
							fakeExperimentArgs,
						},
						SecurityContext: v1alpha1.SecurityContext{
							PodSecurityContext: v1.PodSecurityContext{
								RunAsUser: ptrint64(1000),
							},
						},
						HostPID: fakeHostPID,
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

			if !mock.isErr {
				_, err = client.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(mock.chaosengine.Namespace).Create(mock.chaosengine)
				if err != nil {
					t.Fatalf("engine not created for %v test, err: %v", name, err)
				}
			}

			err = experiment.SetValueFromChaosResources(&engineDetails, client)
			if err != nil && !mock.isErr {
				t.Fatalf("%v Test failed unable to set chaos resources, err: %v", name, err)
			} else if err == nil && mock.isErr {
				t.Fatalf("%v Test failed the expected error should not be nil", name)
			}
			if !mock.isErr {
				expectedResult := experiment.HostPID
				actualResult := mock.chaosexperiment.Spec.Definition.HostPID
				if expectedResult != actualResult {
					t.Fatalf("Test %q failed to set the default env from experiment", name)
				}
			}
		})
	}
}

func TestSetLabels(t *testing.T) {
	experiment := ExperimentDetails{
		Name:               "Fake-Exp-Name",
		Namespace:          "Fake NameSpace",
		JobName:            "fake-job-name",
		StatusCheckTimeout: 10,
		envMap:             map[string]v1.EnvVar{},
		ExpLabels:          map[string]string{},
	}

	engineDetails := EngineDetails{
		Name:            "Fake Engine",
		EngineNamespace: "Fake NameSpace",
		UID:             "fake-chaosuid",
	}

	tests := map[string]struct {
		chaosexperiment *v1alpha1.ChaosExperiment
	}{
		"Test Positive-1": {
			chaosexperiment: &v1alpha1.ChaosExperiment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      experiment.Name,
					Namespace: experiment.Namespace,
				},
				Spec: v1alpha1.ChaosExperimentSpec{
					Definition: v1alpha1.ExperimentDef{
						Labels: map[string]string{
							"fake-label-key": "fake-label-value",
						},
					},
				},
			},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateFakeClient(t)

			_, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(mock.chaosexperiment.Namespace).Create(mock.chaosexperiment)
			if err != nil {
				t.Fatalf("experiment not created for %v test, err: %v", name, err)
			}
			experimentSpec, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(mock.chaosexperiment.Namespace).Get(mock.chaosexperiment.Name, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("fail to get the chaosexperiment for %v test, err: %v", name, err)
			}

			expDetails := experiment.SetLabels(experimentSpec, &engineDetails)
			expectedResult := expDetails.ExpLabels["fake-label-key"]
			actualResult := mock.chaosexperiment.Spec.Definition.Labels["fake-label-key"]
			if !reflect.DeepEqual(expectedResult, actualResult) {
				t.Fatalf("Test %q failed to set the default env from experiment", name)
			}

		})
	}
}

func TestSetImage(t *testing.T) {
	fakeExperimentImage := "fake-exp-image"
	experiment := ExperimentDetails{
		Name:               "Fake-Exp-Name",
		Namespace:          "Fake NameSpace",
		JobName:            "fake-job-name",
		StatusCheckTimeout: 10,
		envMap:             map[string]v1.EnvVar{},
		ExpLabels:          map[string]string{},
	}

	tests := map[string]struct {
		chaosexperiment *v1alpha1.ChaosExperiment
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
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateFakeClient(t)

			_, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(mock.chaosexperiment.Namespace).Create(mock.chaosexperiment)
			if err != nil {
				t.Fatalf("experiment not created for %v test, err: %v", name, err)
			}
			experimentSpec, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(mock.chaosexperiment.Namespace).Get(mock.chaosexperiment.Name, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("fail to get the chaosexperiment for %v test, err: %v", name, err)
			}

			expDetails := experiment.SetImage(experimentSpec)
			expectedResult := expDetails.ExpImage
			actualResult := mock.chaosexperiment.Spec.Definition.Image
			if expectedResult != actualResult {
				t.Fatalf("Test %q failed to set the default env from experiment", name)
			}
		})
	}
}

func TestSetImagePullPolicy(t *testing.T) {
	fakeExperimentImage := "fake-exp-image"
	fakeExperimentImagePullPolicy := v1.PullPolicy("fake-exp-image-pull-policy")
	experiment := ExperimentDetails{
		Name:               "Fake-Exp-Name",
		Namespace:          "Fake NameSpace",
		JobName:            "fake-job-name",
		StatusCheckTimeout: 10,
		envMap:             map[string]v1.EnvVar{},
		ExpLabels:          map[string]string{},
	}

	tests := map[string]struct {
		chaosexperiment *v1alpha1.ChaosExperiment
	}{
		"Test Positive-1": {
			chaosexperiment: &v1alpha1.ChaosExperiment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      experiment.Name,
					Namespace: experiment.Namespace,
				},
				Spec: v1alpha1.ChaosExperimentSpec{
					Definition: v1alpha1.ExperimentDef{
						Image:           fakeExperimentImage,
						ImagePullPolicy: fakeExperimentImagePullPolicy,
					},
				},
			},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateFakeClient(t)

			_, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(mock.chaosexperiment.Namespace).Create(mock.chaosexperiment)
			if err != nil {
				t.Fatalf("experiment not created for %v test, err: %v", name, err)
			}
			experimentSpec, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(mock.chaosexperiment.Namespace).Get(mock.chaosexperiment.Name, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("fail to get the chaosexperiment for %v test, err: %v", name, err)
			}

			expDetails := experiment.SetImagePullPolicy(experimentSpec)
			expectedResult := expDetails.ExpImagePullPolicy
			actualResult := mock.chaosexperiment.Spec.Definition.ImagePullPolicy
			if expectedResult != actualResult {
				t.Fatalf("Test %q failed to set the default env from experiment", name)
			}
		})
	}
}

func TestSetArgs(t *testing.T) {
	fakeExperimentArgs := "fake-exp-args"
	experiment := ExperimentDetails{
		Name:               "Fake-Exp-Name",
		Namespace:          "Fake NameSpace",
		JobName:            "fake-job-name",
		StatusCheckTimeout: 10,
		envMap:             map[string]v1.EnvVar{},
		ExpLabels:          map[string]string{},
		ExpArgs:            []string{},
	}

	tests := map[string]struct {
		chaosexperiment *v1alpha1.ChaosExperiment
	}{
		"Test Positive-1": {
			chaosexperiment: &v1alpha1.ChaosExperiment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      experiment.Name,
					Namespace: experiment.Namespace,
				},
				Spec: v1alpha1.ChaosExperimentSpec{
					Definition: v1alpha1.ExperimentDef{
						Args: []string{
							fakeExperimentArgs,
						},
					},
				},
			},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateFakeClient(t)

			_, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(mock.chaosexperiment.Namespace).Create(mock.chaosexperiment)
			if err != nil {
				t.Fatalf("experiment not created for %v test, err: %v", name, err)
			}
			experimentSpec, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(mock.chaosexperiment.Namespace).Get(mock.chaosexperiment.Name, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("fail to get the chaosexperiment for %v test, err: %v", name, err)
			}

			expDetails := experiment.SetArgs(experimentSpec)
			expectedResult := expDetails.ExpArgs
			actualResult := mock.chaosexperiment.Spec.Definition.Args
			if !reflect.DeepEqual(expectedResult, actualResult) {
				t.Fatalf("Test %q failed to set the default env from experiment", name)
			}
		})
	}
}

func TestSetCommand(t *testing.T) {
	fakeExperimentCommand := "fake-exp-command"
	experiment := ExperimentDetails{
		Name:               "Fake-Exp-Name",
		Namespace:          "Fake NameSpace",
		JobName:            "fake-job-name",
		StatusCheckTimeout: 10,
		envMap:             map[string]v1.EnvVar{},
		ExpLabels:          map[string]string{},
		ExpArgs:            []string{},
	}

	tests := map[string]struct {
		chaosexperiment *v1alpha1.ChaosExperiment
	}{
		"Test Positive-1": {
			chaosexperiment: &v1alpha1.ChaosExperiment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      experiment.Name,
					Namespace: experiment.Namespace,
				},
				Spec: v1alpha1.ChaosExperimentSpec{
					Definition: v1alpha1.ExperimentDef{
						Command: []string{
							fakeExperimentCommand,
						},
					},
				},
			},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateFakeClient(t)

			_, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(mock.chaosexperiment.Namespace).Create(mock.chaosexperiment)
			if err != nil {
				t.Fatalf("experiment not created for %v test, err: %v", name, err)
			}
			experimentSpec, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(mock.chaosexperiment.Namespace).Get(mock.chaosexperiment.Name, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("fail to get the chaosexperiment for %v test, err: %v", name, err)
			}

			expDetails := experiment.SetCommand(experimentSpec)
			expectedResult := expDetails.ExpCommand
			actualResult := mock.chaosexperiment.Spec.Definition.Command
			if !reflect.DeepEqual(expectedResult, actualResult) {
				t.Fatalf("Test %q failed to set the command from experiment", name)
			}
		})
	}
}

func TestSetSecurityContext(t *testing.T) {
	experiment := ExperimentDetails{
		Name:               "Fake-Exp-Name",
		Namespace:          "Fake NameSpace",
		JobName:            "fake-job-name",
		StatusCheckTimeout: 10,
		envMap:             map[string]v1.EnvVar{},
		ExpLabels:          map[string]string{},
		ExpArgs:            []string{},
		SecurityContext:    v1alpha1.SecurityContext{},
	}

	tests := map[string]struct {
		chaosexperiment *v1alpha1.ChaosExperiment
	}{
		"Test Positive-1": {
			chaosexperiment: &v1alpha1.ChaosExperiment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      experiment.Name,
					Namespace: experiment.Namespace,
				},
				Spec: v1alpha1.ChaosExperimentSpec{
					Definition: v1alpha1.ExperimentDef{
						SecurityContext: v1alpha1.SecurityContext{
							PodSecurityContext: v1.PodSecurityContext{
								RunAsUser: ptrint64(1000),
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

			_, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(mock.chaosexperiment.Namespace).Create(mock.chaosexperiment)
			if err != nil {
				t.Fatalf("experiment not created for %v test, err: %v", name, err)
			}
			experimentSpec, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(mock.chaosexperiment.Namespace).Get(mock.chaosexperiment.Name, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("fail to get the chaosexperiment for %v test, err: %v", name, err)
			}

			expDetails := experiment.SetSecurityContext(experimentSpec)
			expectedResult := expDetails.SecurityContext
			actualResult := mock.chaosexperiment.Spec.Definition.SecurityContext
			if !reflect.DeepEqual(expectedResult, actualResult) {
				t.Fatalf("Test %q failed to set the default env from experiment", name)
			}
		})
	}
}

func TestHostPID(t *testing.T) {
	fakeHostPID := false
	experiment := ExperimentDetails{
		Name:               "Fake-Exp-Name",
		Namespace:          "Fake NameSpace",
		JobName:            "fake-job-name",
		StatusCheckTimeout: 10,
		envMap:             map[string]v1.EnvVar{},
		ExpLabels:          map[string]string{},
		ExpArgs:            []string{},
	}

	tests := map[string]struct {
		chaosexperiment *v1alpha1.ChaosExperiment
	}{
		"Test Positive-1": {
			chaosexperiment: &v1alpha1.ChaosExperiment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      experiment.Name,
					Namespace: experiment.Namespace,
				},
				Spec: v1alpha1.ChaosExperimentSpec{
					Definition: v1alpha1.ExperimentDef{
						HostPID: fakeHostPID,
					},
				},
			},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateFakeClient(t)

			_, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(mock.chaosexperiment.Namespace).Create(mock.chaosexperiment)
			if err != nil {
				t.Fatalf("experiment not created for %v test, err: %v", name, err)
			}
			experimentSpec, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(mock.chaosexperiment.Namespace).Get(mock.chaosexperiment.Name, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("fail to get the chaosexperiment for %v test, err: %v", name, err)
			}

			expDetails := experiment.SetHostPID(experimentSpec)
			expectedResult := expDetails.HostPID
			actualResult := mock.chaosexperiment.Spec.Definition.HostPID
			if expectedResult != actualResult {
				t.Fatalf("Test %q failed to set the default env from experiment", name)
			}
		})
	}
}

func TestHandleChaosExperimentExistence(t *testing.T) {
	fakeHostPID := false
	experiment := ExperimentDetails{
		Name:               "Fake-Exp-Name",
		Namespace:          "Fake NameSpace",
		JobName:            "fake-job-name",
		StatusCheckTimeout: 10,
		envMap:             map[string]v1.EnvVar{},
		ExpLabels:          map[string]string{},
		ExpArgs:            []string{},
	}
	engineDetails := EngineDetails{
		Name:            "Fake Engine",
		EngineNamespace: "Fake NameSpace",
		UID:             "fake-chaosuid",
	}
	expStatus := ExperimentStatus{
		Name:   "Fake-Exp-Name",
		Status: v1alpha1.ExperimentStatusRunning,
	}

	tests := map[string]struct {
		chaosexperiment *v1alpha1.ChaosExperiment
		chaosengine     *v1alpha1.ChaosEngine
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
						HostPID: fakeHostPID,
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
				Spec: v1alpha1.ChaosEngineSpec{},
				Status: v1alpha1.ChaosEngineStatus{
					EngineStatus: v1alpha1.EngineStatusInitialized,
					Experiments: []v1alpha1.ExperimentStatuses{
						{
							Name: expStatus.Name,
						},
					},
				},
			},
			chaosexperiment: nil,
			isErr:           true,
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
			} else {
				_, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(mock.chaosengine.Namespace).Create(mock.chaosengine)
				if err != nil {
					t.Fatalf("engine not created for %v test, err: %v", name, err)
				}
			}

			err := experiment.HandleChaosExperimentExistence(engineDetails, client)
			if err != nil && !mock.isErr {
				t.Fatalf("%v Test Failed, err: %v", name, err)
			} else if err == nil && mock.isErr {
				t.Fatalf("%v Test Failed the expected error should not ne nil", name)
			}

			if mock.isErr {
				chaosEngine, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(engineDetails.EngineNamespace).Get(engineDetails.Name, metav1.GetOptions{})
				if err != nil {
					t.Fatalf("%v test failed engine not found, err: %v", name, err)
				}
				if chaosEngine.Status.Experiments[0].Status != "ChaosExperiment Not Found" {
					t.Fatalf("%v test failed experiment status in engine is not updated to not found when no experiment is there, err: %v", name, err)
				}
			}
		})
	}
}

func ptrint64(p int64) *int64 {
	return &p
}
