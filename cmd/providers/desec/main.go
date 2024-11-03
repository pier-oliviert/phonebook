package main

import (
	"context"

	"github.com/pier-oliviert/phonebook/pkg/providers/desec"
	"github.com/pier-oliviert/phonebook/pkg/server"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func main() {
	var err error

	ctx := context.Background()
	logger := log.FromContext(ctx)

	logger.Info("Initializing deSEC Client")
	p, err := desec.NewClient(ctx)
	if err != nil {
		panic(err)
	}

	srv := server.NewServer(p)
	if err := srv.Run(); err != nil {
		panic(err)
	}
}
