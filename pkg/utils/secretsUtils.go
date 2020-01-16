package utils

import (
	"errors"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PatchSecrets patches secrets in experimentDetails.
func (expDetails *ExperimentDetails) PatchSecrets(clients ClientSets, engineName string) error {
	expDetails.SetSecrets(clients, engineName)
	log.Infof("Validating secrets specified in the ChaosExperiment & chaosEngine")
	err := expDetails.ValidateSecrets(clients)
	if err != nil {
		log.Infof("Error Validating secrets, skipping Execution")
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
func (expDetails *ExperimentDetails) SetSecrets(clients ClientSets, engineName string) {

	chaosExperimentObj, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		log.Infof("Unable to get ChaosEXperiment Resource, wouldn't not be able to patch Secrets")
	}
	experimentSecrets := chaosExperimentObj.Spec.Definition.Secrets

	chaosEngineObj, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(expDetails.Namespace).Get(engineName, metav1.GetOptions{})
	if err != nil {
		log.Infof("Unable to get ChaosEngine Resource, wouldn't not be able to patch Secrets")
	}
	experimentsList := chaosEngineObj.Spec.Experiments
	for i := range experimentsList {
		if experimentsList[i].Name == expDetails.Name {
			engineSecrets := experimentsList[i].Spec.Components.Secrets
			for j := range engineSecrets {
				flag := false
				for k := range experimentSecrets {
					if engineSecrets[j].Name == experimentSecrets[k].Name {
						flag = true
						if engineSecrets[j].MountPath != experimentSecrets[k].MountPath {
							experimentSecrets[k].MountPath = engineSecrets[j].MountPath
						}
					}
				}
				if !flag {
					experimentSecrets = append(experimentSecrets, engineSecrets[j])
				}
			}
		}
	}

	expDetails.Secrets = experimentSecrets
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
			log.Infof("Unable to list Secret: %v, in namespace: %v, skipping execution", v.Name, expDetails.Namespace)
		} else {
			log.Infof("Succesfully Validated Secret: %v", v.Name)
		}
	}
	return nil
}
