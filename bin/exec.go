package main

import (
	"flag"
	"fmt"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"time"

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

	klog.InitFlags(nil)
	var Logger utils.LogStruct
	engineDetails := utils.EngineDetails{}
	clients := utils.ClientSets{}

	// Getting the kubeconfig
	config, err := getKubeConfig()
	if err != nil {
		Logger.WithString(fmt.Sprintf("Error in fetching kubeconfig, unable to proceed further due to error: %v", err)).WithVerbosity(1).Log()
		return
	}

	// clientSet creation for further use.
	if err = clients.GenerateClientSets(config); err != nil {
		Logger.WithString(fmt.Sprintf("Unable to generate clientSet, due to error: %v", err)).WithVerbosity(1).Log()
	}

	// Fetching all the ENV's needed
	utils.GetOsEnv(&engineDetails)
	Logger.WithString(fmt.Sprintf("Chaos Experiments List: %v, ChaosEngine Name: %v, Chaos AppNameSpace: %v, Chaos AppLabels: %v, Chaos AppKind: %v, ChaosService AccountName: %v", engineDetails.Experiments, engineDetails.Name, engineDetails.AppNamespace, engineDetails.AppLabel, engineDetails.AppKind, engineDetails.SvcAccount)).WithVerbosity(1).Log()
	Logger.WithString(fmt.Sprintf("Executor trigger by chaosEngine: %v, list of ChaosExperiment to be executed: %v", engineDetails.Name, engineDetails.Experiments)).WithVerbosity(1).Log()

	// Steps for each Experiment
	for i := range engineDetails.Experiments {

		// Sending event to GA instance
		if engineDetails.ClientUUID != "" {
			analytics.TriggerAnalytics(engineDetails.Experiments[i], engineDetails.ClientUUID)
		}

		experiment := utils.NewExperimentDetails()

		experiment.SetValueFromChaosEngine(engineDetails, i)
		Logger.WithString(fmt.Sprintf("Values set are Experiment Name: %v, Experiment Namespace: %v, Experiment ServiceAccount: %v", experiment.Name, experiment.Namespace, experiment.SvcAccount)).WithVerbosity(2).Log()

		experiment.SetValueFromChaosExperiment(clients)
		Logger.WithString(fmt.Sprintf("Values set are Experiment Image: %v, Experiment Args: %v, Experiment Labels: %v, Experiment JobName: %v", experiment.ExpImage, experiment.ExpArgs, experiment.ExpLabels, experiment.JobName)).WithVerbosity(2).Log()

		experiment.SetENV(engineDetails, clients)
		Logger.WithString(fmt.Sprintf("ENV Values set in ChaosExperiment are: %v", experiment.Env)).WithVerbosity(2).Log()

		experimentStatus := utils.ExperimentStatus{}

		experimentStatus.IntialExperimentStatus(experiment)

		experimentStatus.InitialPatchEngine(engineDetails, clients)

		Logger.WithString(fmt.Sprintf("Preparing to run Chaos Experiment: %v", experiment.Name)).WithVerbosity(0).Log()
		// isFound will return the status of experiment in that namespace
		// 1 -> found, 0 -> not-found
		isFound, err := experiment.CheckExistence(clients)

		// If not found in AppNamespace skip the further steps
		if !isFound {
			engineDetails.ExperimentNotFoundPatchEngine(experiment, clients)
			Logger.WithResourceName(experiment.Name).WithResourceType("Chaos Experiment").WithNameSpace(experiment.Namespace).WithOperation("Get").WithVerbosity(0).Log()
			break
		}

		Logger.WithString(fmt.Sprintf("Chaos Experiment Validated in Application Namespace %v", experiment.Namespace)).WithVerbosity(0).Log()

		// Patch ConfigMaps to ChaosExperiment Job
		if err := experiment.PatchConfigMaps(clients); err != nil {
			Logger.WithNameSpace(experiment.Namespace).WithResourceName(experiment.Name).WithString(err.Error()).WithOperation("Patch").WithResourceType("ConfigMaps").WithVerbosity(0).Log()
			break
		}

		// Patch Secrets to ChaosExperiment Job
		if err = experiment.PatchSecrets(clients); err != nil {
			Logger.WithNameSpace(experiment.Namespace).WithResourceName(experiment.Name).WithString(err.Error()).WithOperation("Patch").WithVerbosity(0).WithResourceType("Secrets").Log()
			break
		}

		experiment.VolumeOpts.VolumeOperations(experiment.ConfigMaps, experiment.Secrets)

		// Creation of PodTemplateSpec, and Final Job
		if err = utils.BuildingAndLaunchJob(experiment, clients); err != nil {
			Logger.WithNameSpace(experiment.Namespace).WithResourceName(experiment.JobName).WithString(err.Error()).WithOperation("Construct").WithVerbosity(0).WithResourceType("Job").Log()
			break
		}

		time.Sleep(5 * time.Second)

		// Watching the Job till Completion
		if err = engineDetails.WatchJobForCompletion(experiment, clients); err != nil {
			Logger.WithNameSpace(experiment.Namespace).WithResourceName(experiment.JobName).WithString(err.Error()).WithOperation("Watch").WithVerbosity(0).WithResourceType("Job").Log()
			break
		}

		// Will Update the chaosEngine Status
		if err = engineDetails.UpdateEngineWithResult(experiment, clients); err != nil {
			Logger.WithNameSpace(experiment.Namespace).WithResourceName(engineDetails.Name + "-" + experiment.Name).WithString(err.Error()).WithOperation("Patch").WithVerbosity(0).WithResourceType("ChaosEngine").Log()
		}

		// Delete / retain the Job, using the jobCleanUpPolicy
		if err = engineDetails.DeleteJobAccordingToJobCleanUpPolicy(experiment, clients); err != nil {
			Logger.WithNameSpace(experiment.Namespace).WithResourceName(experiment.JobName).WithString(err.Error()).WithOperation("Delete").WithVerbosity(0).WithResourceType("Job").Log()
		}
	}
}
