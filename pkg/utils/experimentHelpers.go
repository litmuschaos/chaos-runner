package utils

import (
	"os"
	"reflect"
	"strconv"

	litmuschaosv1alpha1 "github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	"github.com/litmuschaos/chaos-runner/pkg/log"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateExperimentList make the list of all experiment, provided inside chaosengine
func CreateExperimentList(engineDetails *EngineDetails) []ExperimentDetails {
	var ExperimentDetailsList []ExperimentDetails
	for i := range engineDetails.Experiments {
		ExperimentDetailsList = append(ExperimentDetailsList, NewExperimentDetails(engineDetails, i))
	}
	return ExperimentDetailsList
}

//SetValueFromChaosExperiment sets value in experimentDetails struct from chaosExperiment
func (expDetails *ExperimentDetails) SetValueFromChaosExperiment(clients ClientSets, engine *EngineDetails) error {

	expirementSpec, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		return errors.Errorf("Unable to get ChaosExperiment instance in namespace: %v", expDetails.Namespace)
	}

	if err := expDetails.SetImage(clients, expirementSpec); err != nil {
		return err
	}
	if err := expDetails.SetImagePullPolicy(clients, expirementSpec); err != nil {
		return err
	}
	if err := expDetails.SetArgs(clients, expirementSpec); err != nil {
		return err
	}
	if err := expDetails.SetLabels(engine, clients, expirementSpec); err != nil {
		return err
	}
	if err := expDetails.SetSecurityContext(clients, expirementSpec); err != nil {
		return err
	}
	if err := expDetails.SetHostPID(clients, expirementSpec); err != nil {
		return err
	}
	return nil
}

//SetENV sets ENV values in experimentDetails struct.
func (expDetails *ExperimentDetails) SetENV(engineDetails EngineDetails, clients ClientSets) error {
	// Get the Default ENV's from ChaosExperiment
	log.Info("Getting the ENV Variables")
	if err := expDetails.SetDefaultEnv(clients); err != nil {
		return err
	}

	// OverWriting the Defaults Varibles from the ChaosEngine ENV
	if err := expDetails.SetEnvFromEngine(engineDetails.Name, clients); err != nil {
		return err
	}
	// Store ENV in a map
	ENVList := map[string]string{
		"CHAOSENGINE":       engineDetails.Name,
		"APP_LABEL":         engineDetails.AppLabel,
		"CHAOS_NAMESPACE":   engineDetails.EngineNamespace,
		"APP_NAMESPACE":     os.Getenv("APP_NAMESPACE"),
		"APP_KIND":          engineDetails.AppKind,
		"AUXILIARY_APPINFO": engineDetails.AuxiliaryAppInfo,
		"CHAOS_UID":         engineDetails.UID,
		"EXPERIMENT_NAME":   expDetails.Name,
	}
	// Adding some addition ENV's from spec.AppInfo of ChaosEngine
	for key, value := range ENVList {
		expDetails.Env[key] = value
	}
	return nil
}

// NewExperimentDetails initilizes the structure
func NewExperimentDetails(engineDetails *EngineDetails, i int) ExperimentDetails {
	var experimentDetails ExperimentDetails
	experimentDetails.Env = make(map[string]string)
	experimentDetails.ExpLabels = make(map[string]string)

	// Initial set to values from EngineDetails Struct
	experimentDetails.Name = engineDetails.Experiments[i]
	experimentDetails.SvcAccount = engineDetails.SvcAccount
	experimentDetails.Namespace = os.Getenv("CHAOS_NAMESPACE")

	// Generation of Random String for appending it into Job Name
	randomString := RandomString()
	// Setting the JobName in Experiment Realted struct
	experimentDetails.JobName = experimentDetails.Name + "-" + randomString
	return experimentDetails
}

// HandleChaosExperimentExistence will check the experiment in the app namespace
func (expDetails *ExperimentDetails) HandleChaosExperimentExistence(engineDetails EngineDetails, clients ClientSets) error {

	_, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		if err := engineDetails.ExperimentNotFoundPatchEngine(expDetails, clients); err != nil {
			return errors.Errorf("Unable to patch Chaos Engine Name: %v, namespace: %v, error: %v", engineDetails.Name, engineDetails.EngineNamespace, err)
		}
		return errors.Errorf("Unable to list Chaos Experiment Name: %v,in Namespace: %v, due to error: %v", expDetails.Name, expDetails.Namespace, err)
	}

	return nil
}

// SetDefaultEnv sets the Env's in Experiment Structure
func (expDetails *ExperimentDetails) SetDefaultEnv(clients ClientSets) error {
	experimentEnv, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		return errors.Errorf("Unable to get the Default ENV from ChaosExperiment, error: %v", err)
	}

	envList := experimentEnv.Spec.Definition.ENVList
	expDetails.Env = make(map[string]string)
	for i := range envList {
		key := envList[i].Name
		value := envList[i].Value
		expDetails.Env[key] = value
	}
	return nil
}

