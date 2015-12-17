package render

import (
	"strings"

	"github.com/weaveworks/scope/probe"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

// DetailedNode is the data type that's yielded to the JavaScript layer when
// we want deep information about an individual node.
type DetailedNode struct {
	ID       string             `json:"id"`
	Label    string             `json:"label"`
	Rank     string             `json:"rank,omitempty"`
	Pseudo   bool               `json:"pseudo,omitempty"`
	Controls []ControlInstance  `json:"controls"`
	Metadata []MetadataRow      `json:"metadata,omitempty"`
	Metrics  []MetricRow        `json:"metrics,omitempty"`
	Children []NodeSummaryGroup `json:"children,omitempty"`
	Parents  []Parent           `json:"parents,omitempty"`
}

type NodeSummaryGroup struct {
	Label      string        `json:"label"`
	Nodes      []NodeSummary `json:"nodes"`
	TopologyID string        `json:"topologyId"`
}

type NodeSummary struct {
	ID       string        `json:"id"`
	Label    string        `json:"label"`
	Metadata []MetadataRow `json:"metadata,omitempty"`
	Metrics  []MetricRow   `json:"metrics,omitempty"`
}

type Parent struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	TopologyID string `json:"topologyId"`
}

// ControlInstance contains a control description, and all the info
// needed to execute it.
type ControlInstance struct {
	ProbeID string `json:"probeId"`
	NodeID  string `json:"nodeId"`
	report.Control
}

// MakeDetailedNode transforms a renderable node to a detailed node. It uses
// aggregate metadata, plus the set of origin node IDs, to produce tables.
func MakeDetailedNode(r report.Report, n RenderableNode) DetailedNode {
	var hosts, pods, containerImages, containers, applications []NodeSummary
	for id, child := range n.Children {
		if id == n.ID {
			continue
		}

		switch child.Metadata[probe.Topology] {
		case "process":
			applications = append(applications, processNodeSummary(child))
		case "container":
			containers = append(containers, containerNodeSummary(child))
		case "container_image":
			containerImages = append(containerImages, containerImageNodeSummary(child))
		case "pod":
			pods = append(pods, podNodeSummary(child))
		case "host":
			hosts = append(hosts, hostNodeSummary(child))
		}
	}
	children := []NodeSummaryGroup{}
	if len(hosts) > 0 {
		children = append(children, NodeSummaryGroup{TopologyID: "hosts", Label: "Hosts", Nodes: hosts})
	}
	if len(pods) > 0 {
		children = append(children, NodeSummaryGroup{TopologyID: "pods", Label: "Pods", Nodes: pods})
	}
	if len(containerImages) > 0 {
		children = append(children, NodeSummaryGroup{TopologyID: "containers-by-image", Label: "Container Images", Nodes: containerImages})
	}
	if len(containers) > 0 {
		children = append(children, NodeSummaryGroup{TopologyID: "containers", Label: "Containers", Nodes: containers})
	}
	if len(applications) > 0 {
		children = append(children, NodeSummaryGroup{TopologyID: "applications", Label: "Applications", Nodes: applications})
	}

	return DetailedNode{
		ID:       n.ID,
		Label:    n.LabelMajor,
		Rank:     n.Rank,
		Pseudo:   n.Pseudo,
		Controls: controls(r, n),
		Metadata: nodeMetadata(r, n),
		Metrics:  nodeMetrics(r, n),
		Children: children,
		Parents:  parents(r, n),
	}
}

func controlsFor(topology report.Topology, nodeID string) []ControlInstance {
	result := []ControlInstance{}
	node, ok := topology.Nodes[nodeID]
	if !ok {
		return result
	}

	for _, id := range node.Controls.Controls {
		if control, ok := topology.Controls[id]; ok {
			result = append(result, ControlInstance{
				ProbeID: node.Metadata[report.ProbeID],
				NodeID:  nodeID,
				Control: control,
			})
		}
	}
	return result
}

func controls(r report.Report, n RenderableNode) []ControlInstance {
	if _, ok := r.Process.Nodes[n.ControlNode]; ok {
		return controlsFor(r.Process, n.ControlNode)
	} else if _, ok := r.Container.Nodes[n.ControlNode]; ok {
		return controlsFor(r.Container, n.ControlNode)
	} else if _, ok := r.ContainerImage.Nodes[n.ControlNode]; ok {
		return controlsFor(r.ContainerImage, n.ControlNode)
	} else if _, ok := r.Host.Nodes[n.ControlNode]; ok {
		return controlsFor(r.Host, n.ControlNode)
	}
	return []ControlInstance{}
}

