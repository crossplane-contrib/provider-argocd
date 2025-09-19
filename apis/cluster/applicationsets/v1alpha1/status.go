package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ArgoApplicationSetStatus are the observable fields of a ApplicationSet.
type ArgoApplicationSetStatus struct {
	Conditions        []ApplicationSetCondition         `json:"conditions,omitempty" protobuf:"bytes,1,name=conditions"`
	ApplicationStatus []ApplicationSetApplicationStatus `json:"applicationStatus,omitempty" protobuf:"bytes,2,name=applicationStatus"`
	// Resources is a list of Applications resources managed by this application set.
	Resources []ResourceStatus `json:"resources,omitempty" protobuf:"bytes,3,opt,name=resources"`
}

// ApplicationSetCondition contains details about an applicationset condition, which is usually an error or warning
type ApplicationSetCondition struct {
	// Type is an applicationset condition type
	Type ApplicationSetConditionType `json:"type" protobuf:"bytes,1,opt,name=type"`
	// Message contains human-readable message indicating details about condition
	Message string `json:"message" protobuf:"bytes,2,opt,name=message"`
	// LastTransitionTime is the time the condition was last observed
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,3,opt,name=lastTransitionTime"`
	// True/False/Unknown
	Status ApplicationSetConditionStatus `json:"status" protobuf:"bytes,4,opt,name=status"`
	// Single word camelcase representing the reason for the status eg ErrorOccurred
	Reason string `json:"reason" protobuf:"bytes,5,opt,name=reason"`
}

// ApplicationSetApplicationStatus contains details about each Application managed by the ApplicationSet
type ApplicationSetApplicationStatus struct {
	// Application contains the name of the Application resource
	Application string `json:"application" protobuf:"bytes,1,opt,name=application"`
	// LastTransitionTime is the time the status was last updated
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,2,opt,name=lastTransitionTime"`
	// Message contains human-readable message indicating details about the status
	Message string `json:"message" protobuf:"bytes,3,opt,name=message"`
	// Status contains the AppSet's perceived status of the managed Application resource: (Waiting, Pending, Progressing, Healthy)
	Status string `json:"status" protobuf:"bytes,4,opt,name=status"`
	// Step tracks which step this Application should be updated in
	Step string `json:"step" protobuf:"bytes,5,opt,name=step"`
	// TargetRevision tracks the desired revisions the Application should be synced to.
	TargetRevisions []string `json:"targetRevisions" protobuf:"bytes,6,opt,name=targetrevisions"`
}

// ApplicationSetConditionStatus is a type which represents possible comparison results
type ApplicationSetConditionStatus string

// ApplicationSetConditionType represents type of application condition. Type name has following convention:
// prefix "Error" means error condition
// prefix "Warning" means warning condition
// prefix "Info" means informational condition
type ApplicationSetConditionType string

type ResourceStatus struct {
	Group                        string        `json:"group,omitempty" protobuf:"bytes,1,opt,name=group"`
	Version                      string        `json:"version,omitempty" protobuf:"bytes,2,opt,name=version"`
	Kind                         string        `json:"kind,omitempty" protobuf:"bytes,3,opt,name=kind"`
	Namespace                    string        `json:"namespace,omitempty" protobuf:"bytes,4,opt,name=namespace"`
	Name                         string        `json:"name,omitempty" protobuf:"bytes,5,opt,name=name"`
	Status                       string        `json:"status,omitempty" protobuf:"bytes,6,opt,name=status"`
	Health                       *HealthStatus `json:"health,omitempty" protobuf:"bytes,7,opt,name=health"`
	Hook                         bool          `json:"hook,omitempty" protobuf:"bytes,8,opt,name=hook"`
	RequiresPruning              bool          `json:"requiresPruning,omitempty" protobuf:"bytes,9,opt,name=requiresPruning"`
	SyncWave                     int64         `json:"syncWave,omitempty" protobuf:"bytes,10,opt,name=syncWave"`
	RequiresDeletionConfirmation bool          `json:"requiresDeletionConfirmation,omitempty" protobuf:"bytes,11,opt,name=requiresDeletionConfirmation"`
}

type HealthStatus struct {
	// Status holds the status code of the application or resource
	Status string `json:"status,omitempty" protobuf:"bytes,1,opt,name=status"`
	// Message is a human-readable informational message describing the health status
	Message string `json:"message,omitempty" protobuf:"bytes,2,opt,name=message"`
	// LastTransitionTime is the time the HealthStatus was set or updated
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,3,opt,name=lastTransitionTime"`
}
