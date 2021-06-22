package utils

import (
	"reflect"

	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/litmuschaos/elves/kubernetes/container"
	"github.com/litmuschaos/elves/kubernetes/job"
	"github.com/litmuschaos/elves/kubernetes/jobspec"
	"github.com/litmuschaos/elves/kubernetes/podtemplatespec"
)

// PodTemplateSpec is struct for creating the *core1.PodTemplateSpec
type PodTemplateSpec struct {
	Object *corev1.PodTemplateSpec
}

// BuildContainerSpec builds a Container with following properties
func buildContainerSpec(experiment *ExperimentDetails, envVars []corev1.EnvVar) (*container.Builder, error) {
	containerSpec := container.NewBuilder().
		WithName(experiment.JobName).
		WithImage(experiment.ExpImage).
		WithCommandNew(experiment.ExpCommand).
		WithArgumentsNew(experiment.ExpArgs).
		WithImagePullPolicy(experiment.ExpImagePullPolicy).
		WithEnvsNew(envVars)

	if !reflect.DeepEqual(experiment.SecurityContext.ContainerSecurityContext, corev1.SecurityContext{}) {

		containerSpec.WithSecurityContext(experiment.SecurityContext.ContainerSecurityContext)

	}

	if !reflect.DeepEqual(experiment.ResourceRequirements, corev1.ResourceRequirements{}) {

		containerSpec.WithResourceRequirements(experiment.ResourceRequirements)

	}

	if experiment.VolumeOpts.VolumeMounts != nil {
		containerSpec.WithVolumeMountsNew(experiment.VolumeOpts.VolumeMounts)
	}

	_, err := containerSpec.Build()

	if err != nil {
		return nil, err
	}

	return containerSpec, err

}

func getEnvFromMap(m map[string]corev1.EnvVar) []corev1.EnvVar {
	var envVars []corev1.EnvVar
	for _, v := range m {
		envVars = append(envVars, v)
	}

	// Add env for getting pod name using downward API
	envVars = append(envVars, corev1.EnvVar{
		Name: "POD_NAME",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.name",
			},
		},
	})

	return envVars
}

// BuildingAndLaunchJob builds Job, and then launch it.
func BuildingAndLaunchJob(experiment *ExperimentDetails, clients ClientSets) error {
	experiment.VolumeOpts.VolumeOperations(experiment)

	envVars := getEnvFromMap(experiment.envMap)
	//Build Container to add in the Pod
	containerForPod, err := buildContainerSpec(experiment, envVars)
	if err != nil {
		return errors.Errorf("unable to build Container for Chaos Experiment, error: %v", err)
	}
	// Will build a PodSpecTemplate
	pod, err := buildPodTemplateSpec(experiment, containerForPod)
	if err != nil {

		return errors.Errorf("unable to build PodTemplateSpec for Chaos Experiment, error: %v", err)
	}
	// Build JobSpec Template
	jobspec, err := buildJobSpec(pod)
	if err != nil {
		return errors.Errorf("unable to build JobSpec for Chaos Experiment, error: %v", err)
	}
	//Build Job
	job, err := experiment.buildJob(jobspec)
	if err != nil {
		return errors.Errorf("unable to Build ChaosExperiment Job, error: %v", err)
	}
	// Creating the Job
	if err = experiment.launchJob(job, clients); err != nil {
		return errors.Errorf("unable to launch ChaosExperiment Job, error: %v", err)
	}
	return nil
}

// launchJob spawn a kubernetes Job using the job Object received.
func (expDetails *ExperimentDetails) launchJob(job *batchv1.Job, clients ClientSets) error {
	_, err := clients.KubeClient.BatchV1().Jobs(expDetails.Namespace).Create(job)
	return err
}

// BuildPodTemplateSpec return a PodTempplateSpec
func buildPodTemplateSpec(experiment *ExperimentDetails, containerForPod *container.Builder) (*podtemplatespec.Builder, error) {
	podtemplate := podtemplatespec.NewBuilder().
		WithName(experiment.JobName).
		WithNamespace(experiment.Namespace).
		WithLabels(experiment.ExpLabels).
		WithServiceAccountName(experiment.SvcAccount).
		WithRestartPolicy(corev1.RestartPolicyNever).
		WithVolumeBuilders(experiment.VolumeOpts.VolumeBuilders).
		WithAnnotations(experiment.Annotations).
		WithContainerBuildersNew(containerForPod)

	if experiment.TerminationGracePeriodSeconds != 0 {
		podtemplate.WithTerminationGracePeriodSeconds(experiment.TerminationGracePeriodSeconds)
	}

	if !reflect.DeepEqual(experiment.SecurityContext.PodSecurityContext, corev1.PodSecurityContext{}) {

		podtemplate.WithSecurityContext(experiment.SecurityContext.PodSecurityContext)

	}

	if experiment.HostPID {
		podtemplate.WithHostPID(experiment.HostPID)
	}

	if experiment.ImagePullSecrets != nil {
		podtemplate.WithImagePullSecrets(experiment.ImagePullSecrets)
	}

	if len(experiment.NodeSelector) != 0 {
		podtemplate.WithNodeSelector(experiment.NodeSelector)
	}

	if experiment.Tolerations != nil {
		podtemplate.WithTolerations(experiment.Tolerations...)
	}

	if _, err := podtemplate.Build(); err != nil {
		return nil, err
	}
	return podtemplate, nil
}

// BuildJobSpec returns a JobSpec
func buildJobSpec(pod *podtemplatespec.Builder) (*jobspec.Builder, error) {
	jobSpecObj := jobspec.NewBuilder().
		WithPodTemplateSpecBuilder(pod)
	_, err := jobSpecObj.Build()
	if err != nil {
		return nil, err
	}
	return jobSpecObj, nil
}

// BuildJob will build the JobObject for creation
func (expDetails *ExperimentDetails) buildJob(jobspec *jobspec.Builder) (*batchv1.Job, error) {
	jobObj, err := job.NewBuilder().
		WithJobSpecBuilder(jobspec).
		WithAnnotations(expDetails.Annotations).
		WithName(expDetails.JobName).
		WithNamespace(expDetails.Namespace).
		WithLabels(expDetails.ExpLabels).
		Build()
	return jobObj, err
}
