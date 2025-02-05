// Copyright 2015 go-dockerclient authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package docker

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ErrContainerAlreadyExists is the error returned by CreateContainer when the
// container already exists.
var ErrContainerAlreadyExists = errors.New("container already exists")

// ListContainersOptions specify parameters to the ListContainers function.
//
// See https://goo.gl/47a6tO for more details.
type ListContainersOptions struct {
	All     bool
	Size    bool
	Limit   int
	Since   string
	Before  string
	Filters map[string][]string
}

// APIPort is a type that represents a port mapping returned by the Docker API
type APIPort struct {
	PrivatePort int64  `json:"PrivatePort,omitempty" yaml:"PrivatePort,omitempty"`
	PublicPort  int64  `json:"PublicPort,omitempty" yaml:"PublicPort,omitempty"`
	Type        string `json:"Type,omitempty" yaml:"Type,omitempty"`
	IP          string `json:"IP,omitempty" yaml:"IP,omitempty"`
}

// APIContainers represents each container in the list returned by
// ListContainers.
type APIContainers struct {
	ID         string            `json:"Id" yaml:"Id"`
	Image      string            `json:"Image,omitempty" yaml:"Image,omitempty"`
	Command    string            `json:"Command,omitempty" yaml:"Command,omitempty"`
	Created    int64             `json:"Created,omitempty" yaml:"Created,omitempty"`
	Status     string            `json:"Status,omitempty" yaml:"Status,omitempty"`
	Ports      []APIPort         `json:"Ports,omitempty" yaml:"Ports,omitempty"`
	SizeRw     int64             `json:"SizeRw,omitempty" yaml:"SizeRw,omitempty"`
	SizeRootFs int64             `json:"SizeRootFs,omitempty" yaml:"SizeRootFs,omitempty"`
	Names      []string          `json:"Names,omitempty" yaml:"Names,omitempty"`
	Labels     map[string]string `json:"Labels,omitempty" yaml:"Labels, omitempty"`
}

