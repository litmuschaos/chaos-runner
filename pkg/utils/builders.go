package utils

import (
	"github.com/litmuschaos/kube-helper/kubernetes/container"
	"github.com/litmuschaos/kube-helper/kubernetes/job"
	jobspec "github.com/litmuschaos/kube-helper/kubernetes/jobspec"
	"github.com/litmuschaos/kube-helper/kubernetes/podtemplatespec"
	volume "github.com/litmuschaos/kube-helper/kubernetes/volume/v1alpha1"
	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodTemplateSpec is struct for creating the *core1.PodTemplateSpec
type PodTemplateSpec struct {
	Object *corev1.PodTemplateSpec
}

// Builder struct for getting the error as well with the template
type Builder struct {
	podtemplatespec *PodTemplateSpec
	errs            []error
}

// BuildContainerSpec builds a Container with following properties
func BuildContainerSpec(perExperiment ExperimentDetails, engineDetails EngineDetails, envVar []corev1.EnvVar, volumeMounts []corev1.VolumeMount) *container.Builder {
	containerSpec := container.NewBuilder().
		WithName(perExperiment.JobName).
		WithImage(perExperiment.ExpImage).
		WithCommandNew([]string{"/bin/bash"}).
		WithArgumentsNew(perExperiment.ExpArgs).
		WithImagePullPolicy("Always").
		//WithVolumeMountsNew(volumeMounts).
		WithEnvsNew(envVar)

	if volumeMounts != nil {
		log.Info("Building Container with VolumeMounts")
		//log.Info(volumeMounts)
		containerSpec.WithVolumeMountsNew(volumeMounts)
	}

	_, err := containerSpec.Build()

	if err != nil {
		log.Info(err)
	}
	return containerSpec

}

// DeployJob the Job using all the details gathered
func DeployJob(perExperiment ExperimentDetails, engineDetails EngineDetails, envVar []corev1.EnvVar, volumeMounts []corev1.VolumeMount, volumeBuilders []*volume.Builder) error {

	// Will build a PodSpecTemplate
	// For creating the spec.template of the Job
	pod := BuildPodTemplateSpec(perExperiment, engineDetails, volumeBuilders)

	//Build Container to add in the Pod
	containerForPod := BuildContainerSpec(perExperiment, engineDetails, envVar, volumeMounts)
	pod.WithContainerBuildersNew(containerForPod)

	// Build JobSpec Template
	jobspec := BuildJobSpec(pod)

	job, err := BuildJob(pod, perExperiment, engineDetails, jobspec)
	//log.Infof("%+v\n", job)
	if err != nil {
		log.Info("Unable to build Job")
		return err
	}

	// Creating the Job
	//log.Infoln("Printing the Job Object : ", job)
	_, err = engineDetails.Clients.KubeClient.BatchV1().Jobs(engineDetails.AppNamespace).Create(job)
	if err != nil {
		log.Info("Unable to create the Job with the clientSet : ", err)
	}
	return nil
}

// BuildPodTemplateSpec return a PodTempplateSpec
func BuildPodTemplateSpec(perExperiment ExperimentDetails, engineDetails EngineDetails, volumeBuilders []*volume.Builder) *podtemplatespec.Builder {
	podtemplate := podtemplatespec.NewBuilder().
		WithName(perExperiment.JobName).
		WithNamespace(engineDetails.AppNamespace).
		WithLabels(perExperiment.ExpLabels).
		WithServiceAccountName(engineDetails.SvcAccount).
		WithRestartPolicy(corev1.RestartPolicyOnFailure)

	// Add VolumeBuilders, if exists
	if volumeBuilders != nil {
		log.Info("Building Pod with VolumeBuilders")
		//log.Info(volumeBuilders)
		podtemplate.WithVolumeBuilders(volumeBuilders)
	}

	_, err := podtemplate.Build()

	if err != nil {
		log.Info(err)
	}
	return podtemplate
}

// BuildJobSpec returns a JobSpec
func BuildJobSpec(pod *podtemplatespec.Builder) *jobspec.Builder {
	jobSpecObj := jobspec.NewBuilder().
		WithPodTemplateSpecBuilder(pod)
	_, err := jobSpecObj.Build()
	if err != nil {
		log.Errorln(err)
	}
	return jobSpecObj
}

func GetOwnerPod(engineDetails EngineDetails) *corev1.Pod {
	pod, err := engineDetails.Clients.KubeClient.CoreV1().Pods(engineDetails.AppNamespace).Get(engineDetails.Name+"-runner", metav1.GetOptions{})
	log.Infof("Errors in Getting Owner/Runner Pod : %v", err)
	return pod

}

func GetPodOwnerRef(ownerPod *corev1.Pod) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(ownerPod,
			corev1.SchemeGroupVersion.WithKind("Pod")),
	}

}

// BuildJob will build the JobObject (*batchv1.Job) for creation
func BuildJob(pod *podtemplatespec.Builder, perExperiment ExperimentDetails, engineDetails EngineDetails, jobspec *jobspec.Builder) (*batchv1.Job, error) {
	//restartPolicy := corev1.RestartPolicyOnFailure
	jobObj, err := job.NewBuilder().
		WithJobSpecBuilder(jobspec).
		WithName(perExperiment.JobName).
		WithNamespace(engineDetails.AppNamespace).
		WithLabels(perExperiment.ExpLabels).
		WithOwnerReferenceNew(GetPodOwnerRef(GetOwnerPod(engineDetails))).
		//WithOwnerReferenceNew(GetOwnerPod(engineDetails)).
		Build()
	if err != nil {
		log.Errorln(err)
		return jobObj, err
	}
	log.Info(jobObj)
	return jobObj, nil
}
