package utils

import (
	log "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
)

// InitialPatchEngine patches the chaosEngine with the initial ExperimentStatuses
func InitialPatchEngine(engineDetails EngineDetails, clients ClientSets) {

	for i := range engineDetails.Experiments {
		log.Info("Initial Patch for Experiment : ", engineDetails.Experiments[i])
		expName := engineDetails.Experiments[i]
		var currExpStatus v1alpha1.ExperimentStatuses
		currExpStatus.Name = expName
		currExpStatus.Status = "Waiting"
		currExpStatus.Verdict = "Wait for Completion"
		currExpStatus.LastUpdateTime = metav1.Now()

		expEngine, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(engineDetails.AppNamespace).Get(engineDetails.Name, metav1.GetOptions{})
		if err != nil {
			log.Infoln("Could'nt Get the Engine : ", err)
		}
		expEngine.Status.Experiments = append(expEngine.Status.Experiments, currExpStatus)
		log.Info("Patching Engine")
		_, updateErr := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(engineDetails.AppNamespace).Update(expEngine)
		if updateErr != nil {
			log.Infoln("Unable to Patch Engine, Update Error : ", updateErr)
		}
	}
}