// ListContainers returns a slice of containers matching the given criteria.
//
// See https://goo.gl/47a6tO for more details.
func (c *Client) ListContainers(opts ListContainersOptions) ([]APIContainers, error) {
	path := "/containers/json?" + queryString(opts)
	resp, err := c.do("GET", path, doOptions{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var containers []APIContainers
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return nil, err
	}
	return containers, nil
}

// Port represents the port number and the protocol, in the form
// <number>/<protocol>. For example: 80/tcp.
type Port string

// Port returns the number of the port.
func (p Port) Port() string {
	return strings.Split(string(p), "/")[0]
}

// Proto returns the name of the protocol.
func (p Port) Proto() string {
	parts := strings.Split(string(p), "/")
	if len(parts) == 1 {
		return "tcp"
	}
	return parts[1]
}

// State represents the state of a container.
type State struct {
	Running    bool      `json:"Running,omitempty" yaml:"Running,omitempty"`
	Paused     bool      `json:"Paused,omitempty" yaml:"Paused,omitempty"`
	Restarting bool      `json:"Restarting,omitempty" yaml:"Restarting,omitempty"`
	OOMKilled  bool      `json:"OOMKilled,omitempty" yaml:"OOMKilled,omitempty"`
	Pid        int       `json:"Pid,omitempty" yaml:"Pid,omitempty"`
	ExitCode   int       `json:"ExitCode,omitempty" yaml:"ExitCode,omitempty"`
	Error      string    `json:"Error,omitempty" yaml:"Error,omitempty"`
	StartedAt  time.Time `json:"StartedAt,omitempty" yaml:"StartedAt,omitempty"`
	FinishedAt time.Time `json:"FinishedAt,omitempty" yaml:"FinishedAt,omitempty"`
}

// String returns the string representation of a state.
func (s *State) String() string {
	if s.Running {
		if s.Paused {
			return "paused"
		}
		return fmt.Sprintf("Up %s", time.Now().UTC().Sub(s.StartedAt))
	}
	return fmt.Sprintf("Exit %d", s.ExitCode)
}

// PortBinding represents the host/container port mapping as returned in the
// `docker inspect` json
type PortBinding struct {
	HostIP   string `json:"HostIP,omitempty" yaml:"HostIP,omitempty"`
	HostPort string `json:"HostPort,omitempty" yaml:"HostPort,omitempty"`
}

// PortMapping represents a deprecated field in the `docker inspect` output,
// and its value as found in NetworkSettings should always be nil
type PortMapping map[string]string

// ContainerNetwork represents the networking settings of a container per network.
type ContainerNetwork struct {
	MacAddress          string `json:"MacAddress,omitempty" yaml:"MacAddress,omitempty"`
	GlobalIPv6PrefixLen int    `json:"GlobalIPv6PrefixLen,omitempty" yaml:"GlobalIPv6PrefixLen,omitempty"`
	GlobalIPv6Address   string `json:"GlobalIPv6Address,omitempty" yaml:"GlobalIPv6Address,omitempty"`
	IPv6Gateway         string `json:"IPv6Gateway,omitempty" yaml:"IPv6Gateway,omitempty"`
	IPPrefixLen         int    `json:"IPPrefixLen,omitempty" yaml:"IPPrefixLen,omitempty"`
	IPAddress           string `json:"IPAddress,omitempty" yaml:"IPAddress,omitempty"`
	Gateway             string `json:"Gateway,omitempty" yaml:"Gateway,omitempty"`
	EndpointID          string `json:"EndpointID,omitempty" yaml:"EndpointID,omitempty"`
}

// NetworkSettings contains network-related information about a container
type NetworkSettings struct {
	Networks               map[string]ContainerNetwork `json:"Networks,omitempty" yaml:"Networks,omitempty"`
	IPAddress              string                      `json:"IPAddress,omitempty" yaml:"IPAddress,omitempty"`
	IPPrefixLen            int                         `json:"IPPrefixLen,omitempty" yaml:"IPPrefixLen,omitempty"`
	MacAddress             string                      `json:"MacAddress,omitempty" yaml:"MacAddress,omitempty"`
	Gateway                string                      `json:"Gateway,omitempty" yaml:"Gateway,omitempty"`
	Bridge                 string                      `json:"Bridge,omitempty" yaml:"Bridge,omitempty"`
	PortMapping            map[string]PortMapping      `json:"PortMapping,omitempty" yaml:"PortMapping,omitempty"`
	Ports                  map[Port][]PortBinding      `json:"Ports,omitempty" yaml:"Ports,omitempty"`
	NetworkID              string                      `json:"NetworkID,omitempty" yaml:"NetworkID,omitempty"`
	EndpointID             string                      `json:"EndpointID,omitempty" yaml:"EndpointID,omitempty"`
	SandboxKey             string                      `json:"SandboxKey,omitempty" yaml:"SandboxKey,omitempty"`
	GlobalIPv6Address      string                      `json:"GlobalIPv6Address,omitempty" yaml:"GlobalIPv6Address,omitempty"`
	GlobalIPv6PrefixLen    int                         `json:"GlobalIPv6PrefixLen,omitempty" yaml:"GlobalIPv6PrefixLen,omitempty"`
	IPv6Gateway            string                      `json:"IPv6Gateway,omitempty" yaml:"IPv6Gateway,omitempty"`
	LinkLocalIPv6Address   string                      `json:"LinkLocalIPv6Address,omitempty" yaml:"LinkLocalIPv6Address,omitempty"`
	LinkLocalIPv6PrefixLen int                         `json:"LinkLocalIPv6PrefixLen,omitempty" yaml:"LinkLocalIPv6PrefixLen,omitempty"`
	SecondaryIPAddresses   []string                    `json:"SecondaryIPAddresses,omitempty" yaml:"SecondaryIPAddresses,omitempty"`
	SecondaryIPv6Addresses []string                    `json:"SecondaryIPv6Addresses,omitempty" yaml:"SecondaryIPv6Addresses,omitempty"`
}

// PortMappingAPI translates the port mappings as contained in NetworkSettings
// into the format in which they would appear when returned by the API
func (settings *NetworkSettings) PortMappingAPI() []APIPort {
	var mapping []APIPort
	for port, bindings := range settings.Ports {
		p, _ := parsePort(port.Port())
		if len(bindings) == 0 {
			mapping = append(mapping, APIPort{
				PrivatePort: int64(p),
				Type:        port.Proto(),
			})
			continue
		}
		for _, binding := range bindings {
			p, _ := parsePort(port.Port())
			h, _ := parsePort(binding.HostPort)
			mapping = append(mapping, APIPort{
				PrivatePort: int64(p),
				PublicPort:  int64(h),
				Type:        port.Proto(),
				IP:          binding.HostIP,
			})
		}
	}
	return mapping
}

func parsePort(rawPort string) (int, error) {
	port, err := strconv.ParseUint(rawPort, 10, 16)
	if err != nil {
		return 0, err
	}
	return int(port), nil
}

// Config is the list of configuration options used when creating a container.
// Config does not contain the options that are specific to starting a container on a
// given host.  Those are contained in HostConfig
type Config struct {
	Hostname          string              `json:"Hostname,omitempty" yaml:"Hostname,omitempty"`
	Domainname        string              `json:"Domainname,omitempty" yaml:"Domainname,omitempty"`
	User              string              `json:"User,omitempty" yaml:"User,omitempty"`
	Memory            int64               `json:"Memory,omitempty" yaml:"Memory,omitempty"`
	MemorySwap        int64               `json:"MemorySwap,omitempty" yaml:"MemorySwap,omitempty"`
	MemoryReservation int64               `json:"MemoryReservation,omitempty" yaml:"MemoryReservation,omitempty"`
	KernelMemory      int64               `json:"KernelMemory,omitempty" yaml:"KernelMemory,omitempty"`
	CPUShares         int64               `json:"CpuShares,omitempty" yaml:"CpuShares,omitempty"`
	CPUSet            string              `json:"Cpuset,omitempty" yaml:"Cpuset,omitempty"`
	AttachStdin       bool                `json:"AttachStdin,omitempty" yaml:"AttachStdin,omitempty"`
	AttachStdout      bool                `json:"AttachStdout,omitempty" yaml:"AttachStdout,omitempty"`
	AttachStderr      bool                `json:"AttachStderr,omitempty" yaml:"AttachStderr,omitempty"`
	PortSpecs         []string            `json:"PortSpecs,omitempty" yaml:"PortSpecs,omitempty"`
	ExposedPorts      map[Port]struct{}   `json:"ExposedPorts,omitempty" yaml:"ExposedPorts,omitempty"`
	StopSignal        string              `json:"StopSignal,omitempty" yaml:"StopSignal,omitempty"`
	Tty               bool                `json:"Tty,omitempty" yaml:"Tty,omitempty"`
	OpenStdin         bool                `json:"OpenStdin,omitempty" yaml:"OpenStdin,omitempty"`
	StdinOnce         bool                `json:"StdinOnce,omitempty" yaml:"StdinOnce,omitempty"`
	Env               []string            `json:"Env,omitempty" yaml:"Env,omitempty"`
	Cmd               []string            `json:"Cmd" yaml:"Cmd"`
	DNS               []string            `json:"Dns,omitempty" yaml:"Dns,omitempty"` // For Docker API v1.9 and below only
	Image             string              `json:"Image,omitempty" yaml:"Image,omitempty"`
	Volumes           map[string]struct{} `json:"Volumes,omitempty" yaml:"Volumes,omitempty"`
	VolumeDriver      string              `json:"VolumeDriver,omitempty" yaml:"VolumeDriver,omitempty"`
	VolumesFrom       string              `json:"VolumesFrom,omitempty" yaml:"VolumesFrom,omitempty"`
	WorkingDir        string              `json:"WorkingDir,omitempty" yaml:"WorkingDir,omitempty"`
	MacAddress        string              `json:"MacAddress,omitempty" yaml:"MacAddress,omitempty"`
	Entrypoint        []string            `json:"Entrypoint" yaml:"Entrypoint"`
	NetworkDisabled   bool                `json:"NetworkDisabled,omitempty" yaml:"NetworkDisabled,omitempty"`
	SecurityOpts      []string            `json:"SecurityOpts,omitempty" yaml:"SecurityOpts,omitempty"`
	OnBuild           []string            `json:"OnBuild,omitempty" yaml:"OnBuild,omitempty"`
	Mounts            []Mount             `json:"Mounts,omitempty" yaml:"Mounts,omitempty"`
	Labels            map[string]string   `json:"Labels,omitempty" yaml:"Labels,omitempty"`
}

// Mount represents a mount point in the container.
//
// It has been added in the version 1.20 of the Docker API, available since
// Docker 1.8.
type Mount struct {
	Source      string
	Destination string
	Mode        string
	RW          bool
}

// LogConfig defines the log driver type and the configuration for it.
type LogConfig struct {
	Type   string            `json:"Type,omitempty" yaml:"Type,omitempty"`
	Config map[string]string `json:"Config,omitempty" yaml:"Config,omitempty"`
}

// ULimit defines system-wide resource limitations
// This can help a lot in system administration, e.g. when a user starts too many processes and therefore makes the system unresponsive for other users.
type ULimit struct {
	Name string `json:"Name,omitempty" yaml:"Name,omitempty"`
	Soft int64  `json:"Soft,omitempty" yaml:"Soft,omitempty"`
	Hard int64  `json:"Hard,omitempty" yaml:"Hard,omitempty"`
}

// SwarmNode containers information about which Swarm node the container is on
type SwarmNode struct {
	ID     string            `json:"ID,omitempty" yaml:"ID,omitempty"`
	IP     string            `json:"IP,omitempty" yaml:"IP,omitempty"`
	Addr   string            `json:"Addr,omitempty" yaml:"Addr,omitempty"`
	Name   string            `json:"Name,omitempty" yaml:"Name,omitempty"`
	CPUs   int64             `json:"CPUs,omitempty" yaml:"CPUs,omitempty"`
	Memory int64             `json:"Memory,omitempty" yaml:"Memory,omitempty"`
	Labels map[string]string `json:"Labels,omitempty" yaml:"Labels,omitempty"`
}

// Container is the type encompasing everything about a container - its config,
// hostconfig, etc.
type Container struct {
	ID string `json:"Id" yaml:"Id"`

	Created time.Time `json:"Created,omitempty" yaml:"Created,omitempty"`

	Path string   `json:"Path,omitempty" yaml:"Path,omitempty"`
	Args []string `json:"Args,omitempty" yaml:"Args,omitempty"`

	Config *Config `json:"Config,omitempty" yaml:"Config,omitempty"`
	State  State   `json:"State,omitempty" yaml:"State,omitempty"`
	Image  string  `json:"Image,omitempty" yaml:"Image,omitempty"`

	Node *SwarmNode `json:"Node,omitempty" yaml:"Node,omitempty"`

	NetworkSettings *NetworkSettings `json:"NetworkSettings,omitempty" yaml:"NetworkSettings,omitempty"`

	SysInitPath    string  `json:"SysInitPath,omitempty" yaml:"SysInitPath,omitempty"`
	ResolvConfPath string  `json:"ResolvConfPath,omitempty" yaml:"ResolvConfPath,omitempty"`
	HostnamePath   string  `json:"HostnamePath,omitempty" yaml:"HostnamePath,omitempty"`
	HostsPath      string  `json:"HostsPath,omitempty" yaml:"HostsPath,omitempty"`
	LogPath        string  `json:"LogPath,omitempty" yaml:"LogPath,omitempty"`
	Name           string  `json:"Name,omitempty" yaml:"Name,omitempty"`
	Driver         string  `json:"Driver,omitempty" yaml:"Driver,omitempty"`
	Mounts         []Mount `json:"Mounts,omitempty" yaml:"Mounts,omitempty"`

	Volumes    map[string]string `json:"Volumes,omitempty" yaml:"Volumes,omitempty"`
	VolumesRW  map[string]bool   `json:"VolumesRW,omitempty" yaml:"VolumesRW,omitempty"`
	HostConfig *HostConfig       `json:"HostConfig,omitempty" yaml:"HostConfig,omitempty"`
	ExecIDs    []string          `json:"ExecIDs,omitempty" yaml:"ExecIDs,omitempty"`

	RestartCount int `json:"RestartCount,omitempty" yaml:"RestartCount,omitempty"`

	AppArmorProfile string `json:"AppArmorProfile,omitempty" yaml:"AppArmorProfile,omitempty"`
}

// RenameContainerOptions specify parameters to the RenameContainer function.
//
// See https://goo.gl/laSOIy for more details.
type RenameContainerOptions struct {
	// ID of container to rename
	ID string `qs:"-"`

	// New name
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
}

// RenameContainer updates and existing containers name
//
// See https://goo.gl/laSOIy for more details.
func (c *Client) RenameContainer(opts RenameContainerOptions) error {
	resp, err := c.do("POST", fmt.Sprintf("/containers/"+opts.ID+"/rename?%s", queryString(opts)), doOptions{})
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// InspectContainer returns information about a container by its ID.
//
// See https://goo.gl/RdIq0b for more details.
func (c *Client) InspectContainer(id string) (*Container, error) {
	path := "/containers/" + id + "/json"
	resp, err := c.do("GET", path, doOptions{})
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return nil, &NoSuchContainer{ID: id}
		}
		return nil, err
	}
	defer resp.Body.Close()
	var container Container
	if err := json.NewDecoder(resp.Body).Decode(&container); err != nil {
		return nil, err
	}
	return &container, nil
}

