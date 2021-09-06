package utils

import (
	"reflect"
	"strconv"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	litmuschaosv1alpha1 "github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
)

// SetInstanceAttributeValuesFromChaosEngine set the value from the chaosengine
func (expDetails *ExperimentDetails) SetInstanceAttributeValuesFromChaosEngine(engine *EngineDetails, clients ClientSets) error {
	chaosEngine, err := engine.GetChaosEngine(clients)
	if err != nil {
		return errors.Errorf("unable to get chaosEngine in namespace: %s", engine.EngineNamespace)
	}
	// fetch all the values from chaosengine and set into expDetails struct
	expDetails.SetExpAnnotationFromEngine(chaosEngine).
		SetExpNodeSelectorFromEngine(chaosEngine).
		SetResourceRequirementsFromEngine(chaosEngine).
		SetImagePullSecretsFromEngine(chaosEngine).
		SetTolerationsFromEngine(chaosEngine).
		SetExpImageFromEngine(chaosEngine).
		SetTerminationGracePeriodSecondsFromEngine(chaosEngine).
		SetDefaultAppHealthCheck(chaosEngine)

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

// SetTerminationGracePeriodSecondsFromEngine set the terminationGracePeriodSeconds for experiment pod provided in the chaosEngine
func (expDetails *ExperimentDetails) SetTerminationGracePeriodSecondsFromEngine(engine *litmuschaosv1alpha1.ChaosEngine) *ExperimentDetails {
	expDetails.TerminationGracePeriodSeconds = engine.Spec.TerminationGracePeriodSeconds
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

// SetDefaultAppHealthCheck sets th default health checks provided inside the chaosEngine
func (expDetails *ExperimentDetails) SetDefaultAppHealthCheck(engine *litmuschaosv1alpha1.ChaosEngine) *ExperimentDetails {
	expDetails.DefaultAppHealthCheck = engine.Spec.DefaultAppHealthCheck
	return expDetails
}

// SetOverrideEnvFromChaosEngine override the default envs with envs passed inside the chaosengine
func (expDetails *ExperimentDetails) SetOverrideEnvFromChaosEngine(engineName string, clients ClientSets) error {

	engineSpec, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(expDetails.Namespace).Get(engineName, metav1.GetOptions{})
	if err != nil {
		return errors.Errorf("unable to get ChaosEngine Resource in namespace: %v", expDetails.Namespace)
	}
	expList := engineSpec.Spec.Experiments
	for _, exp := range expList {
		if exp.Name == expDetails.Name {
			envVars := exp.Spec.Components.ENV
			for _, env := range envVars {
				expDetails.envMap[env.Name] = env
				// extracting and storing instance id explicitly
				// as we need this variable while generating chaos-result name
				if env.Name == "INSTANCE_ID" {
					expDetails.InstanceID = env.Value
				}
			}

			delay, timeout := 2, 180
			if sc := exp.Spec.Components.StatusCheckTimeouts; !reflect.DeepEqual(sc, litmuschaosv1alpha1.StatusCheckTimeout{}) {
				delay = sc.Delay
				timeout = sc.Timeout
			}

			expDetails.envMap["STATUS_CHECK_DELAY"] = v1.EnvVar{
				Name:  "STATUS_CHECK_DELAY",
				Value: strconv.Itoa(delay),
			}
			expDetails.envMap["STATUS_CHECK_TIMEOUT"] = v1.EnvVar{
				Name:  "STATUS_CHECK_TIMEOUT",
				Value: strconv.Itoa(timeout),
			}
			expDetails.StatusCheckTimeout = timeout
		}
	}

	// set the job cleanup policy
	expDetails.envMap["JOB_CLEANUP_POLICY"] = v1.EnvVar{
		Name:  "JOB_CLEANUP_POLICY",
		Value: string(engineSpec.Spec.JobCleanUpPolicy),
	}

	return nil
}
