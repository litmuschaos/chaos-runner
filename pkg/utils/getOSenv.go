package utils

import (
	"os"
	"strings"
)

// GetOsEnv adds the ENV's to EngineDetails
func GetOsEnv(engineDetails *EngineDetails) {
	experimentList := os.Getenv("EXPERIMENT_LIST")
	engineDetails.Name = os.Getenv("CHAOSENGINE")
	engineDetails.AppLabel = os.Getenv("APP_LABEL")
	engineDetails.AppNamespace = os.Getenv("APP_NAMESPACE")
	engineDetails.AppKind = os.Getenv("APP_KIND")
	engineDetails.SvcAccount = os.Getenv("CHAOS_SVC_ACC")
	engineDetails.ClientUUID = os.Getenv("CLIENT_UUID")
	engineDetails.Experiments = strings.Split(experimentList, ",")
	//rand := os.Getenv("RANDOM")
	//max := os.Getenv("MAX_DURATION")
}
