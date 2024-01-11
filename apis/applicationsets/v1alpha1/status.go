package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// ArgoApplicationSetStatus are the observable fields of a ApplicationSet.
type ArgoApplicationSetStatus struct {
	Conditions        []ApplicationSetCondition         `json:"conditions,omitempty" protobuf:"bytes,1,name=conditions"`
	ApplicationStatus []ApplicationSetApplicationStatus `json:"applicationStatus,omitempty" protobuf:"bytes,2,name=applicationStatus"`
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
}

// ApplicationSetConditionStatus is a type which represents possible comparison results
type ApplicationSetConditionStatus string

// ApplicationSetConditionType represents type of application condition. Type name has following convention:
// prefix "Error" means error condition
// prefix "Warning" means warning condition
// prefix "Info" means informational condition
type ApplicationSetConditionType string
