package utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func newDeadmanTestClient(t *testing.T, handler http.Handler) (*httptest.Server, ClientSets) {
	t.Helper()

	server := httptest.NewServer(handler)
	client, err := kubernetes.NewForConfig(&rest.Config{Host: server.URL})
	if err != nil {
		server.Close()
		t.Fatalf("create Kubernetes client: %v", err)
	}

	return server, ClientSets{KubeClient: client}
}

func setDeadmanTestTimings(t *testing.T, outage, interval, probe string) {
	t.Helper()
	t.Setenv(deadmanTimeoutEnv, outage)
	t.Setenv(deadmanIntervalEnv, interval)
	t.Setenv(deadmanProbeTimeoutEnv, probe)
}

func TestStartDeadmanSwitchTripsAfterSustainedAPIOutage(t *testing.T) {
	setDeadmanTestTimings(t, "50ms", "5ms", "10ms")

	server, clients := newDeadmanTestClient(t, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
		},
	))
	defer server.Close()

	ctx, failures, cancel := StartDeadmanSwitch(context.Background(), clients)
	defer cancel()

	select {
	case err, ok := <-failures:
		if !ok {
			t.Fatal("failure channel closed without reporting the outage")
		}
		if err == nil || !strings.Contains(err.Error(), "control-plane heartbeat failed") {
			t.Fatalf("unexpected deadman error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("deadman switch did not trip during sustained API outage")
	}

	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("deadman context was not cancelled after the outage")
	}
}

func TestStartDeadmanSwitchRecoversBeforeTimeout(t *testing.T) {
	setDeadmanTestTimings(t, "300ms", "5ms", "20ms")

	var mu sync.RWMutex
	status := http.StatusServiceUnavailable
	firstRequest := make(chan struct{}, 1)

	server, clients := newDeadmanTestClient(t, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			select {
			case firstRequest <- struct{}{}:
			default:
			}
			mu.RLock()
			currentStatus := status
			mu.RUnlock()

			if currentStatus != http.StatusOK {
				http.Error(w, "temporarily unavailable", currentStatus)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"gitVersion":"test"}`))
		},
	))
	defer server.Close()

	ctx, failures, cancel := StartDeadmanSwitch(context.Background(), clients)
	defer cancel()

	select {
	case <-firstRequest:
	case <-time.After(time.Second):
		t.Fatal("deadman switch did not probe the API server")
	}

	time.Sleep(20 * time.Millisecond)
	mu.Lock()
	status = http.StatusOK
	mu.Unlock()

	select {
	case err, ok := <-failures:
		if ok && err != nil {
			t.Fatalf("deadman tripped despite API recovery: %v", err)
		}
	case <-time.After(150 * time.Millisecond):
		// No failure while the API remains healthy.
	}

	if err := ctx.Err(); err != nil {
		t.Fatalf("context cancelled despite API recovery: %v", err)
	}
}

func TestStartDeadmanSwitchStopsOnParentCancellation(t *testing.T) {
	setDeadmanTestTimings(t, "1s", "10ms", "10ms")

	server, clients := newDeadmanTestClient(t, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"gitVersion":"test"}`))
		},
	))
	defer server.Close()

	parent, cancelParent := context.WithCancel(context.Background())
	ctx, failures, stop := StartDeadmanSwitch(parent, clients)
	defer stop()

	cancelParent()

	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("deadman context did not follow parent cancellation")
	}

	select {
	case err, ok := <-failures:
		if ok && err != nil {
			t.Fatalf("normal parent cancellation reported a deadman failure: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("deadman monitor did not stop after parent cancellation")
	}
}