// ContainerChanges returns changes in the filesystem of the given container.
//
// See https://goo.gl/9GsTIF for more details.
func (c *Client) ContainerChanges(id string) ([]Change, error) {
	path := "/containers/" + id + "/changes"
	resp, err := c.do("GET", path, doOptions{})
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return nil, &NoSuchContainer{ID: id}
		}
		return nil, err
	}
	defer resp.Body.Close()
	var changes []Change
	if err := json.NewDecoder(resp.Body).Decode(&changes); err != nil {
		return nil, err
	}
	return changes, nil
}

// CreateContainerOptions specify parameters to the CreateContainer function.
//
// See https://goo.gl/WxQzrr for more details.
type CreateContainerOptions struct {
	Name       string
	Config     *Config     `qs:"-"`
	HostConfig *HostConfig `qs:"-"`
}

// CreateContainer creates a new container, returning the container instance,
// or an error in case of failure.
//
// See https://goo.gl/WxQzrr for more details.
func (c *Client) CreateContainer(opts CreateContainerOptions) (*Container, error) {
	path := "/containers/create?" + queryString(opts)
	resp, err := c.do(
		"POST",
		path,
		doOptions{
			data: struct {
				*Config
				HostConfig *HostConfig `json:"HostConfig,omitempty" yaml:"HostConfig,omitempty"`
			}{
				opts.Config,
				opts.HostConfig,
			},
		},
	)

	if e, ok := err.(*Error); ok {
		if e.Status == http.StatusNotFound {
			return nil, ErrNoSuchImage
		}
		if e.Status == http.StatusConflict {
			return nil, ErrContainerAlreadyExists
		}
	}

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var container Container
	if err := json.NewDecoder(resp.Body).Decode(&container); err != nil {
		return nil, err
	}

	container.Name = opts.Name

	return &container, nil
}

