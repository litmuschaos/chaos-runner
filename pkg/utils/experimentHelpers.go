package utils

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
func (expDetails *ExperimentDetails) SetENV(engineDetails EngineDetails, clients ClientSets) {
	// Get the Default ENV's from ChaosExperiment
	Logger.WithString(fmt.Sprintf("Getting the ENV Variables for chaosJob")).WithVerbosity(0).Log()
	expDetails.SetDefaultEnv(clients)
	// OverWriting the Defaults Varibles from the ChaosEngine ENV
	expDetails.SetEnvFromEngine(engineDetails.Name, clients)
	// Adding some addition ENV's from spec.AppInfo of ChaosEngine
	expDetails.Env["CHAOSENGINE"] = engineDetails.Name
	expDetails.Env["APP_LABEL"] = engineDetails.AppLabel
	expDetails.Env["APP_NAMESPACE"] = engineDetails.AppNamespace
	expDetails.Env["APP_KIND"] = engineDetails.AppKind
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

// CheckExistence will check the experiment in the app namespace
func (expDetails *ExperimentDetails) CheckExistence(clients ClientSets) (bool, error) {

	_, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	return true, nil
}

// SetDefaultEnv sets the Env's in Experiment Structure
func (expDetails *ExperimentDetails) SetDefaultEnv(clients ClientSets) {
	experimentEnv, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		Logger.WithNameSpace(expDetails.Namespace).WithResourceName(expDetails.Name).WithString(err.Error()).WithOperation("Get").WithVerbosity(1).WithResourceType("ChaosExperiment").Log()
	}

	envList := experimentEnv.Spec.Definition.ENVList
	expDetails.Env = make(map[string]string)
	for i := range envList {
		key := envList[i].Name
		value := envList[i].Value
		expDetails.Env[key] = value
	}
}

// SetEnvFromEngine will over-ride the default variables from one provided in the chaosEngine
func (expDetails *ExperimentDetails) SetEnvFromEngine(engineName string, clients ClientSets) {

	engineSpec, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(expDetails.Namespace).Get(engineName, metav1.GetOptions{})
	if err != nil {
		Logger.WithNameSpace(expDetails.Namespace).WithResourceName(engineName).WithString(err.Error()).WithOperation("Get").WithVerbosity(1).WithResourceType("ChaosEngine").Log()
	}
	envList := engineSpec.Spec.Experiments
	for i := range envList {
		if envList[i].Name == expDetails.Name {
			keyValue := envList[i].Spec.Components
			for j := range keyValue {
				expDetails.Env[keyValue[j].Name] = keyValue[j].Value
			}
		}
	}
}

// SetLabels sets the Experiment Labels, in Experiment Structure
func (expDetails *ExperimentDetails) SetLabels(clients ClientSets) {
	expirementSpec, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		Logger.WithNameSpace(expDetails.Namespace).WithResourceName(expDetails.Name).WithString(err.Error()).WithOperation("Get").WithVerbosity(1).WithResourceType("ChaosExperiment").Log()
	}
	expDetails.ExpLabels = expirementSpec.Spec.Definition.Labels

}

// SetImage sets the Experiment Image, in Experiment Structure
func (expDetails *ExperimentDetails) SetImage(clients ClientSets) {
	expirementSpec, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		Logger.WithNameSpace(expDetails.Namespace).WithResourceName(expDetails.Name).WithString(err.Error()).WithOperation("Get").WithVerbosity(1).WithResourceType("ChaosExperiment").Log()
	}
	expDetails.ExpImage = expirementSpec.Spec.Definition.Image
}

// SetArgs sets the Experiment Image, in Experiment Structure
func (expDetails *ExperimentDetails) SetArgs(clients ClientSets) {
	expirementSpec, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		Logger.WithNameSpace(expDetails.Namespace).WithResourceName(expDetails.Name).WithString(err.Error()).WithOperation("Get").WithVerbosity(1).WithResourceType("ChaosExperiment").Log()
	}
	expDetails.ExpArgs = expirementSpec.Spec.Definition.Args
}
