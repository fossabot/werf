package instruction

import (
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
)

type Healthcheck struct {
	dockerfile_instruction.Healthcheck
}

type HealthcheckType string

var (
	HealthcheckTypeNone     HealthcheckType = "NONE"
	HealthcheckTypeCmd      HealthcheckType = "CMD"
	HealthcheckTypeCmdShell HealthcheckType = "CMD-SHELL"
)

// TODO
