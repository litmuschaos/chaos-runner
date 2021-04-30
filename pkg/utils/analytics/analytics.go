package analytics

import (
	ga "github.com/jpillora/go-ogle-analytics"
	"github.com/litmuschaos/chaos-runner/pkg/log"
)

const (
	// clientID contains TrackingID of the application
	clientID string = "UA-92076314-21"

	// supported event categories

	// category category notifies installation of a Litmus Experiment
	category string = "Chaos-Experiment"

	// supported event actions

	// action is sent when the installation is triggered
	action string = "Execution"
)

// TriggerAnalytics is responsible for sending out events
func TriggerAnalytics(experimentName string, uuid string) {
	client, err := ga.NewClient(clientID)
	if err != nil {
		log.Errorf("unable to create GA client, error: %v", err)
	}
	client.ClientID(uuid)
	err = client.Send(ga.NewEvent(category, action).Label(experimentName))
	if err != nil {
		log.Errorf("unable to send GA event, error: %v", err)
	}
}
