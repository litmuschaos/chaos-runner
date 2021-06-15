package utils

import (
	"strings"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientTypes "k8s.io/apimachinery/pkg/types"
)

func TestCreateEvents(t *testing.T) {
	engineDetails := EngineDetails{
		Name:            "Fake Engine",
		EngineNamespace: "Fake NameSpace",
		UID:             "",
	}

	eventAtr := EventAttributes{
		Reason:  "fake-reason",
		Message: "fake-message",
		Type:    "fake-type",
		Name:    "fake-name",
	}
	client := CreateFakeClient(t)
	err := engineDetails.CreateEvents(&eventAtr, client)
	if err != nil {
		t.Fatalf("TestCreateEvents failed unable to get event, err: %v", err)
	}

	events, err := client.KubeClient.CoreV1().Events(engineDetails.EngineNamespace).List(metav1.ListOptions{})
	if err != nil || len(events.Items) == 0 {
		t.Fatalf("TestCreateEvents failed to get events, err: %v", err)
	}
}

func TestGenerateEvents(t *testing.T) {
	engineDetails := EngineDetails{
		Name:            "Fake Engine",
		EngineNamespace: "Fake NameSpace",
		UID:             "",
	}

	eventAtr := EventAttributes{
		Reason:  "fake-reason",
		Message: "fake-message",
		Type:    "fake-type",
		Name:    "fake-name",
	}

	tests := map[string]struct {
		events v1.Event
		isErr  bool
	}{
		"Test Positive-1": {
			isErr: false,
		},
		"Test Positive-2": {
			events: v1.Event{
				ObjectMeta: metav1.ObjectMeta{
					Name:      eventAtr.Name,
					Namespace: engineDetails.EngineNamespace,
				},
				Source: v1.EventSource{
					Component: engineDetails.Name + "-runner",
				},
				Message:        eventAtr.Message,
				Reason:         eventAtr.Reason,
				Type:           eventAtr.Type,
				Count:          1,
				FirstTimestamp: metav1.Time{Time: time.Now()},
				LastTimestamp:  metav1.Time{Time: time.Now()},
				InvolvedObject: v1.ObjectReference{
					APIVersion: "litmuschaos.io/v1alpha1",
					Kind:       "ChaosEngine",
					Name:       engineDetails.Name,
					Namespace:  engineDetails.EngineNamespace,
					UID:        clientTypes.UID(engineDetails.UID),
				},
			},
			isErr: true,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateFakeClient(t)
			if mock.isErr {
				_, err := client.KubeClient.CoreV1().Events(mock.events.Namespace).Create(&mock.events)
				if err != nil {
					t.Fatalf("fail to create event for %v test, err: %v", name, err)
				}
			}
			err := engineDetails.GenerateEvents(&eventAtr, client)
			if err != nil {
				t.Fatalf("%v fail to generate events, err: %v", name, err)
			}
			events, err := client.KubeClient.CoreV1().Events(engineDetails.EngineNamespace).List(metav1.ListOptions{})
			if err != nil || len(events.Items) == 0 {
				t.Fatalf("%v fail to get events, err: %v", name, err)
			}
		})
	}
}

