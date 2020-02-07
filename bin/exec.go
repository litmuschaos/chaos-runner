package main

import (
	"flag"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

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
		log.Fatalf("Error in fetching the config, error : %v", err)
	}
	// clientSet creation for further use.
	if err = clients.GenerateClientSets(config); err != nil {
		log.Fatalf("Unable to create ClientSets")
	}

	// Fetching all the ENV's needed
	utils.GetOsEnv(&engineDetails)
	log.Infoln("Experiments List: ", engineDetails.Experiments, " ", "Engine Name: ", engineDetails.Name, " ", "appLabels : ", engineDetails.AppLabel, " ", "appNamespace: ", engineDetails.AppNamespace, " ", "appKind: ", engineDetails.AppKind, " ", "Service Account Name: ", engineDetails.SvcAccount)

	// Steps for each Experiment
	for i := range engineDetails.Experiments {

		// Sending event to GA instance
		if engineDetails.ClientUUID != "" {
			analytics.TriggerAnalytics(engineDetails.Experiments[i], engineDetails.ClientUUID)
		}

		experiment := utils.NewExperimentDetails()
		experiment.SetValueFromChaosEngine(engineDetails, i)
		experiment.SetValueFromChaosExperiment(clients)
		err := experiment.SetENV(&engineDetails, clients)
		if err != nil {
			log.Infof("Unable to set ENV %v", err)
			break
		}
		experimentStatus := utils.ExperimentStatus{}
		experimentStatus.IntialExperimentStatus(experiment)
		experimentStatus.InitialPatchEngine(engineDetails, clients)

		log.Infof("Preparing to run Chaos Experiment: %v", experiment.Name)

		// isFound will return the status of experiment in that namespace
		// 1 -> found, 0 -> not-found
		isFound, err := experiment.CheckExistence(clients)
		log.Infoln("Experiment Found Status : ", isFound)

		// If not found in AppNamespace skip the further steps
		if !isFound {
			engineDetails.ExperimentNotFoundPatchEngine(experiment, clients)
			log.Infof("Unable to list Chaos Experiment: %v, in Namespace: %v, skipping execution, with error: %v", experiment.Name, experiment.Namespace, err)
			break
		}

		// Patch ConfigMaps to ChaosExperiment Job
		if err := experiment.PatchConfigMaps(clients, engineDetails); err != nil {
			log.Infof("Unable to patch ConfigMaps, due to: %v", err)
			break
		}

		// Patch Secrets to ChaosExperiment Job
		if err = experiment.PatchSecrets(clients, engineDetails); err != nil {
			log.Infof("Unable to patch Secrets, due to: %v", err)
			break
		}

		experiment.VolumeOpts.VolumeOperations(experiment.ConfigMaps, experiment.Secrets)

		// Creation of PodTemplateSpec, and Final Job
		if err = utils.BuildingAndLaunchJob(experiment, clients); err != nil {
			log.Infof("Unable to construct chaos experiment job due to: %v", err)
			break
		}

		time.Sleep(5 * time.Second)

		// Watching the Job till Completion
		if err = engineDetails.WatchJobForCompletion(experiment, clients); err != nil {
			log.Infof("Unable to Watch the Job, error: %v", err)
			break
		}

		// Will Update the chaosEngine Status
		if err = engineDetails.UpdateEngineWithResult(experiment, clients); err != nil {
			log.Infof("Unable to Update ChaosEngine Status due to: %v", err)
		}

		// Delete / retain the Job, using the jobCleanUpPolicy
		if err = engineDetails.DeleteJobAccordingToJobCleanUpPolicy(experiment, clients); err != nil {
			log.Infof("Unable to Delete chaosExperiment Job due to: %v", err)
		}
	}
}
