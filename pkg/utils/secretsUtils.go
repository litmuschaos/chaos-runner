package utils

import (
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
