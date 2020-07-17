package main

import (
	"k8s.io/klog"

	"github.com/litmuschaos/chaos-runner/pkg/utils"
	"github.com/litmuschaos/chaos-runner/pkg/utils/analytics"
)

func main() {

	engineDetails := utils.EngineDetails{}
	clients := utils.ClientSets{}
	// Getting kubeConfig and Generate ClientSets
	if err := clients.GenerateClientSetFromKubeConfig(); err != nil {
		klog.Errorf("Unable to create ClientSets, error: %v", err)
		return
	}
	// Fetching all the ENV's needed
	utils.GetOsEnv(&engineDetails)
	klog.V(0).Infoln("Experiments List: ", engineDetails.Experiments, " ", "Engine Name: ", engineDetails.Name, " ", "appLabels : ", engineDetails.AppLabel, " ", "appKind: ", engineDetails.AppKind, " ", "Service Account Name: ", engineDetails.SvcAccount, "Engine Namespace: ", engineDetails.EngineNamespace)
	experimentList := utils.CreateExperimentList(&engineDetails)
	if err := utils.InitialPatchEngine(engineDetails, clients, experimentList); err != nil {
		klog.Errorf("Unable to create Initial ExpeirmentStatus in ChaosEngine, due to error: %v", err)
	}
	recorder, err := utils.NewEventRecorder(clients, engineDetails)
	if err != nil {
		klog.Errorf("Unable to initiate EventRecorder for Chaos-Runner, would not be able to add events")
	}
	// Steps for each Experiment
	for _, experiment := range experimentList {

		// Sending event to GA instance
		if engineDetails.ClientUUID != "" {
			analytics.TriggerAnalytics(experiment.Name, engineDetails.ClientUUID)
		}

		if err := experiment.SetValueFromChaosResources(&engineDetails, clients); err != nil {
			klog.V(0).Infof("Unable to set values from Chaos Resources due to error: %v", err)
			recorder.ExperimentSkipped(experiment.Name, utils.ExperimentNotFoundErrorReason)
			continue
		}

		if err := experiment.SetENV(engineDetails, clients); err != nil {
			klog.V(0).Infof("Unable to patch ENV due to error: %v", err)
			recorder.ExperimentSkipped(experiment.Name, utils.ExperimentEnvParseErrorReason)
			continue
		}

		klog.V(0).Infof("Preparing to run Chaos Experiment: %v", experiment.Name)

		if err := experiment.PatchResources(engineDetails, clients); err != nil {
			klog.V(0).Infof("Unable to patch Chaos Resources required for Chaos Experiment: %v, due to error: %v", experiment.Name, err)

		}
		recorder.ExperimentDepedencyCheck(experiment.Name)

		// Creation of PodTemplateSpec, and Final Job
		if err := utils.BuildingAndLaunchJob(&experiment, clients); err != nil {
			klog.V(0).Infof("Unable to construct chaos experiment job due to: %v", err)
			recorder.ExperimentSkipped(experiment.Name, utils.ExperimentJobCreationErrorReason)
			continue
		}
		recorder.ExperimentJobCreate(experiment.Name, experiment.JobName)

		klog.V(0).Infof("Started Chaos Experiment Name: %v, with Job Name: %v", experiment.Name, experiment.JobName)
		// Watching the chaos container till Completion
		if err := engineDetails.WatchChaosContainerForCompletion(&experiment, clients); err != nil {
			klog.V(0).Infof("Unable to Watch the chaos container, error: %v", err)
			recorder.ExperimentSkipped(experiment.Name, utils.ExperimentChaosContainerWatchErrorReason)
			continue
		}

		// Will Update the chaosEngine Status
		if err := engineDetails.UpdateEngineWithResult(&experiment, clients); err != nil {
			klog.V(0).Infof("Unable to Update ChaosEngine Status due to: %v", err)
		}

		// Delete / retain the Job, using the jobCleanUpPolicy
		jobCleanUpPolicy, err := engineDetails.DeleteJobAccordingToJobCleanUpPolicy(&experiment, clients)
		if err != nil {
			klog.V(0).Infof("Unable to Delete ChaosExperiment Job due to: %v", err)
		}
		recorder.ExperimentJobCleanUp(&experiment, jobCleanUpPolicy)
	}
}
