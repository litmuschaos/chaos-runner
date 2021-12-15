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
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"testing"
	"time"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	"github.com/litmuschaos/chaos-runner/pkg/log"
	"github.com/litmuschaos/chaos-runner/pkg/utils"
	"github.com/litmuschaos/chaos-runner/pkg/utils/k8s"
	"github.com/litmuschaos/chaos-runner/pkg/utils/litmus"
	"github.com/litmuschaos/litmus-go/pkg/utils/retry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	clients    utils.ClientSets
	kubeconfig string
)

func TestChaos(t *testing.T) {

	RegisterFailHandler(Fail)
	RunSpecs(t, "BDD test")
}
func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", os.Getenv("HOME")+"/.kube/config", "path to kubeconfig to invoke kubernetes API calls")
}

var _ = BeforeSuite(func() {

	// Getting kubeconfig and generate clientSets
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	Expect(err).To(BeNil(), "failed to get config")

	k8sClientSet, err := k8s.GenerateK8sClientSet(config)
	Expect(err).To(BeNil(), "failed to generate k8sClientSet")

	litmusClientSet, err := litmus.GenerateLitmusClientSet(config)
	Expect(err).To(BeNil(), "failed to generate litmusClientSet")

	clients = utils.ClientSets{}
	clients.KubeClient = k8sClientSet
	clients.LitmusClient = litmusClientSet

	//Creating crds
	By("Installing Litmus CRDs")
	err = exec.Command("kubectl", "apply", "-f", "https://raw.githubusercontent.com/litmuschaos/chaos-operator/master/deploy/chaos_crds.yaml").Run()
	Expect(err).To(BeNil(), "unable to create Litmus CRD's")
	log.Info("CRDs created")

	//Creating rbacs
	By("Installing RBAC")
	err = exec.Command("kubectl", "apply", "-f", "https://raw.githubusercontent.com/litmuschaos/chaos-operator/master/deploy/rbac.yaml").Run()
	Expect(err).To(BeNil(), "unable to create RBAC Permissions")
	log.Info("RBAC created")

	//Creating Chaos-Operator
	By("Installing Chaos-Operator")
	err = exec.Command("kubectl", "apply", "-f", "https://raw.githubusercontent.com/litmuschaos/chaos-operator/master/deploy/operator.yaml").Run()
	Expect(err).To(BeNil(), "unable to create Chaos-operator")

	log.Info("Chaos-Operator created")

	err = retry.
		Times(uint(180 / 2)).
		Wait(time.Duration(2) * time.Second).
		Try(func(attempt uint) error {
			podSpec, err := clients.KubeClient.CoreV1().Pods("litmus").List(metav1.ListOptions{LabelSelector: "name=chaos-operator"})
			if err != nil || len(podSpec.Items) == 0 {
				return errors.Errorf("Unable to list chaos-operator, err: %v", err)
			}
			for _, v := range podSpec.Items {
				if v.Status.Phase != "Running" {
					return errors.Errorf("chaos-operator is not in running state, phase: %v", v.Status.Phase)
				}
			}
			return nil
		})

	Expect(err).To(BeNil(), "the chaos-operator is not in running state")
	log.Info("Chaos-Operator is in running state")

	By("Installing Pod Delete Experiment")
	err = exec.Command("kubectl", "apply", "-f", "https://hub.litmuschaos.io/api/chaos/master?file=charts/generic/pod-delete/experiment.yaml", "-n", "litmus").Run()
	Expect(err).To(BeNil(), "unable to create Pod-Delete Experiment")
	log.Info("pod-delete ChaosExperiment created")

	err = exec.Command("kubectl", "apply", "-f", "https://raw.githubusercontent.com/litmuschaos/chaos-operator/master/tests/manifest/pod_delete_rbac.yaml", "-n", "litmus").Run()
	Expect(err).To(BeNil(), "unable to create pod-delete rbac")
	log.Info("pod-delete-sa created")
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
			_, err := clients.KubeClient.AppsV1().Deployments("litmus").Create(deployment)
			Expect(err).To(
				BeNil(),
				"while creating nginx deployment in namespace litmus",
			)
			log.Info("nginx deployment created")
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
			_, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines("litmus").Create(chaosEngine)
			Expect(err).To(
				BeNil(),
				"while building ChaosEngine engine-nginx in namespace litmus",
			)
			log.Info("chaos engine created")

			err = retry.
				Times(uint(180 / 2)).
				Wait(time.Duration(2) * time.Second).
				Try(func(attempt uint) error {
					pod, err := clients.KubeClient.CoreV1().Pods("litmus").Get("engine-nginx-runner", metav1.GetOptions{})
					if err != nil {
						return errors.Errorf("unable to get chaos-runner pod, err: %v", err)
					}
					if pod.Status.Phase != v1.PodRunning && pod.Status.Phase != v1.PodSucceeded {
						return errors.Errorf("chaos runner is not in running state, phase: %v", pod.Status.Phase)
					}
					return nil
				})

			if err != nil {
				log.Errorf("The chaos-runner is not in running state, err: %v", err)
			}
			log.Info("runner pod created")
		})
	})

	When("Check if the Job is spawned by chaos-runner", func() {
		It("Should create a Pod delete Job", func() {

			err := retry.
				Times(uint(180 / 2)).
				Wait(time.Duration(2) * time.Second).
				Try(func(attempt uint) error {
					var jobName string
					jobs, err := clients.KubeClient.BatchV1().Jobs("litmus").List(metav1.ListOptions{})
					if err != nil {
						return err
					}
					regExpr, err := regexp.Compile("pod-delete-.*")
					if err != nil {
						return err
					}
					for _, job := range jobs.Items {
						matched := regExpr.MatchString(job.Name)
						if matched {
							jobName = job.Name
							break
						}
					}
					if jobName == "" {
						return fmt.Errorf("unable to get the job, might be something wrong with chaos-runner")
					}
					return nil
				})

			Expect(err).To(
				BeNil(),
				"while listing experiment job in namespace litmus",
			)
		})
	})

})

//Deleting all unused resources
var _ = AfterSuite(func() {

	By("Deleting chaosengine CRD")
	ceDeleteCRDs := exec.Command("kubectl", "delete", "crds", "chaosengines.litmuschaos.io").Run()
	Expect(ceDeleteCRDs).To(BeNil())

	By("Deleting other CRDs")
	crdDeletion := exec.Command("kubectl", "delete", "crds", "chaosresults.litmuschaos.io", "chaosexperiments.litmuschaos.io").Run()
	Expect(crdDeletion).To(BeNil())

	By("Deleting namespace litmus")
	rbacDeletion := exec.Command("kubectl", "delete", "ns", "litmus").Run()
	Expect(rbacDeletion).To(BeNil())
	log.Info("deleted CRD and Namespace")
})
