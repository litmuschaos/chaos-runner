package utils

import (
	"errors"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PatchSecrets patches secrets in experimentDetails.
func (expDetails *ExperimentDetails) PatchSecrets(clients ClientSets) error {
	log.Infof("Find the Secrets in the chaosExperiments")
	expDetails.SetSecrets(clients)
	log.Infof("Validating Secrets")
	err := expDetails.ValidateSecrets(clients)
	if err != nil {
		log.Infof("Aborting Execution")
		return err
	}
	return nil
}

// ValidateSecrets validates the secrets in application Namespace
func (clientSets ClientSets) ValidateSecrets(secretName string, experiment *ExperimentDetails) error {

	_, err := clientSets.KubeClient.CoreV1().Secrets(experiment.Namespace).Get(secretName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	return nil
}

// SetSecrets sets the value of secrets in Experiment Structure
func (expDetails *ExperimentDetails) SetSecrets(clients ClientSets) {

	chaosExperimentObj, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		log.Infof("Unable to get ChaosEXperiment Resource, wouldn't not be able to patch ConfigMaps")
	}
	secrets := chaosExperimentObj.Spec.Definition.Secrets
	expDetails.Secrets = secrets
}

// ValidateSecrets checks for secrets in the Applicaation Namespace
func (expDetails *ExperimentDetails) ValidateSecrets(clients ClientSets) error {

	for _, v := range expDetails.Secrets {
		if v.Name == "" || v.MountPath == "" {
			log.Infof("Incomplete Information in Secret, skipping execution of this ChaosExperiment")
			return errors.New("Moving to next ChaosExperiment")
		}
		err := clients.ValidateSecrets(v.Name, expDetails)
		if err != nil {
			log.Infof("Unable to list Secret: %v, in namespace: %v, skipping execution", v.Name, expDetails.Namespace)
		} else {
			log.Infof("Succesfully Validate Secret with Name: %v", v.Name)
		}
	}
	return nil
}