// KeyValuePair is a type for generic key/value pairs as used in the Lxc
// configuration
type KeyValuePair struct {
	Key   string `json:"Key,omitempty" yaml:"Key,omitempty"`
	Value string `json:"Value,omitempty" yaml:"Value,omitempty"`
}

// RestartPolicy represents the policy for automatically restarting a container.
//
// Possible values are:
//
//   - always: the docker daemon will always restart the container
//   - on-failure: the docker daemon will restart the container on failures, at
//                 most MaximumRetryCount times
//   - no: the docker daemon will not restart the container automatically
type RestartPolicy struct {
	Name              string `json:"Name,omitempty" yaml:"Name,omitempty"`
	MaximumRetryCount int    `json:"MaximumRetryCount,omitempty" yaml:"MaximumRetryCount,omitempty"`
}

// AlwaysRestart returns a restart policy that tells the Docker daemon to
// always restart the container.
func AlwaysRestart() RestartPolicy {
	return RestartPolicy{Name: "always"}
}

// RestartOnFailure returns a restart policy that tells the Docker daemon to
// restart the container on failures, trying at most maxRetry times.
func RestartOnFailure(maxRetry int) RestartPolicy {
	return RestartPolicy{Name: "on-failure", MaximumRetryCount: maxRetry}
}

// NeverRestart returns a restart policy that tells the Docker daemon to never
// restart the container on failures.
func NeverRestart() RestartPolicy {
	return RestartPolicy{Name: "no"}
}

// Device represents a device mapping between the Docker host and the
// container.
type Device struct {
	PathOnHost        string `json:"PathOnHost,omitempty" yaml:"PathOnHost,omitempty"`
	PathInContainer   string `json:"PathInContainer,omitempty" yaml:"PathInContainer,omitempty"`
	CgroupPermissions string `json:"CgroupPermissions,omitempty" yaml:"CgroupPermissions,omitempty"`
}

// HostConfig contains the container options related to starting a container on
// a given host
type HostConfig struct {
	Binds            []string               `json:"Binds,omitempty" yaml:"Binds,omitempty"`
	CapAdd           []string               `json:"CapAdd,omitempty" yaml:"CapAdd,omitempty"`
	CapDrop          []string               `json:"CapDrop,omitempty" yaml:"CapDrop,omitempty"`
	GroupAdd         []string               `json:"GroupAdd,omitempty" yaml:"GroupAdd,omitempty"`
	ContainerIDFile  string                 `json:"ContainerIDFile,omitempty" yaml:"ContainerIDFile,omitempty"`
	LxcConf          []KeyValuePair         `json:"LxcConf,omitempty" yaml:"LxcConf,omitempty"`
	Privileged       bool                   `json:"Privileged,omitempty" yaml:"Privileged,omitempty"`
	PortBindings     map[Port][]PortBinding `json:"PortBindings,omitempty" yaml:"PortBindings,omitempty"`
	Links            []string               `json:"Links,omitempty" yaml:"Links,omitempty"`
	PublishAllPorts  bool                   `json:"PublishAllPorts,omitempty" yaml:"PublishAllPorts,omitempty"`
	DNS              []string               `json:"Dns,omitempty" yaml:"Dns,omitempty"` // For Docker API v1.10 and above only
	DNSOptions       []string               `json:"DnsOptions,omitempty" yaml:"DnsOptions,omitempty"`
	DNSSearch        []string               `json:"DnsSearch,omitempty" yaml:"DnsSearch,omitempty"`
	ExtraHosts       []string               `json:"ExtraHosts,omitempty" yaml:"ExtraHosts,omitempty"`
	VolumesFrom      []string               `json:"VolumesFrom,omitempty" yaml:"VolumesFrom,omitempty"`
	NetworkMode      string                 `json:"NetworkMode,omitempty" yaml:"NetworkMode,omitempty"`
	IpcMode          string                 `json:"IpcMode,omitempty" yaml:"IpcMode,omitempty"`
	PidMode          string                 `json:"PidMode,omitempty" yaml:"PidMode,omitempty"`
	UTSMode          string                 `json:"UTSMode,omitempty" yaml:"UTSMode,omitempty"`
	RestartPolicy    RestartPolicy          `json:"RestartPolicy,omitempty" yaml:"RestartPolicy,omitempty"`
	Devices          []Device               `json:"Devices,omitempty" yaml:"Devices,omitempty"`
	LogConfig        LogConfig              `json:"LogConfig,omitempty" yaml:"LogConfig,omitempty"`
	ReadonlyRootfs   bool                   `json:"ReadonlyRootfs,omitempty" yaml:"ReadonlyRootfs,omitempty"`
	SecurityOpt      []string               `json:"SecurityOpt,omitempty" yaml:"SecurityOpt,omitempty"`
	CgroupParent     string                 `json:"CgroupParent,omitempty" yaml:"CgroupParent,omitempty"`
	Memory           int64                  `json:"Memory,omitempty" yaml:"Memory,omitempty"`
	MemorySwap       int64                  `json:"MemorySwap,omitempty" yaml:"MemorySwap,omitempty"`
	MemorySwappiness int64                  `json:"MemorySwappiness,omitempty" yaml:"MemorySwappiness,omitempty"`
	OOMKillDisable   bool                   `json:"OomKillDisable,omitempty" yaml:"OomKillDisable"`
	CPUShares        int64                  `json:"CpuShares,omitempty" yaml:"CpuShares,omitempty"`
	CPUSet           string                 `json:"Cpuset,omitempty" yaml:"Cpuset,omitempty"`
	CPUSetCPUs       string                 `json:"CpusetCpus,omitempty" yaml:"CpusetCpus,omitempty"`
	CPUSetMEMs       string                 `json:"CpusetMems,omitempty" yaml:"CpusetMems,omitempty"`
	CPUQuota         int64                  `json:"CpuQuota,omitempty" yaml:"CpuQuota,omitempty"`
	CPUPeriod        int64                  `json:"CpuPeriod,omitempty" yaml:"CpuPeriod,omitempty"`
	BlkioWeight      int64                  `json:"BlkioWeight,omitempty" yaml:"BlkioWeight"`
	Ulimits          []ULimit               `json:"Ulimits,omitempty" yaml:"Ulimits,omitempty"`
	VolumeDriver     string                 `json:"VolumeDriver,omitempty" yaml:"VolumeDriver,omitempty"`
}

