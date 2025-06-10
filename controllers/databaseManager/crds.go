package databasemanager

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DatabaseInstanceSpec struct {
	Type               string         `json:"type"` // Mysql, Postgres, MongoDB
	StatefulSetRef     StatefulSetRef `json:"statefulSetRef"`
	ScaleOnPodPresence bool           `json:"scaleOnPodPresence,omitempty"`
}

type StatefulSetRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type DatabaseInstanceStatus struct {
	Healthy     bool   `json:"healthy,omitempty"`
	LastChecked string `json:"lastChecked,omitempty"`
	Message     string `json:"message,omitempty"`
}

// Crd
type DatabaseInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseInstanceSpec   `json:"spec,omitempty"`
	Status DatabaseInstanceStatus `json:"status,omitempty"`
}
