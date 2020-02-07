package utils

import (
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//SetValueFromChaosExperiment sets value in experimentDetails struct from chaosExperiment
func (expDetails *ExperimentDetails) SetValueFromChaosExperiment(clients ClientSets, engine *EngineDetails) error {
	if err := expDetails.SetImage(clients); err != nil {
		return err
	}
	if err := expDetails.SetArgs(clients); err != nil {
		return err
	}
	// Get engineUID from the chaos-runner's label
	if err := SetEngineUID(engine, clients); err != nil {
		return err
	}
	if err := expDetails.SetLabels(engine, clients); err != nil {
		return nil
	}
	// Generation of Random String for appending it into Job
	randomString := RandomString()
	// Setting the JobName in Experiment Realted struct
	expDetails.JobName = expDetails.Name + "-" + randomString
	return nil
}

//SetENV sets ENV values in experimentDetails struct.
func (expDetails *ExperimentDetails) SetENV(engineDetails *EngineDetails, clients ClientSets) error {
	// Get the Default ENV's from ChaosExperiment
	log.Infoln("Getting the ENV Variables")
	if err := expDetails.SetDefaultEnv(clients); err != nil {
		return err
	}
	// OverWriting the Defaults Varibles from the ChaosEngine ENV
	if err := expDetails.SetEnvFromEngine(engineDetails.Name, clients); err != nil {
		return err
	}
	// Adding some addition ENV's from spec.AppInfo of ChaosEngine
	expDetails.Env["CHAOSENGINE"] = engineDetails.Name
	expDetails.Env["APP_LABEL"] = engineDetails.AppLabel
	expDetails.Env["APP_NAMESPACE"] = engineDetails.AppNamespace
	expDetails.Env["APP_KIND"] = engineDetails.AppKind
	expDetails.Env["AUXILIARY_APPINFO"] = engineDetails.AuxiliaryAppInfo
	expDetails.Env["CHAOS_UID"] = engineDetails.UID
	return nil
}

//SetValueFromChaosEngine sets value in experimentDetails struct from chaosEngine
func (expDetails *ExperimentDetails) SetValueFromChaosEngine(engineDetails EngineDetails, i int) {
	expDetails.Name = engineDetails.Experiments[i]
	expDetails.Namespace = engineDetails.AppNamespace
	expDetails.SvcAccount = engineDetails.SvcAccount
}

// NewExperimentDetails initilizes the structure
func NewExperimentDetails() *ExperimentDetails {
	var experimentDetails ExperimentDetails
	experimentDetails.Env = make(map[string]string)
	experimentDetails.ExpLabels = make(map[string]string)
	return &experimentDetails
}

// CheckExistence will check the experiment in the app namespace
func (expDetails *ExperimentDetails) CheckExistence(clients ClientSets) (bool, error) {

	_, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	return true, nil
}

// SetDefaultEnv sets the Env's in Experiment Structure
func (expDetails *ExperimentDetails) SetDefaultEnv(clients ClientSets) error {
	experimentEnv, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		return err
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
		return err
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
		return err
	}
	expDetails.ExpLabels = expirementSpec.Spec.Definition.Labels
	expDetails.ExpLabels["chaosUID"] = engine.UID
	return nil
}

// SetImage sets the Experiment Image, in Experiment Structure
func (expDetails *ExperimentDetails) SetImage(clients ClientSets) error {
	expirementSpec, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	expDetails.ExpImage = expirementSpec.Spec.Definition.Image
	return nil
}

// SetArgs sets the Experiment Image, in Experiment Structure
func (expDetails *ExperimentDetails) SetArgs(clients ClientSets) error {
	expirementSpec, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosExperiments(expDetails.Namespace).Get(expDetails.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	expDetails.ExpArgs = expirementSpec.Spec.Definition.Args
	return nil
}

// SetEngineUID fetch the engineUID from chaos-runner
func SetEngineUID(engine *EngineDetails, clients ClientSets) error {
	runnerName := engine.Name + "-runner"
	runnerSpec, err := clients.KubeClient.CoreV1().Pods(engine.AppNamespace).Get(runnerName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	engineUID := runnerSpec.Labels["engineUID"]
	if engineUID != "" {
		engine.UID = engineUID
	} else {
		log.Infof("Can't find the engineUID")
	}
	return nil
}
