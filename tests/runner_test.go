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

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	chaosClient "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned/typed/litmuschaos/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	scheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

var (
	kubeconfig      string
	config          *restclient.Config
	k8sClientSet    *kubernetes.Clientset
	litmusClientSet *chaosClient.LitmuschaosV1alpha1Client
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", os.Getenv("HOME")+"/.kube/config", "path to kubeconfig to invoke kubernetes API calls")
}
func TestChaos(t *testing.T) {

	RegisterFailHandler(Fail)
	RunSpecs(t, "BDD test")
}

var _ = BeforeSuite(func() {

	flag.Parse()

	var err error
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
	err = exec.Command("kubectl", "apply", "-f", "../build/_output/test/chaos_crds.yaml").Run()
	if err != nil {
		klog.Infof("Unable to create Litmus CRD's, due to error: %v", err)
	}

	//Creating rbacs
	err = exec.Command("kubectl", "apply", "-f", "../build/_output/test/rbac.yaml").Run()
	if err != nil {
		klog.Infof("Unable to create RBAC Permissions, due to error: %v", err)
	}

	//Creating Chaos-Operator
	By("Installing Chaos-Operator")
	err = exec.Command("kubectl", "apply", "-f", "../build/_output/test/operator.yaml").Run()
	if err != nil {
		klog.Infof("Unable to create Chaos-operator, due to error: %v", err)
	}

	klog.Infof("Chaos-Operator installed Successfully")

	//Wait for the creation of chaos-operator
	time.Sleep(40 * time.Second)

	//Check for the status of the chaos-operator
	operator, _ := k8sClientSet.CoreV1().Pods("litmus").List(metav1.ListOptions{LabelSelector: "name=chaos-operator"})
	for _, v := range operator.Items {

		Expect(string(v.Status.Phase)).To(Equal("Running"))
		break
	}

	err = exec.Command("kubectl", "apply", "-f", "https://hub.litmuschaos.io/api/chaos/master?file=charts/generic/experiments.yaml", "-n", "litmus").Run()
	if err != nil {
		klog.Infof("Unable to create Pod-Delete Experiment, due to error: %v", err)
	}

	err = exec.Command("kubectl", "apply", "-f", "../build/_output/test/pod_delete_rbac.yaml", "-n", "litmus").Run()
	if err != nil {
		klog.Infof("Unable to create pod-delete rbac, due to error: %v", err)
	}
})

//BDD Tests to check secondary resources
var _ = Describe("BDD on chaos-runner", func() {

	// BDD TEST CASE 1
	When("Create a test Deployment with nginx image", func() {
		It("Should create Nginx deployment ", func() {
			By("Building a nginx deployment")
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
			By("Creating nginx deployment")
			_, err := k8sClientSet.AppsV1().Deployments("litmus").Create(deployment)
			Expect(err).To(
				BeNil(),
				"while creating nginx deployment in namespace litmus",
			)
		})
	})
	When("Creating ChaosEngine to trigger chaos-runner", func() {
		It("Should create a runnerPod and Service ", func() {

			//Creating chaosEngine
			By("Creating ChaosEngine")
			chaosEngine := &v1alpha1.ChaosEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "engine-nginx",
					Namespace: "litmus",
				},
				Spec: v1alpha1.ChaosEngineSpec{
					Appinfo: v1alpha1.ApplicationParams{
						Appns:    "litmus",
						Applabel: "app=nginx",
						AppKind:  "deployment",
					},
					EngineState:         "active",
					ChaosServiceAccount: "pod-delete-sa",
					Components: v1alpha1.ComponentParams{
						Runner: v1alpha1.RunnerInfo{
							Image: "litmuschaos/chaos-runner:ci",
							Type:  "go",
						},
					},
					Monitoring: true,
					Experiments: []v1alpha1.ExperimentList{
						{
							Name: "pod-delete",
							Spec: v1alpha1.ExperimentAttributes{
								Rank: uint32(1),
							},
						},
					},
				},
			}

			By("Creating ChaosEngine Resource")
			_, err := litmusClientSet.ChaosEngines("litmus").Create(chaosEngine)
			Expect(err).To(
				BeNil(),
				"while building ChaosEngine engine-nginx in namespace litmus",
			)

			time.Sleep(30 * time.Second)

			//Fetching engine-nginx-runner pod
			runner, err := k8sClientSet.CoreV1().Pods("litmus").Get("engine-nginx-runner", metav1.GetOptions{})
			Expect(err).To(BeNil())
			Expect(string(runner.Status.Phase)).To(Or(Equal("Running"), Equal("Succeeded")))
		})
	})
	var jobName string
	When("Check if the Job is spawned by chaos-runner", func() {
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
				"Unable to get the job, might be something wrong with chaos-runner",
			)
		})
	})

})

//Deleting all unused resources
var _ = AfterSuite(func() {
	By("Deleting all CRDs")
	crdDeletion := exec.Command("kubectl", "delete", "-f", "../build/_output/test/chaos_crds.yaml").Run()
	Expect(crdDeletion).To(BeNil())
	By("Deleting RBAC Permissions")
	rbacDeletion := exec.Command("kubectl", "delete", "-f", "../build/_output/test/rbac.yaml").Run()
	Expect(rbacDeletion).To(BeNil())
})
