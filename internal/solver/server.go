package solver

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook"
	whapi "github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apiserver"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
)

// Serve requests on port 4443 for the DNS-01 Solver through Kubernete's
// APIService(1) and using self-signed certificates created by Phonebook's helm chart.
// The Service runs through HTTPS but is only used internally by Kubernetes as described by
// cert-manager's documentation on DNS-01 webhooks integration (2).
//
// Most of the code here was copied from Cert-manager's abstraction layer. By default, cert-manager expects the webhook
// to run isolated as its own binary, which means their webhook abstraction deals with arguments parsing and server configuration.
//
// While this may work for other integrations, it caused a bunch of weird issues for Phonebook as the server runs
// inside the main controller. First, the command line arguments conflicts with the ones provided for the controller
// then, the client Phonebook's solver wants to use is different than the one passed by cert-manager's abstraction.
//
// For this reason, a lot of the code was copied over so the Aggregation layer can be properly configured
// with the APIService and Phonebook can still interact with DNSRecord the same way the rest of the operator is.
//
// 1. https://kubernetes.io/docs/tasks/extend-kubernetes/configure-aggregation-layer/
// 2. https://cert-manager.io/docs/configuration/acme/dns01/webhook/
func Serve(ctx context.Context, slvr *Solver) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	opts := genericoptions.NewRecommendedOptions(
		"<UNUSED>",
		apiserver.Codecs.LegacyCodec(whapi.SchemeGroupVersion),
	)

	opts.Etcd = nil
	opts.Admission = nil
	opts.Features.EnablePriorityAndFairness = false

	opts.SecureServing.BindPort = 4443
	opts.SecureServing.ServerCert.CertKey = genericoptions.CertKey{
		CertFile: "/tls/tls.crt",
		KeyFile:  "/tls/tls.key",
	}

	if err := opts.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return fmt.Errorf("PB-SLV-#0002: error creating self-signed certificates: %v", err)
	}

	serverConfig := genericapiserver.NewRecommendedConfig(apiserver.Codecs)
	if err := opts.ApplyTo(serverConfig); err != nil {
		return fmt.Errorf("PB-SLV-#0002: %w", err)
	}

	if errs := opts.Validate(); len(errs) > 0 {
		return fmt.Errorf("PB-SLV-#0002: error validating recommended options: %v", errs)
	}

	config := &apiserver.Config{
		GenericConfig: serverConfig,
		ExtraConfig: apiserver.ExtraConfig{
			SolverGroup: slvr.Group(),
			Solvers:     []webhook.Solver{slvr},
		},
	}

	server, err := config.Complete().New()
	if err != nil {
		return err
	}

	return server.GenericAPIServer.PrepareRun().RunWithContext(ctx)
}
