package main

import (
	"github.com/litmuschaos/chaos-runner/pkg/log"
	"github.com/litmuschaos/chaos-runner/pkg/utils"
	"github.com/litmuschaos/chaos-runner/pkg/utils/analytics"
	"github.com/sirupsen/logrus"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:          true,
		DisableSorting:         true,
		DisableLevelTruncation: true,
	})
}

func main() {

	engineDetails := utils.EngineDetails{}
	clients := utils.ClientSets{}
	// Getting kubeConfig and Generate ClientSets
	if err := clients.GenerateClientSetFromKubeConfig(); err != nil {
		log.Errorf("unable to create ClientSets, error: %v", err)
		return
	}
	// Fetching all the ENVs passed from the chaos-operator
	// create and initialize the experimentList
	if err := engineDetails.SetEngineDetails().SetEngineUID(clients); err != nil {
		log.Errorf("unable to get ChaosEngineUID, error: %v", err)
		return
	}
	experimentList := engineDetails.CreateExperimentList()
	log.InfoWithValues("Experiments details are as follows", logrus.Fields{
		"Experiments List":     engineDetails.Experiments,
		"Engine Name":          engineDetails.Name,
		"appLabels":            engineDetails.AppLabel,
		"appNs":                engineDetails.AppNs,
		"appKind":              engineDetails.AppKind,
		"Service Account Name": engineDetails.SvcAccount,
		"Engine Namespace":     engineDetails.EngineNamespace,
	})

	if err := utils.InitialPatchEngine(engineDetails, clients, experimentList); err != nil {
		log.Errorf("unable to patch Initial ExperimentStatus in ChaosEngine, error: %v", err)
		return
	}

	// Steps for each Experiment
	for _, experiment := range experimentList {

		// Sending event to GA instance
		if engineDetails.ClientUUID != "" {
			analytics.TriggerAnalytics(experiment.Name, engineDetails.ClientUUID)
		}
		// check the existence of chaosexperiment inside the cluster
		if err := experiment.HandleChaosExperimentExistence(engineDetails, clients); err != nil {
			log.Errorf("unable to get ChaosExperiment name: %v, in namespace: %v, error: %v", experiment.Name, experiment.Namespace, err)
			experiment.ExperimentSkipped(utils.ExperimentNotFoundErrorReason, engineDetails, clients)
			continue
		}
		// derive the required field from the experiment & engine and set into experimentDetails struct
		if err := experiment.SetValueFromChaosResources(&engineDetails, clients); err != nil {
			log.Errorf("unable to set values from Chaos Resources, error: %v", err)
			experiment.ExperimentSkipped(utils.ExperimentNotFoundErrorReason, engineDetails, clients)
			engineDetails.ExperimentSkippedPatchEngine(&experiment, clients)
			continue
		}
		// derive the envs from the chaos experiment and override their values from chaosengine if any
		if err := experiment.SetENV(engineDetails, clients); err != nil {
			log.Errorf("unable to patch ENV, error: %v", err)
			experiment.ExperimentSkipped(utils.ExperimentEnvParseErrorReason, engineDetails, clients)
			engineDetails.ExperimentSkippedPatchEngine(&experiment, clients)
			continue
		}

		log.Infof("Preparing to run Chaos Experiment: %v", experiment.Name)

		if err := experiment.PatchResources(engineDetails, clients); err != nil {
			log.Errorf("unable to patch Chaos Resources required for Chaos Experiment: %v, error: %v", experiment.Name, err)
			experiment.ExperimentSkipped(utils.ExperimentDependencyCheckReason, engineDetails, clients)
			engineDetails.ExperimentSkippedPatchEngine(&experiment, clients)
			continue
		}
		// generating experiment dependency check event inside chaosengine
		experiment.ExperimentDependencyCheck(engineDetails, clients)

		// Creation of PodTemplateSpec, and Final Job
		if err := utils.BuildingAndLaunchJob(&experiment, clients); err != nil {
			log.Errorf("unable to construct chaos experiment job, error: %v", err)
			experiment.ExperimentSkipped(utils.ExperimentDependencyCheckReason, engineDetails, clients)
			engineDetails.ExperimentSkippedPatchEngine(&experiment, clients)
			continue
		}

		experiment.ExperimentJobCreate(engineDetails, clients)

		log.Infof("Started Chaos Experiment Name: %v, with Job Name: %v", experiment.Name, experiment.JobName)
		// Watching the chaos container till Completion
		if err := engineDetails.WatchChaosContainerForCompletion(&experiment, clients); err != nil {
			log.Errorf("unable to Watch the chaos container, error: %v", err)
			experiment.ExperimentSkipped(utils.ExperimentChaosContainerWatchErrorReason, engineDetails, clients)
			engineDetails.ExperimentSkippedPatchEngine(&experiment, clients)
			continue
		}

		log.Infof("Chaos Pod Completed, Experiment Name: %v, with Job Name: %v", experiment.Name, experiment.JobName)

		// Will Update the chaosEngine Status
		if err := engineDetails.UpdateEngineWithResult(&experiment, clients); err != nil {
			log.Errorf("unable to Update ChaosEngine Status, error: %v", err)
		}

		log.Infof("Chaos Engine has been updated with result, Experiment Name: %v", experiment.Name)

		// Delete/Retain the Job, based on the jobCleanUpPolicy
		jobCleanUpPolicy, err := engineDetails.DeleteJobAccordingToJobCleanUpPolicy(&experiment, clients)
		if err != nil {
			log.Errorf("unable to Delete ChaosExperiment Job, error: %v", err)
		}
		experiment.ExperimentJobCleanUp(string(jobCleanUpPolicy), engineDetails, clients)
	}
}
