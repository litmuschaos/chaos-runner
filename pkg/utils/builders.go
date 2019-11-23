package utils

import (
	"github.com/litmuschaos/kube-helper/kubernetes/container"
	"github.com/litmuschaos/kube-helper/kubernetes/job"
	"github.com/litmuschaos/kube-helper/kubernetes/podtemplatespec"
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

// DeployJob the Job using all the details gathered
func DeployJob(perExperiment ExperimentDetails, engineDetails EngineDetails, envVar []corev1.EnvVar) error {

	// Will build a PodSpecTemplate
	// For creating the spec.template of the Job
	pod := BuildPodTemplateSpec(perExperiment, engineDetails, envVar)

	// Building the Job Object using the podSpecTemplate
	job, err := BuildJob(pod, perExperiment, engineDetails)
	if err != nil {
		log.Info("Unable to build Job")
		return err
	}

	// Generation of ClientSet for creation
	clientSet, _, err := GenerateClientSets(engineDetails.Config)
	if err != nil {
		log.Info("Unable to generate ClientSet while Creating Job")
		return err
	}

	jobsClient := clientSet.BatchV1().Jobs(engineDetails.AppNamespace)

	// Creating the Job
<<<<<<< HEAD
	jobCreationResult, err := jobsClient.Create(job)
	log.Info("Jobcreation log : ", jobCreationResult)
=======
	_, err = jobsClient.Create(job)
	//log.Info("Jobcreation log : ", jobCreationResult)
>>>>>>> 7fc58d356d7f488f50b0af0134e6d881b469225b
	if err != nil {
		log.Info("Unable to create the Job with the clientSet")
	}
	return nil
}

// BuildPodTemplateSpec will build the PodTemplateSpec for further usage
func BuildPodTemplateSpec(perExperiment ExperimentDetails, engineDetails EngineDetails, envVar []corev1.EnvVar) *podtemplatespec.Builder {

	podtemplate := podtemplatespec.NewBuilder().
		WithName(perExperiment.JobName).
		WithNamespace(engineDetails.AppNamespace).
		WithLabels(perExperiment.ExpLabels).
		WithServiceAccountName(engineDetails.SvcAccount).
		WithContainerBuilders(
			container.NewBuilder().
				WithName(perExperiment.JobName).
				WithImage(perExperiment.ExpImage).
				WithCommandNew([]string{"/bin/bash"}).
				WithArgumentsNew(perExperiment.ExpArgs).
				WithImagePullPolicy("Always").
				WithEnvsNew(envVar),
		)
	return podtemplate
}

// BuildJob will build the JobObject (*batchv1.Job) for creation
func BuildJob(pod *podtemplatespec.Builder, perExperiment ExperimentDetails, engineDetails EngineDetails) (*batchv1.Job, error) {
	restartPolicy := corev1.RestartPolicyOnFailure
	jobObj, err := job.NewBuilder().
		WithName(perExperiment.JobName).
		WithNamespace(engineDetails.AppNamespace).
		WithLabels(perExperiment.ExpLabels).
		WithPodTemplateSpecBuilder(pod).
		WithRestartPolicy(restartPolicy).
		Build()
	if err != nil {
		return jobObj, err
	}
	return jobObj, nil
}