func TestExperimentSkipped(t *testing.T) {
	engineDetails := EngineDetails{
		Name:            "Fake Engine",
		EngineNamespace: "Fake NameSpace",
		UID:             "",
	}

	eventAtr := EventAttributes{
		Reason:  "fake-reason",
		Message: "fake-message",
		Type:    "fake-type",
		Name:    "fake-name",
	}
	experiment := ExperimentDetails{
		Name:               "Fake-Exp-Name",
		Namespace:          "Fake NameSpace",
		JobName:            "fake-jobs-name-12345",
		StatusCheckTimeout: 2,
	}

	tests := map[string]struct {
		events v1.Event
		isErr  bool
	}{
		"Test Positive-1": {
			isErr: false,
		},
		"Test Positive-2": {
			events: v1.Event{
				ObjectMeta: metav1.ObjectMeta{
					Name:      eventAtr.Name,
					Namespace: engineDetails.EngineNamespace,
				},
				Source: v1.EventSource{
					Component: engineDetails.Name + "-runner",
				},
				Message:        eventAtr.Message,
				Reason:         eventAtr.Reason,
				Type:           eventAtr.Type,
				Count:          1,
				FirstTimestamp: metav1.Time{Time: time.Now()},
				LastTimestamp:  metav1.Time{Time: time.Now()},
				InvolvedObject: v1.ObjectReference{
					APIVersion: "litmuschaos.io/v1alpha1",
					Kind:       "ChaosEngine",
					Name:       engineDetails.Name,
					Namespace:  engineDetails.EngineNamespace,
					UID:        clientTypes.UID(engineDetails.UID),
				},
			},
			isErr: true,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateFakeClient(t)
			if mock.isErr {
				_, err := client.KubeClient.CoreV1().Events(mock.events.Namespace).Create(&mock.events)
				if err != nil {
					t.Fatalf("fail to create event for %v test, err: %v", name, err)
				}
			}
			experiment.ExperimentSkipped(eventAtr.Reason, engineDetails, client)

			events, err := client.KubeClient.CoreV1().Events(engineDetails.EngineNamespace).List(metav1.ListOptions{})
			if err != nil || len(events.Items) == 0 {
				t.Fatalf("%v fail to get events, err: %v", name, err)
			}

			if mock.isErr && !strings.Contains(events.Items[1].Message, "Experiment Job creation failed, skipping Chaos Experiment") {
				t.Fatalf("%v failed to get the skip event message", name)
			}
		})
	}
}

func TestExperimentDependencyCheck(t *testing.T) {
	engineDetails := EngineDetails{
		Name:            "Fake Engine",
		EngineNamespace: "Fake NameSpace",
		UID:             "",
	}

	eventAtr := EventAttributes{
		Reason:  "fake-reason",
		Message: "fake-message",
		Type:    "fake-type",
		Name:    "fake-name",
	}
	experiment := ExperimentDetails{
		Name:               "Fake-Exp-Name",
		Namespace:          "Fake NameSpace",
		JobName:            "fake-jobs-name-12345",
		StatusCheckTimeout: 2,
	}

	tests := map[string]struct {
		events v1.Event
		isErr  bool
	}{
		"Test Positive-1": {
			isErr: false,
		},
		"Test Positive-2": {
			events: v1.Event{
				ObjectMeta: metav1.ObjectMeta{
					Name:      eventAtr.Name,
					Namespace: engineDetails.EngineNamespace,
				},
				Source: v1.EventSource{
					Component: engineDetails.Name + "-runner",
				},
				Message:        eventAtr.Message,
				Reason:         eventAtr.Reason,
				Type:           eventAtr.Type,
				Count:          1,
				FirstTimestamp: metav1.Time{Time: time.Now()},
				LastTimestamp:  metav1.Time{Time: time.Now()},
				InvolvedObject: v1.ObjectReference{
					APIVersion: "litmuschaos.io/v1alpha1",
					Kind:       "ChaosEngine",
					Name:       engineDetails.Name,
					Namespace:  engineDetails.EngineNamespace,
					UID:        clientTypes.UID(engineDetails.UID),
				},
			},
			isErr: true,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateFakeClient(t)
			if mock.isErr {
				_, err := client.KubeClient.CoreV1().Events(mock.events.Namespace).Create(&mock.events)
				if err != nil {
					t.Fatalf("fail to create event for %v test, err: %v", name, err)
				}
			}
			experiment.ExperimentDependencyCheck(engineDetails, client)

			events, err := client.KubeClient.CoreV1().Events(engineDetails.EngineNamespace).List(metav1.ListOptions{})
			if err != nil || len(events.Items) == 0 {
				t.Fatalf("%v fail to get events, err: %v", name, err)
			}

			if mock.isErr && !strings.Contains(events.Items[1].Message, "Experiment resources validated for Chaos Experiment") {
				t.Fatalf("%v failed to get the validate event message", name)
			}
		})
	}
}

