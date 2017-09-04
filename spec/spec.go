package spec

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

type CouchDB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              CouchDBSpec   `json:"spec"`
	Status            CouchDBStatus `json:"status,omitempty"`
}

type CouchDBSpec struct {
	Version   string     `json:"version"`
	BaseImage string     `json:"baseImage"`
	Size      int        `json:"size"`
	Pod       *PodPolicy `json:"pod,omitempty"`
}

type PodPolicy struct {
	// Labels specifies the labels to attach to pods the operator creates for the cluster.
	Labels map[string]string `json:"labels,omitempty"`

	// NodeSelector specifies a map of key-value pairs. For the pod to be eligible
	// to run on a node, the node must have each of the indicated key-value pairs as
	// labels.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// AntiAffinity determines if the couchdb-operator tries to avoid putting
	// the couchdb members in the same cluster onto the same node.
	AntiAffinity bool `json:"antiAffinity,omitempty"`

	// List of environment variables to set in the couchdb container.
	// should container COUCHDB_USER and COUCHDB_PASSWORD. If it doesn't,
	// admin/admin will be choosen
	CouchDBEnv []apiv1.EnvVar `json:"couchdbEnv,omitempty"`
}

type CouchDBStatus struct {
	State   CouchDBState `json:"state,omitempty"`
	Message string       `json:"message,omitempty"`
}
type CouchDBState string

const (
	CouchDBStateNone      CouchDBState = "None"
	CouchDBStateProcessed CouchDBState = "Processed"
)

type CouchDBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []CouchDB `json:"items"`
}
