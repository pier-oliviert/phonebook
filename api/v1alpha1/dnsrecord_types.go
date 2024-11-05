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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// The main condition to talk to the provider. Each provider have finer
	// states that will be reflected as status for this condition.
	ProviderCondition konditions.ConditionType = "Provider"

	IntegrationCondition konditions.ConditionType = "Integration"
)

// DNSRecordSpec defines the desired state of DNSRecord and represents
// a single DNS Record. It is expected that each DNS Record won't conflict with each other
// and it's the user's job to make sure that each record have a unique spec.
type DNSRecordSpec struct {
	// Zone is the the DNS Zone that you want to create a record for.
	// If you want to create a CNAME called foo.mydomain.com,
	// "mydomain.com" would be your zone.
	//
	// The Zone needs to find a match in one of the DNSProvider configured in your
	// cluster. Unless the optional `Provider` field is set, Phonebook will look
	// at all the providers configured to try to find a match for the zone.
	//
	// If no provider matches the zone, the record won't be created.
	Zone string `json:"zone"`

	// RecordType represent the type for the Record you want to create.
	// Can be A, AAAA, CNAME, TXT, etc.
	RecordType string `json:"recordType"`

	// Name of the record represents the subdomain in the CNAME example used for zone.
	// In that example, the `Name` would be `foo`
	Name string `json:"name"`

	// Targets represents where the record should point to. Depending on the record type,
	// it can be an IP address or some text value.
	// The reason why targets is plural is because some provider support multiple values for
	// a given record types. For most cases, it's expected to only have 1 value.
	Targets []string `json:"targets"`

	// Provider specific configuration settings that can be used
	// to configure a DNS Record in accordance to the provider used.
	// Each provider provides its own set of custom fields.
	Properties map[string]string `json:"properties,omitempty"`

	// TTL is the Time To Live for the record. It represents the time
	// in seconds that the record is cached by resolvers.
	// If not set, the provider will use its default value (60 seconds).
	TTL *int64 `json:"ttl,omitempty"`

	// Optional field to be more specific about which Provider you want to use for
	// this record. This field is useful if you have more than one Provider serving
	// the same Zone (ie. Split-Horizon DNS).
	//
	// In most cases, this field isn't necessary as the Zone field should be enough
	// to let Phonebook find the proper Provider. This field only gives a hint to Phonebook
	// and the Zones has to match as well.
	Integration *string `json:"integration,omitempty"`
}

// Optional field that a provider can use to keep track of remote data it might need in the future, eg. Remote ID for deleting the
// record. Values can only be string.
type IntegrationInfo map[string]string

// StagingUpdater is an interface used by providers to safely update DNSRecord's status. Since
// DNSRecord can interact with multiple DNSIntegrations, the DNSRecord's status needs to be scoped for
// each DNSIntegration so they can all keep the DNSRecord status updated without conflicting with each other.
//
// StagingUpdater is a proxy that will scope all changes to the specific condition/IntegrationInfo. It's rather simple,
// each DNSIntegration that has authority over the zone will have its own Condition in the DSNRecord's Conditions as well
// its own entry in the IntegrationInfo map. Both of those will have the name of the DNSIntegration as the unique key,
// which means that even if there's more than one integration for a given provider (aws, cloudflare, etc.), the uniqueness
// of the key is still valid.
//
// As the name of the methods and interface suggest, these operations are only staging. As a result, they aren't persisted
// when those method returns. In fact, multiple calls overwrite the previously set values during the **same reconciliation loop**.
//
// When Create/Destroy returns, the server's reconciliation loop will update the Condition and the IntegrationInfo.
//
// It is important to note that in case of an error returning from Create/Destroy, the error will take precedence over
// the staged condition. The Status will be set to Error and the reason will be set to the error message.
// +kubebuilder:object:generate=false
type StagingUpdater interface {
	// StageCondition lets an integration update the condition that is attached to the
	// DNSIntegration.
	// For instance, if the DNSIntegration was created with the name `my-test-123` and the provider
	// is AWS, calling
	//   StageCondition(konditions.ConditionCreated, "Resource created")
	//
	// would mean that the condition with the condition type `my-test-123` will
	// have the status and reason set at the end of the reconciliation loop.
	StageCondition(status konditions.ConditionStatus, reason string)

	// StagingRemoteInfo lets the provider store provider-related information in
	// the DNSRecord's RemoteInfo field. This is an optional field.
	StageRemoteInfo(IntegrationInfo)
}

// DNSRecordStatus defines the observed state of DNSRecord
type DNSRecordStatus struct {
	// Set of conditions that the DNSRecord will go through during its
	// lifecycle.
	Conditions konditions.Conditions `json:"conditions,omitempty"`

	// RemoteInfo is a field that can be used by DNSIntegration's provider to
	// store information as the Record is created. Each integration has its own map it can
	// populate with arbitrary data. Each entries in the root RemoteInfo refers to the name of
	// the integration that stored the intormation. For instance, if you have a DNSRecord that
	// is shared between 2 integrations named `cloudflare-dev` and `aws-prod`, RemoteInfo would
	// look like this:
	//    map[string]map[string]string{
	//      "cloudflare-dev": map[string]string{
	//        // cloudflare related information about the record
	//      },
	//      "aws-prod": map[string]string{
	//        // aws related information about the record
	//      }
	//    }
	//
	// A DNSIntegration can have multiple entries stored in this field and it's up the integration
	// to make sure those fields are not stale.
	RemoteInfo map[string]IntegrationInfo `json:"remoteInfo,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DNSRecord is the Schema for the dnsrecords API
type DNSRecord struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DNSRecordSpec   `json:"spec,omitempty"`
	Status DNSRecordStatus `json:"status,omitempty"`
}

// This helper method is added to DNSRecord to make it match the
// konditions.ConditionalObject interface to use the Lock mechanism
// with konditionner.
func (d *DNSRecord) Conditions() *konditions.Conditions {
	return &d.Status.Conditions
}

// +kubebuilder:object:root=true

// DNSRecordList contains a list of DNSRecord
type DNSRecordList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DNSRecord `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DNSRecord{}, &DNSRecordList{})
}
