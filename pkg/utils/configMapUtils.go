package utils

import (
	"errors"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//PatchConfigMaps patches configmaps in experimentDetails struct.
func (expDetails *ExperimentDetails) PatchConfigMaps(clients ClientSets) error {
	expDetails.SetConfigMaps(clients)
	Logger.WithString(fmt.Sprintf("Validating configmaps specified in the ChaosExperiment")).WithVerbosity(0).Log()
	err := expDetails.ValidateConfigMaps(clients)
	if err != nil {
		Logger.WithString(fmt.Sprintf("Error Validating configMaps, skipping Execution")).WithVerbosity(1).Log()
		return err
	}
	return nil
}

// SetConfigMaps sets the value of configMaps in Experiment Structure
func (expDetails *ExperimentDetails) SetConfigMaps(clients ClientSets) {

	chaosExperimentObj, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		Logger.WithNameSpace(expDetails.Namespace).WithResourceName(expDetails.Name).WithString(err.Error()).WithOperation("Get").WithVerbosity(1).WithResourceType("ChaosExperiment").Log()
	}
	configMaps := chaosExperimentObj.Spec.Definition.ConfigMaps
	expDetails.ConfigMaps = configMaps
}

// ValidateConfigMap validates the configMap, before checking or creating them.
func (clientSets ClientSets) ValidateConfigMap(configMapName string, experiment *ExperimentDetails) error {

	_, err := clientSets.KubeClient.CoreV1().ConfigMaps(experiment.Namespace).Get(configMapName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	return nil

}

// ValidateConfigMaps checks for configMaps in the Application Namespace
func (expDetails *ExperimentDetails) ValidateConfigMaps(clients ClientSets) error {

	for _, v := range expDetails.ConfigMaps {
		if v.Name == "" || v.MountPath == "" {
			//log.Infof("Incomplete Information in ConfigMap, will skip execution")
			return errors.New("Incomplete Information in ConfigMap, will skip execution")
		}
		err := clients.ValidateConfigMap(v.Name, expDetails)
		if err != nil {
			Logger.WithNameSpace(expDetails.Namespace).WithResourceName(v.Name).WithString(err.Error()).WithOperation("List").WithVerbosity(1).WithResourceType("ConfigMap").Log()
		} else {
			Logger.WithString(fmt.Sprintf("Succesfully Validated ConfigMap: %v", v.Name)).WithVerbosity(0).Log()
		}
	}
	return nil
}
