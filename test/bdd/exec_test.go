package bdd

/*
Copyright 2019 LitmusChaos Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"flag"
	"os"
	"os/exec"
	"regexp"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	scheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	chaosClient "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned/typed/litmuschaos/v1alpha1"
	restclient "k8s.io/client-go/rest"
)

var (
	kubeconfig      string
	config          *restclient.Config
	k8sClientSet    *kubernetes.Clientset
	litmusClientSet *chaosClient.LitmuschaosV1alpha1Client
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig to invoke kubernetes API calls")
}
func TestChaos(t *testing.T) {

	RegisterFailHandler(Fail)
	RunSpecs(t, "BDD test")
}

var _ = BeforeSuite(func() {

	var err error
	kubeconfig = os.Getenv("HOME") + "/.kube/config"
	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		Expect(err).To(BeNil(), "failed to get config")
	}

	k8sClientSet, err = kubernetes.NewForConfig(config)

	if err != nil {
		Expect(err).To(BeNil(), "failed to get k8sClientSet")
	}

	litmusClientSet, err = chaosClient.NewForConfig(config)

	if err != nil {
		Expect(err).To(BeNil(), "failed to get litmusClientSet")
	}

	err = v1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		klog.Infof("Error to add to Scheme: %v", err)
	}

	//Creating crds
	By("Installing Litmus CRDs")
	err = exec.Command("kubectl", "create", "-f", "../../deploy/chaos_crds.yaml").Run()
	if err != nil {
		klog.Infof("Unable to execute command, due to error: %v", err)
	}

	//Creating rbacs
	err = exec.Command("kubectl", "create", "-f", "../../deploy/rbac.yaml").Run()
	if err != nil {
		klog.Infof("Unable to execute command, due to error: %v", err)
	}

	//Creating Chaos-Operator
	By("Installing Chaos-Operator")
	err = exec.Command("kubectl", "create", "-f", "../../deploy/operator.yaml").Run()
	if err != nil {
		klog.Infof("Unable to execute command, due to error: %v", err)
	}

	klog.Infof("Chaos-Operator installed Successfully")

	//Wait for the creation of chaos-operator
	time.Sleep(10 * time.Second)

	//Check for the status of the chaos-operator
	operator, _ := k8sClientSet.CoreV1().Pods("litmus").List(metav1.ListOptions{LabelSelector: "name=chaos-operator"})
	for _, v := range operator.Items {

		Expect(string(v.Status.Phase)).To(Equal("Running"))
		break
	}
})

//BDD Tests to check secondary resources
var _ = Describe("BDD on chaos-executor", func() {

	var err error
	// BDD TEST CASE 1
	When("Create a test Deployment with nginx image", func() {
		It("Should create Nginx deployment ", func() {
			By("building a deployment")
			//creating nginx deployment
			deployment := &appv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nginx",
					Namespace: "litmus",
					Labels: map[string]string{
						"app": "nginx",
					},
					Annotations: map[string]string{
						"litmuschaos.io/chaos": "true",
					},
				},
				Spec: appv1.DeploymentSpec{
					Replicas: func(i int32) *int32 { return &i }(3),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "nginx",
						},
					},
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app": "nginx",
							},
						},
						Spec: v1.PodSpec{
							ServiceAccountName: "litmus",
							Containers: []v1.Container{
								{
									Name:  "nginx",
									Image: "nginx:latest",
									Ports: []v1.ContainerPort{
										{

											ContainerPort: 80,
										},
									},
								},
							},
						},
					},
				},
			}
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building deployment nginx in namespace litmus",
			)

			By("creating above deployment")
			_, err := k8sClientSet.AppsV1().Deployments("litmus").Create(deployment)
			Expect(err).To(
				BeNil(),
				"while creating deployment nginx in namespace litmus",
			)
		})
	})

	When("Create Pod-delete chaosExperiment in litmus Namespace", func() {
		It("should create a CustomResource ChaosExperiment Pod-delete", func() {

			By("building a chaosExperiment")
			ChaosExperiment := &v1alpha1.ChaosExperiment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-delete",
					Namespace: "litmus",
					Labels: map[string]string{
						"litmuschaos.io/name": "kubernetes",
					},
				},
				Spec: v1alpha1.ChaosExperimentSpec{
					Definition: v1alpha1.ExperimentDef{

						Args:    []string{"-c", "ansible-playbook ./experiments/generic/pod_delete/pod_delete_ansible_logic.yml -i /etc/ansible/hosts -vv; exit 0"},
						Command: []string{"/bin/bash"},
						Image:   "litmuschaos/ansible-runner:ci",
						ENVList: []v1alpha1.ENVPair{
							{
								Name:  "ANSIBLE_STDOUT_CALLBACK",
								Value: "default",
							},
							{
								Name:  "TOTAL_CHAOS_DURATION",
								Value: "15",
							},
							{
								Name:  "CHAOS_INTERVAL",
								Value: "5",
							},
							{
								Name:  "LIB",
								Value: "",
							},
						},
						Labels: map[string]string{
							"name": "pod-delete",
						},
					},
				},
			}
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building pod-delete chaosExperiment",
			)

			By("creating above chaosExperiment")
			_, err = litmusClientSet.ChaosExperiments("litmus").Create(ChaosExperiment)
			Expect(err).To(
				BeNil(),
				"while creating chaosExerpiment Pod-delete in namespace litmus",
			)
		})
	})
	When("Creating ChaosEngine to trigger chaos-executor", func() {
		It("should create a runnerPod and Service ", func() {

			By("Building a ChaosEngine")
			chaosEngine := &v1alpha1.ChaosEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "engine-nginx",
					Namespace: "litmus",
				},
				Spec: v1alpha1.ChaosEngineSpec{
					Components: v1alpha1.ComponentParams{
						Runner: v1alpha1.RunnerInfo{
							Type:  "go",
							Image: "litmuschaos/chaos-executor:ci",
						},
					},
					Appinfo: v1alpha1.ApplicationParams{
						Appns:    "litmus",
						Applabel: "app=nginx",
						AppKind:  "deployment",
					},
					ChaosServiceAccount: "litmus",
					Monitoring:          true,
					Experiments: []v1alpha1.ExperimentList{
						{
							Name: "pod-delete",
						},
					},
				},
			}
			Expect(err).ShouldNot(
				HaveOccurred(),
				"while building ChaosEngine engine-nginx in namespace litmus",
			)

			By("Creating ChaosEngine Resource")
			_, err = litmusClientSet.ChaosEngines("litmus").Create(chaosEngine)
			Expect(err).To(
				BeNil(),
				"while building ChaosEngine engine-nginx in namespace litmus",
			)

			time.Sleep(30 * time.Second)

			//Fetching engine-nginx-runner pod
			runner, err := k8sClientSet.CoreV1().Pods("litmus").Get("engine-nginx-runner", metav1.GetOptions{})
			Expect(err).To(BeNil())
			//Fetching engine-nginx-exporter pod
			exporter, err := k8sClientSet.CoreV1().Pods("litmus").Get("engine-nginx-monitor", metav1.GetOptions{})
			Expect(err).To(BeNil())
			Expect(string(runner.Status.Phase)).To(Or(Equal("Running"), Equal("Succeeded")))
			Expect(string(exporter.Status.Phase)).To(Equal("Running"))

		})
	})
	var jobName string
	When("Check if the Job is spawned by chaos-executor", func() {
		It("Should create a Pod delete Job", func() {
			jobs, _ := k8sClientSet.BatchV1().Jobs("litmus").List(metav1.ListOptions{})
			for i := range jobs.Items {
				matched, _ := regexp.MatchString("pod-delete-.*", jobs.Items[i].Name)
				if matched == true {
					jobName = jobs.Items[i].Name
				}
			}
			Expect(jobName).To(
				Not(BeEmpty()),
				"Unable to get the job, might be something wrong with chaos-executor",
			)
		})
	})

})

//Deleting all unused resources
var _ = AfterSuite(func() {

	By("Deleting Litmus NameSpace")
	deleteErr := k8sClientSet.CoreV1().Namespaces().Delete("litmus", &metav1.DeleteOptions{})
	Expect(deleteErr).To(
		BeNil(),
		"Unable to delete Litmus, might be lack of permissions",
	)

	By("Deleting all CRDs")
	crdDeletion := exec.Command("kubectl", "delete", "-f", "../../deploy/chaos_crds.yaml").Run()
	Expect(crdDeletion).To(BeNil())
})
