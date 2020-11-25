package utils

import (
	"reflect"
	"strconv"

	litmuschaosv1alpha1 "github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SetValueFromChaosEngine set the value from the chaosengine
func (expDetails *ExperimentDetails) SetValueFromChaosEngine(engine *EngineDetails, clients ClientSets) error {
	chaosEngine, err := engine.GetChaosEngine(clients)
	if err != nil {
		return errors.Errorf("Unable to get chaosEngine in namespace: %s", engine.EngineNamespace)
	}
	// fetch all the values from chaosengine and set into expDetails struct
	expDetails.SetExpAnnotationFromEngine(chaosEngine).
		SetExpNodeSelectorFromEngine(chaosEngine).
		SetResourceRequirementsFromEngine(chaosEngine).
		SetImagePullSecretsFromEngine(chaosEngine).
		SetTolerationsFromEngine(chaosEngine).
		SetExpImageFromEngine(chaosEngine)

	engine.SetChaosUIDFromEngine(chaosEngine)

	return nil
}

// SetExpImageFromEngine will override the default exp image with the one provided in the chaosEngine
func (expDetails *ExperimentDetails) SetExpImageFromEngine(engine *litmuschaosv1alpha1.ChaosEngine) *ExperimentDetails {
	expRefList := engine.Spec.Experiments
	for i := range expRefList {
		if expRefList[i].Name == expDetails.Name {

			if expRefList[i].Spec.Components.ExperimentImage != "" {
				expDetails.ExpImage = expRefList[i].Spec.Components.ExperimentImage
			}
		}
	}
	return expDetails
}

// SetExpAnnotationFromEngine override the default exp annotation with the one provided in the chaosEngine
func (expDetails *ExperimentDetails) SetExpAnnotationFromEngine(engine *litmuschaosv1alpha1.ChaosEngine) *ExperimentDetails {
	expRefList := engine.Spec.Experiments
	for i := range expRefList {
		if expRefList[i].Name == expDetails.Name {
			expDetails.Annotations = expRefList[i].Spec.Components.ExperimentAnnotations
		}
	}
	return expDetails
}

// SetResourceRequirementsFromEngine add the resource requirements provided inside chaosengine
func (expDetails *ExperimentDetails) SetResourceRequirementsFromEngine(engine *litmuschaosv1alpha1.ChaosEngine) *ExperimentDetails {
	expRefList := engine.Spec.Experiments
	for i := range expRefList {
		if expRefList[i].Name == expDetails.Name {
			expDetails.ResourceRequirements = expRefList[i].Spec.Components.Resources
		}
	}
	return expDetails
}

// SetImagePullSecretsFromEngine add the image pull secrets provided inside chaosengine
func (expDetails *ExperimentDetails) SetImagePullSecretsFromEngine(engine *litmuschaosv1alpha1.ChaosEngine) *ExperimentDetails {
	expRefList := engine.Spec.Experiments
	for i := range expRefList {
		if expRefList[i].Name == expDetails.Name {
			if expRefList[i].Spec.Components.ExperimentImagePullSecrets != nil {
				expDetails.ImagePullSecrets = expRefList[i].Spec.Components.ExperimentImagePullSecrets
			}
		}
	}
	return expDetails
}

// SetExpNodeSelectorFromEngine add the nodeSelector attribute based on the key/value provided in the chaosEngine
func (expDetails *ExperimentDetails) SetExpNodeSelectorFromEngine(engine *litmuschaosv1alpha1.ChaosEngine) *ExperimentDetails {
	expRefList := engine.Spec.Experiments
	for i := range expRefList {
		if expRefList[i].Name == expDetails.Name {
			expDetails.NodeSelector = expRefList[i].Spec.Components.NodeSelector
		}
	}
	return expDetails
}

// SetTolerationsFromEngine add the tolerations based on the key/operator/effect provided in the chaosEngine
func (expDetails *ExperimentDetails) SetTolerationsFromEngine(engine *litmuschaosv1alpha1.ChaosEngine) *ExperimentDetails {
	expRefList := engine.Spec.Experiments
	for i := range expRefList {
		if expRefList[i].Name == expDetails.Name {
			expDetails.Tolerations = expRefList[i].Spec.Components.Tolerations
		}
	}
	return expDetails
}

// SetChaosUIDFromEngine sets the chaosuid from the chaosengine
func (engineDetails *EngineDetails) SetChaosUIDFromEngine(engine *litmuschaosv1alpha1.ChaosEngine) {
	engineDetails.UID = string(engine.UID)
}

// SetEnvFromEngine override the default envs with envs passed inside the chaosengine
func (expDetails *ExperimentDetails) SetEnvFromEngine(engineName string, clients ClientSets) error {

	engineSpec, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(expDetails.Namespace).Get(engineName, metav1.GetOptions{})
	if err != nil {
		return errors.Errorf("Unable to get ChaosEngine Resource in namespace: %v", expDetails.Namespace)
	}
	expList := engineSpec.Spec.Experiments
	for i := range expList {
		if expList[i].Name == expDetails.Name {
			keyValue := expList[i].Spec.Components.ENV
			for j := range keyValue {
				expDetails.Env[keyValue[j].Name] = keyValue[j].Value
				// extracting and storing instance id explicitly
				// as we need this variable while generating chaos-result name
				if keyValue[j].Name == "INSTANCE_ID" {
					expDetails.InstanceID = keyValue[j].Value
				}
			}
			statusCheckTimeout := expList[i].Spec.Components.StatusCheckTimeouts
			if !reflect.DeepEqual(statusCheckTimeout, litmuschaosv1alpha1.StatusCheckTimeout{}) {

				expDetails.Env["STATUS_CHECK_DELAY"] = strconv.Itoa(statusCheckTimeout.Delay)
				expDetails.Env["STATUS_CHECK_TIMEOUT"] = strconv.Itoa(statusCheckTimeout.Timeout)
				expDetails.StatusCheckTimeout = statusCheckTimeout.Timeout
			} else {
				expDetails.Env["STATUS_CHECK_DELAY"] = "2"
				expDetails.Env["STATUS_CHECK_TIMEOUT"] = "180"
				expDetails.StatusCheckTimeout = 180
			}

		}
	}
	// set the job cleanup policy
	expDetails.Env["JOB_CLEANUP_POLICY"] = string(engineSpec.Spec.JobCleanUpPolicy)
	return nil
}