// parents is a total a hack to find the parents of a node (which is
// ill-defined).
func parents(r report.Report, n RenderableNode) []Parent {
	result := []Parent{}

	// Add a host if we have one
	if hostNodeID, ok := n.Node.Metadata[report.HostNodeID]; ok {
		if hostNode, ok := r.Host.Nodes[hostNodeID]; ok {
			result = append(result, Parent{
				ID:         MakeHostID(hostNode.Metadata[host.HostName]),
				Label:      hostNode.Metadata[host.HostName],
				TopologyID: "hosts",
			})
		}
	}

	hostID := report.ExtractHostID(n.Node)

	if namespaceID, ok := n.Node.Metadata[kubernetes.Namespace]; ok {
		// Add kubernetes services if we have them
		if serviceIDs, ok := n.Node.Metadata[kubernetes.ServiceIDs]; ok {
			for _, serviceID := range strings.Fields(serviceIDs) {
				serviceNodeID := report.MakeServiceNodeID(namespaceID, serviceID)
				if serviceNode, ok := r.Service.Nodes[serviceNodeID]; ok {
					result = append(result, Parent{
						ID:         MakeServiceID(serviceID),
						Label:      serviceNode.Metadata[kubernetes.ServiceName],
						TopologyID: "pods-by-service",
					})
				}
			}
		}
		// add kubernetes pod if we have one
		if podID, ok := n.Node.Metadata[kubernetes.PodID]; ok {
			podNodeID := report.MakePodNodeID(namespaceID, podID)
			if podNode, ok := r.Pod.Nodes[podNodeID]; ok {
				// Add a pod if we have one
				result = append(result, Parent{
					ID:         MakePodID(podID),
					Label:      podNode.Metadata[kubernetes.PodName],
					TopologyID: "pods",
				})
			}
		}
	}

	// add container if we have one
	if containerID, ok := n.Node.Metadata[docker.ContainerID]; ok {
		containerNodeID := report.MakeContainerNodeID(hostID, containerID)
		if containerNode, ok := r.Container.Nodes[containerNodeID]; ok {
			label, _ := GetRenderableContainerName(containerNode)
			result = append(result, Parent{
				ID:         MakeContainerID(containerID),
				Label:      label,
				TopologyID: "containers",
			})
		}
	}

	// add container image
	if containerImageID, ok := n.Node.Metadata[docker.ImageID]; ok {
		if containerImageNode, ok := r.ContainerImage.Nodes[containerImageID]; ok {
			result = append(result, Parent{
				ID:         MakeContainerImageID(containerImageID),
				Label:      containerImageNode.Metadata[docker.ImageName],
				TopologyID: "containers-by-image",
			})
		}
	}

	for i, parent := range result {
		if parent.ID == n.ID {
			result = append(result[:i], result[i+1:]...)
		}
	}

	return result
}

// nodeTopology (optimistically) tells us which origin topology an ID belongs
// to.
func nodeTopology(r report.Report, id string) (report.Topology, string, report.Node, bool) {
	// TODO(paulbellamy): This is a duplicate of report.Topologies. Needs a
	// cleanup/refactor. It's duplicated because Topology structs are not
	// comparable in go, so we need an identity. Better yet, replace this with a
	// polymorphism on either the topology, or the node type.

	topologies := map[string]report.Topology{
		"endpoint":        r.Endpoint,
		"address":         r.Address,
		"process":         r.Process,
		"container":       r.Container,
		"container_image": r.ContainerImage,
		"pod":             r.Pod,
		"service":         r.Service,
		"host":            r.Host,
		"overlay":         r.Overlay,
	}
	for tName, t := range topologies {
		if n, ok := t.Nodes[id]; ok {
			return t, tName, n, true
		}
	}
	return report.Topology{}, "", report.Node{}, false
}

func processNodeSummary(nmd report.Node) NodeSummary {
	var (
		id    string
		label = nmd.Metadata[process.Comm]
	)
	if pid, ok := nmd.Metadata[process.PID]; ok {
		if label == "" {
			label = pid
		}
		id = MakeProcessID(report.ExtractHostID(nmd), pid)
	}
	return NodeSummary{
		ID:       id,
		Label:    label,
		Metadata: processNodeMetadata(nmd),
		Metrics:  processNodeMetrics(nmd),
	}
}

func containerNodeSummary(nmd report.Node) NodeSummary {
	label, _ := GetRenderableContainerName(nmd)
	return NodeSummary{
		ID:       MakeContainerID(nmd.Metadata[docker.ContainerID]),
		Label:    label,
		Metadata: containerNodeMetadata(nmd),
		Metrics:  containerNodeMetrics(nmd),
	}
}

func podNodeSummary(nmd report.Node) NodeSummary {
	return NodeSummary{
		ID:       MakePodID(nmd.Metadata[kubernetes.PodID]),
		Label:    nmd.Metadata[kubernetes.PodName],
		Metadata: podNodeMetadata(nmd),
	}
}

func containerImageNodeSummary(nmd report.Node) NodeSummary {
	return NodeSummary{
		ID:       MakeContainerImageID(nmd.Metadata[docker.ImageID]),
		Label:    nmd.Metadata[docker.ImageName],
		Metadata: containerImageNodeMetadata(nmd),
	}
}

func hostNodeSummary(nmd report.Node) NodeSummary {
	return NodeSummary{
		ID:       MakeHostID(nmd.Metadata[host.HostName]),
		Label:    nmd.Metadata[host.HostName],
		Metadata: hostNodeMetadata(nmd),
		Metrics:  hostNodeMetrics(nmd),
	}
}
