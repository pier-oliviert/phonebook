package integrations

import (
	"context"
	"fmt"
	"strings"

	"github.com/pier-oliviert/konditionner/pkg/konditions"
	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
	"github.com/pier-oliviert/phonebook/api/v1alpha1/integrations"
	"github.com/pier-oliviert/phonebook/api/v1alpha1/references"
	"github.com/pier-oliviert/phonebook/pkg/providers"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/env"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type deployment struct {
	ctx         context.Context
	integration *phonebook.DNSIntegration
	client.Client
}

func DeploymentTask(ctx context.Context, c client.Client, integration *phonebook.DNSIntegration) konditions.Task {
	t := deployment{
		ctx:         ctx,
		integration: integration,
		Client:      c,
	}

	return t.Run
}

func (t deployment) Run(condition konditions.Condition) (konditions.Condition, error) {
	d := t.deployment()
	if err := t.Create(t.ctx, d); err != nil {
		return condition, err
	}

	t.integration.Status.Deployment = references.NewReference(d)
	condition.Status = konditions.ConditionCreated
	condition.Reason = fmt.Sprintf("Deployment Created: %s", d.Name)

	return condition, nil
}

func (t deployment) deployment() *apps.Deployment {
	img := ""
	if t.integration.Spec.Provider.Image != nil {
		img = *t.integration.Spec.Provider.Image
	} else {
		img = providers.ProviderImages[t.integration.Spec.Provider.Name]
	}

	envs := []core.EnvVar{}
	if len(t.integration.Spec.Env) != 0 {
		envs = append(envs, t.integration.Spec.Env...)
	}

	if t.integration.Spec.SecretRef != nil {
		secret := t.integration.Spec.SecretRef
		for _, sk := range secret.Keys {
			envs = append(envs, core.EnvVar{
				Name: sk.Name,
				ValueFrom: &core.EnvVarSource{
					SecretKeyRef: &core.SecretKeySelector{
						LocalObjectReference: secret.Selector(),
						Key:                  sk.Key,
					},
				},
			})
		}
	}

	envs = append(envs,
		core.EnvVar{
			Name:  "PB_INTEGRATION",
			Value: t.integration.Name,
		}, core.EnvVar{
			Name:  "PB_ZONES",
			Value: strings.Join(t.integration.Spec.Zones, ","),
		},
	)

	container := core.Container{
		Name:            "provider",
		Env:             envs,
		Image:           img,
		ImagePullPolicy: core.PullIfNotPresent,
	}

	if len(t.integration.Spec.Provider.Command) != 0 {
		container.Command = t.integration.Spec.Provider.Command
	}

	if len(t.integration.Spec.Provider.Args) != 0 {
		container.Args = t.integration.Spec.Provider.Args
	}

	var replicaCount int32 = 1

	deployment := &apps.Deployment{
		ObjectMeta: meta.ObjectMeta{
			Name:      fmt.Sprintf("provider-%s", t.integration.Name),
			Namespace: env.GetString("PB_NAMESPACE", "phonebook-system"),
			OwnerReferences: []meta.OwnerReference{{
				APIVersion: t.integration.APIVersion,
				Kind:       t.integration.Kind,
				Name:       t.integration.Name,
				UID:        t.integration.UID,
			}},
			Labels: map[string]string{
				integrations.DeploymentLabel: t.integration.Name,
			},
		},
		Spec: apps.DeploymentSpec{
			Replicas: &replicaCount,
			Selector: &meta.LabelSelector{
				MatchLabels: map[string]string{
					integrations.DeploymentLabel: t.integration.Name,
				},
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: meta.ObjectMeta{
					Labels: map[string]string{
						integrations.DeploymentLabel: t.integration.Name,
					},
				},
				Spec: core.PodSpec{
					ServiceAccountName: env.GetString("PB_PROVIDER_SERVICE_ACC", "phonebook-providers"),
					Containers:         []core.Container{container},
				},
			},
		},
	}

	return deployment
}
