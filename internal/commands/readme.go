package commands

import (
	"fmt"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/chapmanjacobd/discotheque/internal/models"
)

type ReadmeCmd struct {
	models.GlobalFlags
}

func (c *ReadmeCmd) Run(ctx *kong.Context) error {
	var sb strings.Builder

	sb.WriteString("# discotheque\n\n")
	sb.WriteString("Golang implementation of xklb/library\n\n")
	sb.WriteString("## Install\n\n")
	sb.WriteString("    go install github.com/chapmanjacobd/discotheque/cmd/disco@latest\n\n")
	sb.WriteString("## Usage\n\n")

	// Iterate through subcommands
	for _, node := range ctx.Model.Children {
		if node.Hidden {
			continue
		}
		sb.WriteString(fmt.Sprintf("### %s\n\n", node.Name))
		sb.WriteString(fmt.Sprintf("%s\n\n", node.Help))
		sb.WriteString("<details><summary>Usage</summary>\n\n")
		sb.WriteString("```bash\n")
		sb.WriteString(fmt.Sprintf("$ disco %s -h\n", node.Name))
		sb.WriteString("```\n\n")
		sb.WriteString("</details>\n\n")
	}

	fmt.Print(sb.String())
	return nil
}