// StartContainer starts a container, returning an error in case of failure.
//
// See https://goo.gl/MrBAJv for more details.
func (c *Client) StartContainer(id string, hostConfig *HostConfig) error {
	path := "/containers/" + id + "/start"
	resp, err := c.do("POST", path, doOptions{data: hostConfig, forceJSON: true})
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return &NoSuchContainer{ID: id, Err: err}
		}
		return err
	}
	if resp.StatusCode == http.StatusNotModified {
		return &ContainerAlreadyRunning{ID: id}
	}
	resp.Body.Close()
	return nil
}

// StopContainer stops a container, killing it after the given timeout (in
// seconds).
//
// See https://goo.gl/USqsFt for more details.
func (c *Client) StopContainer(id string, timeout uint) error {
	path := fmt.Sprintf("/containers/%s/stop?t=%d", id, timeout)
	resp, err := c.do("POST", path, doOptions{})
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return &NoSuchContainer{ID: id}
		}
		return err
	}
	if resp.StatusCode == http.StatusNotModified {
		return &ContainerNotRunning{ID: id}
	}
	resp.Body.Close()
	return nil
}

// RestartContainer stops a container, killing it after the given timeout (in
// seconds), during the stop process.
//
// See https://goo.gl/QzsDnz for more details.
func (c *Client) RestartContainer(id string, timeout uint) error {
	path := fmt.Sprintf("/containers/%s/restart?t=%d", id, timeout)
	resp, err := c.do("POST", path, doOptions{})
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return &NoSuchContainer{ID: id}
		}
		return err
	}
	resp.Body.Close()
	return nil
}

// PauseContainer pauses the given container.
//
// See https://goo.gl/OF7W9X for more details.
func (c *Client) PauseContainer(id string) error {
	path := fmt.Sprintf("/containers/%s/pause", id)
	resp, err := c.do("POST", path, doOptions{})
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return &NoSuchContainer{ID: id}
		}
		return err
	}
	resp.Body.Close()
	return nil
}

// UnpauseContainer unpauses the given container.
//
// See https://goo.gl/7dwyPA for more details.
func (c *Client) UnpauseContainer(id string) error {
	path := fmt.Sprintf("/containers/%s/unpause", id)
	resp, err := c.do("POST", path, doOptions{})
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return &NoSuchContainer{ID: id}
		}
		return err
	}
	resp.Body.Close()
	return nil
}

// TopResult represents the list of processes running in a container, as
// returned by /containers/<id>/top.
//
// See https://goo.gl/Rb46aY for more details.
type TopResult struct {
	Titles    []string
	Processes [][]string
}

// TopContainer returns processes running inside a container
//
// See https://goo.gl/Rb46aY for more details.
func (c *Client) TopContainer(id string, psArgs string) (TopResult, error) {
	var args string
	var result TopResult
	if psArgs != "" {
		args = fmt.Sprintf("?ps_args=%s", psArgs)
	}
	path := fmt.Sprintf("/containers/%s/top%s", id, args)
	resp, err := c.do("GET", path, doOptions{})
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return result, &NoSuchContainer{ID: id}
		}
		return result, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, err
	}
	return result, nil
}

