# chaos-executor

**Why does this repo Exists**
This repo, is all about converting the current ansible-based executor to a go-based executor.
It is a experimental repo, for implementing all the functions of ansible-based executor, to go-based executor

**How to use chaos-executor**
To use these follow these steps:
    - Clone the chaos-operator
    - Change it -runner image in the builder function, to litmuschaos/chaos-executor:ci
    - Check the logs, of the runner

**Functionality Implemented**
    - All Executor functionality implemented
    - Added the status paching while job execution

**Limitations**
    - Currently, it can only run Generic chaos, as it does'nt have the ability to read the config maps, will add that as soon possible
    - status might go empty after the experiment execution
    - Errors aren't handled gracefully, will have to use some other logger, rather than logrus