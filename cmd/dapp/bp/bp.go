package bp

import (
	"fmt"
	"os"

	"github.com/flant/dapp/cmd/dapp/common"
	"github.com/flant/dapp/cmd/dapp/docker_authorizer"
	"github.com/flant/dapp/pkg/build"
	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/docker"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/dapp/pkg/ssh_agent"
	"github.com/flant/dapp/pkg/true_git"
	"github.com/spf13/cobra"
)

var CmdData struct {
	Repo       string
	WithStages bool

	PullUsername     string
	PullPassword     string
	PushUsername     string
	PushPassword     string
	RegistryUsername string
	RegistryPassword string
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "bp [DIMG_NAME...]",
		RunE: func(cmd *cobra.Command, args []string) error {
			if CmdData.PullUsername == "" {
				CmdData.PullUsername = CmdData.RegistryUsername
			}
			if CmdData.PullPassword == "" {
				CmdData.PullPassword = CmdData.RegistryPassword
			}
			if CmdData.PushUsername == "" {
				CmdData.PushUsername = CmdData.RegistryUsername
			}
			if CmdData.PushPassword == "" {
				CmdData.PushPassword = CmdData.RegistryPassword
			}

			err := runBP(args)
			if err != nil {
				return fmt.Errorf("bp failed: %s", err)
			}
			return nil
		},
	}

	common.SetupName(&CommonCmdData, cmd)
	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)
	common.SetupSSHKey(&CommonCmdData, cmd)

	cmd.PersistentFlags().StringVarP(&CmdData.Repo, "repo", "", "", "Docker repository name to push images to. CI_REGISTRY_IMAGE will be used by default if available.")
	cmd.PersistentFlags().BoolVarP(&CmdData.WithStages, "with-stages", "", false, "Push images with stages cache")

	cmd.PersistentFlags().StringVarP(&CmdData.PullUsername, "pull-username", "", "", "Docker registry username to authorize pull of base images")
	cmd.PersistentFlags().StringVarP(&CmdData.PullPassword, "pull-password", "", "", "Docker registry password to authorize pull of base images")
	cmd.PersistentFlags().StringVarP(&CmdData.PushUsername, "push-username", "", "", "Docker registry username to authorize push to the docker repo")
	cmd.PersistentFlags().StringVarP(&CmdData.PushPassword, "push-password", "", "", "Docker registry password to authorize push to the docker repo")
	cmd.PersistentFlags().StringVarP(&CmdData.RegistryUsername, "registry-username", "", "", "Docker registry username to authorize pull of base images and push to the docker repo")
	cmd.PersistentFlags().StringVarP(&CmdData.RegistryUsername, "registry-password", "", "", "Docker registry password to authorize pull of base images and push to the docker repo")

	common.SetupTag(&CommonCmdData, cmd)

	return cmd
}

func runBP(dimgsToProcess []string) error {
	if err := dapp.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
		return err
	}

	if err := true_git.Init(); err != nil {
		return err
	}

	projectDir, err := common.GetProjectDir(&CommonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	projectName, err := common.GetProjectName(&CommonCmdData, projectDir)
	if err != nil {
		return fmt.Errorf("getting project name failed: %s", err)
	}

	projectBuildDir, err := common.GetProjectBuildDir(projectName)
	if err != nil {
		return fmt.Errorf("getting project build dir failed: %s", err)
	}

	projectTmpDir, err := common.GetProjectTmpDir()
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	if !docker.Debug() {
		defer common.RemoveProjectTmpDir(projectTmpDir)
	}

	dappfile, err := common.GetDappfile(projectDir)
	if err != nil {
		return fmt.Errorf("dappfile parsing failed: %s", err)
	}

	repo, err := common.GetRequiredRepoName(projectName, CmdData.Repo)
	if err != nil {
		return err
	}

	dockerAuthorizer, err := docker_authorizer.GetBPDockerAuthorizer(projectTmpDir, CmdData.PullUsername, CmdData.PullPassword, CmdData.PushUsername, CmdData.PushPassword, repo)
	if err != nil {
		return err
	}

	if err := docker.Init(dockerAuthorizer.HostDockerConfigDir); err != nil {
		return err
	}

	if err := ssh_agent.Init(*CommonCmdData.SSHKeys); err != nil {
		return fmt.Errorf("cannot initialize ssh agent: %s", err)
	}
	defer func() {
		err := ssh_agent.Terminate()
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: ssh agent termination failed: %s", err)
		}
	}()

	tagOpts, err := common.GetTagOptions(&CommonCmdData, projectDir)
	if err != nil {
		return err
	}

	pushOpts := build.PushOptions{TagOptions: tagOpts, WithStages: CmdData.WithStages}

	c := build.NewConveyor(dappfile, dimgsToProcess, projectDir, projectName, projectBuildDir, projectTmpDir, ssh_agent.SSHAuthSock, dockerAuthorizer)
	if err = c.BP(repo, pushOpts); err != nil {
		return err
	}

	return nil
}
