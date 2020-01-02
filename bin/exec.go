package main

import (
	"flag"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/litmuschaos/chaos-executor/pkg/utils"
	"github.com/litmuschaos/chaos-executor/pkg/utils/analytics"
	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
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

// checkStatusListForExp will loook for the Experiment in AppNamespace
func checkStatusListForExp(status []v1alpha1.ExperimentStatuses, jobName string) int {
	for i := range status {
		if status[i].Name == jobName {
			return i
		}
	}
	return -1
}

func main() {

	engineDetails := utils.EngineDetails{}
	clients := utils.ClientSets{}
	// Getting the config
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

		experiment := utils.NewExperimentDetails()
		experiment.Name = engineDetails.Experiments[i]
		experiment.Namespace = engineDetails.AppNamespace
		experiment.SvcAccount = engineDetails.SvcAccount
		log.Infof("Printing experiment.Name: %v, experiment.Namespace : %v", experiment.Name, experiment.Namespace)

		log.Infoln("Going with the experiment Name: " + engineDetails.Experiments[i])

		// isFound will return the status of experiment in that namespace
		// 1 -> found, 0 -> not-found
		isFound := experiment.CheckExistence(clients)
		log.Infoln("Experiment Found Status : ", isFound)

		// If not found in AppNamespace skip the further steps
		if !isFound {
			log.Infoln("Can't Find Experiment Name : "+engineDetails.Experiments[i], "In Namespace : "+engineDetails.AppNamespace)
			log.Infoln("Not Executing the Experiment : " + engineDetails.Experiments[i])
			break
		}

		log.Infoln("Getting the Default ENV Variables")

		// Get the Default ENV's from ChaosExperiment
		experiment.SetDefaultEnv(clients)

		log.Info("Printing the Default Variables", experiment.Env)

		// Get the ConfigMaps for patching them in the job creation
		log.Infoln("Find the configMaps in the chaosExperiments")

		experiment.SetConfigMaps(clients)

		err := experiment.ValidateConfigMaps(clients)
		if err != nil {
			log.Infof("Aborting Execution")
			return
		}

		experiment.SetSecrets(clients)
		err = experiment.ValidateSecrets(clients)
		if err != nil {
			log.Infof("Aborting Execution")
			return
		}

		log.Infof("Validated ConfigMaps: %v", experiment.ConfigMaps)
		log.Infof("Validated Secrets: %v", experiment.Secrets)

		// Adding VolumeBuilders, according to the ConfigMaps, and Secrets from ChaosEXperiment
		experiment.VolumeOpts.VolumeBuilders = utils.CreateVolumeBuilder(experiment.ConfigMaps, experiment.Secrets)

		// Adding VoulmeMounts, according to the configMaps, and Secrets from ChaosEXperiment
		experiment.VolumeOpts.VolumeMounts = utils.CreateVolumeMounts(experiment.ConfigMaps, experiment.Secrets)

		// OverWriting the Defaults Varibles from the ChaosEngine ENV
		log.Infoln("Patching some required ENV's")
		experiment.SetEnvFromEngine(engineDetails.Name, clients)

		// Adding some addition necessary ENV's
		experiment.Env["CHAOSENGINE"] = engineDetails.Name
		experiment.Env["APP_LABEL"] = engineDetails.AppLabel
		experiment.Env["APP_NAMESPACE"] = engineDetails.AppNamespace
		experiment.Env["APP_KIND"] = engineDetails.AppKind

		log.Info("Printing the Over-ridden Variables")
		log.Infoln(experiment.Env)

		log.Infoln("Getting all the details of the experiment Name : " + engineDetails.Experiments[i])

		// Fetching more details from the ChaosExperiment needed for execution
		experiment.SetImage(clients)
		experiment.SetArgs(clients)
		experiment.SetLabels(clients)
		log.Infof("Variables for ChaosJob: Experiment Labels: %v, Experiment Image: %v, experiment.Args: %v", experiment.ExpLabels, experiment.ExpImage, experiment.ExpArgs)

		// Generation of Random String for appending it into Job
		randomString := utils.RandomString()

		// Setting the JobName in Experiment Realted struct
		experiment.JobName = engineDetails.Experiments[i] + "-" + randomString

		// Sending event to GA instance
		if engineDetails.ClientUUID != "" {
			analytics.TriggerAnalytics(experiment.JobName, engineDetails.ClientUUID)
		}
		
		log.Infof("JobName for this Experiment : %v", experiment.JobName)

		// Creation of PodTemplateSpec, and Final Job
		err = utils.DeployJob(experiment, clients)
		if err != nil {
			log.Infof("Error while building Job : %v", err)
		}

		time.Sleep(5 * time.Second)
		// Getting the Experiment Result Name
		resultName := utils.GetResultName(engineDetails, i)

		// Watching the Job till Completion
		err = utils.WatchingJobtillCompletion(experiment, engineDetails, clients)
		if err != nil {
			log.Infof("Unable to Watch the Job, error: %v", err)
		}

		// Will Update the result,
		// Delete / retain the Job, using the jobCleanUpPolicy
		err = utils.UpdateResultWithJobAndDeletingJob(engineDetails, resultName, experiment, clients)
		if err != nil {
			log.Infof("Unable to Update ChaosResult, error: %v", err)
		}
	}
}
