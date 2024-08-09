package utils

import (
	"context"
	"reflect"

	"github.com/litmuschaos/chaos-operator/api/litmuschaos/v1alpha1"
	"github.com/litmuschaos/chaos-runner/pkg/telemetry"
	"github.com/litmuschaos/elves/kubernetes/container"
	"github.com/litmuschaos/elves/kubernetes/job"
	"github.com/litmuschaos/elves/kubernetes/jobspec"
	"github.com/litmuschaos/elves/kubernetes/podtemplatespec"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// buildSideCarSpec builds a Container with following properties
func buildSideCarSpec(experiment *ExperimentDetails) ([]*container.Builder, error) {
	var sidecarContainers []*container.Builder

	for _, sidecar := range experiment.SideCars {
		var volumeOpts VolumeOpts

		if len(sidecar.Secrets) != 0 {
			volumeOpts.NewVolumeMounts().BuildVolumeMountsForSecrets(sidecar.Secrets)
		}

		containerSpec := container.NewBuilder().
			WithName(experiment.JobName + "-sidecar-" + RandomString(6)).
			WithImage(sidecar.Image).
			WithImagePullPolicy(sidecar.ImagePullPolicy).
			WithEnvsNew(sidecar.ENV)

		if !reflect.DeepEqual(experiment.ResourceRequirements, corev1.ResourceRequirements{}) {
			containerSpec.WithResourceRequirements(experiment.ResourceRequirements)
		}

		if volumeOpts.VolumeMounts != nil {
			containerSpec.WithVolumeMountsNew(volumeOpts.VolumeMounts)
		}

		if len(sidecar.EnvFrom) != 0 {
			containerSpec.WithEnvsFrom(sidecar.EnvFrom)
		}

		if _, err := containerSpec.Build(); err != nil {
			return nil, err
		}

		sidecarContainers = append(sidecarContainers, containerSpec)
	}

	return sidecarContainers, err
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
func BuildingAndLaunchJob(ctx context.Context, experiment *ExperimentDetails, clients ClientSets) error {
	ctx, span := otel.Tracer(telemetry.TracerName).Start(ctx, "CreateExperimentJob")
	defer span.End()

	experiment.VolumeOpts.VolumeOperations(experiment)

	envVars := getEnvFromMap(experiment.envMap)
	//Build Container to add in the Pod
	containerForPod, err := buildContainerSpec(experiment, envVars)
	if err != nil {
		return errors.Errorf("unable to build Container for Chaos Experiment, error: %v", err)
	}

	containers := []*container.Builder{containerForPod}

	if len(experiment.SideCars) != 0 {
		sidecars, err := buildSideCarSpec(experiment)
		if err != nil {
			return errors.Errorf("unable to build sidecar Container for Chaos Experiment, error: %v", err)
		}
		containers = append(containers, sidecars...)
	}

	// Will build a PodSpecTemplate
	pod, err := buildPodTemplateSpec(experiment, containers...)
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
	_, err := clients.KubeClient.BatchV1().Jobs(expDetails.Namespace).Create(context.Background(), job, v1.CreateOptions{})
	return err
}

// BuildPodTemplateSpec return a PodTemplateSpec
func buildPodTemplateSpec(experiment *ExperimentDetails, containers ...*container.Builder) (*podtemplatespec.Builder, error) {
	podtemplate := podtemplatespec.NewBuilder().
		WithName(experiment.JobName).
		WithNamespace(experiment.Namespace).
		WithLabels(experiment.ExpLabels).
		WithServiceAccountName(experiment.SvcAccount).
		WithRestartPolicy(corev1.RestartPolicyNever).
		WithVolumeBuilders(experiment.VolumeOpts.VolumeBuilders).
		WithAnnotations(experiment.Annotations).
		WithContainerBuildersNew(containers...)

	if experiment.TerminationGracePeriodSeconds != 0 {
		podtemplate.WithTerminationGracePeriodSeconds(experiment.TerminationGracePeriodSeconds)
	}

	if !reflect.DeepEqual(experiment.SecurityContext.PodSecurityContext, corev1.PodSecurityContext{}) {

		podtemplate.WithSecurityContext(experiment.SecurityContext.PodSecurityContext)

	}

	if len(experiment.SideCars) != 0 {
		secrets := setSidecarSecrets(experiment)
		if len(secrets) != 0 {
			var volumeOpts VolumeOpts
			volumeOpts.NewVolumeBuilder().BuildVolumeBuilderForSecrets(secrets)
			podtemplate.WithVolumeBuilders(volumeOpts.VolumeBuilders)
		}
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

func setSidecarSecrets(experiment *ExperimentDetails) []v1alpha1.Secret {
	var secrets []v1alpha1.Secret
	secretMap := make(map[string]bool)
	for _, sidecar := range experiment.SideCars {
		for _, secret := range sidecar.Secrets {
			if _, ok := secretMap[secret.Name]; !ok {
				secretMap[secret.Name] = true
				secrets = append(secrets, secret)
			}
		}
	}
	return secrets
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