// Stats represents container statistics, returned by /containers/<id>/stats.
//
// See https://goo.gl/GNmLHb for more details.
type Stats struct {
	Read    time.Time `json:"read,omitempty" yaml:"read,omitempty"`
	Network struct {
		RxDropped uint64 `json:"rx_dropped,omitempty" yaml:"rx_dropped,omitempty"`
		RxBytes   uint64 `json:"rx_bytes,omitempty" yaml:"rx_bytes,omitempty"`
		RxErrors  uint64 `json:"rx_errors,omitempty" yaml:"rx_errors,omitempty"`
		TxPackets uint64 `json:"tx_packets,omitempty" yaml:"tx_packets,omitempty"`
		TxDropped uint64 `json:"tx_dropped,omitempty" yaml:"tx_dropped,omitempty"`
		RxPackets uint64 `json:"rx_packets,omitempty" yaml:"rx_packets,omitempty"`
		TxErrors  uint64 `json:"tx_errors,omitempty" yaml:"tx_errors,omitempty"`
		TxBytes   uint64 `json:"tx_bytes,omitempty" yaml:"tx_bytes,omitempty"`
	} `json:"network,omitempty" yaml:"network,omitempty"`
	MemoryStats struct {
		Stats struct {
			TotalPgmafault          uint64 `json:"total_pgmafault,omitempty" yaml:"total_pgmafault,omitempty"`
			Cache                   uint64 `json:"cache,omitempty" yaml:"cache,omitempty"`
			MappedFile              uint64 `json:"mapped_file,omitempty" yaml:"mapped_file,omitempty"`
			TotalInactiveFile       uint64 `json:"total_inactive_file,omitempty" yaml:"total_inactive_file,omitempty"`
			Pgpgout                 uint64 `json:"pgpgout,omitempty" yaml:"pgpgout,omitempty"`
			Rss                     uint64 `json:"rss,omitempty" yaml:"rss,omitempty"`
			TotalMappedFile         uint64 `json:"total_mapped_file,omitempty" yaml:"total_mapped_file,omitempty"`
			Writeback               uint64 `json:"writeback,omitempty" yaml:"writeback,omitempty"`
			Unevictable             uint64 `json:"unevictable,omitempty" yaml:"unevictable,omitempty"`
			Pgpgin                  uint64 `json:"pgpgin,omitempty" yaml:"pgpgin,omitempty"`
			TotalUnevictable        uint64 `json:"total_unevictable,omitempty" yaml:"total_unevictable,omitempty"`
			Pgmajfault              uint64 `json:"pgmajfault,omitempty" yaml:"pgmajfault,omitempty"`
			TotalRss                uint64 `json:"total_rss,omitempty" yaml:"total_rss,omitempty"`
			TotalRssHuge            uint64 `json:"total_rss_huge,omitempty" yaml:"total_rss_huge,omitempty"`
			TotalWriteback          uint64 `json:"total_writeback,omitempty" yaml:"total_writeback,omitempty"`
			TotalInactiveAnon       uint64 `json:"total_inactive_anon,omitempty" yaml:"total_inactive_anon,omitempty"`
			RssHuge                 uint64 `json:"rss_huge,omitempty" yaml:"rss_huge,omitempty"`
			HierarchicalMemoryLimit uint64 `json:"hierarchical_memory_limit,omitempty" yaml:"hierarchical_memory_limit,omitempty"`
			TotalPgfault            uint64 `json:"total_pgfault,omitempty" yaml:"total_pgfault,omitempty"`
			TotalActiveFile         uint64 `json:"total_active_file,omitempty" yaml:"total_active_file,omitempty"`
			ActiveAnon              uint64 `json:"active_anon,omitempty" yaml:"active_anon,omitempty"`
			TotalActiveAnon         uint64 `json:"total_active_anon,omitempty" yaml:"total_active_anon,omitempty"`
			TotalPgpgout            uint64 `json:"total_pgpgout,omitempty" yaml:"total_pgpgout,omitempty"`
			TotalCache              uint64 `json:"total_cache,omitempty" yaml:"total_cache,omitempty"`
			InactiveAnon            uint64 `json:"inactive_anon,omitempty" yaml:"inactive_anon,omitempty"`
			ActiveFile              uint64 `json:"active_file,omitempty" yaml:"active_file,omitempty"`
			Pgfault                 uint64 `json:"pgfault,omitempty" yaml:"pgfault,omitempty"`
			InactiveFile            uint64 `json:"inactive_file,omitempty" yaml:"inactive_file,omitempty"`
			TotalPgpgin             uint64 `json:"total_pgpgin,omitempty" yaml:"total_pgpgin,omitempty"`
		} `json:"stats,omitempty" yaml:"stats,omitempty"`
		MaxUsage uint64 `json:"max_usage,omitempty" yaml:"max_usage,omitempty"`
		Usage    uint64 `json:"usage,omitempty" yaml:"usage,omitempty"`
		Failcnt  uint64 `json:"failcnt,omitempty" yaml:"failcnt,omitempty"`
		Limit    uint64 `json:"limit,omitempty" yaml:"limit,omitempty"`
	} `json:"memory_stats,omitempty" yaml:"memory_stats,omitempty"`
	BlkioStats struct {
		IOServiceBytesRecursive []BlkioStatsEntry `json:"io_service_bytes_recursive,omitempty" yaml:"io_service_bytes_recursive,omitempty"`
		IOServicedRecursive     []BlkioStatsEntry `json:"io_serviced_recursive,omitempty" yaml:"io_serviced_recursive,omitempty"`
		IOQueueRecursive        []BlkioStatsEntry `json:"io_queue_recursive,omitempty" yaml:"io_queue_recursive,omitempty"`
		IOServiceTimeRecursive  []BlkioStatsEntry `json:"io_service_time_recursive,omitempty" yaml:"io_service_time_recursive,omitempty"`
		IOWaitTimeRecursive     []BlkioStatsEntry `json:"io_wait_time_recursive,omitempty" yaml:"io_wait_time_recursive,omitempty"`
		IOMergedRecursive       []BlkioStatsEntry `json:"io_merged_recursive,omitempty" yaml:"io_merged_recursive,omitempty"`
		IOTimeRecursive         []BlkioStatsEntry `json:"io_time_recursive,omitempty" yaml:"io_time_recursive,omitempty"`
		SectorsRecursive        []BlkioStatsEntry `json:"sectors_recursive,omitempty" yaml:"sectors_recursive,omitempty"`
	} `json:"blkio_stats,omitempty" yaml:"blkio_stats,omitempty"`
	CPUStats    CPUStats `json:"cpu_stats,omitempty" yaml:"cpu_stats,omitempty"`
	PreCPUStats CPUStats `json:"precpu_stats,omitempty"`
}

// CPUStats is a stats entry for cpu stats
type CPUStats struct {
	CPUUsage struct {
		PercpuUsage       []uint64 `json:"percpu_usage,omitempty" yaml:"percpu_usage,omitempty"`
		UsageInUsermode   uint64   `json:"usage_in_usermode,omitempty" yaml:"usage_in_usermode,omitempty"`
		TotalUsage        uint64   `json:"total_usage,omitempty" yaml:"total_usage,omitempty"`
		UsageInKernelmode uint64   `json:"usage_in_kernelmode,omitempty" yaml:"usage_in_kernelmode,omitempty"`
	} `json:"cpu_usage,omitempty" yaml:"cpu_usage,omitempty"`
	SystemCPUUsage uint64 `json:"system_cpu_usage,omitempty" yaml:"system_cpu_usage,omitempty"`
	ThrottlingData struct {
		Periods          uint64 `json:"periods,omitempty"`
		ThrottledPeriods uint64 `json:"throttled_periods,omitempty"`
		ThrottledTime    uint64 `json:"throttled_time,omitempty"`
	} `json:"throttling_data,omitempty" yaml:"throttling_data,omitempty"`
}

// BlkioStatsEntry is a stats entry for blkio_stats
type BlkioStatsEntry struct {
	Major uint64 `json:"major,omitempty" yaml:"major,omitempty"`
	Minor uint64 `json:"minor,omitempty" yaml:"minor,omitempty"`
	Op    string `json:"op,omitempty" yaml:"op,omitempty"`
	Value uint64 `json:"value,omitempty" yaml:"value,omitempty"`
}

// StatsOptions specify parameters to the Stats function.
//
// See https://goo.gl/GNmLHb for more details.
type StatsOptions struct {
	ID     string
	Stats  chan<- *Stats
	Stream bool
	// A flag that enables stopping the stats operation
	Done <-chan bool
	// Initial connection timeout
	Timeout time.Duration
}

