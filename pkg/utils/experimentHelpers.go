package utils

import (
	"errors"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (expDetails *ExperimentDetails) SetValueFromChaosExperiment(clients ClientSets) {
	expDetails.SetImage(clients)
	expDetails.SetArgs(clients)
	expDetails.SetLabels(clients)
	// Generation of Random String for appending it into Job
	randomString := RandomString()
	// Setting the JobName in Experiment Realted struct
	expDetails.JobName = expDetails.Name + "-" + randomString
}
func (expDetails *ExperimentDetails) SetENV(engineDetails EngineDetails, clients ClientSets) {
	// Get the Default ENV's from ChaosExperiment
	log.Infoln("Getting the Default ENV Variables")
	expDetails.SetDefaultEnv(clients)
	// OverWriting the Defaults Varibles from the ChaosEngine ENV
	log.Infoln("Patching some required ENV's")
	expDetails.SetEnvFromEngine(engineDetails.Name, clients)
	// Adding some addition necessary ENV's
	expDetails.Env["CHAOSENGINE"] = engineDetails.Name
	expDetails.Env["APP_LABEL"] = engineDetails.AppLabel
	expDetails.Env["APP_NAMESPACE"] = engineDetails.AppNamespace
	expDetails.Env["APP_KIND"] = engineDetails.AppKind
}
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
func (expDetails *ExperimentDetails) CheckExistence(clients ClientSets) bool {

	_, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	//log.Infof("error while getting exp: %v", err)
	if err != nil {
		return false
	}
	return true
}

// SetDefaultEnv sets the Env's in Experiment Structure
func (expDetails *ExperimentDetails) SetDefaultEnv(clients ClientSets) {
	experimentEnv, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		log.Infof("Unable to get the Default ENV from ChaosExperiment, error : %v", err)
	}

	envList := experimentEnv.Spec.Definition.ENVList
	expDetails.Env = make(map[string]string)
	for i := range envList {
		key := envList[i].Name
		value := envList[i].Value
		expDetails.Env[key] = value
	}
}

// SetConfigMaps sets the value of configMaps in Experiment Structure
func (expDetails *ExperimentDetails) SetConfigMaps(clients ClientSets) {

	chaosExperimentObj, _ := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	configMaps := chaosExperimentObj.Spec.Definition.ConfigMaps
	expDetails.ConfigMaps = configMaps
}

// ValidateConfigMaps checks for configMaps in the Application Namespace
func (expDetails *ExperimentDetails) ValidateConfigMaps(clients ClientSets) error {

	for _, v := range expDetails.ConfigMaps {
		if v.Name == "" || v.MountPath == "" {
			log.Infof("Incomplete Information in ConfigMap, will abort execution")
			return errors.New("Abort Execution")
		}
		err := clients.ValidateConfigMap(v.Name, expDetails)
		if err != nil {
			log.Infof("Unable to get ConfigMap with Name: %v, in namespace: %v", v.Name, expDetails.Namespace)
		} else {
			log.Infof("Succesfully Validate ConfigMap with Name: %v", v.Name)
		}
	}
	return nil
}

// SetSecrets sets the value of secrets in Experiment Structure
func (expDetails *ExperimentDetails) SetSecrets(clients ClientSets) {

	chaosExperimentObj, _ := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	secrets := chaosExperimentObj.Spec.Definition.Secrets
	expDetails.Secrets = secrets
}

// ValidateSecrets checks for secrets in the Applicaation Namespace
func (expDetails *ExperimentDetails) ValidateSecrets(clients ClientSets) error {

	for _, v := range expDetails.Secrets {
		if v.Name == "" || v.MountPath == "" {
			log.Infof("Incomplete Information in Secret, will abort execution")
			return errors.New("Abort Execution")
		}
		err := clients.ValidateSecrets(v.Name, expDetails)
		if err != nil {
			log.Infof("Unable to get Secrets with Name: %v, in namespace: %v", v.Name, expDetails.Namespace)
		} else {
			log.Infof("Succesfully Validate Secret with Name: %v", v.Name)
		}
	}
	return nil
}

// SetEnvFromEngine will over-ride the default variables from one provided in the chaosEngine
func (expDetails *ExperimentDetails) SetEnvFromEngine(engineName string, clients ClientSets) {

	engineSpec, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(expDetails.Namespace).Get(engineName, metav1.GetOptions{})
	if err != nil {
		log.Infof("Unable to get the chaosEngine, error : %v", err)
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
		log.Infoln(err)
	}
	expDetails.ExpLabels = expirementSpec.Spec.Definition.Labels

}

// SetImage sets the Experiment Image, in Experiment Structure
func (expDetails *ExperimentDetails) SetImage(clients ClientSets) {
	expirementSpec, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		log.Infoln(err)
	}
	expDetails.ExpImage = expirementSpec.Spec.Definition.Image
}

// SetArgs sets the Experiment Image, in Experiment Structure
func (expDetails *ExperimentDetails) SetArgs(clients ClientSets) {
	expirementSpec, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		log.Infoln(err)
	}
	expDetails.ExpArgs = expirementSpec.Spec.Definition.Args
}
