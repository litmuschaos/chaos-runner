package utils

import (
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

//SetValueFromChaosExperiment sets value in experimentDetails struct from chaosExperiment
func (expDetails *ExperimentDetails) SetValueFromChaosExperiment(clients ClientSets, engine *EngineDetails) error {
	if err := expDetails.SetImage(clients); err != nil {
		return err
	}
	if err := expDetails.SetArgs(clients); err != nil {
		return err
	}
	if err := expDetails.SetLabels(engine, clients); err != nil {
		return err
	}
	// Generation of Random String for appending it into Job
	randomString := RandomString()
	// Setting the JobName in Experiment Realted struct
	expDetails.JobName = expDetails.Name + "-" + randomString
	return nil
}

//SetENV sets ENV values in experimentDetails struct.
func (expDetails *ExperimentDetails) SetENV(engineDetails EngineDetails, clients ClientSets) error {
	// Get the Default ENV's from ChaosExperiment
	klog.V(0).Infof("Getting the ENV Variables")
	if err := expDetails.SetDefaultEnv(clients); err != nil {
		return err
	}

	// OverWriting the Defaults Varibles from the ChaosEngine ENV
	if err := expDetails.SetEnvFromEngine(engineDetails.Name, clients); err != nil {
		return err
	}
	// Store ENV in a map
	ENVList := map[string]string{"CHAOSENGINE": engineDetails.Name, "APP_LABEL": engineDetails.AppLabel, "APP_NAMESPACE": engineDetails.AppNamespace, "APP_KIND": engineDetails.AppKind, "AUXILIARY_APPINFO": engineDetails.AuxiliaryAppInfo, "CHAOS_UID": engineDetails.UID}
	// Adding some addition ENV's from spec.AppInfo of ChaosEngine
	for key, value := range ENVList {
		expDetails.Env[key] = value
	}
	return nil
}

//SetValueFromChaosEngine sets value in experimentDetails struct from chaosEngine
func (expDetails *ExperimentDetails) SetValueFromChaosEngine(engineDetails *EngineDetails, i int, clients ClientSets) error {
	expDetails.Name = engineDetails.Experiments[i]
	expDetails.Namespace = engineDetails.AppNamespace
	expDetails.SvcAccount = engineDetails.SvcAccount
	// Get engineUID from the chaos-runner's label
	if err := SetEngineUID(engineDetails, clients); err != nil {
		return err
	}
	return nil
}

// NewExperimentDetails initilizes the structure
func NewExperimentDetails() *ExperimentDetails {
	var experimentDetails ExperimentDetails
	experimentDetails.Env = make(map[string]string)
	experimentDetails.ExpLabels = make(map[string]string)
	return &experimentDetails
}

// HandleChaosExperimentExistence will check the experiment in the app namespace
func (expDetails *ExperimentDetails) HandleChaosExperimentExistence(engineDetails EngineDetails, clients ClientSets) error {

	_, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		if err := engineDetails.ExperimentNotFoundPatchEngine(expDetails, clients); err != nil {
			return errors.Wrapf(err, "Unable to patch Chaos Engine Name: %v, in namespace: %v, due to error: %v", engineDetails.Name, engineDetails.AppNamespace, err)
		}
		return errors.Wrapf(err, "Unable to list Chaos Experiment Name: %v,in Namespace: %v, due to error: %v", expDetails.Name, expDetails.Namespace, err)
	}

	return nil
}

// SetDefaultEnv sets the Env's in Experiment Structure
func (expDetails *ExperimentDetails) SetDefaultEnv(clients ClientSets) error {
	experimentEnv, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "Unable to get the Default ENV from ChaosExperiment, error : %v", err)
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
		return errors.Wrapf(err, "Unable to get ChaosEngine Resource in namespace: %v", expDetails.Namespace)
	}
	envList := engineSpec.Spec.Experiments
	for i := range envList {
		if envList[i].Name == expDetails.Name {
			keyValue := envList[i].Spec.Components.ENV
			for j := range keyValue {
				expDetails.Env[keyValue[j].Name] = keyValue[j].Value
			}
		}
	}
	return nil
}

// SetLabels sets the Experiment Labels, in Experiment Structure
func (expDetails *ExperimentDetails) SetLabels(engine *EngineDetails, clients ClientSets) error {
	expirementSpec, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "Unable to get ChaosExperiment instance in namespace: %v", expDetails.Namespace)
	}
	expDetails.ExpLabels = expirementSpec.Spec.Definition.Labels
	expDetails.ExpLabels["chaosUID"] = engine.UID
	return nil
}

// SetImage sets the Experiment Image, in Experiment Structure
func (expDetails *ExperimentDetails) SetImage(clients ClientSets) error {
	expirementSpec, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "Unable to get ChaosExperiment instance in namespace: %v", expDetails.Namespace)
	}
	expDetails.ExpImage = expirementSpec.Spec.Definition.Image
	return nil
}

// SetArgs sets the Experiment Image, in Experiment Structure
func (expDetails *ExperimentDetails) SetArgs(clients ClientSets) error {
	expirementSpec, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "Unable to get ChaosExperiment instance in namespace: %v", expDetails.Namespace)
	}
	expDetails.ExpArgs = expirementSpec.Spec.Definition.Args
	return nil
}

// SetEngineUID fetch the engineUID from chaos-runner
func SetEngineUID(engine *EngineDetails, clients ClientSets) error {
	runnerName := engine.Name + "-runner"
	runnerSpec, err := clients.KubeClient.CoreV1().Pods(engine.AppNamespace).Get(runnerName, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "Unable to get runner pod in namespace: %v", engine.AppNamespace)
	}
	engineUID := runnerSpec.Labels["engineUID"]
	if engineUID != "" {
		engine.UID = engineUID
	} else {
		return errors.Wrapf(err, "Unable to get ChaosEngine UID, due to error: %v", err)
	}
	return nil
}
