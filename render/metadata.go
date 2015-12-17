package render

import (
	"fmt"
	"sort"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/probe/overlay"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

// TODO(paulbellamy): append docker labels to the metadata for containers and container images
var (
	processNodeMetadata = renderMetadata([]MetadataRow{
		{ID: process.PPID, Label: "Parent PID"},
		{ID: process.Cmdline, Label: "Command"},
		{ID: process.Threads, Label: "# Threads"},
	})
	containerNodeMetadata = renderMetadata([]MetadataRow{
		{ID: docker.ContainerID, Label: "ID"},
		{ID: docker.ImageID, Label: "Image ID"},
		{ID: docker.ContainerState, Label: "State"},
		{ID: docker.ContainerPorts, Label: "Ports"},
		{ID: docker.ContainerCreated, Label: "Created"},
		{ID: docker.ContainerCommand, Label: "Command"},
		{ID: overlay.WeaveMACAddress, Label: "Weave MAC"},
		{ID: overlay.WeaveDNSHostname, Label: "Weave DNS Hostname"},
	}, getDockerLabelRows)
	containerImageNodeMetadata = renderMetadata([]MetadataRow{
		{ID: docker.ImageID, Label: "Image ID"},
	})
	podNodeMetadata = renderMetadata([]MetadataRow{
		{ID: kubernetes.PodID, Label: "ID"},
		{ID: kubernetes.Namespace, Label: "Namespace"},
		{ID: kubernetes.PodCreated, Label: "Created"},
	})
	hostNodeMetadata = renderMetadata([]MetadataRow{
		{ID: host.HostName, Label: "Hostname"},
		{ID: host.OS, Label: "Operating system"},
		{ID: host.KernelVersion, Label: "Kernel version"},
		{ID: host.Uptime, Label: "Uptime"},
	})
)

type MetadataRow struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Value string `json:"value"`
}

// NodeMetadata produces a table (to be consumed directly by the UI) based on
// an origin ID, which is (optimistically) a node ID in one of our topologies.
func NodeMetadata(r report.Report, n RenderableNode) []MetadataRow {
	renderers := map[string]struct {
		t report.Topology
		r func(report.Node) []MetadataRow
	}{
		"process":         {r.Process, processNodeMetadata},
		"container":       {r.Container, containerNodeMetadata},
		"container_image": {r.ContainerImage, containerImageNodeMetadata},
		"pod":             {r.Pod, podNodeMetadata},
		"host":            {r.Host, hostNodeMetadata},
	}
	if renderer, ok := renderers[n.SummaryTopology]; ok {
		if nmd, ok := renderer.t.Nodes[n.SummaryID]; ok {
			return renderer.r(nmd)
		}
	}
	return nil
}

func renderMetadata(templates []MetadataRow, extras ...func(report.Node) []MetadataRow) func(report.Node) []MetadataRow {
	return func(nmd report.Node) []MetadataRow {
		rows := []MetadataRow{}
		for _, tuple := range templates {
			if val, ok := nmd.Metadata[tuple.ID]; ok {
				rows = append(rows, MetadataRow{ID: tuple.ID, Label: tuple.Label, Value: val})
			}
		}
		for _, extra := range extras {
			rows = append(rows, extra(nmd)...)
		}
		return rows
	}
}

func getDockerLabelRows(nmd report.Node) []MetadataRow {
	rows := []MetadataRow{}
	// Add labels in alphabetical order
	labels := docker.ExtractLabels(nmd)
	labelKeys := make([]string, 0, len(labels))
	for k := range labels {
		labelKeys = append(labelKeys, k)
	}
	sort.Strings(labelKeys)
	for _, labelKey := range labelKeys {
		rows = append(rows, MetadataRow{ID: "label_" + labelKey, Label: fmt.Sprintf("Label %q", labelKey), Value: labels[labelKey]})
	}
	return rows
}
