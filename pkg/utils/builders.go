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
		log.Info(volumeMounts)
		containerSpec.WithVolumeMountsNew(volumeMounts)
	}

	return containerSpec

}

// DeployJob the Job using all the details gathered
func DeployJob(perExperiment ExperimentDetails, engineDetails EngineDetails, envVar []corev1.EnvVar, volumeMounts []corev1.VolumeMount, volumeBuilders []*volume.Builder) error {

	// Will build a PodSpecTemplate
	// For creating the spec.template of the Job
	pod := BuildPodTemplateSpec(perExperiment, engineDetails, envVar, volumeMounts, volumeBuilders)
	jobspec := BuildJobSpec(pod)

	// Generation of ClientSet for creation
	clientSet, _, err := GenerateClientSets(engineDetails.Config)
	if err != nil {
		log.Info("Unable to generate ClientSet while Creating Job")
		return err
	}

	jobsClient := clientSet.BatchV1().Jobs(engineDetails.AppNamespace)

	job, err := BuildJob(pod, perExperiment, engineDetails, jobspec)
	if err != nil {
		log.Info("Unable to build Job")
		return err
	}

	// Creating the Job
	//log.Infoln("Printing the Job Object : ", job)
	_, err = jobsClient.Create(job)
	if err != nil {
		log.Info("Unable to create the Job with the clientSet : ", err)
	}
	return nil
}

// BuildPodTemplateSpec return a PodTempplateSpec
func BuildPodTemplateSpec(perExperiment ExperimentDetails, engineDetails EngineDetails, envVar []corev1.EnvVar) *podtemplatespec.Builder {
	podtemplate := podtemplatespec.NewBuilder().
		WithName(perExperiment.JobName).
		WithNamespace(engineDetails.AppNamespace).
		WithLabels(perExperiment.ExpLabels).
		WithServiceAccountName(engineDetails.SvcAccount).
		WithRestartPolicy(corev1.RestartPolicyOnFailure)
	return podtemplate
}

// BuildJobSpec returns a JobSpec
func BuildJobSpec(pod *podtemplatespec.Builder) *jobspec.Builder {
	jobSpecObj := jobspec.NewBuilder().
		WithPodTemplateSpecBuilder(pod)
	_, err := jobSpecObj.Build()
	if err != nil {
		log.Info(err)
	}
	return jobSpecObj
}

// BuildJob will build the JobObject (*batchv1.Job) for creation
func BuildJob(pod *podtemplatespec.Builder, perExperiment ExperimentDetails, engineDetails EngineDetails, jobspec *jobspec.Builder) (*batchv1.Job, error) {
	//restartPolicy := corev1.RestartPolicyOnFailure
	jobObj, err := job.NewBuilder().
		WithJobSpecBuilder(jobspec).
		WithName(perExperiment.JobName).
		WithNamespace(engineDetails.AppNamespace).
		WithLabels(perExperiment.ExpLabels).
		Build()
	if err != nil {
		log.Info(err)
		return jobObj, err
	}
	return jobObj, nil
}
