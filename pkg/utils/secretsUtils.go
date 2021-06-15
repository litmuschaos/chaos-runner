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

	if len(expDetails.Secrets) != 0 {
		log.Infof("Validating secrets specified in the ChaosExperiment & ChaosEngine")
		if err = expDetails.ValidateSecrets(clients); err != nil {
			return err
		}
	}
	return nil
}

// ValidatePresenceOfSecretResourceInCluster validates the secret in Chaos Namespace
func (clientSets ClientSets) ValidatePresenceOfSecretResourceInCluster(secretName, namespace string) error {
	_, err := clientSets.KubeClient.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
	return err
}

// SetSecrets sets the value of secrets in Experiment Structure
func (expDetails *ExperimentDetails) SetSecrets(clients ClientSets, engineDetails EngineDetails) error {
	experimentSecrets, err := expDetails.getSecretsFromChaosExperiment(clients)
	if err != nil {
		return err
	}
	engineSecrets, err := expDetails.getSecretsFromChaosEngine(clients, engineDetails)
	if err != nil {
		return err
	}
	// Overriding the Secrets from the ChaosEngine
	expDetails.getOverridingSecretsFromChaosEngine(experimentSecrets, engineSecrets)

	return nil
}

// ValidateSecrets checks for secrets in the Chaos Namespace
func (expDetails *ExperimentDetails) ValidateSecrets(clients ClientSets) error {
	for _, v := range expDetails.Secrets {
		if v.Name == "" || v.MountPath == "" {
			return errors.Errorf("Incomplete Information in Secret, will skip execution")
		}
		err := clients.ValidatePresenceOfSecretResourceInCluster(v.Name, expDetails.Namespace)
		if err != nil {
			return errors.Errorf("unable to get Secret with Name: %v, in namespace: %v, error: %v", v.Name, expDetails.Namespace, err)
		}
		log.Infof("Successfully Validated Secret: %v", v.Name)
	}
	return nil
}

func (expDetails *ExperimentDetails) getSecretsFromChaosExperiment(clients ClientSets) ([]v1alpha1.Secret, error) {
	chaosExperimentObj, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Errorf("unable to get ChaosExperiment Resource, error: %v", err)
	}
	experimentSecrets := chaosExperimentObj.Spec.Definition.Secrets

	return experimentSecrets, nil
}

func (expDetails *ExperimentDetails) getSecretsFromChaosEngine(clients ClientSets, engineDetails EngineDetails) ([]v1alpha1.Secret, error) {
	chaosEngineObj, err := engineDetails.GetChaosEngine(clients)
	if err != nil {
		return nil, errors.Errorf("unable to get ChaosEngine Resource, error: %v", err)
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

// getOverridingSecretsFromChaosEngine will override secrets from ChaosEngine
func (expDetails *ExperimentDetails) getOverridingSecretsFromChaosEngine(experimentSecrets []v1alpha1.Secret, engineSecrets []v1alpha1.Secret) {
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
