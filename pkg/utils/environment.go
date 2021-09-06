package utils

import (
	"os"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"

	"github.com/litmuschaos/chaos-runner/pkg/log"
)

// SetEngineDetails adds the ENV's to EngineDetails
func (engineDetails *EngineDetails) SetEngineDetails() *EngineDetails {
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

// SetEngineUID set the chaosengine UID
func (engineDetails *EngineDetails) SetEngineUID(clients ClientSets) error {
	chaosEngine, err := engineDetails.GetChaosEngine(clients)
	if err != nil {
		return err
	}
	engineDetails.UID = string(chaosEngine.UID)
	return nil
}

//SetENV sets ENV values in experimentDetails struct.
func (expDetails *ExperimentDetails) SetENV(engineDetails EngineDetails, clients ClientSets) error {
	// Setting envs from engine fields other than env
	expDetails.setEnv("CHAOSENGINE", engineDetails.Name).
		setEnv("APP_LABEL", engineDetails.AppLabel).
		setEnv("CHAOS_NAMESPACE", engineDetails.EngineNamespace).
		setEnv("APP_NAMESPACE", engineDetails.AppNs).
		setEnv("APP_KIND", engineDetails.AppKind).
		setEnv("AUXILIARY_APPINFO", engineDetails.AuxiliaryAppInfo).
		setEnv("CHAOS_UID", engineDetails.UID).
		setEnv("EXPERIMENT_NAME", expDetails.Name).
		setEnv("ANNOTATION_KEY", engineDetails.AnnotationKey).
		setEnv("ANNOTATION_CHECK", engineDetails.AnnotationCheck).
		setEnv("LIB_IMAGE_PULL_POLICY", string(expDetails.ExpImagePullPolicy)).
		setEnv("TERMINATION_GRACE_PERIOD_SECONDS", strconv.Itoa(int(expDetails.TerminationGracePeriodSeconds))).
		setEnv("DEFAULT_APP_HEALTH_CHECK", expDetails.DefaultAppHealthCheck)

	// Get the Default ENV's from ChaosExperiment
	log.Info("Getting the ENV Variables")
	if err := expDetails.SetDefaultEnvFromChaosExperiment(clients); err != nil {
		return err
	}

	// OverWriting the Defaults Varibles from the ChaosEngine ENV
	if err := expDetails.SetOverrideEnvFromChaosEngine(engineDetails.Name, clients); err != nil {
		return err
	}
	return nil
}

// setEnv set the env inside experimentDetails struct
func (expDetails *ExperimentDetails) setEnv(key, value string) *ExperimentDetails {

	if value == "" {
		return expDetails
	}
	expDetails.envMap[key] = v1.EnvVar{
		Name:  key,
		Value: value,
	}
	return expDetails
}
