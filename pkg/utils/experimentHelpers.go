package utils

import (
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

//SetValueFromChaosExperiment sets value in experimentDetails struct from chaosExperiment
func (expDetails *ExperimentDetails) SetValueFromChaosExperiment(clients ClientSets) {
	expDetails.SetImage(clients)
	expDetails.SetArgs(clients)
	expDetails.SetLabels(clients)
	// Generation of Random String for appending it into Job
	randomString := RandomString()
	// Setting the JobName in Experiment Realted struct
	expDetails.JobName = expDetails.Name + "-" + randomString
}

//SetENV sets ENV values in experimentDetails struct.
func (expDetails *ExperimentDetails) SetENV(engineDetails EngineDetails, clients ClientSets) error {
	// Get the Default ENV's from ChaosExperiment
	klog.V(0).Infof("Getting the ENV Variables")
	if err := expDetails.SetDefaultEnv(clients); err != nil {
		return err
	}

	// OverWriting the Defaults Varibles from the ChaosEngine ENV
	if err := expDetails.SetEnvFromEngine(engineDetails.Name, clients); err != nil {
		return err
	}

	// Adding some addition ENV's from spec.AppInfo of ChaosEngine
	expDetails.Env["CHAOSENGINE"] = engineDetails.Name
	expDetails.Env["APP_LABEL"] = engineDetails.AppLabel
	expDetails.Env["APP_NAMESPACE"] = engineDetails.AppNamespace
	expDetails.Env["APP_KIND"] = engineDetails.AppKind
	expDetails.Env["AUXILIARY_APPINFO"] = engineDetails.AuxiliaryAppInfo

	return nil
}

//SetValueFromChaosEngine sets value in experimentDetails struct from chaosEngine
func (expDetails *ExperimentDetails) SetValueFromChaosEngine(engineDetails EngineDetails, i int) {
	expDetails.Name = engineDetails.Experiments[i]
	expDetails.Namespace = engineDetails.AppNamespace
	expDetails.SvcAccount = engineDetails.SvcAccount
}

// NewExperimentDetails initilizes the structure
func NewExperimentDetails() *ExperimentDetails {
	var experimentDetails ExperimentDetails
	experimentDetails.Env = make(map[string]string)
	experimentDetails.ExpLabels = make(map[string]string)
	return &experimentDetails
}

// HandleChaosExperimentExistence will check the experiment in the app namespace
func (expDetails *ExperimentDetails) HandleChaosExperimentExistence(engineDetails EngineDetails, clients ClientSets) error {

	_, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		if err := engineDetails.ExperimentNotFoundPatchEngine(expDetails, clients); err != nil {
			return errors.Wrapf(err, "Unable to patch Chaos Engine Name: %v, in namespace: %v, due to error: %v", engineDetails.Name, engineDetails.AppNamespace, err)
		}
		return errors.Wrapf(err, "Unable to list Chaos Experiment Name: %v,in Namespace: %v, due to error: %v", expDetails.Name, expDetails.Namespace, err)
	}

	return nil
}

// SetDefaultEnv sets the Env's in Experiment Structure
func (expDetails *ExperimentDetails) SetDefaultEnv(clients ClientSets) error {
	experimentEnv, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "Unable to get the Default ENV from ChaosExperiment, error : %v", err)
	}

	envList := experimentEnv.Spec.Definition.ENVList
	expDetails.Env = make(map[string]string)
	for i := range envList {
		key := envList[i].Name
		value := envList[i].Value
		expDetails.Env[key] = value
	}
	return nil
}

// SetEnvFromEngine will over-ride the default variables from one provided in the chaosEngine
func (expDetails *ExperimentDetails) SetEnvFromEngine(engineName string, clients ClientSets) error {

	engineSpec, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(expDetails.Namespace).Get(engineName, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "Unable to get ChaosEngine Resource in namespace: %v", expDetails.Namespace)
	}
	envList := engineSpec.Spec.Experiments
	for i := range envList {
		if envList[i].Name == expDetails.Name {
			keyValue := envList[i].Spec.Components.ENV
			for j := range keyValue {
				expDetails.Env[keyValue[j].Name] = keyValue[j].Value
			}
		}
	}
	return nil

}

// SetLabels sets the Experiment Labels, in Experiment Structure
func (expDetails *ExperimentDetails) SetLabels(clients ClientSets) {
	expirementSpec, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		klog.V(0).Infoln(err)
	}
	expDetails.ExpLabels = expirementSpec.Spec.Definition.Labels

}

// SetImage sets the Experiment Image, in Experiment Structure
func (expDetails *ExperimentDetails) SetImage(clients ClientSets) {
	expirementSpec, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		klog.V(0).Infoln(err)
	}
	expDetails.ExpImage = expirementSpec.Spec.Definition.Image
}

// SetArgs sets the Experiment Image, in Experiment Structure
func (expDetails *ExperimentDetails) SetArgs(clients ClientSets) {
	expirementSpec, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		klog.V(0).Infoln(err)
	}
	expDetails.ExpArgs = expirementSpec.Spec.Definition.Args
}
