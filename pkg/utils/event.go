package utils

import (
	"time"

	"github.com/litmuschaos/chaos-runner/pkg/log"
	apiv1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientTypes "k8s.io/apimachinery/pkg/types"
)

//CreateEvents create the events in the desired resource
func (engineDetails EngineDetails) CreateEvents(eventAttributes *EventAttributes, clients ClientSets) error {

	events := &apiv1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      eventAttributes.Name,
			Namespace: engineDetails.EngineNamespace,
		},
		Source: apiv1.EventSource{
			Component: engineDetails.Name + "-runner",
		},
		Message:        eventAttributes.Message,
		Reason:         eventAttributes.Reason,
		Type:           eventAttributes.Type,
		Count:          1,
		FirstTimestamp: metav1.Time{Time: time.Now()},
		LastTimestamp:  metav1.Time{Time: time.Now()},
		InvolvedObject: apiv1.ObjectReference{
			APIVersion: "litmuschaos.io/v1alpha1",
			Kind:       "ChaosEngine",
			Name:       engineDetails.Name,
			Namespace:  engineDetails.EngineNamespace,
			UID:        clientTypes.UID(engineDetails.UID),
		},
	}

	_, err := clients.KubeClient.CoreV1().Events(engineDetails.EngineNamespace).Create(events)
	return err

}

//GenerateEvents update the events and increase the count by 1, if already present
// else it will create a new event
func (engineDetails EngineDetails) GenerateEvents(eventAttributes *EventAttributes, clients ClientSets) error {

	event, err := clients.KubeClient.CoreV1().Events(engineDetails.EngineNamespace).Get(eventAttributes.Name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			if err := engineDetails.CreateEvents(eventAttributes, clients); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		event.LastTimestamp = metav1.Time{Time: time.Now()}
		event.Count = event.Count + 1
		event.Message = eventAttributes.Message
		_, err = clients.KubeClient.CoreV1().Events(engineDetails.EngineNamespace).Update(event)
		return err
	}
	return nil
}

// ExperimentSkipped is an standard event spawned just after a ChaosExperiment is skipped
func (expDetails ExperimentDetails) ExperimentSkipped(reason string, engineDetails EngineDetails, clients ClientSets) {
	event := EventAttributes{}
	msg := "Experiment Job creation failed, skipping Chaos Experiment: " + expDetails.Name
	event.SetEventAttributes(reason, "Warning", msg)
	event.Name = event.Reason + expDetails.Name + string(engineDetails.UID)
	if err := engineDetails.GenerateEvents(&event, clients); err != nil {
		log.Errorf("unable to create event, err: %v", err)
	}
}

// ExperimentDependencyCheck is an standard event spawned just after validating
// experiment dependent resources such as ChaosExperiment, ConfigMaps and Secrets.
func (expDetails ExperimentDetails) ExperimentDependencyCheck(engineDetails EngineDetails, clients ClientSets) {
	event := EventAttributes{}
	msg := "Experiment resources validated for Chaos Experiment: " + expDetails.Name
	event.SetEventAttributes(ExperimentDependencyCheckReason, "Normal", msg)
	event.Name = event.Reason + expDetails.Name + string(engineDetails.UID)
	if err := engineDetails.GenerateEvents(&event, clients); err != nil {
		log.Errorf("unable to create event, err: %v", err)
	}
}

// ExperimentJobCreate is an standard event spawned just after starting ChaosExperiment Job
func (expDetails ExperimentDetails) ExperimentJobCreate(engineDetails EngineDetails, clients ClientSets) {
	event := EventAttributes{}
	msg := "Experiment Job " + expDetails.JobName + " for Chaos Experiment: " + expDetails.Name
	event.SetEventAttributes(ExperimentJobCreateReason, "Normal", msg)
	event.Name = event.Reason + expDetails.Name + string(engineDetails.UID)
	if err := engineDetails.GenerateEvents(&event, clients); err != nil {
		log.Errorf("unable to create event, err: %v", err)
	}
}

// ExperimentJobCleanUp is an standard event spawned just after deleting ChaosExperiment Job
func (expDetails ExperimentDetails) ExperimentJobCleanUp(jobCleanUpPolicy string, engineDetails EngineDetails, clients ClientSets) {
	event := EventAttributes{}
	msg := "Experiment Job " + expDetails.JobName + " will be retained"
	if jobCleanUpPolicy == "delete" {
		msg = "Experiment Job: " + expDetails.JobName + " will be deleted"
	}
	event.SetEventAttributes(ExperimentJobCleanUpReason, "Normal", msg)
	event.Name = event.Reason + expDetails.Name + string(engineDetails.UID)
	if err := engineDetails.GenerateEvents(&event, clients); err != nil {
		log.Errorf("unable to create event, err: %v", err)
	}
}

// SetEventAttributes set the event attributes for each
func (event *EventAttributes) SetEventAttributes(reason, eventType, msg string) {
	event.Message = msg
	event.Reason = reason
	event.Type = eventType
}
