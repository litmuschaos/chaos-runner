package analytics

import (
	ga "github.com/jpillora/go-ogle-analytics"
	"github.com/litmuschaos/chaos-executor/pkg/utils"
	log "github.com/sirupsen/logrus"
)

const (
	// clientID contains TrackingID of the application
	clientID string = "UA-92076314-21"

	// supported event categories

	// category category notifies installation of a Litmus Experiment
	category string = "Chaos-Experiment"

	// supported event actions

	// action is sent when the installation is triggered
	action string = "Installation"
)
// TriggerAnalytics is reponsible for sending out events
func TriggerAnalytics(experimentName string) {
	engineDetails := utils.EngineDetails{}
	utils.GetOsEnv(&engineDetails)
	client, err := ga.NewClient(clientID)
	if err != nil {
		log.Error(err, "GA Client ID Error")
	}
	uuid := engineDetails.ClientUUID
	client.ClientID(uuid)
	err = client.Send(ga.NewEvent(category, action).Label(experimentName))
	if err != nil {
		log.Info("Unable to send GA event")
	}
}
