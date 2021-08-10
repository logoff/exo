package container

import (
	"context"
	"fmt"
	"time"

	core "github.com/deref/exo/internal/core/api"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	docker "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func (c *Container) Initialize(ctx context.Context, input *core.InitializeInput) (output *core.InitializeOutput, err error) {
	if err := c.ensureImage(ctx); err != nil {
		return nil, fmt.Errorf("ensuring image: %w", err)
	}

	if err := c.create(ctx); err != nil {
		return nil, fmt.Errorf("creating container: %w", err)
	}

	if err := c.start(ctx); err != nil {
		c.Logger.Infof("starting container %q: %v", c.ContainerID, err)
	}

	return &core.InitializeOutput{}, nil
}

func (c *Container) create(ctx context.Context) error {
	var healthCfg *container.HealthConfig
	if c.Healthcheck != nil {
		healthCfg = &container.HealthConfig{
			Test:        c.Healthcheck.Test,
			Interval:    time.Duration(c.Healthcheck.Interval),
			Timeout:     time.Duration(c.Healthcheck.Timeout),
			Retries:     c.Healthcheck.Retries,
			StartPeriod: time.Duration(c.Healthcheck.StartPeriod),
		}
	}
	containerCfg := &container.Config{
		Hostname:     c.Hostname,
		Domainname:   c.Domainname,
		User:         c.User,
		ExposedPorts: make(nat.PortSet),
		Tty:          c.TTY,
		OpenStdin:    c.StdinOpen,
		// StdinOnce       bool                // If true, close stdin after the 1 attached client disconnects.
		Env:         c.Environment.Slice(),
		Cmd:         strslice.StrSlice(c.Command),
		Healthcheck: healthCfg,
		// ArgsEscaped     bool                `json:",omitempty"` // True if command is already escaped (meaning treat as a command line) (Windows specific).
		Image: c.State.ImageID,
		// Volumes         map[string]struct{} // List of volumes (mounts) used for the container
		WorkingDir: c.WorkingDir,
		Entrypoint: strslice.StrSlice(c.Entrypoint),
		// NetworkDisabled bool                `json:",omitempty"` // Is network disabled
		MacAddress: c.MacAddress,
		// OnBuild         []string            // ONBUILD metadata that were defined on the image Dockerfile
		Labels:     c.Labels.WithoutNils(),
		StopSignal: c.StopSignal,
		// Shell           strslice.StrSlice   `json:",omitempty"` // Shell for shell-form of RUN, CMD, ENTRYPOINT
	}
	if c.StopGracePeriod != nil {
		timeout := int(time.Duration(*c.StopGracePeriod).Round(time.Second).Seconds())
		containerCfg.StopTimeout = &timeout
	}
	for _, mapping := range c.Ports {
		target := nat.Port(mapping.Target) // TODO: Handle port ranges.
		containerCfg.ExposedPorts[target] = struct{}{}
	}
	logCfg := container.LogConfig{}
	if c.Logging.Driver == "" && (c.Logging.Options == nil || len(c.Logging.Options) == 0) {
		// No logging configuration specified, so default to logging to exo's
		// syslog service.
		logCfg.Type = "syslog"
		logCfg.Config = map[string]string{
			"syslog-address":  fmt.Sprintf("udp://localhost:%d", c.SyslogPort),
			"syslog-facility": "1", // "user-level messages"
			"tag":             c.ComponentID,
			"syslog-format":   "rfc5424micro",
		}
	} else {
		logCfg.Type = c.Logging.Driver
		logCfg.Config = c.Logging.Options
	}
	hostCfg := &container.HostConfig{
		//// Applicable to all platforms
		//Binds           []string      // List of volume bindings for this container
		//ContainerIDFile string        // File (path) where the containerId is written
		LogConfig: logCfg,
		//NetworkMode     NetworkMode   // Network mode to use for the container
		PortBindings: make(nat.PortMap),
		//RestartPolicy   RestartPolicy // Restart policy to be used for the container
		// TODO: Potentially inherit from deploy's restart_policy.
		RestartPolicy: container.RestartPolicy{
			Name: c.Spec.Restart,
		},
		//AutoRemove      bool          // Automatically remove container when it exits
		//VolumeDriver    string        // Name of the volume driver used to mount volumes
		//VolumesFrom     []string      // List of volumes to take from other container

		//// Applicable to UNIX platforms
		//CapAdd          strslice.StrSlice // List of kernel capabilities to add to the container
		//CapDrop         strslice.StrSlice // List of kernel capabilities to remove from the container
		//CgroupnsMode    CgroupnsMode      // Cgroup namespace mode to use for the container
		//DNS             []string          `json:"Dns"`        // List of DNS server to lookup
		//DNSOptions      []string          `json:"DnsOptions"` // List of DNSOption to look for
		//DNSSearch       []string          `json:"DnsSearch"`  // List of DNSSearch to look for
		//ExtraHosts      []string          // List of extra hosts
		//GroupAdd        []string          // List of additional groups that the container process will run as
		//IpcMode         IpcMode           // IPC namespace to use for the container
		//Cgroup          CgroupSpec        // Cgroup to use for the container
		//Links           []string          // List of links (in the name:alias form)
		//OomScoreAdj     int               // Container preference for OOM-killing
		//PidMode         PidMode           // PID namespace to use for the container
		Privileged: c.Privileged,
		//PublishAllPorts bool              // Should docker publish all exposed port for the container
		//ReadonlyRootfs  bool              // Is the container root filesystem in read-only
		//SecurityOpt     []string          // List of string values to customize labels for MLS systems, such as SELinux.
		//StorageOpt      map[string]string `json:",omitempty"` // Storage driver options per container.
		//Tmpfs           map[string]string `json:",omitempty"` // List of tmpfs (mounts) used for the container
		//UTSMode         UTSMode           // UTS namespace to use for the container
		//UsernsMode      UsernsMode        // The user namespace to use for the container
		ShmSize: int64(c.ShmSize),
		//Sysctls         map[string]string `json:",omitempty"` // List of Namespaced sysctls used for the container
		Runtime: c.Runtime,

		//// Applicable to Windows
		//ConsoleSize [2]uint   // Initial console size (height,width)
		//Isolation   Isolation // Isolation technology of the container (e.g. default, hyperv)

		//// Contains container's resources (cgroups, ulimits)
		//Resources

		//// Mounts specs used by the container
		//Mounts []mount.Mount `json:",omitempty"`

		//// MaskedPaths is the list of paths to be masked inside the container (this overrides the default set of paths)
		//MaskedPaths []string

		//// ReadonlyPaths is the list of paths to be set as read-only inside the container (this overrides the default set of paths)
		//ReadonlyPaths []string

		//// Run a custom init inside the container, if null, use the daemon's configured settings
		//Init *bool `json:",omitempty"`
	}
	for _, mapping := range c.Ports {
		target := nat.Port(mapping.Target) // TODO: Handle ranges.
		bindings := hostCfg.PortBindings[target]
		bindings = append(bindings, nat.PortBinding{
			HostIP:   mapping.HostIP,
			HostPort: mapping.Published,
		})
		// TODO: Handle mapping.Mode and mapping.Protocol.
		hostCfg.PortBindings[target] = bindings
	}
	networkCfg := &network.NetworkingConfig{
		//EndpointsConfig map[string]*EndpointSettings // Endpoint configs for each connecting network
	}
	var platform *v1.Platform
	//platform := &v1.Platform{
	//	//// Architecture field specifies the CPU architecture, for example
	//	//// `amd64` or `ppc64`.
	//	//Architecture string `json:"architecture"`

	//	//// OS specifies the operating system, for example `linux` or `windows`.
	//	//OS string `json:"os"`

	//	//// OSVersion is an optional field specifying the operating system
	//	//// version, for example on Windows `10.0.14393.1066`.
	//	//OSVersion string `json:"os.version,omitempty"`

	//	//// OSFeatures is an optional field specifying an array of strings,
	//	//// each listing a required OS feature (for example on Windows `win32k`).
	//	//OSFeatures []string `json:"os.features,omitempty"`

	//	//// Variant is an optional field specifying a variant of the CPU, for
	//	//// example `v7` to specify ARMv7 when architecture is `arm`.
	//	//Variant string `json:"variant,omitempty"`
	//}
	createdBody, err := c.Docker.ContainerCreate(ctx, containerCfg, hostCfg, networkCfg, platform, c.ContainerName)
	if err != nil {
		return err
	}
	c.ContainerID = createdBody.ID
	return nil
}

