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

	"github.com/pier-oliviert/phonebook/pkg/providers/aws"
	"github.com/pier-oliviert/phonebook/pkg/server"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func main() {
	var err error

	ctx := context.Background()
	logger := log.FromContext(ctx)

	logger.Info("Initializing AWS Client")
	p, err := aws.NewClient(ctx)
	if err != nil {
		panic(err)
	}

	srv := server.NewServer(p)
	if err := srv.Run(); err != nil {
		panic(err)
	}
}
