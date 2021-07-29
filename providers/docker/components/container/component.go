package container

import (
	"github.com/deref/exo/providers/docker/compose"
	docker "github.com/docker/docker/client"
)

type Container struct {
	ComponentID string
	Spec
	State

	Docker *docker.Client
}

type Spec compose.Service

type State struct {
	ContainerID string `json:"containerId"`
}
