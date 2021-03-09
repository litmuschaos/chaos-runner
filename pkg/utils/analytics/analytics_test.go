package analytics

import (
	"testing"
)

func TestPatchConfigMaps(t *testing.T) {
	fakeExperimentName := "fake-experiment-name"
	fakeUUID := "fake-uuid-12345"
	TriggerAnalytics(fakeExperimentName, fakeUUID)
}