// Stats sends container statistics for the given container to the given channel.
//
// This function is blocking, similar to a streaming call for logs, and should be run
// on a separate goroutine from the caller. Note that this function will block until
// the given container is removed, not just exited. When finished, this function
// will close the given channel. Alternatively, function can be stopped by
// signaling on the Done channel.
//
// See https://goo.gl/GNmLHb for more details.
func (c *Client) Stats(opts StatsOptions) (retErr error) {
	errC := make(chan error, 1)
	readCloser, writeCloser := io.Pipe()

	defer func() {
		close(opts.Stats)

		select {
		case err := <-errC:
			if err != nil && retErr == nil {
				retErr = err
			}
		default:
			// No errors
		}

		if err := readCloser.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	go func() {
		err := c.stream("GET", fmt.Sprintf("/containers/%s/stats?stream=%v", opts.ID, opts.Stream), streamOptions{
			rawJSONStream:  true,
			useJSONDecoder: true,
			stdout:         writeCloser,
			timeout:        opts.Timeout,
		})
		if err != nil {
			dockerError, ok := err.(*Error)
			if ok {
				if dockerError.Status == http.StatusNotFound {
					err = &NoSuchContainer{ID: opts.ID}
				}
			}
		}
		if closeErr := writeCloser.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		errC <- err
		close(errC)
	}()

	quit := make(chan struct{})
	defer close(quit)
	go func() {
		// block here waiting for the signal to stop function
		select {
		case <-opts.Done:
			readCloser.Close()
		case <-quit:
			return
		}
	}()

	decoder := json.NewDecoder(readCloser)
	stats := new(Stats)
	for err := decoder.Decode(stats); err != io.EOF; err = decoder.Decode(stats) {
		if err != nil {
			return err
		}
		opts.Stats <- stats
		stats = new(Stats)
	}
	return nil
}

// KillContainerOptions represents the set of options that can be used in a
// call to KillContainer.
//
// See https://goo.gl/hkS9i8 for more details.
type KillContainerOptions struct {
	// The ID of the container.
	ID string `qs:"-"`

	// The signal to send to the container. When omitted, Docker server
	// will assume SIGKILL.
	Signal Signal
}

// KillContainer sends a signal to a container, returning an error in case of
// failure.
//
// See https://goo.gl/hkS9i8 for more details.
func (c *Client) KillContainer(opts KillContainerOptions) error {
	path := "/containers/" + opts.ID + "/kill" + "?" + queryString(opts)
	resp, err := c.do("POST", path, doOptions{})
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return &NoSuchContainer{ID: opts.ID}
		}
		return err
	}
	resp.Body.Close()
	return nil
}

// RemoveContainerOptions encapsulates options to remove a container.
//
// See https://goo.gl/RQyX62 for more details.
type RemoveContainerOptions struct {
	// The ID of the container.
	ID string `qs:"-"`

	// A flag that indicates whether Docker should remove the volumes
	// associated to the container.
	RemoveVolumes bool `qs:"v"`

	// A flag that indicates whether Docker should remove the container
	// even if it is currently running.
	Force bool
}

// RemoveContainer removes a container, returning an error in case of failure.
//
// See https://goo.gl/RQyX62 for more details.
func (c *Client) RemoveContainer(opts RemoveContainerOptions) error {
	path := "/containers/" + opts.ID + "?" + queryString(opts)
	resp, err := c.do("DELETE", path, doOptions{})
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return &NoSuchContainer{ID: opts.ID}
		}
		return err
	}
	resp.Body.Close()
	return nil
}

// UploadToContainerOptions is the set of options that can be used when
// uploading an archive into a container.
//
// See https://goo.gl/Ss97HW for more details.
type UploadToContainerOptions struct {
	InputStream          io.Reader `json:"-" qs:"-"`
	Path                 string    `qs:"path"`
	NoOverwriteDirNonDir bool      `qs:"noOverwriteDirNonDir"`
}

// UploadToContainer uploads a tar archive to be extracted to a path in the
// filesystem of the container.
//
// See https://goo.gl/Ss97HW for more details.
func (c *Client) UploadToContainer(id string, opts UploadToContainerOptions) error {
	url := fmt.Sprintf("/containers/%s/archive?", id) + queryString(opts)

	return c.stream("PUT", url, streamOptions{
		in: opts.InputStream,
	})
}

// DownloadFromContainerOptions is the set of options that can be used when
// downloading resources from a container.
//
// See https://goo.gl/KnZJDX for more details.
type DownloadFromContainerOptions struct {
	OutputStream io.Writer `json:"-" qs:"-"`
	Path         string    `qs:"path"`
}

// DownloadFromContainer downloads a tar archive of files or folders in a container.
//
// See https://goo.gl/KnZJDX for more details.
func (c *Client) DownloadFromContainer(id string, opts DownloadFromContainerOptions) error {
	url := fmt.Sprintf("/containers/%s/archive?", id) + queryString(opts)

	return c.stream("GET", url, streamOptions{
		setRawTerminal: true,
		stdout:         opts.OutputStream,
	})
}

// CopyFromContainerOptions has been DEPRECATED, please use DownloadFromContainerOptions along with DownloadFromContainer.
//
// See https://goo.gl/R2jevW for more details.
type CopyFromContainerOptions struct {
	OutputStream io.Writer `json:"-"`
	Container    string    `json:"-"`
	Resource     string
}

// CopyFromContainer has been DEPRECATED, please use DownloadFromContainerOptions along with DownloadFromContainer.
//
// See https://goo.gl/R2jevW for more details.
func (c *Client) CopyFromContainer(opts CopyFromContainerOptions) error {
	if opts.Container == "" {
		return &NoSuchContainer{ID: opts.Container}
	}
	url := fmt.Sprintf("/containers/%s/copy", opts.Container)
	resp, err := c.do("POST", url, doOptions{data: opts})
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return &NoSuchContainer{ID: opts.Container}
		}
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(opts.OutputStream, resp.Body)
	return err
}

// WaitContainer blocks until the given container stops, return the exit code
// of the container status.
//
// See https://goo.gl/Gc1rge for more details.
func (c *Client) WaitContainer(id string) (int, error) {
	resp, err := c.do("POST", "/containers/"+id+"/wait", doOptions{})
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return 0, &NoSuchContainer{ID: id}
		}
		return 0, err
	}
	defer resp.Body.Close()
	var r struct{ StatusCode int }
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return 0, err
	}
	return r.StatusCode, nil
}

