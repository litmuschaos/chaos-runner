package utils

import (
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (expDetails *ExperimentDetails) PatchConfigMaps(clients ClientSets) error {
	log.Infof("Finding configMaps in the chaosExperiments")
	expDetails.SetConfigMaps(clients)
	log.Infof("Validating ConfigMaps")
	err := expDetails.ValidateConfigMaps(clients)
	if err != nil {
		log.Infof("Aborting Execution")
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
