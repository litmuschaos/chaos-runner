package utils

import (
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	"github.com/litmuschaos/chaos-runner/pkg/log"
)

// PatchSecrets patches secrets in experimentDetails.
func (expDetails *ExperimentDetails) PatchSecrets(clients ClientSets, engineDetails EngineDetails) error {
	err := expDetails.SetSecrets(clients, engineDetails)
	if err != nil {
		return err
	}

	log.Infof("Validating secrets specified in the ChaosExperiment & ChaosEngine")
	err = expDetails.ValidateSecrets(clients)
	if err != nil {
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
func (expDetails *ExperimentDetails) SetSecrets(clients ClientSets, engineDetails EngineDetails) error {
	experimentSecrets, err := getExperimentSecrets(clients, expDetails)
	if err != nil {
		return err
	}
	engineSecrets, err := getEngineSecrets(clients, engineDetails, expDetails)
	if err != nil {
		return err
	}

	// Overriding the Secrets from the ChaosEngine
	OverridingSecrets(experimentSecrets, engineSecrets, expDetails)

	return nil
}

// ValidateSecrets checks for secrets in the Applicaation Namespace
func (expDetails *ExperimentDetails) ValidateSecrets(clients ClientSets) error {

	for _, v := range expDetails.Secrets {
		if v.Name == "" || v.MountPath == "" {
			return errors.Errorf("Incomplete Information in Secret, will skip execution")
		}
		err := clients.ValidateSecrets(v.Name, expDetails)
		if err != nil {
			return errors.Errorf("Unable to get Secret with Name: %v, in namespace: %v", v.Name, expDetails.Namespace)
		}
		log.Infof("Successfully Validated Secret: %v", v.Name)
	}
	return nil
}

func getExperimentSecrets(clients ClientSets, expDetails *ExperimentDetails) ([]v1alpha1.Secret, error) {
	chaosExperimentObj, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Errorf("Unable to get ChaosExperiment Resource, error: %v", err)
	}
	experimentSecrets := chaosExperimentObj.Spec.Definition.Secrets

	return experimentSecrets, nil
}

func getEngineSecrets(clients ClientSets, engineDetails EngineDetails, expDetails *ExperimentDetails) ([]v1alpha1.Secret, error) {
	chaosEngineObj, err := engineDetails.GetChaosEngine(clients)
	if err != nil {
		return nil, errors.Errorf("Unable to get ChaosEngine Resource, error: %v", err)
	}
	experimentsList := chaosEngineObj.Spec.Experiments
	for i := range experimentsList {
		if experimentsList[i].Name == expDetails.Name {
			engineSecrets := experimentsList[i].Spec.Components.Secrets
			return engineSecrets, nil
		}
	}
	return nil, errors.Errorf("No experiment found with %v name in ChaosEngine", expDetails.Name)
}

// OverridingSecrets will override secrets from ChaosEngine
func OverridingSecrets(experimentSecrets []v1alpha1.Secret, engineSecrets []v1alpha1.Secret, expDetails *ExperimentDetails) {

	for i := range engineSecrets {
		flag := false
		for j := range experimentSecrets {
			if engineSecrets[i].Name == experimentSecrets[j].Name {
				flag = true
				if engineSecrets[i].MountPath != experimentSecrets[j].MountPath {
					experimentSecrets[j].MountPath = engineSecrets[i].MountPath
				}
			}
		}
		if !flag {
			experimentSecrets = append(experimentSecrets, engineSecrets[i])
		}
	}
	expDetails.Secrets = experimentSecrets
}
