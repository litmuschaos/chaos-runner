package utils

import (
	"os"
	"strings"

	v1 "k8s.io/api/core/v1"

	"github.com/litmuschaos/chaos-runner/pkg/log"
)

// GetOsEnv adds the ENV's to EngineDetails
func (engineDetails *EngineDetails) GetOsEnv() *EngineDetails {
	engineDetails.Experiments = strings.Split(os.Getenv("EXPERIMENT_LIST"), ",")
	engineDetails.Name = os.Getenv("CHAOSENGINE")
	engineDetails.AppLabel = os.Getenv("APP_LABEL")
	engineDetails.AppNs = os.Getenv("APP_NAMESPACE")
	engineDetails.EngineNamespace = os.Getenv("CHAOS_NAMESPACE")
	engineDetails.AppKind = os.Getenv("APP_KIND")
	engineDetails.SvcAccount = os.Getenv("CHAOS_SVC_ACC")
	engineDetails.ClientUUID = os.Getenv("CLIENT_UUID")
	engineDetails.AuxiliaryAppInfo = os.Getenv("AUXILIARY_APPINFO")
	engineDetails.AnnotationKey = os.Getenv("ANNOTATION_KEY")
	engineDetails.AnnotationCheck = os.Getenv("ANNOTATION_CHECK")
	return engineDetails
}

// GetEngineUID get the chaosengine UID
func (engineDetails *EngineDetails) GetEngineUID(clients ClientSets) *EngineDetails {
	chaosEngine, err := engineDetails.GetChaosEngine(clients)
	if err != nil {
		log.Errorf("Unable to get chaosEngine in namespace: %s", engineDetails.EngineNamespace)
	}
	engineDetails.UID = string(chaosEngine.UID)
	return engineDetails
}

//SetENV sets ENV values in experimentDetails struct.
func (expDetails *ExperimentDetails) SetENV(engineDetails EngineDetails, clients ClientSets) error {
	// Get the Default ENV's from ChaosExperiment
	log.Info("Getting the ENV Variables")
	if err := expDetails.SetDefaultEnvFromChaosExperiment(clients); err != nil {
		return err
	}

	// OverWriting the Defaults Varibles from the ChaosEngine ENV
	if err := expDetails.SetOverrideEnvFromChaosEngine(engineDetails.Name, clients); err != nil {
		return err
	}
	// Store ENV in a map
	ENVList := map[string]string{
		"CHAOSENGINE":           engineDetails.Name,
		"APP_LABEL":             engineDetails.AppLabel,
		"CHAOS_NAMESPACE":       engineDetails.EngineNamespace,
		"APP_NAMESPACE":         engineDetails.AppNs,
		"APP_KIND":              engineDetails.AppKind,
		"AUXILIARY_APPINFO":     engineDetails.AuxiliaryAppInfo,
		"CHAOS_UID":             engineDetails.UID,
		"EXPERIMENT_NAME":       expDetails.Name,
		"ANNOTATION_KEY":        engineDetails.AnnotationKey,
		"ANNOTATION_CHECK":      engineDetails.AnnotationCheck,
		"LIB_IMAGE_PULL_POLICY": string(expDetails.ExpImagePullPolicy),
	}
	// Adding some additional ENV's from spec.AppInfo of ChaosEngine// Adding some additional ENV's from spec.AppInfo of ChaosEngine
	for key, value := range ENVList {
		expDetails.envMap[key] = v1.EnvVar{
			Name:  key,
			Value: value,
		}
	}
	return nil
}
