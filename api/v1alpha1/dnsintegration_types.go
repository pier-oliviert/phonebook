/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"github.com/pier-oliviert/konditionner/pkg/konditions"
	"github.com/pier-oliviert/phonebook/api/v1alpha1/references"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// A DNSIntegrationSpec represents the bridge between Phonebook's DNSRecord
// and the cloud provider's client that will be in charge of those Records.
// When a DNSIntegration is created, it will create a new deployment using a
// Provider's image. The Deployment will then be in charge of any DNSRecord that
// matches its Provider and Zone, as specified in the DNSRecord.
type DNSIntegrationSpec struct {
	// Provider that backs this DNSIntegration, ie. cloudflare, aws, azure, etc.
	// This field is used to figure out what Client to initialize and configure.
	Provider DNSProviderSpec `json:"provider"`

	// Zones for which this integration has authority over. However, it doesn't mean
	// that this provider has exclusivity over the zones. One example would be for
	// Split-Horizon DNS (1) where the same Zone can be managed by different providers.
	//
	// A Provider can own multiple zones. When a DNSRecord is created, it will look for
	// a provider if the optional value is set. After, it will look at the DNSRecord's zone
	// and attempt to match it against one of the zone listed here. If there's a match,
	// the record will be processed by the Provider.
	//
	// 1. https://en.wikipedia.org/wiki/Split-horizon_DNS
	Zones []string `json:"zones"`

	// Env are passed directly to the Provider as Environment Variables for the deployment. This can
	// be useful for configurations. It uses native env structure as defined in K8s' docs(1).
	//
	// It can be useful to source environment variables from config or to set them directly too.
	// If you want to source environment variable from secrets, you may use `secretRef` instead as
	// it is simpler to use, but that's up to you.
	//
	// 1. https://kubernetes.io/docs/tasks/inject-data-application/define-environment-variable-container/
	Env []core.EnvVar `json:"env,omitempty"`

	// A reference to a Kubernetes Secret that will be passed to the Provider. Each keys
	// defined will be exported as an environment variable to the provider's deployment.
	// The SecretRef has precedence over the `Env` field so any keys specified here will override
	// values that would otherwise be defined in the `Env` field.
	SecretRef *references.SecretRef `json:"secretRef,omitempty"`
}

type DNSProviderSpec struct {
	// Name of the provider as specified in the documentation, ie. cloudflare, aws, azure, etc.
	// The name has to be a direct match
	Name string `json:"name"`

	// Image name if you want to use a different image name than the default one used
	// by Phonebook. If this value isn't set, Phonebook will generate an image name
	// based off the Provider's name and Phonebook's default repository.
	// It will also always use the `latest` tag
	Image *string `json:"image,omitempty"`

	// Command can be spceifici
	Command []string `json:"cmd,omitempty"`

	Args []string `json:"args,omitempty"`
}

// DNSProviderStatus defines the observed state of DNSProvider
type DNSIntegrationStatus struct {
	// Set of conditions that the DNSRecord will go through during its
	// lifecycle.
	Conditions konditions.Conditions `json:"conditions,omitempty"`

	// Reference to the deployment that was created for this
	// Integration.
	Deployment *references.Reference `json:"deployment,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=dnsintegrations,scope=Cluster
// +kubebuilder:subresource:status
// DNSProvider is the Schema for the dnsproviders API
type DNSIntegration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DNSIntegrationSpec   `json:"spec,omitempty"`
	Status DNSIntegrationStatus `json:"status,omitempty"`
}

func (d *DNSIntegration) Conditions() *konditions.Conditions {
	return &d.Status.Conditions
}

// +kubebuilder:object:root=true

// DNSProviderList contains a list of DNSProvider
type DNSIntegrationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DNSIntegration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DNSIntegration{}, &DNSIntegrationList{})
}
