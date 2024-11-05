package server

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"strings"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/utils/env"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
	reconcilers "github.com/pier-oliviert/phonebook/internal/reconcilers/provider"
	"github.com/pier-oliviert/phonebook/pkg/providers"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(phonebook.AddToScheme(scheme))
}

type Server interface {
	Run() error
}

func NewServer(p providers.Provider) Server {
	s := &server{}
	s.Store(p)
	return s
}

type server struct {
	providers.ProviderStore
}

func (s *server) Run() error {
	var tlsOpts []func(*tls.Config)

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	logger := log.FromContext(context.Background())

	disableHTTP2 := func(c *tls.Config) {
		logger.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	tlsOpts = append(tlsOpts, disableHTTP2)
	integration := env.GetString("PB_INTEGRATION", "")
	zones := strings.Split(env.GetString("PB_ZONES", ""), ",")

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                        scheme,
		HealthProbeBindAddress:        ":8081",
		LeaderElection:                true,
		LeaderElectionID:              fmt.Sprintf("%s-provider.phonebook.se.quencer.io", integration),
		LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		return fmt.Errorf("PB#0004: Unable to start manager -- %w", err)
	}

	if err = s.Provider().Configure(context.Background(), integration, zones); err != nil {
		// Error coming from a Provider should already be coded, so returning it as is.
		return err
	}

	if err = (&reconcilers.ProviderReconciler{
		Integration:   integration,
		Store:         &s.ProviderStore,
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		EventRecorder: mgr.GetEventRecorderFor("dnsrecord"),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("PB#0004: Unable to create controller -- %w", err)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return fmt.Errorf("PB#0004: Unable to set up health check -- %w", err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return fmt.Errorf("PB#0004: Unable to set up ready check -- %w", err)
	}

	logger.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("PB#0004: Could not start controller -- %w", err)
	}

	return nil
}
