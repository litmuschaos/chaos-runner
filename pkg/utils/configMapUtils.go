package utils

import (
	"errors"
	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	"github.com/litmuschaos/kube-helper/kubernetes/configmap"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// ValidateConfigMaps validates the configMap, before checking or creating them.
func ValidateConfigMaps(configMaps []v1alpha1.ConfigMap, engineDetails EngineDetails, clients ClientSets) ([]v1alpha1.ConfigMap, []error) {

	var errorList []error
	var validConfigMaps []v1alpha1.ConfigMap

	for _, v := range configMaps {
		if v.Name == "" || v.MountPath == "" {
			log.Infof("Unable to validate the configMap, with Name: %v , with mountPath: %v", v.Name, v.MountPath)
			e := errors.New("Aborting Execution, configMap Name or mountPath is invalid")
			errorList = append(errorList, e)
			return nil, errorList
		}

		_, err := clients.KubeClient.CoreV1().ConfigMaps(engineDetails.AppNamespace).Get(v.Name, metav1.GetOptions{})

		if err != nil {
			//errors = append(errors, err)
			log.Infof("Unable to find ConfigMap with Name: %v", v.Name)

			if v.Data != nil {
				log.Infof("Will try to build configMap with Name : %v", v.Name)
				configMapObject := createConfigMapObject(v)

				_, err = clients.KubeClient.CoreV1().ConfigMaps(engineDetails.AppNamespace).Create(configMapObject)

				if err != nil {
					log.Errorf("Unable to create ConfigMap Error : %v", err)
					errorList = append(errorList, err)
					return nil, errorList
				}
				validConfigMaps = append(validConfigMaps, v)
				log.Infof("Successfully created ConfigMap with Name: %v", v.Name)

			} else {
				log.Infof("configMap with name: %v not found. unable to create this configMap as no data is specified. Aborting Execution", v.Name)
				e := errors.New("Aborting Execution, configMap not found & doesn't contain Data")
				errorList = append(errorList, e)
				return nil, errorList
			}

		}

		validConfigMaps = append(validConfigMaps, v)
		log.Infof("Successfully Validated the ConfigMap with Name: %v", v.Name)

	}

	return validConfigMaps, errorList
}

// CheckConfigMaps checks for the configMaps embedded inside the chaosExperiments
func CheckConfigMaps(engineDetails EngineDetails, config *rest.Config, experimentName string) (bool, []v1alpha1.ConfigMap) {
	_, litmusClientSet, err := GenerateClientSets(config)
	if err != nil {
		log.Info(err)
	}
	chaosExperimentObj, err := litmusClientSet.LitmuschaosV1alpha1().ChaosExperiments(engineDetails.AppNamespace).Get(experimentName, metav1.GetOptions{})
	check := chaosExperimentObj.Spec.Definition.ConfigMaps
	if len(check) != 0 {
		return true, check
	} else {
		return false, nil
	}
}

// createConfigMapObject creates configMap
func createConfigMapObject(configMap v1alpha1.ConfigMap) *corev1.ConfigMap {
	// Create label maps
	labels := make(map[string]string)
	labels["Experiment"] = configMap.Name
	configMapObj, err := configmap.NewBuilder().
		WithName(configMap.Name).
		WithData(configMap.Data).
		WithLabels(labels).
		Build()

	if err != nil {
		log.Infoln("Unable to create the ConfigMap Object : ", err)
		return nil
	} else {
		return configMapObj
	}

}

// CreateConfigMaps builds configMaps
func CreateConfigMaps(configMaps []v1alpha1.ConfigMap, engineDetails EngineDetails) error {
	//var dataList []map[string]string
	// Generation of ClientSet for creation
	clientSet, _, err := GenerateClientSets(engineDetails.Config)
	if err != nil {
		log.Info("Unable to generate ClientSet while Creating Job : ", err)
		return err
	}
	for i := range configMaps {
		configMapObject := createConfigMapObject(configMaps[i])
		_, err = clientSet.CoreV1().ConfigMaps(engineDetails.AppNamespace).Create(configMapObject)
		if err != nil {
			log.Infoln("Unable to create ConfigMap using the KubeConfig : ", err)
			return err
		}
	}
	return err
}
