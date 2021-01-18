package utils

import (
	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	volume "github.com/litmuschaos/elves/kubernetes/volume/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

var (
	// hostpathTypeFile represents the hostpath type
	hostpathTypeFile = corev1.HostPathFile
)

// VolumeOperations filles up VolumeOpts strucuture
func (volumeOpts *VolumeOpts) VolumeOperations(experiment *ExperimentDetails) {
	volumeOpts.NewVolumeBuilder().
		BuildVolumeBuilderForConfigMaps(experiment.ConfigMaps).
		BuildVolumeBuilderForSecrets(experiment.Secrets).
		BuildVolumeBuilderForHostFileVolumes(experiment.HostFileVolumes)

	volumeOpts.NewVolumeMounts().
		BuildVolumeMountsForConfigMaps(experiment.ConfigMaps).
		BuildVolumeMountsForSecrets(experiment.Secrets).
		BuildVolumeMountsForHostFileVolumes(experiment.HostFileVolumes)
}

// NewVolumeMounts initialize the volume builder
func (volumeOpts *VolumeOpts) NewVolumeMounts() *VolumeOpts {
	var volumeMountsList []corev1.VolumeMount
	volumeOpts.VolumeMounts = volumeMountsList
	return volumeOpts
}

// NewVolumeBuilder initialize the volume builder
func (volumeOpts *VolumeOpts) NewVolumeBuilder() *VolumeOpts {
	volumeBuilderList := []*volume.Builder{}
	volumeOpts.VolumeBuilders = volumeBuilderList
	return volumeOpts
}

// BuildVolumeMountsForConfigMaps builds VolumeMounts for ConfigMaps
func (volumeOpts *VolumeOpts) BuildVolumeMountsForConfigMaps(configMaps []v1alpha1.ConfigMap) *VolumeOpts {
	var volumeMountsList []corev1.VolumeMount
	if configMaps == nil {
		return volumeOpts
	}
	for _, v := range configMaps {
		var volumeMount corev1.VolumeMount
		volumeMount.Name = v.Name
		volumeMount.MountPath = v.MountPath
		volumeMountsList = append(volumeMountsList, volumeMount)
	}

	volumeOpts.VolumeMounts = append(volumeOpts.VolumeMounts, volumeMountsList...)
	return volumeOpts
}

// BuildVolumeMountsForSecrets builds VolumeMounts for Secrets
func (volumeOpts *VolumeOpts) BuildVolumeMountsForSecrets(secrets []v1alpha1.Secret) *VolumeOpts {
	var volumeMountsList []corev1.VolumeMount
	if secrets == nil {
		return volumeOpts
	}
	for _, v := range secrets {
		var volumeMount corev1.VolumeMount
		volumeMount.Name = v.Name
		volumeMount.MountPath = v.MountPath
		volumeMountsList = append(volumeMountsList, volumeMount)
	}
	volumeOpts.VolumeMounts = append(volumeOpts.VolumeMounts, volumeMountsList...)
	return volumeOpts
}

// BuildVolumeMountsForHostFileVolumes  builds VolumeMounts for HostFileVolume
func (volumeOpts *VolumeOpts) BuildVolumeMountsForHostFileVolumes(hostFileVolumes []v1alpha1.HostFile) *VolumeOpts {
	var volumeMountsList []corev1.VolumeMount
	if hostFileVolumes == nil {
		return volumeOpts
	}
	for _, v := range hostFileVolumes {
		var volumeMount corev1.VolumeMount
		volumeMount.Name = v.Name
		volumeMount.MountPath = v.MountPath
		volumeMountsList = append(volumeMountsList, volumeMount)
	}
	volumeOpts.VolumeMounts = append(volumeOpts.VolumeMounts, volumeMountsList...)
	return volumeOpts
}

// BuildVolumeBuilderForConfigMaps builds VolumeBuilders for ConfigMaps
func (volumeOpts *VolumeOpts) BuildVolumeBuilderForConfigMaps(configMaps []v1alpha1.ConfigMap) *VolumeOpts {
	volumeBuilderList := []*volume.Builder{}
	if configMaps == nil {
		return volumeOpts
	}
	for _, v := range configMaps {
		volumeBuilder := volume.NewBuilder().
			WithConfigMap(v.Name)
		volumeBuilderList = append(volumeBuilderList, volumeBuilder)
	}
	volumeOpts.VolumeBuilders = append(volumeOpts.VolumeBuilders, volumeBuilderList...)
	return volumeOpts
}

// BuildVolumeBuilderForSecrets builds VolumeBuilders for Secrets
func (volumeOpts *VolumeOpts) BuildVolumeBuilderForSecrets(secrets []v1alpha1.Secret) *VolumeOpts {
	volumeBuilderList := []*volume.Builder{}
	if secrets == nil {
		return volumeOpts
	}
	for _, v := range secrets {
		volumeBuilder := volume.NewBuilder().
			WithSecret(v.Name)
		volumeBuilderList = append(volumeBuilderList, volumeBuilder)
	}
	volumeOpts.VolumeBuilders = append(volumeOpts.VolumeBuilders, volumeBuilderList...)
	return volumeOpts
}

// BuildVolumeBuilderForHostFileVolumes builds VolumeBuilders for HostFileVolume
func (volumeOpts *VolumeOpts) BuildVolumeBuilderForHostFileVolumes(hostFileVolumes []v1alpha1.HostFile) *VolumeOpts {
	volumeBuilderList := []*volume.Builder{}
	if hostFileVolumes == nil {
		return volumeOpts
	}

	for _, v := range hostFileVolumes {
		fileType := &hostpathTypeFile
		if v.Type != "" {
			fileType = &v.Type
		}
		volumeBuilder := volume.NewBuilder().
			WithName(v.Name).
			WithHostPathAndType(
				v.NodePath,
				fileType,
			)
		volumeBuilderList = append(volumeBuilderList, volumeBuilder)
	}
	volumeOpts.VolumeBuilders = append(volumeOpts.VolumeBuilders, volumeBuilderList...)
	return volumeOpts
}
