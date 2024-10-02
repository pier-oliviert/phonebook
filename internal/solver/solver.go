package solver

import (
	"context"
	"fmt"

	whapi "github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// The Label's key that identify a DNSRecord as part of this solver
	kChallengeLabel string = "phonebook.se.quencer.io/solver"

	// The type of challenge the solver has created. Currently, only
	// one type of challenge is supported (dns-01) and the way the label is
	// created isn't really to be extensible but rather be human-readable so someone
	// can look at the resources in their cluster and clearly understands
	// what this label is for.
	kChallengeKey string = "dns-01-challenge"
)

type Solver struct {
	group string
	name  string

	client.Client
}

// Create a new solver to be used with Phonebook.
//
// The Name of the solver is the name of the interface as defined by RFC2136.
// This is not a user-defined value but rather the name defined in the helm chart that is
// Phonebook specific.
//
// The manager is used to get information about the API Group and to retrieve its fully
// configured client.
//
// The Solver returned is fully configured and ready to go. It won't start
// accepting challenges until `Run()` is called on the solver.
func NewSolver(name string, c client.Client) *Solver {
	return &Solver{
		name:   name,
		group:  fmt.Sprintf("phonebook.%s", phonebook.GroupVersion.Group),
		Client: c,
	}
}

// Initialize is a no-op to complete the Solver's Interface.
// The client is not needed as one was already provided by the manager
// during the creation of the solver.
func (s *Solver) Initialize(c *rest.Config, stopCh <-chan struct{}) error {
	return nil
}

// Runs the Webhook Server with the solver
//
// This method blocks as it starts the server.
func (s *Solver) Run(ctx context.Context) error {
	return Serve(ctx, s)
}

// Name is the name specified by Phonebook and is required to match
// the APIServer resource created by the helm chart.
func (s *Solver) Name() string {
	return s.name
}

func (s *Solver) Group() string {
	return s.group
}

// Present a challenge to the solver. The Challenge Request comes from
// cert-manager through the webhook integration. Once presented with a challenge,
// a DNS Record needs to be created with the challenge information so cert-manager
// can assert that the domain is owned by user of Phonebook.
//
// The DNSRecord created for a challenge will have a generic "challenge" label
// added to it. This will allow cleanup operations to get a list of all challenges
// and remove those that needs to be cleaned up. Now, in an ideal world, the label
// would have been specific enough so cleanup duties would be able to retrieve only
// the proper challenge records. For example, the label could hold the ResolvedFQDN and
// the cleanup could retrieve DNS Records by ResolvedFQDN. It is, however, impossible
// to implement safely as Labels have a stricter validation set(1) compared to FQDN(2).
//
// The tradeoff is to label challenges and do client-side filtering inside the CleanUp
// method instead. It is reasonable to think that the number of challenges at any given
// time is pretty low, the bandwidth/CPU to filter those records on the client side seems
// acceptable at this time.
//
// 1. https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set
// 2. https://www.ietf.org/rfc/rfc1034.txt
func (s *Solver) Present(ch *whapi.ChallengeRequest) error {
	ctx := context.Background()

	ep := &phonebook.DNSRecord{
		ObjectMeta: meta.ObjectMeta{
			Namespace:    ch.ResourceNamespace,
			GenerateName: "challenge-",
			Labels: map[string]string{
				kChallengeLabel: kChallengeKey,
			},
		},
		Spec: phonebook.DNSRecordSpec{
			RecordType: "TXT",
			Name:       ch.ResolvedFQDN,
			Targets:    []string{ch.Key},
		},
	}
	return s.Create(ctx, ep)
}

// Request to clean up the request after a success/failure.
// At this point, the DNSRecord was possibly created and it needs to be
// deleted so Phonebook can remove it from the Provider.
//
// As described in the Present method, records for challenges have a generic label
// associated with them. The CleanUp method needs to retrieve all dsn records that includes
// this label and do client side filtering to only delete records that have a matching
// challenge Key.
//
// It's possible that the DNSRecord was already deleted, or that more than 1 record with the same key
// exists. All in all, any record that has a matching label & challenge key needs to be deleted
// when this method returns.
//
// Deleting the record will have Phonebook run through the finalizer and delete the record
// on the provider's side.
func (s *Solver) CleanUp(ch *whapi.ChallengeRequest) error {
	ctx := context.Background()

	var challenges phonebook.DNSRecordList

	label, err := labels.Parse(fmt.Sprintf("%s=%s", kChallengeLabel, kChallengeKey))
	if err != nil {
		return fmt.Errorf("PB-SLV-#0003: failed to parse the label selector -- %w", err)
	}

	opts := client.ListOptions{
		LabelSelector: label,
		Namespace:     ch.ResourceNamespace,
	}

	err = s.List(ctx, &challenges, &opts)
	if err != nil {
		return err
	}

	// Delete only the DNSEndpoint that has the same key. It's unlikely there's more than one, but since
	// it's a list, let's process all of them.
	for _, record := range challenges.Items {
		if len(record.Spec.Targets) != 1 {
			// Return error
			return fmt.Errorf("PB-SLV-0001: Record unexpectedly had more than one target: %v", record.Spec.Targets)
		}

		if record.Spec.Targets[0] == ch.Key {
			if err := s.Delete(ctx, &record); err != nil {
				return err
			}
		}
	}

	return nil
}
