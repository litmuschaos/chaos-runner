package utils

import (
	"k8s.io/klog"
)

func (log *LogStruct) Log() {
	if log.Operation == "" && log.ResourceName == "" && log.ResourceType == "" && log.Namespace == "" {
		if log.Verbosity == 0 {
			klog.V(0).Infof(log.String)
		} else if log.Verbosity == 1 {
			klog.V(1).Infof(log.String)
		} else if log.Verbosity == 2 {
			klog.V(2).Infof(log.String)
		}

	}
	if log.Verbosity == 1 {
		klog.V(0).Infof("Unable to %v chaos resource: %v, of type: %v, in Namespace: %v", log.Operation, log.ResourceName, log.ResourceType, log.Namespace)
		klog.V(1).Infof("Unable to %v chaos resource: %v, of type: %v, in Namespace: %v, due to error: %v", log.Operation, log.ResourceName, log.ResourceType, log.Namespace, log.String)

	}
	clearLog(log)
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
func (log *LogStruct) WithString(str string) *LogStruct {
	log.String = str
	return log
}

func clearLog(log *LogStruct) {
	log.Namespace = ""
	log.Operation = ""
	log.ResourceName = ""
	log.ResourceType = ""
	log.Verbosity = -1
}
