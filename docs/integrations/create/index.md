---
title: 'Create a provider'
date: 2024-10-18T10:38:15-04:00
draft: false
weight: 100
---

The `DNSIntegration` CRD was created to allow anyone to create their own provider and manage DNSRecord using Phonebook. Here you will find all the information to create your own provider.

## Concepts

To understand how integrations work, it's best to visualize Phonebook as a disjointed Kubernetes Operator. Each integration runs a deployment that register a new operator that will listen for `DNSRecord` and only act on the ones that fits their integration profile.

Phonebook's main controller has the responsibility of validating new DNSRecord as well as ensuring that records can be safely garbage collected when deleted. All the actual operations between a DNSRecord and a DNS Provider is done through the DNSIntegration's deployment.

## A Provider's main function

```go {filename="main.go"}
package main

import (
	"context"

	"github.com/pier-oliviert/phonebook/pkg/providers/cloudflare"
	"github.com/pier-oliviert/phonebook/pkg/server"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func main() {
	var err error

	ctx := context.Background()
	logger := log.FromContext(ctx)

	logger.Info("Initializing My New Client")

    // Replace this with your provider's client
    // The client needs to implement the providers.Provider interface
	p, err := cloudflare.NewClient(ctx)
	if err != nil {
		panic(err)
	}

	srv := server.NewServer(p)
	if err := srv.Run(); err != nil {
		panic(err)
	}
}
```

The server that Phonebook provides is a fully configured operator. The call `srv.Run()` will block and will then pass off all the request to the client.
