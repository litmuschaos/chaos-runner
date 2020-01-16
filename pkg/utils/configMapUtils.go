package utils

import (
	"errors"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//PatchConfigMaps patches configmaps in experimentDetails struct.
func (expDetails *ExperimentDetails) PatchConfigMaps(clients ClientSets, engineName string) error {
	expDetails.SetConfigMaps(clients, engineName)
	log.Infof("Validating configmaps specified in the ChaosExperiment & chaosEngine")
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
func (expDetails *ExperimentDetails) SetConfigMaps(clients ClientSets, engineName string) {

	chaosExperimentObj, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		log.Infof("Unable to get ChaosExperiment Resource, wouldn't not be able to patch ConfigMaps")
	}
	experimentConfigMaps := chaosExperimentObj.Spec.Definition.ConfigMaps

	chaosEngineObj, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(expDetails.Namespace).Get(engineName, metav1.GetOptions{})
	if err != nil {
		log.Infof("Unable to get ChaosEngine Resource, wouldn't not be able to patch ConfigMaps")
	}
	experimentsList := chaosEngineObj.Spec.Experiments
	for i := range experimentsList {
		if experimentsList[i].Name == expDetails.Name {
			engineConfigMaps := experimentsList[i].Spec.Components.ConfigMaps
			for j := range engineConfigMaps {
				flag := false
				for k := range experimentConfigMaps {
					if engineConfigMaps[j].Name == experimentConfigMaps[k].Name {
						flag = true
						if engineConfigMaps[j].MountPath != experimentConfigMaps[k].MountPath {
							experimentConfigMaps[k].MountPath = engineConfigMaps[j].MountPath
						}
					}
				}
				if !flag {
					experimentConfigMaps = append(experimentConfigMaps, engineConfigMaps[j])
				}
			}
		}
	}

	expDetails.ConfigMaps = experimentConfigMaps
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
			log.Infof("Unable to get ConfigMap with Name: %v, in namespace: %v", v.Name, expDetails.Namespace)
		} else {
			log.Infof("Succesfully Validated ConfigMap: %v", v.Name)
		}
	}
	return nil
}
