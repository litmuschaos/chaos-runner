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
	engineDetails.EngineNamespace = os.Getenv("CHAOS_NAMESPACE")
	engineDetails.AppKind = os.Getenv("APP_KIND")
	engineDetails.SvcAccount = os.Getenv("CHAOS_SVC_ACC")
	engineDetails.ClientUUID = os.Getenv("CLIENT_UUID")
	engineDetails.Experiments = strings.Split(experimentList, ",")
	engineDetails.AuxiliaryAppInfo = os.Getenv("AUXILIARY_APPINFO")

	//TODO: Use engineDetails.AdminMode, to change behaviour of chaos-runner
	// engineDetails.AdminMode = false
	// if os.Getenv("ADMIN_MODE") == "true" {
	// 	engineDetails.AdminMode = true
	// }
}
