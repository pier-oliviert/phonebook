package integrations

import (
	"context"
	"fmt"

	"github.com/pier-oliviert/konditionner/pkg/konditions"
	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
	"github.com/pier-oliviert/phonebook/api/v1alpha1/integrations"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type health struct {
	ctx         context.Context
	integration *phonebook.DNSIntegration
	client.Client
}

func HealthTask(ctx context.Context, c client.Client, integration *phonebook.DNSIntegration) konditions.Task {
	t := health{
		ctx:         ctx,
		integration: integration,
		Client:      c,
	}

	return t.Run
}

func (t health) Run(condition konditions.Condition) (konditions.Condition, error) {
	var deployments apps.DeploymentList
	var err error
	var label labels.Selector

	label, err = labels.Parse(fmt.Sprintf("%s=%s", integrations.DeploymentLabel, t.integration.Name))
	if err != nil {
		return condition, fmt.Errorf("PB#0006: Could not parse the label selector: %w", err)
	}

	err = t.List(t.ctx, &deployments, &client.ListOptions{
		LabelSelector: label,
		Namespace:     t.integration.Namespace,
	})
	if err != nil {
		goto Done
	}

	if deployments.Size() == 0 {
		condition.Status = konditions.ConditionStatus("Waiting")
		condition.Reason = "Waiting for deployment to be available"
		goto Done
	}
	// Although there should always only be one deployment for a given integration, the call
	// to retrieve deployments with labels returns a list, so let's pretend there's more than one
	// and if there's an unhealthy deployment, stop the loop right there and then.
	for _, d := range deployments.Items {
		log.FromContext(t.ctx).Info("Deployment", "Deployment", d)
		for _, c := range d.Status.Conditions {
			log.FromContext(t.ctx).Info("Condition", "Condition", c)
			if c.Status != core.ConditionTrue {
				err = fmt.Errorf("PB#0005: Deployment(%s) is not healthy. Check the logs of the pods for more info.", d.Name)
				goto Done
			}
		}
	}

	condition.Status = konditions.ConditionCompleted
	condition.Reason = "Healthy"

Done:
	return condition, err
}
