package references

import (
	core "k8s.io/api/core/v1"
)

// +kubebuilder:object:generate=true
type SecretRef struct {
	Keys []SecretKey `json:"keys"`
	Name string      `json:"name"`
}

func (sr SecretRef) Selector() core.LocalObjectReference {
	return core.LocalObjectReference{
		Name: sr.Name,
	}
}

// +kubebuilder:object:generate=true
type SecretKey struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}
