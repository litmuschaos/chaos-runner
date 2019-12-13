package utils

import (
	"errors"
	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// ValidateSecrets validates the secrets, before checking them.
func ValidateSecrets(secrets []v1alpha1.Secret, engineDetails EngineDetails, clients ClientSets) ([]v1alpha1.Secret, []error) {
	var errorList []error
	var validSecrets []v1alpha1.Secret

	for _, v := range secrets {
		if v.Name == "" || v.MountPath == "" {
			log.Infof("Unable to validate the Secret, with Name: %v , with mountPath: %v", v.Name, v.MountPath)
			e := errors.New("Aborting Execution, Secret Name or mountPath is invalid")
			errorList = append(errorList, e)
			return nil, errorList
		}

		_, err := clients.KubeClient.CoreV1().Secrets(engineDetails.AppNamespace).Get(v.Name, metav1.GetOptions{})

		if err != nil {
			//errors = append(errors, err)
			log.Infof("Unable to find Secret with Name: %v . Aborting Execution", v.Name)

			e := errors.New("Aborting Execution, Secret not found")
			errorList = append(errorList, e)
			return nil, errorList

		}

		validSecrets = append(validSecrets, v)
		log.Infof("Successfully Validated the Secret with Name: %v", v.Name)

	}

	return validSecrets, errorList
}

// CheckSecrets checks for the secrets embedded inside the chaosExperiments
func CheckSecrets(engineDetails EngineDetails, config *rest.Config, experimentName string) (bool, []v1alpha1.Secret) {
	_, litmusClientSet, err := GenerateClientSets(config)
	if err != nil {
		log.Info(err)
	}
	chaosExperimentObj, err := litmusClientSet.LitmuschaosV1alpha1().ChaosExperiments(engineDetails.AppNamespace).Get(experimentName, metav1.GetOptions{})
	if err != nil {
		log.Info(err)
	}
	check := chaosExperimentObj.Spec.Definition.Secrets
	if len(check) != 0 {
		return true, check
	} else {
		return false, nil
	}
}
