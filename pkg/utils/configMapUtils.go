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
	/*_, litmusClientSet, err := GenerateClientSets(config)
	if err != nil {
		log.Info(err)
	}*/
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
	/*_, litmusClientSet, err := GenerateClientSets(config)
	if err != nil {
		log.Info(err)
	}*/
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
	// Generation of ClientSet for validation
	/*clientSet, _, err := GenerateClientSets(engineDetails.Config)
	if err != nil {
		log.Info("Unable to generate ClientSet while Creating Job : ", err)
		return err
	}
	for i := range configMaps {
		if configMaps[i].Name == "" || configMaps[i].MountPath == "" {
			log.Errorf("Unable to create this configMap with Name: %v , with MoountPath: %v", configMaps[i].Name, configMaps[i].MountPath)
			continue
		}
		_, err := clientSet.CoreV1().ConfigMaps(engineDetails.AppNamespace).Get(configMaps[i].Name, metav1.GetOptions{})
		if err != nil {
			log.Errorf("Unable to find the ConfigMap with name : %v", configMaps[i].Name)

			// Will check for configMap Data, if found create configMap
			if configMaps[i].Data != nil {
				log.Infof("Will try to build configMap with Name : %v", configMaps[i].Name)
				configMapObject := createConfigMapObject(configMaps[i])
				_, err = clientSet.CoreV1().ConfigMaps(engineDetails.AppNamespace).Create(configMapObject)
				if err != nil {
					//log.Errorf("Unable to create ConfigMap Error : %v", err)
					return err
				}

			}
		} else {
			log.Infof("ConfigMap with Name : %v , found", configMaps[i].Name)
		}
	}
	return nil*/
	//return nil
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

/*func CreateVolumes(configMaps []v1alpha1.ConfigMap) []corev1.Volume {
	//volumesList := make([]corev1.Volume, len(configMaps))
	log.Infoln("-------------------------------------")
	log.Infoln(configMaps)
	log.Infoln("--------------------------")
	volumesList := []corev1.Volume{}

	for _, v := range configMaps {
		log.Infoln("RETURNING OBJECT")
		log.Infoln(v)
		volume := corev1.Volume{}
		volume.Name = v.Name
		log.Infoln("!!!!!!!!!!!!!!!!!!!!!!!!!!!! ->")
		log.Infoln(v.Name)
		volume.ConfigMap.Name = v.Name
		var i int = 420
		var k int32 = int32(i)
		volume.ConfigMap.DefaultMode = &k
		volumesList = append(volumesList, volume)
		//volumesList[i] = volume
	}
	return volumesList
}*/

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
