package utils

import (
	clientV1alpha1 "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func getLabels(appns string, chaosExperiment string, litmusClientSet *clientV1alpha1.Clientset) map[string]string {
	expirementSpec, err := litmusClientSet.LitmuschaosV1alpha1().ChaosExperiments(appns).Get(chaosExperiment, metav1.GetOptions{})
	if err != nil {
		log.Infoln(err)
	}
	return expirementSpec.Spec.Definition.Labels

}
func getImage(appns string, chaosExperiment string, litmusClientSet *clientV1alpha1.Clientset) string {
	expirementSpec, err := litmusClientSet.LitmuschaosV1alpha1().ChaosExperiments(appns).Get(chaosExperiment, metav1.GetOptions{})
	if err != nil {
		log.Infoln(err)
	}
	image := expirementSpec.Spec.Definition.Image
	return image
}
func getArgs(appns string, chaosExperiment string, litmusClientSet *clientV1alpha1.Clientset) []string {
	expirementSpec, err := litmusClientSet.LitmuschaosV1alpha1().ChaosExperiments(appns).Get(chaosExperiment, metav1.GetOptions{})
	if err != nil {
		log.Infoln(err)
	}
	args := expirementSpec.Spec.Definition.Args
	return args
}

// GetDetails will use the above functions to get all the necessary details needed for Job Execution
func GetDetails(appns string, chaosExperiment string, config *rest.Config) (map[string]string, string, []string) {
	_, litmusClientSet, err := GenerateClientSets(config)
	if err != nil {
		log.Info(err)
	}
	labels := getLabels(appns, chaosExperiment, litmusClientSet)
	image := getImage(appns, chaosExperiment, litmusClientSet)
	args := getArgs(appns, chaosExperiment, litmusClientSet)
	return labels, image, args

}
