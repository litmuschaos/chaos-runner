package utils

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	deadmanTimeoutEnv      = "CONTROL_PLANE_HEARTBEAT_TIMEOUT"
	deadmanIntervalEnv     = "CONTROL_PLANE_HEARTBEAT_INTERVAL"
	deadmanProbeTimeoutEnv = "CONTROL_PLANE_HEARTBEAT_PROBE_TIMEOUT"
)

func durationFromEnv(name string, fallback time.Duration) time.Duration {
	raw := os.Getenv(name)
	if raw == "" {
		return fallback
	}
	if d, err := time.ParseDuration(raw); err == nil && d > 0 {
		return d
	}
	if seconds, err := strconv.Atoi(raw); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	return fallback
}

// StartDeadmanSwitch watches Kubernetes API-server reachability. A sustained
// outage cancels the returned context and publishes one error on the channel.
// Defaults: 30s outage timeout, 5s probe interval, 3s per probe.
func StartDeadmanSwitch(parent context.Context, clients ClientSets) (context.Context, <-chan error, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)
	failures := make(chan error, 1)
	outageTimeout := durationFromEnv(deadmanTimeoutEnv, 30*time.Second)
	interval := durationFromEnv(deadmanIntervalEnv, 5*time.Second)
	probeTimeout := durationFromEnv(deadmanProbeTimeoutEnv, 3*time.Second)

	go func() {
		defer close(failures)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		var outageStarted time.Time

		probe := func() error {
			probeCtx, probeCancel := context.WithTimeout(ctx, probeTimeout)
			defer probeCancel()
			// /version is a lightweight API-server health/reachability probe.
			_, err := clients.KubeClient.Discovery().RESTClient().Get().
				AbsPath("/version").
				Timeout(probeTimeout).
				Do(probeCtx).Raw()
			return err
		}

		for {
			if err := ctx.Err(); err != nil {
				return
			}
			err := probe()
			if err == nil {
				outageStarted = time.Time{}
			} else {
				if outageStarted.IsZero() {
					outageStarted = time.Now()
				}
				if time.Since(outageStarted) >= outageTimeout {
					deadmanErr := fmt.Errorf("control-plane heartbeat failed continuously for %s: %w", outageTimeout, err)
					failures <- deadmanErr
					cancel()
					return
				}
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	return ctx, failures, cancel
}

// BestEffortCleanupChaosJob attempts to terminate an active chaos Job after the
// deadman switch trips. If the API server is down, deletion cannot be guaranteed;
// Kubernetes will process the deletion when the API becomes reachable.
func BestEffortCleanupChaosJob(experiment *ExperimentDetails, clients ClientSets) error {
	if experiment == nil || experiment.JobName == "" || clients.KubeClient == nil {
		return nil
	}
	deletePolicy := metav1.DeletePropagationForeground
	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		lastErr = clients.KubeClient.BatchV1().Jobs(experiment.Namespace).Delete(
			cleanupCtx,
			experiment.JobName,
			metav1.DeleteOptions{PropagationPolicy: &deletePolicy},
		)
		cancel()
		if lastErr == nil {
			return nil
		}
		// NotFound means the desired cleanup state has already been reached.
		if apierrors.IsNotFound(lastErr) {
			return nil
		}
		if attempt < 4 {
			time.Sleep(2 * time.Second)
		}
	}
	return fmt.Errorf("best-effort cleanup failed for chaos Job %s/%s after retries: %w", experiment.Namespace, experiment.JobName, lastErr)
}
