package integrations

import "github.com/pier-oliviert/konditionner/pkg/konditions"

const (
	DeploymentCondition konditions.ConditionType = "Deployment"
	ServiceCondition    konditions.ConditionType = "Service"

	DeploymentFinalizer = "phonebook.se.quencer.io/deployment"
	DeploymentLabel     = "phonebook.se.quencer.io/deployment"
)
