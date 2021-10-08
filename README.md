# CHAOS RUNNER

[![Slack Channel](https://img.shields.io/badge/Slack-Join-purple)](https://slack.litmuschaos.io)
![GitHub Workflow](https://github.com/litmuschaos/chaos-runner/actions/workflows/push.yml/badge.svg?branch=master)
[![Docker Pulls](https://img.shields.io/docker/pulls/litmuschaos/chaos-runner.svg)](https://hub.docker.com/r/litmuschaos/chaos-runner)
[![GitHub issues](https://img.shields.io/github/issues/litmuschaos/chaos-runner)](https://github.com/litmuschaos/chaos-runner/issues)
[![Twitter Follow](https://img.shields.io/twitter/follow/litmuschaos?style=social)](https://twitter.com/LitmusChaos)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/5291/badge)](https://bestpractices.coreinfrastructure.org/projects/5291)
[![Go Report Card](https://goreportcard.com/badge/github.com/litmuschaos/chaos-runner)](https://goreportcard.com/report/github.com/litmuschaos/chaos-runner)
[![BCH compliance](https://bettercodehub.com/edge/badge/litmuschaos/chaos-runner?branch=master)](https://bettercodehub.com/)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Flitmuschaos%2Fchaos-runner.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Flitmuschaos%2Fchaos-runner?ref=badge_shield)
[![YouTube Channel](https://img.shields.io/badge/YouTube-Subscribe-red)](https://www.youtube.com/channel/UCa57PMqmz_j0wnteRa9nCaw)
<br><br>

The chaos Runner is an operational bridge between the Chaos-Operator and the LitmusChaos experiment jobs. 

- It is launched as a pod in the chaos namespace(where chaosengine is running) & reconciled by the Litmus Chaos Operator
- Reads the chaos parameters from the experiment CR & overrides with values from the ChaosEngine, constructs the experiment job after 
  validating dependencies such as configmap/secret volumes & launches it (along with the monitor/chaos-exporter deployment if engine's monitoring policy is true)
- Monitors the experiment pod until completion
- Cleans up the experiment job post completion based on the engine's jobCleanUpPolicy (delete or retain)
- Patches the ChaosEngine with the verdict of the experiment and creates the events for the different phases inside chaosengine. 

Objective behind chaos-runner creation:

- Support a contextual/audit logging framework in litmus where the sequence of events from creation of the engine to its eventual removal 
  (with the experiment execution summary in b/w) is traceable

- Support termination/abort of experiments in progress, Removal of all chaos residue with single operation etc., One of the ways to achieve 
  this, is to ensure that the OwnerReference of the ChaosEngine is passed to the experiment jobs (which can be arguably termed the child resources 
  along with the runner itself) to allow the garbage collection to take care of the deletePropagation.

- Create and/or mount volume (configmaps, secrets) resources with validation for availability of these resources.

- Support dependency management of experiments in case of batch runs with possible parallel / asynchronous execution & thereby patching of the ChaosEngine.

- Allow multiple combinations of random execution in case of future support for Chaos Scheduling, where it may be necessary for the job execution to be 
  randomized based on different conditions (iteration count, minimum intervals etc.,)

## Further Improvements 

- The Go Chaos Runner is in beta stage with further improvements coming soon!! 

## How to get started?

Refer the [LitmusChaos documentation](https://docs.litmuschaos.io) and [Experiment Documentation](https://litmuschaos.github.io/litmus/experiments/concepts/chaos-resources/contents/)

## How do I contribute?

You can contribute by raising issues, improving the documentation, contributing to the core framework and tooling, etc.

Head over to the [Contribution guide](CONTRIBUTING.md)

## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Flitmuschaos%2Fchaos-runner.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Flitmuschaos%2Fchaos-runner?ref=badge_large)