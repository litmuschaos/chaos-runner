package utils

import (
	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

//PatchConfigMaps patches configmaps in experimentDetails struct.
func (expDetails *ExperimentDetails) PatchConfigMaps(clients ClientSets, engineDetails EngineDetails) error {
	err := expDetails.SetConfigMaps(clients, engineDetails)
	if err != nil {
		return err
	}

	klog.V(0).Infof("Validating configmaps specified in the ChaosExperiment & ChaosEngine")
	err = expDetails.ValidateConfigMaps(clients)
	if err != nil {
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
func (expDetails *ExperimentDetails) SetConfigMaps(clients ClientSets, engineDetails EngineDetails) error {

	experimentConfigMaps, err := getExperimentConfigmaps(clients, expDetails)
	if err != nil {
		return err
	}
	engineConfigMaps, err := getEngineConfigmaps(clients, engineDetails, expDetails)
	if err != nil {
		return err
	}
	// Overriding the ConfigMaps from the ChaosEngine
	OverridingConfigMaps(experimentConfigMaps, engineConfigMaps, expDetails)

	return nil
}

// ValidateConfigMaps checks for configMaps in the Application Namespace
func (expDetails *ExperimentDetails) ValidateConfigMaps(clients ClientSets) error {

	for _, v := range expDetails.ConfigMaps {
		if v.Name == "" || v.MountPath == "" {
			return errors.New("Incomplete Information in ConfigMap, will skip execution")
		}
		err := clients.ValidateConfigMap(v.Name, expDetails)
		if err != nil {
			return errors.Wrapf(err, "Unable to get ConfigMap with Name: %v, in namespace: %v", v.Name, expDetails.Namespace)
		}
		klog.V(0).Infof("Succesfully Validated ConfigMap: %v", v.Name)
	}
	return nil
}

func getExperimentConfigmaps(clients ClientSets, expDetails *ExperimentDetails) ([]v1alpha1.ConfigMap, error) {
	chaosExperimentObj, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to get ChaosExperiment Resource,  error: %v", err)
	}
	experimentConfigMaps := chaosExperimentObj.Spec.Definition.ConfigMaps

	return experimentConfigMaps, nil
}

func getEngineConfigmaps(clients ClientSets, engineDetails EngineDetails, expDetails *ExperimentDetails) ([]v1alpha1.ConfigMap, error) {

	chaosEngineObj, err := engineDetails.GetChaosEngine(clients)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to get ChaosEngine Resource,  error: %v", err)
	}
	experimentsList := chaosEngineObj.Spec.Experiments
	for i := range experimentsList {
		if experimentsList[i].Name == expDetails.Name {
			engineConfigMaps := experimentsList[i].Spec.Components.ConfigMaps
			return engineConfigMaps, nil
		}
	}
	return nil, errors.Wrapf(err, "No experiment found with %v name in ChaosEngine", expDetails.Name)
}

// OverridingConfigMaps will override configmaps from ChaosEngine
func OverridingConfigMaps(experimentConfigMaps []v1alpha1.ConfigMap, engineConfigMaps []v1alpha1.ConfigMap, expDetails *ExperimentDetails) {

	for i := range engineConfigMaps {
		flag := false
		for j := range experimentConfigMaps {
			if engineConfigMaps[i].Name == experimentConfigMaps[j].Name {
				flag = true
				if engineConfigMaps[i].MountPath != experimentConfigMaps[j].MountPath {
					experimentConfigMaps[j].MountPath = engineConfigMaps[i].MountPath
				}
			}
		}
		if !flag {
			experimentConfigMaps = append(experimentConfigMaps, engineConfigMaps[i])
		}
	}
	expDetails.ConfigMaps = experimentConfigMaps
}
