package analytics

import (
	ga "github.com/jpillora/go-ogle-analytics"
	"k8s.io/klog"
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

// TriggerAnalytics is reponsible for sending out events
func TriggerAnalytics(experimentName string, uuid string) {
	client, err := ga.NewClient(clientID)
	if err != nil {
		klog.Error(err, "GA Client ID Error")
	}
	client.ClientID(uuid)
	err = client.Send(ga.NewEvent(category, action).Label(experimentName))
	if err != nil {
		klog.Infoln("Unable to send GA event", err)
	}
}
