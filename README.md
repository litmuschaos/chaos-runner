# chaos-executor

**Why does this repo Exists**
This repo, is all about converting the current ansible-based executor to a go-based executor.
It is a experimental repo, for implementing all the functions of ansible-based executor, to go-based executor
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
      image: "rahulchheda1997/chaos-executor:ci"
  experiments:
  - name: pod-delete 
    spec:
      components: null
  monitoring: false

**How to use chaos-executor**
To use these follow these steps:
    - Edit the chaosEngine YAML, and add type: "go", in components.runner, and change the image: "rahulchheda1997/chaos-executor:ci"
    - Here is a sample ChaosEngine YAML for reference: 
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
        image: "rahulchheda1997/chaos-executor:ci"
    experiments:
    - name: pod-delete 
      spec:
        components: null
      monitoring: false



**Sample ChaosEngine YAML**

**Functionality Implemented**
    - All Executor functionality implemented
    - Added the status paching while job execution

**Limitations**
    - Currently, it can only run Generic chaos, as it does'nt have the ability to read the config maps, will add that as soon possible
    - status might go empty after the experiment execution
    - Errors aren't handled gracefully, will have to use some other logger, rather than logrus