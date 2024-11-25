package integrations

import "github.com/pier-oliviert/konditionner/pkg/konditions"

const (
	DeploymentCondition konditions.ConditionType = "Deployment"
	HealthCondition     konditions.ConditionType = "Health"

	DeploymentFinalizer = "phonebook.se.quencer.io/deployment"
	DeploymentLabel     = "phonebook.se.quencer.io/deployment"
)