func TestExperimentJobCreate(t *testing.T) {
	engineDetails := EngineDetails{
		Name:            "Fake Engine",
		EngineNamespace: "Fake NameSpace",
		UID:             "",
	}

	eventAtr := EventAttributes{
		Reason:  "fake-reason",
		Message: "fake-message",
		Type:    "fake-type",
		Name:    "fake-name",
	}
	experiment := ExperimentDetails{
		Name:               "Fake-Exp-Name",
		Namespace:          "Fake NameSpace",
		JobName:            "fake-jobs-name-12345",
		StatusCheckTimeout: 2,
	}

	tests := map[string]struct {
		events v1.Event
		isErr  bool
	}{
		"Test Positive-1": {
			isErr: false,
		},
		"Test Positive-2": {
			events: v1.Event{
				ObjectMeta: metav1.ObjectMeta{
					Name:      eventAtr.Name,
					Namespace: engineDetails.EngineNamespace,
				},
				Source: v1.EventSource{
					Component: engineDetails.Name + "-runner",
				},
				Message:        eventAtr.Message,
				Reason:         eventAtr.Reason,
				Type:           eventAtr.Type,
				Count:          1,
				FirstTimestamp: metav1.Time{Time: time.Now()},
				LastTimestamp:  metav1.Time{Time: time.Now()},
				InvolvedObject: v1.ObjectReference{
					APIVersion: "litmuschaos.io/v1alpha1",
					Kind:       "ChaosEngine",
					Name:       engineDetails.Name,
					Namespace:  engineDetails.EngineNamespace,
					UID:        clientTypes.UID(engineDetails.UID),
				},
			},
			isErr: true,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateFakeClient(t)
			if mock.isErr {
				_, err := client.KubeClient.CoreV1().Events(mock.events.Namespace).Create(&mock.events)
				if err != nil {
					t.Fatalf("fail to create event for %v test, err: %v", name, err)
				}
			}
			experiment.ExperimentJobCreate(engineDetails, client)

			events, err := client.KubeClient.CoreV1().Events(engineDetails.EngineNamespace).List(metav1.ListOptions{})
			if err != nil || len(events.Items) == 0 {
				t.Fatalf("%v fail to get events, err: %v", name, err)
			}

			if mock.isErr && !strings.Contains(events.Items[1].Message, "Experiment Job "+experiment.JobName+" for Chaos Experiment") {
				t.Fatalf("%v failed to get the validate event message", name)
			}
		})
	}
}

func TestExperimentJobCleanUp(t *testing.T) {
	fakeJobCleanupPolicy := "delete"
	engineDetails := EngineDetails{
		Name:            "Fake Engine",
		EngineNamespace: "Fake NameSpace",
		UID:             "",
	}

	eventAtr := EventAttributes{
		Reason:  "fake-reason",
		Message: "fake-message",
		Type:    "fake-type",
		Name:    "fake-name",
	}
	experiment := ExperimentDetails{
		Name:               "Fake-Exp-Name",
		Namespace:          "Fake NameSpace",
		JobName:            "fake-jobs-name-12345",
		StatusCheckTimeout: 2,
	}

	tests := map[string]struct {
		events v1.Event
		isErr  bool
	}{
		"Test Positive-1": {
			isErr: false,
		},
		"Test Positive-2": {
			events: v1.Event{
				ObjectMeta: metav1.ObjectMeta{
					Name:      eventAtr.Name,
					Namespace: engineDetails.EngineNamespace,
				},
				Source: v1.EventSource{
					Component: engineDetails.Name + "-runner",
				},
				Message:        eventAtr.Message,
				Reason:         eventAtr.Reason,
				Type:           eventAtr.Type,
				Count:          1,
				FirstTimestamp: metav1.Time{Time: time.Now()},
				LastTimestamp:  metav1.Time{Time: time.Now()},
				InvolvedObject: v1.ObjectReference{
					APIVersion: "litmuschaos.io/v1alpha1",
					Kind:       "ChaosEngine",
					Name:       engineDetails.Name,
					Namespace:  engineDetails.EngineNamespace,
					UID:        clientTypes.UID(engineDetails.UID),
				},
			},
			isErr: true,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateFakeClient(t)
			if mock.isErr {
				_, err := client.KubeClient.CoreV1().Events(mock.events.Namespace).Create(&mock.events)
				if err != nil {
					t.Fatalf("fail to create event for %v test, err: %v", name, err)
				}
			}
			experiment.ExperimentJobCleanUp(fakeJobCleanupPolicy, engineDetails, client)

			events, err := client.KubeClient.CoreV1().Events(engineDetails.EngineNamespace).List(metav1.ListOptions{})
			if err != nil || len(events.Items) == 0 {
				t.Fatalf("%v fail to get events, err: %v", name, err)
			}

			if mock.isErr && !strings.Contains(events.Items[1].Message, "Experiment Job: "+experiment.JobName+" will be deleted") {
				t.Fatalf("%v failed to get the validate event message", name)
			}
		})
	}
}
