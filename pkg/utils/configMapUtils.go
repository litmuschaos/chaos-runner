package utils

import (
	"errors"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//PatchConfigMaps patches configmaps in experimentDetails struct.
func (expDetails *ExperimentDetails) PatchConfigMaps(clients ClientSets) error {
	expDetails.SetConfigMaps(clients)
	log.Infof("Validating configmaps specified in the ChaosExperiment")
	err := expDetails.ValidateConfigMaps(clients)
	if err != nil {
		log.Infof("Error Validating configMaps, skipping Execution")
		return err
	}
	return nil
}

// ValidateConfigMap validates the configMap, before checking or creating them.
func (clientSets ClientSets) ValidateConfigMap(configMapName string, experiment *ExperimentDetails) error {

	_, err := clientSets.KubeClient.CoreV1().ConfigMaps(experiment.Namespace).Get(configMapName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	return nil

}

// SetConfigMaps sets the value of configMaps in Experiment Structure
func (expDetails *ExperimentDetails) SetConfigMaps(clients ClientSets) {

	chaosExperimentObj, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		log.Infof("Unable to get ChaosEXperiment Resource, wouldn't not be able to patch ConfigMaps")
	}
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
