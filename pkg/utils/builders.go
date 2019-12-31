package utils

import (
	log "github.com/sirupsen/logrus"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/litmuschaos/kube-helper/kubernetes/container"
	"github.com/litmuschaos/kube-helper/kubernetes/job"
	jobspec "github.com/litmuschaos/kube-helper/kubernetes/jobspec"
	"github.com/litmuschaos/kube-helper/kubernetes/podtemplatespec"
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
func BuildContainerSpec(experiment *ExperimentDetails, envVar []corev1.EnvVar) *container.Builder {
	containerSpec := container.NewBuilder().
		WithName(experiment.JobName).
		WithImage(experiment.ExpImage).
		WithCommandNew([]string{"/bin/bash"}).
		WithArgumentsNew(experiment.ExpArgs).
		WithImagePullPolicy("Always").
		//WithVolumeMountsNew(volumeMounts).
		WithEnvsNew(envVar)

	if experiment.VolumeOpts.VolumeMounts != nil {
		log.Info("Building Container with VolumeMounts")
		//log.Info(volumeMounts)
		containerSpec.WithVolumeMountsNew(experiment.VolumeOpts.VolumeMounts)
	}

	_, err := containerSpec.Build()

	if err != nil {
		log.Info(err)
	}

	return containerSpec

}

func getEnvFromMap(env map[string]string) []corev1.EnvVar {
	var envVar []corev1.EnvVar
	for k, v := range env {
		var perEnv corev1.EnvVar
		perEnv.Name = k
		perEnv.Value = v
		envVar = append(envVar, perEnv)
	}
	return envVar
}

// DeployJob the Job using all the details gathered
func DeployJob(experiment *ExperimentDetails, clients ClientSets) error {

	envVar := getEnvFromMap(experiment.Env)
	// Will build a PodSpecTemplate
	// For creating the spec.template of the Job
	pod := BuildPodTemplateSpec(experiment)

	//Build Container to add in the Pod
	containerForPod := BuildContainerSpec(experiment, envVar)
	pod.WithContainerBuildersNew(containerForPod)

	// Build JobSpec Template
	jobspec := BuildJobSpec(pod)

	job, err := experiment.BuildJob(pod, jobspec)
	if err != nil {
		log.Info("Unable to build Job")
		return err
	}

	// Creating the Job
	_, err = clients.KubeClient.BatchV1().Jobs(experiment.Namespace).Create(job)
	if err != nil {
		log.Info("Unable to create the Job with the clientSet : ", err)
	}
	return nil
}

// BuildPodTemplateSpec return a PodTempplateSpec
func BuildPodTemplateSpec(experiment *ExperimentDetails) *podtemplatespec.Builder {
	podtemplate := podtemplatespec.NewBuilder().
		WithName(experiment.JobName).
		WithNamespace(experiment.Namespace).
		WithLabels(experiment.ExpLabels).
		WithServiceAccountName(experiment.SvcAccount).
		WithRestartPolicy(corev1.RestartPolicyOnFailure)

	// Add VolumeBuilders, if exists
	if experiment.VolumeOpts.VolumeBuilders != nil {
		log.Info("Building Pod with VolumeBuilders")
		//log.Info(volumeBuilders)
		podtemplate.WithVolumeBuilders(experiment.VolumeOpts.VolumeBuilders)
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

// BuildJob will build the JobObject for creation
func (experiment ExperimentDetails) BuildJob(pod *podtemplatespec.Builder, jobspec *jobspec.Builder) (*batchv1.Job, error) {
	//restartPolicy := corev1.RestartPolicyOnFailure
	jobObj, err := job.NewBuilder().
		WithJobSpecBuilder(jobspec).
		WithName(experiment.JobName).
		WithNamespace(experiment.Namespace).
		WithLabels(experiment.ExpLabels).
		Build()
	if err != nil {
		log.Errorln(err)
		return jobObj, err
	}
	return jobObj, nil
}