// SetEnvFromEngine will over-ride the default variables from one provided in the chaosEngine
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
	return nil
}

// SetLabels sets the Experiment Labels, in Experiment Structure
func (expDetails *ExperimentDetails) SetLabels(engine *EngineDetails, clients ClientSets, expirementSpec *litmuschaosv1alpha1.ChaosExperiment) error {
	expDetails.ExpLabels = expirementSpec.Spec.Definition.Labels
	expDetails.ExpLabels["chaosUID"] = engine.UID
	return nil
}

// SetImage sets the Experiment Image, in Experiment Structure
func (expDetails *ExperimentDetails) SetImage(clients ClientSets, expirementSpec *litmuschaosv1alpha1.ChaosExperiment) error {
	expDetails.ExpImage = expirementSpec.Spec.Definition.Image
	return nil
}

// SetImagePullPolicy sets the Experiment ImagePullPolicy, in Experiment Structure
func (expDetails *ExperimentDetails) SetImagePullPolicy(clients ClientSets, expirementSpec *litmuschaosv1alpha1.ChaosExperiment) error {
	if expirementSpec.Spec.Definition.ImagePullPolicy == "" {
		expDetails.ExpImagePullPolicy = DefaultExpImagePullPolicy
	} else {
		expDetails.ExpImagePullPolicy = expirementSpec.Spec.Definition.ImagePullPolicy
	}
	return nil
}

// SetArgs sets the Experiment Image, in Experiment Structure
func (expDetails *ExperimentDetails) SetArgs(clients ClientSets, expirementSpec *litmuschaosv1alpha1.ChaosExperiment) error {
	expDetails.ExpArgs = expirementSpec.Spec.Definition.Args
	return nil
}

// SetValueFromChaosResources fetches required values from various Chaos Resources
func (expDetails *ExperimentDetails) SetValueFromChaosResources(engineDetails *EngineDetails, clients ClientSets) error {
	if err := expDetails.SetValueFromChaosEngine(engineDetails, clients); err != nil {
		return errors.Errorf("Unable to set value from Chaos Engine, error: %v", err)
	}

	if err := engineDetails.SetValueFromChaosRunner(clients); err != nil {
		return errors.Errorf("Unable to set value from Chaos Runner, error: %v", err)
	}
	if err := expDetails.HandleChaosExperimentExistence(*engineDetails, clients); err != nil {
		return errors.Errorf("Unable to get ChaosExperiment Name: %v, in namespace: %v, error: %v", expDetails.Name, expDetails.Namespace, err)
	}
	if err := expDetails.SetValueFromChaosExperiment(clients, engineDetails); err != nil {
		return errors.Errorf("Unable to set value from Chaos Experiment, error: %v", err)
	}
	if err := expDetails.SetExpImageFromEngine(engineDetails.Name, clients); err != nil {
		return errors.Errorf("Unable to set image from Chaos Engine, error: %v", err)
	}
	return nil
}

// SetValueFromChaosRunner fetch the engineUID from ChaosRunner
func (engine *EngineDetails) SetValueFromChaosRunner(clients ClientSets) error {
	runnerName := engine.Name + "-runner"
	runnerSpec, err := clients.KubeClient.CoreV1().Pods(engine.EngineNamespace).Get(runnerName, metav1.GetOptions{})
	if err != nil {
		return errors.Errorf("Unable to get runner pod in namespace: %v", engine.EngineNamespace)
	}
	chaosUID := runnerSpec.Labels["chaosUID"]
	if chaosUID != "" {
		engine.UID = chaosUID
	} else {
		return errors.Errorf("Unable to get ChaosEngine UID, error: %v", err)
	}
	return nil
}

// SetValueFromChaosEngine set the value from chaosengine
func (expDetails *ExperimentDetails) SetValueFromChaosEngine(engine *EngineDetails, clients ClientSets) error {

	chaosEngine, err := engine.GetChaosEngine(clients)
	if err != nil {
		return errors.Errorf("Unable to get chaosEngine in namespace: %s", engine.EngineNamespace)
	}
	expDetails.Namespace = chaosEngine.Namespace
	if err := expDetails.SetExpAnnotationFromEngine(engine.Name, clients); err != nil {
		return err
	}
	if err := expDetails.SetExpNodeSelectorFromEngine(engine.Name, clients); err != nil {
		return err
	}
	if err := expDetails.SetResourceRequirementsFromEngine(engine.Name, clients); err != nil {
		return err
	}
	if err := expDetails.SetImagePullSecretsFromEngine(engine.Name, clients); err != nil {
		return err
	}
	if err := expDetails.SetTolerationsFromEngine(engine.Name, clients); err != nil {
		return err
	}
	return nil
}

