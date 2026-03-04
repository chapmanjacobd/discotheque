package commands

import (
	"os"

	"github.com/chapmanjacobd/discotheque/internal/utils"
)

type VersionCmd struct{}

func (c *VersionCmd) Run() error {
	utils.RenderVersion(os.Stdout)
	return nil
}