// CommitContainerOptions aggregates parameters to the CommitContainer method.
//
// See https://goo.gl/mqfoCw for more details.
type CommitContainerOptions struct {
	Container  string
	Repository string `qs:"repo"`
	Tag        string
	Message    string `qs:"m"`
	Author     string
	Run        *Config `qs:"-"`
}

// CommitContainer creates a new image from a container's changes.
//
// See https://goo.gl/mqfoCw for more details.
func (c *Client) CommitContainer(opts CommitContainerOptions) (*Image, error) {
	path := "/commit?" + queryString(opts)
	resp, err := c.do("POST", path, doOptions{data: opts.Run})
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return nil, &NoSuchContainer{ID: opts.Container}
		}
		return nil, err
	}
	defer resp.Body.Close()
	var image Image
	if err := json.NewDecoder(resp.Body).Decode(&image); err != nil {
		return nil, err
	}
	return &image, nil
}

// AttachToContainerOptions is the set of options that can be used when
// attaching to a container.
//
// See https://goo.gl/NKpkFk for more details.
type AttachToContainerOptions struct {
	Container    string    `qs:"-"`
	InputStream  io.Reader `qs:"-"`
	OutputStream io.Writer `qs:"-"`
	ErrorStream  io.Writer `qs:"-"`

	// Get container logs, sending it to OutputStream.
	Logs bool

	// Stream the response?
	Stream bool

	// Attach to stdin, and use InputStream.
	Stdin bool

	// Attach to stdout, and use OutputStream.
	Stdout bool

	// Attach to stderr, and use ErrorStream.
	Stderr bool

	// If set, after a successful connect, a sentinel will be sent and then the
	// client will block on receive before continuing.
	//
	// It must be an unbuffered channel. Using a buffered channel can lead
	// to unexpected behavior.
	Success chan struct{}

	// Use raw terminal? Usually true when the container contains a TTY.
	RawTerminal bool `qs:"-"`
}

// AttachToContainer attaches to a container, using the given options.
//
// See https://goo.gl/NKpkFk for more details.
func (c *Client) AttachToContainer(opts AttachToContainerOptions) error {
	if opts.Container == "" {
		return &NoSuchContainer{ID: opts.Container}
	}
	path := "/containers/" + opts.Container + "/attach?" + queryString(opts)
	cw, err := c.hijack("POST", path, hijackOptions{
		success:        opts.Success,
		setRawTerminal: opts.RawTerminal,
		in:             opts.InputStream,
		stdout:         opts.OutputStream,
		stderr:         opts.ErrorStream,
	})
	if err != nil {
		return err
	}
	return cw.Wait()
}

// AttachToContainerNonBlocking attaches to a container, using the given options.
// This function does not block.
//
// See https://goo.gl/NKpkFk for more details.
func (c *Client) AttachToContainerNonBlocking(opts AttachToContainerOptions) (CloseWaiter, error) {
	if opts.Container == "" {
		return nil, &NoSuchContainer{ID: opts.Container}
	}
	path := "/containers/" + opts.Container + "/attach?" + queryString(opts)
	return c.hijack("POST", path, hijackOptions{
		success:        opts.Success,
		setRawTerminal: opts.RawTerminal,
		in:             opts.InputStream,
		stdout:         opts.OutputStream,
		stderr:         opts.ErrorStream,
	})
}

// LogsOptions represents the set of options used when getting logs from a
// container.
//
// See https://goo.gl/yl8PGm for more details.
type LogsOptions struct {
	Container    string    `qs:"-"`
	OutputStream io.Writer `qs:"-"`
	ErrorStream  io.Writer `qs:"-"`
	Follow       bool
	Stdout       bool
	Stderr       bool
	Since        int64
	Timestamps   bool
	Tail         string

	// Use raw terminal? Usually true when the container contains a TTY.
	RawTerminal bool `qs:"-"`
}

// Logs gets stdout and stderr logs from the specified container.
//
// See https://goo.gl/yl8PGm for more details.
func (c *Client) Logs(opts LogsOptions) error {
	if opts.Container == "" {
		return &NoSuchContainer{ID: opts.Container}
	}
	if opts.Tail == "" {
		opts.Tail = "all"
	}
	path := "/containers/" + opts.Container + "/logs?" + queryString(opts)
	return c.stream("GET", path, streamOptions{
		setRawTerminal: opts.RawTerminal,
		stdout:         opts.OutputStream,
		stderr:         opts.ErrorStream,
	})
}

// ResizeContainerTTY resizes the terminal to the given height and width.
//
// See https://goo.gl/xERhCc for more details.
func (c *Client) ResizeContainerTTY(id string, height, width int) error {
	params := make(url.Values)
	params.Set("h", strconv.Itoa(height))
	params.Set("w", strconv.Itoa(width))
	resp, err := c.do("POST", "/containers/"+id+"/resize?"+params.Encode(), doOptions{})
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// ExportContainerOptions is the set of parameters to the ExportContainer
// method.
//
// See https://goo.gl/dOkTyk for more details.
type ExportContainerOptions struct {
	ID           string
	OutputStream io.Writer
}

// ExportContainer export the contents of container id as tar archive
// and prints the exported contents to stdout.
//
// See https://goo.gl/dOkTyk for more details.
func (c *Client) ExportContainer(opts ExportContainerOptions) error {
	if opts.ID == "" {
		return &NoSuchContainer{ID: opts.ID}
	}
	url := fmt.Sprintf("/containers/%s/export", opts.ID)
	return c.stream("GET", url, streamOptions{
		setRawTerminal: true,
		stdout:         opts.OutputStream,
	})
}

// NoSuchContainer is the error returned when a given container does not exist.
type NoSuchContainer struct {
	ID  string
	Err error
}

func (err *NoSuchContainer) Error() string {
	if err.Err != nil {
		return err.Err.Error()
	}
	return "No such container: " + err.ID
}

// ContainerAlreadyRunning is the error returned when a given container is
// already running.
type ContainerAlreadyRunning struct {
	ID string
}

func (err *ContainerAlreadyRunning) Error() string {
	return "Container already running: " + err.ID
}

// ContainerNotRunning is the error returned when a given container is not
// running.
type ContainerNotRunning struct {
	ID string
}

func (err *ContainerNotRunning) Error() string {
	return "Container not running: " + err.ID
}
