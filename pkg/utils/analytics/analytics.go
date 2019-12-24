package analytics

import (
	// "github.com/go-logr/logr"
	ga "github.com/jpillora/go-ogle-analytics"
)

// Test ashdihadihai
func Test(experimentName string) {

	client, err := ga.NewClient("UA-127388617-2")
	if err != nil {
		println("client")
		panic(err)
	}
	client.ClientID("dba81026-a68e-4f67-9f4c-10bfafd97ed4")
	err = client.Send(ga.NewEvent("ExperimentRun", "Total").Label("AppName"))
	if err != nil {
		panic(err)
	}
	err = client.Send(ga.NewEvent("ExperimentRun", experimentName).Label("AppName"))
	if err != nil {
		println("expoervvv")
		panic(err)
	}
	println("Event fired!")

}
