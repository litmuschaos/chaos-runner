package utils

import (
	//"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NotFoundPatchEngine patches the chaosEngine when ChoasExperiment is not Found
func NotFoundPatchEngine(i int, engineDetails EngineDetails) {
	/*_, litmusClient, err := GenerateClientSets(engineDetails.Config)
	if err != nil {
		log.Infoln("Couldn't Create ClientSet")
	}*/
	expEngine, err := engineDetails.Clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(engineDetails.AppNamespace).Get(engineDetails.Name, metav1.GetOptions{})
	if err != nil {
		log.Infoln("Could'nt Get the Engine : ", err)
	}
	expEngine.Status.Experiments[i].Status = "Experiment not Found in this Namespace"
	expEngine.Status.Experiments[i].Verdict = "Not Executed"
	log.Info("Patching Engine")
	_, updateErr := engineDetails.Clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(engineDetails.AppNamespace).Update(expEngine)
	if updateErr != nil {
		log.Infoln("Unable to Patch Engine, Update Error : ", updateErr)
	}
}
