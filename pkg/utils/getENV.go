package utils

import (
	"os"
	"strings"

	"github.com/litmuschaos/chaos-runner/pkg/log"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

//SetENV sets ENV values in experimentDetails struct.
func (expDetails *ExperimentDetails) SetENV(engineDetails EngineDetails, clients ClientSets) error {
	// Get the Default ENV's from ChaosExperiment
	log.Info("Getting the ENV Variables")
	if err := expDetails.SetDefaultEnv(clients); err != nil {
		return err
	}

	// OverWriting the Defaults Varibles from the ChaosEngine ENV
	if err := expDetails.SetEnvFromEngine(engineDetails.Name, clients); err != nil {
		return err
	}
	// Store ENV in a map
	ENVList := map[string]string{
		"CHAOSENGINE":       engineDetails.Name,
		"APP_LABEL":         engineDetails.AppLabel,
		"CHAOS_NAMESPACE":   engineDetails.EngineNamespace,
		"APP_NAMESPACE":     engineDetails.AppNs,
		"APP_KIND":          engineDetails.AppKind,
		"AUXILIARY_APPINFO": engineDetails.AuxiliaryAppInfo,
		"CHAOS_UID":         engineDetails.UID,
		"EXPERIMENT_NAME":   expDetails.Name,
		"ANNOTATION_KEY":    engineDetails.AnnotationKey,
		"ANNOTATION_CHECK":  engineDetails.AnnotationCheck,
	}
	// Adding some addition ENV's from spec.AppInfo of ChaosEngine
	for key, value := range ENVList {
		expDetails.Env[key] = value
	}
	return nil
}

// SetDefaultEnv sets the Env's in Experiment Structure
func (expDetails *ExperimentDetails) SetDefaultEnv(clients ClientSets) error {
	experimentEnv, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		return errors.Errorf("Unable to get the %v ChaosExperiment in %v namespace, error: %v", expDetails.Name, expDetails.Namespace, err)
	}

	expDetails.Env = make(map[string]string)
	envList := experimentEnv.Spec.Definition.ENVList
	for i := range envList {
		key := envList[i].Name
		value := envList[i].Value
		expDetails.Env[key] = value
	}
	return nil
}
