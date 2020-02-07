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
		klog.V(0).Infof("Error in fetching the config, error : %v", err)
		return
	}
	// clientSet creation for further use.
	if err = clients.GenerateClientSets(config); err != nil {
		klog.V(0).Infof("Unable to create ClientSets")
		return
	}

	// Fetching all the ENV's needed
	utils.GetOsEnv(&engineDetails)
	klog.V(0).Infoln("Experiments List: ", engineDetails.Experiments, " ", "Engine Name: ", engineDetails.Name, " ", "appLabels : ", engineDetails.AppLabel, " ", "appNamespace: ", engineDetails.AppNamespace, " ", "appKind: ", engineDetails.AppKind, " ", "Service Account Name: ", engineDetails.SvcAccount)

	// Steps for each Experiment
	for i := range engineDetails.Experiments {

		// Sending event to GA instance
		if engineDetails.ClientUUID != "" {
			analytics.TriggerAnalytics(engineDetails.Experiments[i], engineDetails.ClientUUID)
		}

		experiment := utils.NewExperimentDetails()
		experiment.SetValueFromChaosEngine(engineDetails, i)
		experiment.SetValueFromChaosExperiment(clients)
		if err := experiment.SetENV(engineDetails, clients); err != nil {
			klog.V(0).Infof("Unable to patch ENV, due to error: %v", err)
			break
		}

		experimentStatus := utils.ExperimentStatus{}
		experimentStatus.IntialExperimentStatus(experiment)
		if err := experimentStatus.InitialPatchEngine(engineDetails, clients); err != nil {
			klog.V(0).Infof("Unable to set Intial Status in ChaosEngine, due to error: %v", err)
		}

		klog.V(0).Infof("Preparing to run Chaos Experiment: %v", experiment.Name)

		if err := experiment.HandleChaosExperimentExistence(engineDetails, clients); err != nil {
			klog.V(0).Infof("Unable to get ChaosExperiment Name: %v, in namespace: %v, due to error: %v", experiment.Name, experiment.Namespace, err)
			break
		}

		if err := experiment.PatchResources(engineDetails, clients); err != nil {
			klog.V(0).Infof("Unable to patch Chaos Resources required for Chaos Experiment: %v, due to error: %v", experiment.Name, err)
		}

		// Creation of PodTemplateSpec, and Final Job
		if err = utils.BuildingAndLaunchJob(experiment, clients); err != nil {
			klog.V(0).Infof("Unable to construct chaos experiment job due to: %v", err)
			break
		}

		time.Sleep(5 * time.Second)

		klog.V(0).Infof("Started Chaos Experiment Name: %v, with Job Name: %v", experiment.Name, experiment.JobName)
		// Watching the Job till Completion
		if err = engineDetails.WatchJobForCompletion(experiment, clients); err != nil {
			klog.V(0).Infof("Unable to Watch the Job, error: %v", err)
			break
		}

		// Will Update the chaosEngine Status
		if err = engineDetails.UpdateEngineWithResult(experiment, clients); err != nil {
			klog.V(0).Infof("Unable to Update ChaosEngine Status due to: %v", err)
		}

		// Delete / retain the Job, using the jobCleanUpPolicy
		if err = engineDetails.DeleteJobAccordingToJobCleanUpPolicy(experiment, clients); err != nil {
			klog.V(0).Infof("Unable to Delete chaosExperiment Job due to: %v", err)
		}
	}
}
