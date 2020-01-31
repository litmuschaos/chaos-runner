package utils

import (
	"k8s.io/klog"
)

func (log *LogStruct) Log(logStruct LogStruct) {
	klog.V(0).Infof("Unable to %v chaos resource: %v, of type: %v, in Namespace: %v, due to error: %v", log.Operation, log.ResourceName, log.ResourceType, log.Namespace, log.Error)
	if log.Operation == "" && log.ResourceName == "" && log.ResourceType == "" || log.Namespace == "" {
		klog.V(0).Infof(log.Error)
	}
}

func (log *LogStruct) WithOperation(operation string) *LogStruct {
	log.Operation = operation
	return log
}

func (log *LogStruct) WithResourceType(resourceType string) *LogStruct {
	log.ResourceType = resourceType
	return log
}

func (log *LogStruct) WithResourceName(resourceName string) *LogStruct {
	log.ResourceName = resourceName
	return log
}

func (log *LogStruct) WithVerbosity(verbosity int32) *LogStruct {
	log.Verbosity = verbosity
	return log
}

func (log *LogStruct) WithNameSpace(namespace string) *LogStruct {
	log.Namespace = namespace
	return log
}
func (log *LogStruct) WithError(err string) *LogStruct {
	log.Error = err
	return log
}
