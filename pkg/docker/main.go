package docker

import (
	"os"

	"github.com/docker/cli/cli/command"
	cliconfig "github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/term"
	"github.com/flant/werf/pkg/logger"
	"golang.org/x/net/context"
)

var (
	cli       *command.DockerCli
	apiClient *client.Client
)

func Init(dockerConfigDir string) error {
	if dockerConfigDir != "" {
		cliconfig.SetDir(dockerConfigDir)
	}

	if err := setDockerClient(); err != nil {
		return err
	}

	if err := setDockerApiClient(); err != nil {
		return err
	}

	return nil
}

func ServerVersion() (*types.Version, error) {
	ctx := context.Background()
	version, err := apiClient.ServerVersion(ctx)
	if err != nil {
		return nil, err
	}

	return &version, nil
}

func setDockerClient() error {
	stdIn, _, _ := term.StdStreams()
	cli = command.NewDockerCli(stdIn, logger.GetOutStream(), logger.GetErrStream(), false)
	opts := flags.NewClientOptions()
	if err := cli.Initialize(opts); err != nil {
		return err
	}

	return nil
}

func setDockerApiClient() error {
	ctx := context.Background()
	serverVersion, err := cli.Client().ServerVersion(ctx)
	apiClient, err = client.NewClientWithOpts(client.WithVersion(serverVersion.APIVersion))
	if err != nil {
		return err
	}

	return nil
}

func Debug() bool {
	return os.Getenv("WERF_DEBUG_DOCKER") == "1"
}
