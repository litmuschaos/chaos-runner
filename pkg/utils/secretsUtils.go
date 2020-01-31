package utils

import (
	"errors"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PatchSecrets patches secrets in experimentDetails.
func (expDetails *ExperimentDetails) PatchSecrets(clients ClientSets) error {
	Logger.WithString(fmt.Sprintf("Validating secrets specified in the ChaosExperiment")).WithVerbosity(0).Log()
	expDetails.SetSecrets(clients)
	err := expDetails.ValidateSecrets(clients)
	if err != nil {
		Logger.WithString(fmt.Sprintf("Error Validating secrets, skipping Execution")).WithVerbosity(1).Log()
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
		Logger.WithNameSpace(expDetails.Namespace).WithResourceName(expDetails.Name).WithString(err.Error()).WithOperation("Get").WithVerbosity(1).WithResourceType("ChaosExperiment").Log()

	}
	secrets := chaosExperimentObj.Spec.Definition.Secrets
	expDetails.Secrets = secrets
}

// ValidateSecrets checks for secrets in the Applicaation Namespace
func (expDetails *ExperimentDetails) ValidateSecrets(clients ClientSets) error {

	for _, v := range expDetails.Secrets {
		if v.Name == "" || v.MountPath == "" {
			//log.Infof("Incomplete Information in Secret, skipping execution of this ChaosExperiment")
			return errors.New("Incomplete Information in Secret, will skip execution")
		}
		err := clients.ValidateSecrets(v.Name, expDetails)
		if err != nil {
			Logger.WithNameSpace(expDetails.Namespace).WithResourceName(v.Name).WithString(err.Error()).WithOperation("List").WithVerbosity(1).WithResourceType("Secret").Log()
		} else {
			Logger.WithString(fmt.Sprintf("Succesfully Validated Secret: %v", v.Name)).WithVerbosity(0).Log()
		}
	}
	return nil
}
