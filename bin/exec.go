package main

import (
	"flag"
	"github.com/litmuschaos/chaos-executor/pkg/utils"
	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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

var kubeconfig string
var err error
var config *rest.Config

func main() {

	engineDetails := utils.EngineDetails{}

	// Getting the config
	config, err := getKubeConfig()
	if err != nil {
		log.Info("Error in fetching the config")
		panic(err.Error())
	}

	engineDetails.Config = config

	// Fetching all the ENV's needed
	utils.GetOsEnv(&engineDetails)
	log.Infoln("Experiments List: ", engineDetails.Experiments, " ", "Engine Name: ", engineDetails.Name, " ", "appLabels : ", engineDetails.AppLabel, " ", "appNamespace: ", engineDetails.AppNamespace, " ", "appKind: ", engineDetails.AppKind, " ", "Service Account Name: ", engineDetails.SvcAccount)

	// Steps for each Experiment
	for i := range engineDetails.Experiments {

		log.Infoln("Going with the experiment Name : " + engineDetails.Experiments[i])

		// isFound will return the status of experiment in that namespace
		// 1 -> found, 0 -> not-found
		isFound := !utils.CheckExperimentInAppNamespace("default", engineDetails.Experiments[i], config)
		log.Infoln("Experiment Found Status : ", isFound)

		// If not found in AppNamespace skip the further steps
		if !isFound {
			log.Infoln("Can't Find Experiment Name : "+engineDetails.Experiments[i], "In Namespace : "+engineDetails.AppNamespace)
			log.Infoln("Not Executing the Experiment : " + engineDetails.Experiments[i])
			break
		}

		var perExperiment utils.ExperimentDetails

		log.Infoln("Getting the Default ENV Variables")

		// Get thee Deafult ENV's from ChaosExperiment
		perExperiment.Env = utils.GetEnvFromExperiment(engineDetails.AppNamespace, engineDetails.Experiments[i], engineDetails.Config)

		log.Info("Printing the Default Variables", perExperiment.Env)
		log.Infoln("OverWriting the Default Variables")

		// OverWriting the Deafults Varibles from the ChaosEngine one's
		utils.OverWriteEnvFromEngine(engineDetails.AppNamespace, engineDetails.Name, engineDetails.Config, perExperiment.Env, engineDetails.Experiments[i])

		log.Infoln("Patching some required ENV's")

		// Adding some addition necessary ENV's
		perExperiment.Env["CHAOSENGINE"] = engineDetails.Name
		perExperiment.Env["APP_LABEL"] = engineDetails.AppLabel
		perExperiment.Env["APP_NAMESPACE"] = engineDetails.AppNamespace
		perExperiment.Env["APP_KIND"] = engineDetails.AppKind

		log.Info("Printing the Over-ridden Variables")
		log.Infoln(perExperiment.Env)

		log.Infoln("Converting the Variables using A Range loop to convert the map of ENV to corev1.EnvVar to directly send to the Builder Func")

		// Converting the ENV's (map[string]string)  --> ([]corev1.EnvVar)
		var envVar []corev1.EnvVar
		for k, v := range perExperiment.Env {
			var perEnv corev1.EnvVar
			perEnv.Name = k
			perEnv.Value = v
			envVar = append(envVar, perEnv)
		}

		log.Info("Printing the corev1.EnvVar : ")
		log.Infoln(envVar)

		log.Infoln("getting all the details of the experiment Name : " + engineDetails.Experiments[i])

		// Fetching more details from the CHoasExpeirment needed for execution
		perExperiment.ExpLabels, perExperiment.ExpImage, perExperiment.ExpArgs = utils.GetDetails(engineDetails.AppNamespace, engineDetails.Experiments[i], engineDetails.Config)

		log.Infoln("Variables for ChaosJob : ", "Experiment Labels : ", perExperiment.ExpLabels, " Experiment Image : ", perExperiment.ExpImage, " Experiment Args : ", perExperiment.ExpArgs)

		// Generation of Random String for appending it into Job
		randomString := utils.RandomString()

		// Setting the JobName in Experiment Realted struct
		perExperiment.JobName = engineDetails.Experiments[i] + "-" + randomString

		log.Infoln("JobName for this Experiment : " + perExperiment.JobName)

		// Creation of PodTemplateSpec, and Final Job
		err = utils.DeployJob(perExperiment, engineDetails, envVar)
		if err != nil {
			log.Infoln("Error while building Job : ", err)
		}

		// Genrationg Client Set for more functionality
		var clients utils.ClientSets

		// ClientSet Generation
		clients.KubeClient, clients.LitmusClient, err = utils.GenerateClientSets(engineDetails.Config)
		if err != nil {
			log.Info("Unable to generate ClientSet while Creating Job")
			log.Fatal("Unable to create Client Set : ", err)
		}

		// Getting the Experiment Result Name
		resultName := utils.GetResultName(engineDetails, i)

		// Watching the Job till Completion
		err = utils.WatchingJobtillCompletion(perExperiment, engineDetails, clients)
		if err != nil {
			log.Info("Unable to Watch the Job")
			log.Error(err)
		}

		// Will Update the result,
		// Delete / retain the Job, using the jobCleanUpPolicy
		err = utils.UpdateResultWithJobAndDeletingJob(engineDetails, clients, resultName, perExperiment)
		if err != nil {
			log.Info("Unable to Update Resource")
			log.Error(err)
		}
	}
}
