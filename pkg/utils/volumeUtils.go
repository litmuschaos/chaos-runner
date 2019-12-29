package utils

import (
	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	volume "github.com/litmuschaos/kube-helper/kubernetes/volume/v1alpha1"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

// CreateVolumeBuilder build Volume needed in execution of experiments
func CreateVolumeBuilder(configMaps []v1alpha1.ConfigMap, secrets []v1alpha1.Secret) []*volume.Builder {
	volumeBuilderList := []*volume.Builder{}
	if configMaps == nil {
		log.Infoln("Unable to fetch chaosExperiment ConfigMaps, to create volume")
		return nil
	}
	for _, v := range configMaps {
		log.Infoln("Would create VolumeBuilder for : ", v)
		volumeBuilder := volume.NewBuilder().
			WithConfigMap(v.Name)
		volumeBuilderList = append(volumeBuilderList, volumeBuilder)
	}

	if secrets == nil {
		log.Infoln("Unable to getch Valid chaosExperiment Secrets, to create Volumes")
	}

	for _, v := range secrets {
		log.Infof("Would create VolumeBuilder for Secret Name: %v", v)
		volumeBuilder := volume.NewBuilder().WithSecret(v.Name)
		volumeBuilderList = append(volumeBuilderList, volumeBuilder)
	}
	return volumeBuilderList
}

// CreateVolumeMounts mounts Volume needed in execution of experiments
func CreateVolumeMounts(configMaps []v1alpha1.ConfigMap, secrets []v1alpha1.Secret) []corev1.VolumeMount {
	var volumeMountsList []corev1.VolumeMount
	for _, v := range configMaps {
		var volumeMount corev1.VolumeMount
		volumeMount.Name = v.Name
		volumeMount.MountPath = v.MountPath
		volumeMountsList = append(volumeMountsList, volumeMount)
	}

	for _, v := range secrets {
		var volumeMount corev1.VolumeMount
		volumeMount.Name = v.Name
		volumeMount.MountPath = v.MountPath
		volumeMountsList = append(volumeMountsList, volumeMount)
	}

	return volumeMountsList
}
