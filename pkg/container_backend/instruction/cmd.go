package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend/build_context"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
)

type Cmd struct {
	dockerfile_instruction.Cmd
}

func NewCmd(i dockerfile_instruction.Cmd) *Cmd {
	return &Cmd{Cmd: i}
}

func (i *Cmd) UsesBuildContext() bool {
	return false
}

func (i *Cmd) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContext *build_context.BuildContext) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{CommonOpts: drvOpts, Cmd: i.Cmd.Cmd}); err != nil {
		return fmt.Errorf("error setting cmd %v for container %s: %w", i.Cmd, containerName, err)
	}
	return nil
}