func (c *Container) Refresh(ctx context.Context, input *core.RefreshInput) (*core.RefreshOutput, error) {
	if c.ContainerID == "" {
		c.Running = false
		return &core.RefreshOutput{}, nil
	}

	inspection, err := c.Docker.ContainerInspect(ctx, c.ContainerID)
	if err != nil {
		return nil, fmt.Errorf("inspecting container: %w", err)
	}

	c.Running = inspection.State.Running
	return &core.RefreshOutput{}, nil
}

func (c *Container) Dispose(ctx context.Context, input *core.DisposeInput) (*core.DisposeOutput, error) {
	if c.ContainerID == "" {
		return &core.DisposeOutput{}, nil
	}
	if err := c.stop(ctx); err != nil {
		c.Logger.Infof("stopping container %q: %v", c.ContainerID, err)
	}
	err := c.Docker.ContainerRemove(ctx, c.ContainerID, types.ContainerRemoveOptions{
		// XXX RemoveVolumes: ???,
		// XXX RemoveLinks: ???,
		Force: true, // OK?
	})
	if docker.IsErrNotFound(err) {
		c.Logger.Infof("disposing container not found: %q", c.ContainerID)
		err = nil
	}
	if err != nil {
		return nil, err
	}
	c.ContainerID = ""
	return &core.DisposeOutput{}, nil
}
