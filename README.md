# CHAOS RUNNER

The chaos Runner is an operational bridge between the Chaos-Operator and the LitmusChaos experiment jobs. 

- It is launched as a pod in the app namespace & reconciled by the Litmus Chaos Operator
- Reads the chaos parameters from the experiment CR & overrides with values from the ChaosEngine, constructs the experiment job after 
  validating dependencies such as configmap/secret volumes & launches it (along with the monitor/chaos-exporter deployment if engine's monitoring policy is true)
- Monitors the experiment job until completion
- Cleans up the experiment job post completion based on the engine's jobCleanUpPolicy (delete or retain)
- Patches the ChaosEngine with the verdict of the experiment 

This repo consists of the go-version of the currently used ansible-based runner/executor. The motivation includes:

- Support a contextual/audit logging framework in litmus where the sequence of events from creation of the engine to its eventual removal 
  (with the experiment execution summary in b/w) is traceable

- Support termination/abort of experiments in progress, Removal of all chaos residue with single operation etc., One of the ways to achieve 
  this, is to ensure that the OwnerReference of the ChaosEngine is passed to the experiment jobs (which can be arguably termed the child resources 
  along with the runner itself) to allow the garbage collection to take care of the deletePropagation.

- Create and/or mount volume (configmaps, secrets) resources with validation for availability of these resources.

- Support dependency management of experiments in case of batch runs with possible parallel / asynchronous execution & thereby patching of the ChaosEngine.

- Increased execution performance (today, the time taken to construct and launch the experiment job is directly proportional to the number of environment 
  variables passed to it. The Kafka chaos experiments are seen to take, on an average, ~60-70s before launching the job itself.)

- Allow multiple combinations of random execution in case of future support for Chaos Scheduling, where it may be necessary for the job execution to be 
  randomized based on different conditions (iteration count, minimum intervals etc.,)

## How to use the Go Chaos Runner

- Provide the appropriate values in the ChaosEngine spec as shown below:

```
apiVersion: litmuschaos.io/v1alpha1
kind: ChaosEngine
metadata:
  name: engine
  namespace: litmus
spec:
  appinfo:
    appkind: deployment
    applabel: app=nginx
    appns: litmus
  chaosServiceAccount: litmus
  components:
    runner:
      type: "go"
      image: "litmuschaos/chaos-executor:ci"
  experiments:
  - name: pod-delete 
    spec:
      components: null
  monitoring: false
```

## Further Improvements 

- The Go Chaos Runner is in alpha stage with further improvements coming soon!! 
