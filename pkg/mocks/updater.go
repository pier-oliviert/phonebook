package mocks

import (
	"github.com/pier-oliviert/konditionner/pkg/konditions"
	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
)

type Updater struct {
	Status *konditions.ConditionStatus
	Reason *string

	Info phonebook.IntegrationInfo
}

func (u *Updater) StageCondition(status konditions.ConditionStatus, reason string) {
	u.Status = &status
	u.Reason = &reason
}

func (u *Updater) StageRemoteInfo(info phonebook.IntegrationInfo) {
	u.Info = info
}
