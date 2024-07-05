package utils

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/litmuschaos/chaos-runner/pkg/log"
	"github.com/litmuschaos/chaos-runner/pkg/telemetry"
	v1 "k8s.io/api/core/v1"
)

// SetEngineDetails adds the ENV's to EngineDetails
func (engineDetails *EngineDetails) SetEngineDetails() *EngineDetails {
	engineDetails.Experiments = strings.Split(os.Getenv("EXPERIMENT_LIST"), ",")
	engineDetails.Name = os.Getenv("CHAOSENGINE")
	engineDetails.EngineNamespace = os.Getenv("CHAOS_NAMESPACE")
	engineDetails.SvcAccount = os.Getenv("CHAOS_SVC_ACC")
	engineDetails.ClientUUID = os.Getenv("CLIENT_UUID")
	engineDetails.AuxiliaryAppInfo = os.Getenv("AUXILIARY_APPINFO")
	engineDetails.Targets = os.Getenv("TARGETS")
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

// SetENV sets ENV values in experimentDetails struct.
func (expDetails *ExperimentDetails) SetENV(ctx context.Context, engineDetails EngineDetails, clients ClientSets) error {

	// Setting envs from engine fields other than env
	expDetails.setEnv("CHAOSENGINE", engineDetails.Name).
		setEnv("TARGETS", engineDetails.Targets).
		setEnv("CHAOS_NAMESPACE", engineDetails.EngineNamespace).
		setEnv("AUXILIARY_APPINFO", engineDetails.AuxiliaryAppInfo).
		setEnv("CHAOS_UID", engineDetails.UID).
		setEnv("EXPERIMENT_NAME", expDetails.Name).
		setEnv("LIB_IMAGE_PULL_POLICY", string(expDetails.ExpImagePullPolicy)).
		setEnv("TERMINATION_GRACE_PERIOD_SECONDS", strconv.Itoa(int(expDetails.TerminationGracePeriodSeconds))).
		setEnv("DEFAULT_HEALTH_CHECK", expDetails.DefaultHealthCheck).
		setEnv("CHAOS_SERVICE_ACCOUNT", expDetails.SvcAccount).
		setEnv("OTEL_EXPORTER_OTLP_ENDPOINT", os.Getenv(telemetry.OTELExporterOTLPEndpoint)).
		setEnv("TRACE_PARENT", telemetry.GetMarshalledSpanFromContext(ctx))

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
