package utils

import (
	"errors"
	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	"github.com/litmuschaos/kube-helper/kubernetes/configmap"
	volume "github.com/litmuschaos/kube-helper/kubernetes/volume/v1alpha1"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/client-go/rest"
)

// CheckConfigMaps checks for the configMaps embedded inside the chaosExperiments
func CheckConfigMaps(engineDetails EngineDetails, experimentName string) (bool, []v1alpha1.ConfigMap) {

	chaosExperimentObj, err := engineDetails.Clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(engineDetails.AppNamespace).Get(experimentName, metav1.GetOptions{})
	if err != nil {
		log.Info(err)
	}
	check := chaosExperimentObj.Spec.Definition.ConfigMaps
	if len(check) != 0 {
		return true, check
	} else {
		return false, nil
	}
}

// CheckSecrets checks for the configMaps embedded inside the chaosExperiments
func CheckSecrets(engineDetails EngineDetails, experimentName string) (bool, []v1alpha1.Secret) {

	chaosExperimentObj, err := engineDetails.Clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(engineDetails.AppNamespace).Get(experimentName, metav1.GetOptions{})
	if err != nil {
		log.Info(err)
	}
	check := chaosExperimentObj.Spec.Definition.Secrets
	log.Info(check)
	if len(check) != 0 {
		return true, check
	} else {
		return false, nil
	}
}

// createConfigMapObject
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

// ValidateSecrets validates the secrets, before checking them.
func ValidateSecrets(secrets []v1alpha1.Secret, engineDetails EngineDetails) ([]v1alpha1.Secret, []error) {
	var errorList []error
	var validSecrets []v1alpha1.Secret

	for _, v := range secrets {
		if v.Name == "" || v.MountPath == "" {
			log.Infof("Unable to validate the Secret, with Name: %v , with mountPath: %v", v.Name, v.MountPath)
			e := errors.New("Aborting Execution, Secret Name or mountPath is invalid")
			errorList = append(errorList, e)
			return nil, errorList
		}

		_, err := engineDetails.Clients.KubeClient.CoreV1().Secrets(engineDetails.AppNamespace).Get(v.Name, metav1.GetOptions{})

		if err != nil {
			//errors = append(errors, err)
			log.Infof("Unable to find ConfigMap with Name: %v", v.Name)

			log.Infof("Did'nt find the configMap with Name: %v, and Data is also empty. Aborting Execution", v.Name)

			e := errors.New("Aborting Execution, configMap not found & doesn't contain Data")
			errorList = append(errorList, e)
			return nil, errorList

		}

		validSecrets = append(validSecrets, v)
		log.Infof("Successfully Validated the Secret with Name: %v", v.Name)

	}

	return validSecrets, errorList
}

// ValidateConfigMaps validates the configMap, before checking or creating them.
func ValidateConfigMaps(configMaps []v1alpha1.ConfigMap, engineDetails EngineDetails) ([]v1alpha1.ConfigMap, []error) {

	var errorList []error
	var validConfigMaps []v1alpha1.ConfigMap

	for _, v := range configMaps {
		if v.Name == "" || v.MountPath == "" {
			log.Infof("Unable to validate the configMap, with Name: %v , with mountPath: %v", v.Name, v.MountPath)
			e := errors.New("Aborting Execution, configMap Name or mountPath is invalid")
			errorList = append(errorList, e)
			return nil, errorList
		}

		_, err := engineDetails.Clients.KubeClient.CoreV1().ConfigMaps(engineDetails.AppNamespace).Get(v.Name, metav1.GetOptions{})

		if err != nil {
			//errors = append(errors, err)
			log.Infof("Unable to find ConfigMap with Name: %v", v.Name)

			if v.Data != nil {
				log.Infof("Will try to build configMap with Name : %v", v.Name)
				configMapObject := createConfigMapObject(v)

				_, err = engineDetails.Clients.KubeClient.CoreV1().ConfigMaps(engineDetails.AppNamespace).Create(configMapObject)

				if err != nil {
					log.Errorf("Unable to create ConfigMap Error : %v", err)
					errorList = append(errorList, err)
					return nil, errorList
				}
				validConfigMaps = append(validConfigMaps, v)
				log.Infof("Successfully created ConfigMap with Name: %v", v.Name)

			} else {
				log.Infof("Did'nt find the configMap with Name: %v, and Data is also empty. Aborting Execution", v.Name)
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

// CreateConfigMaps builds configMaps
func CreateConfigMaps(configMaps []v1alpha1.ConfigMap, engineDetails EngineDetails) error {

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

// CreateVolumeBuilder build Volume needed in execution of experiments
func CreateVolumeBuilder(configMaps []v1alpha1.ConfigMap, secrets []v1alpha1.Secret) []*volume.Builder {
	volumeBuilderList := []*volume.Builder{}
	if configMaps == nil {
		log.Infoln("Unable to fetch chaosExperiment ConfigMaps, to create volume")
	}
	for _, v := range configMaps {
		log.Infof("Would create VolumeBuilder for ConfigMap Name: %v", v.Name)
		volumeBuilder := volume.NewBuilder().WithConfigMap(v.Name)
		volumeBuilderList = append(volumeBuilderList, volumeBuilder)
	}

	if secrets == nil {
		log.Infoln("Unable to getch Valid chaosExperiment Secrets, to create Volumes")
	}

	for _, v := range secrets {
		log.Infof("Would create VolumeBuilder for Secret Name: %v", v.Name)
		volumeBuilder := volume.NewBuilder().WithSecret(v.Name)
		volumeBuilderList = append(volumeBuilderList, volumeBuilder)
	}
	return volumeBuilderList
}

// CreateVolumeMounts mounts Volume needed in execution of experiments
func CreateVolumeMounts(configMaps []v1alpha1.ConfigMap, secrets []v1alpha1.Secret) []corev1.VolumeMount {
	var volumeMountsList []corev1.VolumeMount
	for _, v := range configMaps {
		//volumeMount = make(corev1.VolumeMount)
		var volumeMount corev1.VolumeMount
		volumeMount.Name = v.Name
		volumeMount.MountPath = v.MountPath
		volumeMountsList = append(volumeMountsList, volumeMount)
	}

	for _, v := range secrets {
		//volumeMount = make(corev1.VolumeMount)
		var volumeMount corev1.VolumeMount
		volumeMount.Name = v.Name
		volumeMount.MountPath = v.MountPath
		volumeMountsList = append(volumeMountsList, volumeMount)
	}
	return volumeMountsList
}
