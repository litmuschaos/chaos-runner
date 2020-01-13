package main

import (
	"flag"
	"time"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	"github.com/litmuschaos/chaos-executor/pkg/utils"
	"github.com/litmuschaos/chaos-executor/pkg/utils/analytics"
)

// getKubeConfig setup the config for access cluster resource
func getKubeConfig() (*rest.Config, error) {
	kubeconfig := flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()
	// Use in-cluster config if kubeconfig path is specified
	if *kubeconfig == "" {
		config, err := rest.InClusterConfig()
		if err != nil {
			return config, err
		}
	}
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return config, err
	}
	return config, err
}

func main() {

	engineDetails := utils.EngineDetails{}
	clients := utils.ClientSets{}

	// Getting the kubeconfig
	config, err := getKubeConfig()
	if err != nil {
		klog.V(2).Infof("Error in fetching kubeconfig, unable to proceed further due to error: %v", err)
		return
	}

	// clientSet creation for further use.
	if err = clients.GenerateClientSets(config); err != nil {
		klog.V(2).Infof("Unable to generate clientSet, due to error: %v", err)
	}

	// Fetching all the ENV's needed
	utils.GetOsEnv(&engineDetails)
	klog.V(3).Infof("Chaos Experiments List: %v, ChaosEngine Name: %v, Chaos AppNameSpace: %v, Chaos AppLabels: %v, Chaos AppKind: %v, ChaosService AccountName: %v", engineDetails.Experiments, engineDetails.Name, engineDetails.AppNamespace, engineDetails.AppLabel, engineDetails.AppKind, engineDetails.SvcAccount)
	klog.V(2).Infof("Chaos Experiment List: %v, ChaosEngine Name: %v, Chaos AppLabels: %v, Chaos AppKind: %v", engineDetails.Experiments, engineDetails.Name, engineDetails.AppLabel, engineDetails.AppKind)
	klog.V(1).Infof("Executor trigger by chaosEngine: %v, list of ChaosExperiment to be executed: %v", engineDetails.Name, engineDetails.Experiments)

	// Steps for each Experiment
	for i := range engineDetails.Experiments {

		// Sending event to GA instance
		if engineDetails.ClientUUID != "" {
			analytics.TriggerAnalytics(engineDetails.Experiments[i], engineDetails.ClientUUID)
		}

		experiment := utils.NewExperimentDetails()

		klog.V(3).Infof("Setting Values from ChaosEngine")
		experiment.SetValueFromChaosEngine(engineDetails, i)
		klog.V(3).Infof("Values set are Experiment Name: %v, Experiment Namespace: %v, Experiment ServiceAccount: %v", experiment.Name, experiment.Namespace, experiment.SvcAccount)

		klog.V(3).Infof("Setting Value from ChaosExperiment")
		experiment.SetValueFromChaosExperiment(clients)
		klog.V(3).Infof("Values set are Experiment Image: %v, Experiment Args: %v, Experiment Labels: %v, Experiment JobName: %v", experiment.ExpImage, experiment.ExpArgs, experiment.ExpLabels, experiment.JobName)

		klog.V(3).Infof("Setting ENV Values for ChaosExperiment")
		experiment.SetENV(engineDetails, clients)
		klog.V(3).Infof("ENV Values set in ChaosExperiment are: %v", experiment.Env)

		experimentStatus := utils.ExperimentStatus{}

		klog.V(3).Infof("Creating Intial Experiment Status")
		experimentStatus.IntialExperimentStatus(experiment)

		klog.V(2).Infof("ChaosEngine Initial Patching will be done for all the Experiment including the one not found also.")
		experimentStatus.InitialPatchEngine(engineDetails, clients)

		klog.V(1).Infof("Perparing to run Chaos Experiment: %v", experiment.Name)
		klog.V(3).Infof("Printing Experiment Stucture: %v", experiment)
		// isFound will return the status of experiment in that namespace
		// 1 -> found, 0 -> not-found
		isFound, err := experiment.CheckExistence(clients)
		klog.V(1).Infof("Chaos Experiment Validated in Application NamespaceL %v", experiment.Namespace)
		klog.V(2).Infof("Experiment Found Status : %v", isFound)

		// If not found in AppNamespace skip the further steps
		if !isFound {
			klog.V(2).Infof("Will try to patch ChaosEngine, with Not Found Status for Experiment: %v", experiment.Name)
			engineDetails.ExperimentNotFoundPatchEngine(experiment, clients)
			klog.V(1).Infof("Unable to list Chaos Experiment: %v, in Namespace: %v, skipping execution, with error: %v", experiment.Name, experiment.Namespace, err)
			break
		}

		// Patch ConfigMaps to ChaosExperiment Job
		if err := experiment.PatchConfigMaps(clients); err != nil {
			klog.V(1).Infof("Unable to patch ConfigMaps, due to: %v", err)
			break
		}

		// Patch Secrets to ChaosExperiment Job
		if err = experiment.PatchSecrets(clients); err != nil {
			klog.V(1).Infof("Unable to patch Secrets, due to: %v", err)
			break
		}

		experiment.VolumeOpts.VolumeOperations(experiment.ConfigMaps, experiment.Secrets)

		// Creation of PodTemplateSpec, and Final Job
		if err = utils.BuildingAndLaunchJob(experiment, clients); err != nil {
			klog.V(1).Infof("Unable to construct chaos experiment job due to: %v", err)
			break
		}

		time.Sleep(5 * time.Second)

		// Watching the Job till Completion
		if err = engineDetails.WatchJobForCompletion(experiment, clients); err != nil {
			klog.V(1).Infof("Unable to Watch the Job, error: %v", err)
			break
		}

		// Will Update the chaosEngine Status
		if err = engineDetails.UpdateEngineWithResult(experiment, clients); err != nil {
			klog.V(1).Infof("Unable to Update ChaosEngine Status due to: %v", err)
		}

		// Delete / retain the Job, using the jobCleanUpPolicy
		if err = engineDetails.DeleteJobAccordingToJobCleanUpPolicy(experiment, clients); err != nil {
			klog.V(1).Infof("Unable to Delete chaosExperiment Job due to: %v", err)
		}
	}
}
