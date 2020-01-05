package utils

import (
	//log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (expStatus *ExperimentStatus) NotFoundExperimentStatus(experimentDetails *ExperimentDetails) {
	expStatus.Name = experimentDetails.JobName
	expStatus.Status = "ChaosExperiment Not Found"
	expStatus.Verdict = "Failed"
	expStatus.LastUpdateTime = metav1.Now()
}

// NotFoundPatchEngine patches the chaosEngine when ChaosExperiment is not Found
func (engineDetails EngineDetails) ExperimentNotFoundPatchEngine(experiment *ExperimentDetails, clients ClientSets) {

	var expStatus ExperimentStatus
	expStatus.NotFoundExperimentStatus(experiment)
	expStatus.PatchChaosEngineStatus(engineDetails, clients)

	/*expEngine, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(engineDetails.AppNamespace).Get(engineDetails.Name, metav1.GetOptions{})
	if err != nil {
		log.Infoln("Could'nt Get the Engine : ", err)
	}
	expEngine.Status.Experiments[i].Status = "Experiment not Found in this Namespace"
	expEngine.Status.Experiments[i].Verdict = "Not Executed"
	log.Info("Patching Engine")
	_, updateErr := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(engineDetails.AppNamespace).Update(expEngine)
	if updateErr != nil {
		log.Infoln("Unable to Patch Engine, Update Error : ", updateErr)
	}*/
}
