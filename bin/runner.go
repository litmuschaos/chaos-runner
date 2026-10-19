package main

import (
	"context"
	"errors"
	"os"

	"github.com/litmuschaos/chaos-runner/pkg/log"
	"github.com/litmuschaos/chaos-runner/pkg/telemetry"
	"github.com/litmuschaos/chaos-runner/pkg/utils"
	"github.com/litmuschaos/chaos-runner/pkg/utils/analytics"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:          true,
		DisableSorting:         true,
		DisableLevelTruncation: true,
	})
}

func main() {
	ctx := context.Background()

	// Set up Observability.
	if otelExporterEndpoint := os.Getenv(telemetry.OTELExporterOTLPEndpoint); otelExporterEndpoint != "" {
		shutdown, err := telemetry.InitOTelSDK(ctx, otelExporterEndpoint)
		if err != nil {
			log.Errorf("Failed to initialize OTel SDK: %v", err)
			return
		}
		defer func() {
			err = errors.Join(err, shutdown(ctx))
		}()
		ctx = telemetry.GetTraceParentContext()
	}

	engineDetails := utils.EngineDetails{}
	clients := utils.ClientSets{}

	ctx, span := otel.Tracer(telemetry.TracerName).Start(ctx, "ExecuteChaosRunner")
	defer span.End()

	// Getting kubeConfig and Generate ClientSets
	if err := clients.GenerateClientSetFromKubeConfig(); err != nil {
		log.Errorf("unable to create ClientSets, error: %v", err)
		return
	}

	runnerCtx, deadmanErrors, stopDeadman := utils.StartDeadmanSwitch(ctx, clients)
	defer stopDeadman()

	var deadmanFailure error
	checkDeadman := func() error {
		if deadmanFailure != nil {
			return deadmanFailure
		}

		select {
		case err, ok := <-deadmanErrors:
			if ok && err != nil {
				deadmanFailure = err
			}
		default:
		}

		if deadmanFailure != nil {
			return deadmanFailure
		}

		if runnerCtx.Err() != nil {
			// The deadman monitor sends its error before cancelling runnerCtx.
			// Drain the buffered error to preserve the original failure.
			select {
			case err, ok := <-deadmanErrors:
				if ok && err != nil {
					deadmanFailure = err
					return deadmanFailure
				}
			default:
			}
			return runnerCtx.Err()
		}

		return nil
	}

	// Fetching all the ENVs passed from the chaos-operator
	// create and initialize the experimentList
	if err := engineDetails.SetEngineDetails().SetEngineUID(clients); err != nil {
		log.Errorf("unable to get ChaosEngineUID, error: %v", err)
		return
	}

	experimentList := engineDetails.CreateExperimentList()
	log.InfoWithValues("Experiments details are as follows", logrus.Fields{
		"Experiments List":     engineDetails.Experiments,
		"Engine Name":          engineDetails.Name,
		"Targets":              engineDetails.Targets,
		"Service Account Name": engineDetails.SvcAccount,
		"Engine Namespace":     engineDetails.EngineNamespace,
	})

	if err := checkDeadman(); err != nil {
		log.Errorf("deadman switch stopped the chaos runner before execution: %v", err)
		return
	}

	if err := utils.InitialPatchEngine(engineDetails, clients, experimentList); err != nil {
		log.Errorf("unable to patch Initial ExperimentStatus in ChaosEngine, error: %v", err)
		return
	}

	// Steps for each Experiment
	for _, experiment := range experimentList {
		if err := checkDeadman(); err != nil {
			log.Errorf("deadman switch stopped scheduling experiments: %v", err)
			break
		}

		// Sending event to GA instance
		if engineDetails.ClientUUID != "" {
			analytics.TriggerAnalytics(experiment.Name, engineDetails.ClientUUID)
		}

		// Check the existence of ChaosExperiment inside the cluster.
		if err := experiment.HandleChaosExperimentExistence(engineDetails, clients); err != nil {
			if deadmanErr := checkDeadman(); deadmanErr != nil {
				log.Errorf("deadman switch stopped the chaos runner: %v", deadmanErr)
				break
			}

			log.Errorf("unable to get ChaosExperiment name: %v, in namespace: %v, error: %v", experiment.Name, experiment.Namespace, err)
			experiment.ExperimentSkipped(utils.ExperimentNotFoundErrorReason, engineDetails, clients)
			continue
		}

		// Derive the required fields from the experiment and engine.
		if err := experiment.SetValueFromChaosResources(&engineDetails, clients); err != nil {
			if deadmanErr := checkDeadman(); deadmanErr != nil {
				log.Errorf("deadman switch stopped the chaos runner: %v", deadmanErr)
				break
			}

			log.Errorf("unable to set values from Chaos Resources, error: %v", err)
			experiment.ExperimentSkipped(utils.ExperimentNotFoundErrorReason, engineDetails, clients)
			engineDetails.ExperimentSkippedPatchEngine(&experiment, clients)
			continue
		}

		// Derive the environment variables from the experiment and engine.
		if err := experiment.SetENV(runnerCtx, engineDetails, clients); err != nil {
			if deadmanErr := checkDeadman(); deadmanErr != nil {
				log.Errorf("deadman switch stopped the chaos runner: %v", deadmanErr)
				break
			}

			log.Errorf("unable to patch ENV, error: %v", err)
			experiment.ExperimentSkipped(utils.ExperimentEnvParseErrorReason, engineDetails, clients)
			engineDetails.ExperimentSkippedPatchEngine(&experiment, clients)
			continue
		}

		// Derive the sidecar details from ChaosEngine.
		if err := experiment.SetSideCarDetails(engineDetails.Name, clients); err != nil {
			if deadmanErr := checkDeadman(); deadmanErr != nil {
				log.Errorf("deadman switch stopped the chaos runner: %v", deadmanErr)
				break
			}

			log.Errorf("unable to get sidecar details, error: %v", err)
			experiment.ExperimentSkipped(utils.ExperimentSideCarPatchErrorReason, engineDetails, clients)
			engineDetails.ExperimentSkippedPatchEngine(&experiment, clients)
			continue
		}

		log.Infof("Preparing to run Chaos Experiment: %v", experiment.Name)

		if err := experiment.PatchResources(engineDetails, clients); err != nil {
			if deadmanErr := checkDeadman(); deadmanErr != nil {
				log.Errorf("deadman switch stopped the chaos runner: %v", deadmanErr)
				break
			}

			log.Errorf("unable to patch Chaos Resources required for Chaos Experiment: %v, error: %v", experiment.Name, err)
			experiment.ExperimentSkipped(utils.ExperimentDependencyCheckReason, engineDetails, clients)
			engineDetails.ExperimentSkippedPatchEngine(&experiment, clients)
			continue
		}

		// Do not schedule further work after the control plane becomes unreachable.
		if err := checkDeadman(); err != nil {
			log.Errorf("deadman switch stopped scheduling experiments: %v", err)
			break
		}

		// Generate the experiment dependency-check event inside ChaosEngine.
		experiment.ExperimentDependencyCheck(engineDetails, clients)

		// Create the PodTemplateSpec and final Job.
		if err := utils.BuildingAndLaunchJob(runnerCtx, &experiment, clients); err != nil {
			if deadmanErr := checkDeadman(); deadmanErr != nil {
				log.Errorf("deadman switch stopped the chaos runner: %v", deadmanErr)
				break
			}

			log.Errorf("unable to construct chaos experiment job, error: %v", err)
			experiment.ExperimentSkipped(utils.ExperimentDependencyCheckReason, engineDetails, clients)
			engineDetails.ExperimentSkippedPatchEngine(&experiment, clients)
			continue
		}

		// Check again before continuing with the active Job.
		if err := checkDeadman(); err != nil {
			log.Errorf("deadman switch triggered after Job creation: %v", err)
			if cleanupErr := utils.BestEffortCleanupChaosJob(&experiment, clients); cleanupErr != nil {
				log.Errorf("best-effort chaos Job cleanup failed: %v", cleanupErr)
			}
			break
		}

		experiment.ExperimentJobCreate(engineDetails, clients)

		log.Infof("Started Chaos Experiment Name: %v, with Job Name: %v", experiment.Name, experiment.JobName)

		// Watch the chaos container until completion or cancellation.
		if err := engineDetails.WatchChaosContainerForCompletionContext(runnerCtx, &experiment, clients); err != nil {
			if deadmanErr := checkDeadman(); deadmanErr != nil {
				log.Errorf("deadman switch triggered; stopping active chaos Job %s: %v", experiment.JobName, deadmanErr)
				if cleanupErr := utils.BestEffortCleanupChaosJob(&experiment, clients); cleanupErr != nil {
					log.Errorf("best-effort chaos Job cleanup failed: %v", cleanupErr)
				}
				break
			}

			log.Errorf("unable to Watch the chaos container, error: %v", err)
			experiment.ExperimentSkipped(utils.ExperimentChaosContainerWatchErrorReason, engineDetails, clients)
			engineDetails.ExperimentSkippedPatchEngine(&experiment, clients)
			continue
		}

		log.Infof("Chaos Pod Completed, Experiment Name: %v, with Job Name: %v", experiment.Name, experiment.JobName)

		// Update the ChaosEngine status.
		if err := engineDetails.UpdateEngineWithResult(&experiment, clients); err != nil {
			log.Errorf("unable to Update ChaosEngine Status, error: %v", err)
		}

		log.Infof("Chaos Engine has been updated with result, Experiment Name: %v", experiment.Name)

		// Delete or retain the Job based on the job cleanup policy.
		jobCleanUpPolicy, err := engineDetails.DeleteJobAccordingToJobCleanUpPolicy(&experiment, clients)
		if err != nil {
			log.Errorf("unable to Delete ChaosExperiment Job, error: %v", err)
		}

		experiment.ExperimentJobCleanUp(string(jobCleanUpPolicy), engineDetails, clients)
	}
}
