package utils

import (
	litmuschaosv1alpha1 "github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateExperimentList make the list of all experiment, provided inside chaosengine
func (engineDetails *EngineDetails) CreateExperimentList() []ExperimentDetails {
	var ExperimentDetailsList []ExperimentDetails
	for i := range engineDetails.Experiments {
		ExperimentDetailsList = append(ExperimentDetailsList, engineDetails.NewExperimentDetails(i))
	}
	return ExperimentDetailsList
}

// NewExperimentDetails create and initialize the experimentDetails
func (engineDetails *EngineDetails) NewExperimentDetails(i int) ExperimentDetails {
	var experimentDetails ExperimentDetails
	experimentDetails.envMap = make(map[string]v1.EnvVar)
	experimentDetails.ExpLabels = make(map[string]string)

	// set the initial values from the EngineDetails struct
	experimentDetails.Name = engineDetails.Experiments[i]
	experimentDetails.SvcAccount = engineDetails.SvcAccount
	experimentDetails.Namespace = engineDetails.EngineNamespace
	// Setting the JobName in Experiment related struct
	experimentDetails.JobName = experimentDetails.Name + "-" + RandomString()
	return experimentDetails
}

// SetDefaultEnvFromChaosExperiment sets the Env's in Experiment Structure
func (expDetails *ExperimentDetails) SetDefaultEnvFromChaosExperiment(clients ClientSets) error {
	experimentEnv, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		return errors.Errorf("unable to get the %v ChaosExperiment in %v namespace, error: %v", expDetails.Name, expDetails.Namespace, err)
	}

	envList := experimentEnv.Spec.Definition.ENVList
	for _, env := range envList {
		expDetails.envMap[env.Name] = env
	}

	return nil
}

// SetValueFromChaosResources fetches required values from various Chaos Resources
func (expDetails *ExperimentDetails) SetValueFromChaosResources(engineDetails *EngineDetails, clients ClientSets) error {

	if err := expDetails.SetDefaultAttributeValuesFromChaosExperiment(clients, engineDetails); err != nil {
		return errors.Errorf("unable to set value from Chaos Experiment, error: %v", err)
	}
	if err := expDetails.SetInstanceAttributeValuesFromChaosEngine(engineDetails, clients); err != nil {
		return errors.Errorf("unable to set value from Chaos Engine, error: %v", err)
	}
	return nil
}

// HandleChaosExperimentExistence will check the experiment in the app namespace
func (expDetails *ExperimentDetails) HandleChaosExperimentExistence(engineDetails EngineDetails, clients ClientSets) error {

	_, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		if err := engineDetails.ExperimentNotFoundPatchEngine(expDetails, clients); err != nil {
			return errors.Errorf("unable to patch Chaos Engine Name: %v, namespace: %v, error: %v", engineDetails.Name, engineDetails.EngineNamespace, err)
		}
		return errors.Errorf("unable to list Chaos Experiment Name: %v,in Namespace: %v, error: %v", expDetails.Name, expDetails.Namespace, err)
	}

	return nil
}

//SetDefaultAttributeValuesFromChaosExperiment sets value in experimentDetails struct from chaosExperiment
func (expDetails *ExperimentDetails) SetDefaultAttributeValuesFromChaosExperiment(clients ClientSets, engine *EngineDetails) error {

	experimentSpec, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		return errors.Errorf("unable to get %v ChaosExperiment instance in namespace: %v", expDetails.Name, expDetails.Namespace)
	}

	// fetch all the values from chaosexperiment and set into expDetails struct
	expDetails.SetImage(experimentSpec).
		SetImagePullPolicy(experimentSpec).
		SetArgs(experimentSpec).
		SetCommand(experimentSpec).
		SetLabels(experimentSpec, engine).
		SetSecurityContext(experimentSpec).
		SetHostPID(experimentSpec)

	return nil
}

// SetLabels sets the Experiment Labels, in Experiment Structure
func (expDetails *ExperimentDetails) SetLabels(experimentSpec *litmuschaosv1alpha1.ChaosExperiment, engine *EngineDetails) *ExperimentDetails {
	expDetails.ExpLabels = experimentSpec.Spec.Definition.Labels
	expDetails.ExpLabels["chaosUID"] = engine.UID
	return expDetails
}

// SetImage sets the Experiment Image, in Experiment Structure
func (expDetails *ExperimentDetails) SetImage(experimentSpec *litmuschaosv1alpha1.ChaosExperiment) *ExperimentDetails {
	expDetails.ExpImage = experimentSpec.Spec.Definition.Image
	return expDetails
}

// SetImagePullPolicy sets the Experiment ImagePullPolicy, in Experiment Structure
func (expDetails *ExperimentDetails) SetImagePullPolicy(experimentSpec *litmuschaosv1alpha1.ChaosExperiment) *ExperimentDetails {
	if experimentSpec.Spec.Definition.ImagePullPolicy == "" {
		expDetails.ExpImagePullPolicy = DefaultExpImagePullPolicy
	} else {
		expDetails.ExpImagePullPolicy = experimentSpec.Spec.Definition.ImagePullPolicy
	}
	return expDetails
}

// SetArgs sets the Experiment Image, in Experiment Structure
func (expDetails *ExperimentDetails) SetArgs(experimentSpec *litmuschaosv1alpha1.ChaosExperiment) *ExperimentDetails {
	expDetails.ExpArgs = experimentSpec.Spec.Definition.Args
	return expDetails
}

// SetCommand to execute inside the experiment image.
func (expDetails *ExperimentDetails) SetCommand(experimentSpec *litmuschaosv1alpha1.ChaosExperiment) *ExperimentDetails {
	expDetails.ExpCommand = experimentSpec.Spec.Definition.Command
	return expDetails
}

// SetSecurityContext sets the security context, in Experiment Structure
func (expDetails *ExperimentDetails) SetSecurityContext(experimentSpec *litmuschaosv1alpha1.ChaosExperiment) *ExperimentDetails {
	expDetails.SecurityContext = experimentSpec.Spec.Definition.SecurityContext
	return expDetails
}

// SetHostPID sets the hostPID, in Experiment Structure
func (expDetails *ExperimentDetails) SetHostPID(experimentSpec *litmuschaosv1alpha1.ChaosExperiment) *ExperimentDetails {
	expDetails.HostPID = experimentSpec.Spec.Definition.HostPID
	return expDetails
}