// SetExpAnnotationFromEngine will over-ride the default exp annotation with the one provided in the chaosEngine
func (expDetails *ExperimentDetails) SetExpAnnotationFromEngine(engineName string, clients ClientSets) error {

	engineSpec, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(expDetails.Namespace).Get(engineName, metav1.GetOptions{})
	if err != nil {
		return errors.Errorf("Unable to get ChaosEngine Resource in namespace: %v", expDetails.Namespace)
	}

	expRefList := engineSpec.Spec.Experiments
	for i := range expRefList {
		if expRefList[i].Name == expDetails.Name {
			expDetails.Annotations = expRefList[i].Spec.Components.ExperimentAnnotations
		}
	}
	return nil
}

// SetResourceRequirementsFromEngine will add the resource requirements provided inside chaosengine
func (expDetails *ExperimentDetails) SetResourceRequirementsFromEngine(engineName string, clients ClientSets) error {

	engineSpec, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(expDetails.Namespace).Get(engineName, metav1.GetOptions{})
	if err != nil {
		return errors.Errorf("Unable to get ChaosEngine Resource in namespace: %v", expDetails.Namespace)
	}

	expRefList := engineSpec.Spec.Experiments
	for i := range expRefList {
		if expRefList[i].Name == expDetails.Name {
			expDetails.ResourceRequirements = expRefList[i].Spec.Components.Resources
		}
	}
	return nil
}

// SetImagePullSecretsFromEngine will add the image pull secrets provided inside chaosengine
func (expDetails *ExperimentDetails) SetImagePullSecretsFromEngine(engineName string, clients ClientSets) error {

	engineSpec, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(expDetails.Namespace).Get(engineName, metav1.GetOptions{})
	if err != nil {
		return errors.Errorf("Unable to get ChaosEngine Resource in namespace: %v", expDetails.Namespace)
	}

	expRefList := engineSpec.Spec.Experiments
	for i := range expRefList {
		if expRefList[i].Name == expDetails.Name {
			if expRefList[i].Spec.Components.ExperimentImagePullSecrets != nil {
				expDetails.ImagePullSecrets = expRefList[i].Spec.Components.ExperimentImagePullSecrets
			}
		}
	}
	return nil
}

// SetExpNodeSelectorFromEngine will add the nodeSelector attribute based the key/value provided in the chaosEngine
func (expDetails *ExperimentDetails) SetExpNodeSelectorFromEngine(engineName string, clients ClientSets) error {

	engineSpec, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(expDetails.Namespace).Get(engineName, metav1.GetOptions{})
	if err != nil {
		return errors.Errorf("Unable to get ChaosEngine Resource in namespace: %v", expDetails.Namespace)
	}

	expRefList := engineSpec.Spec.Experiments
	for i := range expRefList {
		if expRefList[i].Name == expDetails.Name {
			expDetails.NodeSelector = expRefList[i].Spec.Components.NodeSelector
		}
	}
	return nil
}

// SetTolerationsFromEngine will add the tolerations based on the key/operator/effect provided in the chaosEngine
func (expDetails *ExperimentDetails) SetTolerationsFromEngine(engineName string, clients ClientSets) error {

	engineSpec, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(expDetails.Namespace).Get(engineName, metav1.GetOptions{})
	if err != nil {
		return errors.Errorf("Unable to get ChaosEngine Resource in namespace: %v", expDetails.Namespace)
	}

	expRefList := engineSpec.Spec.Experiments
	for i := range expRefList {
		if expRefList[i].Name == expDetails.Name {
			expDetails.Tolerations = expRefList[i].Spec.Components.Tolerations
		}
	}
	return nil
}

// SetSecurityContext sets the security context, in Experiment Structure
func (expDetails *ExperimentDetails) SetSecurityContext(clients ClientSets, expirementSpec *litmuschaosv1alpha1.ChaosExperiment) error {

	expDetails.SecurityContext = expirementSpec.Spec.Definition.SecurityContext
	return nil
}

// SetHostPID sets the hostPID, in Experiment Structure
func (expDetails *ExperimentDetails) SetHostPID(clients ClientSets, expirementSpec *litmuschaosv1alpha1.ChaosExperiment) error {
	expDetails.HostPID = expirementSpec.Spec.Definition.HostPID

	return nil
}

// SetExpImageFromEngine will over-ride the default exp image with the one provided in the chaosEngine
func (expDetails *ExperimentDetails) SetExpImageFromEngine(engineName string, clients ClientSets) error {

	engineSpec, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(expDetails.Namespace).Get(engineName, metav1.GetOptions{})
	if err != nil {
		return errors.Errorf("Unable to get ChaosEngine Resource in namespace: %v", expDetails.Namespace)
	}

	expRefList := engineSpec.Spec.Experiments
	for i := range expRefList {
		if expRefList[i].Name == expDetails.Name {

			if expRefList[i].Spec.Components.ExperimentImage != "" {
				expDetails.ExpImage = expRefList[i].Spec.Components.ExperimentImage
			}
		}
	}
	return nil
}
