package utils

import (
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

var Logger LogStruct

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
		containerSpec.WithVolumeMountsNew(experiment.VolumeOpts.VolumeMounts)
	}

	_, err := containerSpec.Build()

	if err != nil {
		Logger.WithNameSpace(experiment.Namespace).WithResourceName(experiment.JobName).WithString(err.Error()).WithOperation("Build").WithVerbosity(1).WithResourceType("COntainer").Log()
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

// BuildingAndLaunchJob builds Job, and then launch it.
func BuildingAndLaunchJob(experiment *ExperimentDetails, clients ClientSets) error {
	envVar := getEnvFromMap(experiment.Env)
	//Build Container to add in the Pod
	containerForPod := BuildContainerSpec(experiment, envVar)
	// Will build a PodSpecTemplate
	pod := BuildPodTemplateSpec(experiment, containerForPod)
	// Build JobSpec Template
	jobspec := BuildJobSpec(experiment, pod)
	job, err := experiment.BuildJob(pod, jobspec)
	if err != nil {
		Logger.WithNameSpace(experiment.Namespace).WithResourceName(experiment.JobName).WithString(err.Error()).WithOperation("Build").WithVerbosity(1).WithResourceType("Job").Log()
		return err
	}
	// Creating the Job
	if err = experiment.launchJob(job, clients); err != nil {
		return err
	}
	return nil
}

// launchJob spawn a kubernetes Job using the job Object recieved.
func (experiment *ExperimentDetails) launchJob(job *batchv1.Job, clients ClientSets) error {
	_, err := clients.KubeClient.BatchV1().Jobs(experiment.Namespace).Create(job)
	if err != nil {
		Logger.WithNameSpace(experiment.Namespace).WithResourceName(experiment.JobName).WithString(err.Error()).WithOperation("Create").WithVerbosity(1).WithResourceType("Job").Log()
	}
	return nil
}

// BuildPodTemplateSpec return a PodTempplateSpec
func BuildPodTemplateSpec(experiment *ExperimentDetails, containerForPod *container.Builder) *podtemplatespec.Builder {
	podtemplate := podtemplatespec.NewBuilder().
		WithName(experiment.JobName).
		WithNamespace(experiment.Namespace).
		WithLabels(experiment.ExpLabels).
		WithServiceAccountName(experiment.SvcAccount).
		WithRestartPolicy(corev1.RestartPolicyOnFailure).
		WithVolumeBuilders(experiment.VolumeOpts.VolumeBuilders).
		WithContainerBuildersNew(containerForPod)

	if _, err := podtemplate.Build(); err != nil {
		Logger.WithNameSpace(experiment.Namespace).WithResourceName(experiment.JobName).WithString(err.Error()).WithOperation("Build").WithVerbosity(1).WithResourceType("PodTemplate").Log()
		return nil
	}
	return podtemplate
}

// BuildJobSpec returns a JobSpec
func BuildJobSpec(experiment *ExperimentDetails, pod *podtemplatespec.Builder) *jobspec.Builder {
	jobSpecObj := jobspec.NewBuilder().
		WithPodTemplateSpecBuilder(pod)
	_, err := jobSpecObj.Build()
	if err != nil {
		Logger.WithNameSpace(experiment.Namespace).WithResourceName(experiment.JobName).WithString(err.Error()).WithOperation("Build").WithVerbosity(1).WithResourceType("JobSpec").Log()
		return nil
	}
	return jobSpecObj
}

// BuildJob will build the JobObject for creation
func (experiment *ExperimentDetails) BuildJob(pod *podtemplatespec.Builder, jobspec *jobspec.Builder) (*batchv1.Job, error) {
	//restartPolicy := corev1.RestartPolicyOnFailure
	jobObj, err := job.NewBuilder().
		WithJobSpecBuilder(jobspec).
		WithName(experiment.JobName).
		WithNamespace(experiment.Namespace).
		WithLabels(experiment.ExpLabels).
		Build()
	if err != nil {
		Logger.WithNameSpace(experiment.Namespace).WithResourceName(experiment.JobName).WithString(err.Error()).WithOperation("Build").WithVerbosity(1).WithResourceType("Job").Log()
		return jobObj, err
	}
	return jobObj, nil
}
