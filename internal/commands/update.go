package commands

import (
	"github.com/chapmanjacobd/discotheque/internal/utils"
)

type UpdateCmd struct{}

func (c *UpdateCmd) Run() error {
	utils.MaybeUpdate()
	return nil
}
