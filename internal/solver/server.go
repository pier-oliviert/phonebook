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
		return fmt.Errorf("error creating self-signed certificates: %v", err)
	}

	serverConfig := genericapiserver.NewRecommendedConfig(apiserver.Codecs)
	if err := opts.ApplyTo(serverConfig); err != nil {
		return err
	}

	if errs := opts.Validate(); len(errs) > 0 {
		return fmt.Errorf("error validating recommended options: %v", errs)
	}

	config := &apiserver.Config{
		GenericConfig: serverConfig,
		ExtraConfig: apiserver.ExtraConfig{
			SolverGroup: fmt.Sprintf("phonebook.%s", slvr.Group()),
			Solvers:     []webhook.Solver{slvr},
		},
	}

	server, err := config.Complete().New()
	if err != nil {
		return err
	}

	return server.GenericAPIServer.PrepareRun().RunWithContext(ctx)
}
