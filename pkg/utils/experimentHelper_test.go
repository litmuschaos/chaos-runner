package utils

import (
	"testing"
)

func TestCreateExperimentList(t *testing.T) {

	tests := map[string]struct {
		engineDetails EngineDetails
		isErr         bool
	}{
		"Test Positive-1": {
			engineDetails: EngineDetails{
				Name:            "Fake Engine",
				EngineNamespace: "Fake NameSpace",
				Experiments: []string{
					"fake-exp-1",
					"fake-exp-2",
				},
			},
			isErr: false,
		},
		"Test Negative-1": {
			engineDetails: EngineDetails{
				Name:            "Fake Engine",
				EngineNamespace: "Fake NameSpace",
			},
			isErr: true,
		},
	}

	for name, moke := range tests {
		t.Run(name, func(t *testing.T) {

			ExpList := moke.engineDetails.CreateExperimentList()
			if len(ExpList) == 0 && !moke.isErr {
				t.Fatalf("%v test failed as the experiment list is still empty", name)
			} else if len(ExpList) != 0 && moke.isErr {
				t.Fatalf("%v test failed as the experiment list is non empty for non empty experiment details on engine", name)
			}

			for i := range ExpList {
				if ExpList[i].Name != moke.engineDetails.Experiments[i] && !moke.isErr {
					t.Fatalf("The expected experimentName is %v but got %v", ExpList[i].Name, moke.engineDetails.Experiments[i])
				}
			}
		})
	}
}
