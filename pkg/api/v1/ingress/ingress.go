package ingress

import metav1 "github.com/lacodon/recoon/pkg/api/v1/meta"

type Ingress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec *Spec `json:"spec,omitempty"`
}

type Spec struct {
	Domain        string `json:"domain,omitempty"`
	ContainerPort int32  `json:"containerPort,omitempty"`
	EnableTLS     bool   `json:"enableTLS,omitempty"`
}
