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
	err = clients.GenerateClientSets(config)
	if err != nil {
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
		experiment.SetENV(engineDetails, clients)
		log.Infof("Printing Experiment Structure: %v", experiment)

		experimentStatus := utils.ExperimentStatus{}
		experimentStatus.IntialExperimentStatus(experiment)
		experimentStatus.InitialPatchEngine(engineDetails, clients)

		log.Infof("Printing experiment.Name: %v, experiment.Namespace : %v", experiment.Name, experiment.Namespace)
		log.Infoln("Going with the experiment Name: " + engineDetails.Experiments[i])

		// isFound will return the status of experiment in that namespace
		// 1 -> found, 0 -> not-found
		isFound := experiment.CheckExistence(clients)
		log.Infoln("Experiment Found Status : ", isFound)

		// If not found in AppNamespace skip the further steps
		if !isFound {
			//TODO Patch Engine if chaosExperiment is not found.
			engineDetails.ExperimentNotFoundPatchEngine(experiment, clients)

			log.Infoln("Can't Find Experiment Name : "+engineDetails.Experiments[i], "In Namespace : "+engineDetails.AppNamespace)
			log.Infoln("Not Executing the Experiment : " + engineDetails.Experiments[i])
			break
		}
		log.Info("Printing the ENV Variables", experiment.Env)

		// Patch ConfigMaps to ChaosExperiment Job
		err := experiment.PatchConfigMaps(clients)
		if err != nil {
			break
		}
		// Patch Secrets to ChaosExperiment Job
		err = experiment.PatchSecrets(clients)
		if err != nil {
			break
		}
		// Adding VolumeBuilders, according to the ConfigMaps, and Secrets from ChaosEXperiment
		experiment.VolumeOpts.VolumeBuilders = utils.CreateVolumeBuilder(experiment.ConfigMaps, experiment.Secrets)

		// Adding VoulmeMounts, according to the configMaps, and Secrets from ChaosEXperiment
		experiment.VolumeOpts.VolumeMounts = utils.CreateVolumeMounts(experiment.ConfigMaps, experiment.Secrets)

		// Creation of PodTemplateSpec, and Final Job
		err = utils.DeployJob(experiment, clients)
		if err != nil {
			log.Infof("Error while building Job : %v", err)
			break
		}

		time.Sleep(5 * time.Second)

		err = engineDetails.WatchingJobtillCompletion(experiment, clients)
		// Watching the Job till Completion
		if err != nil {
			log.Infof("Unable to Watch the Job, error: %v", err)
		}

		// Will Update the result,
		// Delete / retain the Job, using the jobCleanUpPolicy
		err = engineDetails.UpdateResultWithJobAndDeletingJob(experiment, clients)
		if err != nil {
			log.Infof("Unable to Update ChaosResult, error: %v", err)
		}
	}
}
