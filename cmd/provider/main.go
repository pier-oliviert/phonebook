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

package main

import (
	"context"
	"fmt"

	"github.com/pier-oliviert/phonebook/pkg/provider/server"
	"github.com/pier-oliviert/phonebook/pkg/providers"
	"github.com/pier-oliviert/phonebook/pkg/providers/aws"
	"github.com/pier-oliviert/phonebook/pkg/providers/cloudflare"
	"github.com/pier-oliviert/phonebook/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	PhonebookProvider string = "PB_PROVIDER"
)

func main() {
	var err error

	name, err := utils.RetrieveValueFromEnvOrFile(PhonebookProvider)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	logger := log.FromContext(ctx)
	logger.Info("Initializing provider", "Name", name)

	var p providers.Provider

	switch name {
	case "aws":
		p, err = aws.NewClient(ctx)
	case "cloudflare":
		p, err = cloudflare.NewClient(ctx)
	case "":
		panic(fmt.Errorf("PB#0001: The environment variable %s need to be set with a valid provider name", PhonebookProvider))
	default:
		panic(fmt.Errorf("PB#0001: The environment variable %s need to be set with a valid provider name, got %s", PhonebookProvider, name))
	}

	if err != nil {
		panic(err)
	}

	srv := server.NewServer(p)
	if err := srv.Run(); err != nil {
		panic(err)
	}
}
