package render_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/fixture"
)

func TestMakeDetailedHostNode(t *testing.T) {
	renderableNode := render.HostRenderer.Render(fixture.Report)[render.MakeHostID(fixture.ClientHostID)]
	have := render.MakeDetailedNode(fixture.Report, renderableNode)
	want := render.DetailedNode{
		ID:       render.MakeHostID(fixture.ClientHostID),
		Label:    "client",
		Rank:     "hostname.com",
		Pseudo:   false,
		Controls: []render.ControlInstance{},
		Metadata: []render.MetadataRow{
			{
				ID:    "host_name",
				Label: "Hostname",
				Value: "client.hostname.com",
			},
			{
				ID:    "os",
				Label: "Operating system",
				Value: "Linux",
			},
		},
		Metrics: []render.MetricRow{
			{
				ID:     "load1",
				Group:  "load",
				Label:  "Load (1m)",
				Value:  0.01,
				Metric: &fixture.LoadMetric,
			},
			{
				ID:     "load5",
				Group:  "load",
				Label:  "Load (5m)",
				Value:  0.01,
				Metric: &fixture.LoadMetric,
			},
			{
				ID:     "load15",
				Label:  "Load (15m)",
				Group:  "load",
				Value:  0.01,
				Metric: &fixture.LoadMetric,
			},
		},
		Children: []render.NodeSummaryGroup{},
		Parents:  []render.Parent{},
		/*
				Connections: []render.MetadataRow{
						{
					Title:   "Connections",
					Numeric: false,
					Rank:    0,
					Rows: []render.Row{
						{
							Key:        "TCP connections",
							Value: "3",
						},
						{
							Key:        "Client",
							Value: "Server",
							Expandable: true,
						},
						{
							Key:        "10.10.10.20",
							Value: "192.168.1.1",
							Expandable: true,
						},
					},
				},
			},
		*/
	}
	if !reflect.DeepEqual(want, have) {
		t.Errorf("%s\nwant %#v\nhave %#v", test.Diff(want, have), want, have)
	}
}

func TestMakeDetailedContainerNode(t *testing.T) {
	id := render.MakeContainerID(fixture.ServerContainerID)
	renderableNode, ok := render.ContainerRenderer.Render(fixture.Report)[id]
	if !ok {
		t.Fatalf("Node not found: %s", id)
	}
	have := render.MakeDetailedNode(fixture.Report, renderableNode)
	want := render.DetailedNode{
		ID:       id,
		Label:    "server",
		Rank:     "imageid456",
		Pseudo:   false,
		Controls: []render.ControlInstance{},
		Metadata: []render.MetadataRow{
			{ID: "docker_container_id", Label: "ID", Value: fixture.ServerContainerID},
			{ID: "docker_image_id", Label: "Image ID", Value: fixture.ServerContainerImageID},
			{ID: "docker_container_state", Label: "State", Value: "running"},
			{ID: "label_" + render.AmazonECSContainerNameLabel, Label: fmt.Sprintf(`Label %q`, render.AmazonECSContainerNameLabel), Value: `server`},
			{ID: "label_foo1", Label: `Label "foo1"`, Value: `bar1`},
			{ID: "label_foo2", Label: `Label "foo2"`, Value: `bar2`},
			{ID: "label_io.kubernetes.pod.name", Label: `Label "io.kubernetes.pod.name"`, Value: "ping/pong-b"},
		},
		Metrics: []render.MetricRow{},
		Children: []render.NodeSummaryGroup{
			{
				Label:      "Applications",
				TopologyID: "applications",
				Nodes: []render.NodeSummary{
					{
						ID:       fmt.Sprintf("process:%s:%s", "server.hostname.com", fixture.ServerPID),
						Label:    "apache",
						Metadata: []render.MetadataRow{},
						Metrics:  []render.MetricRow{},
					},
				},
			},
		},
		Parents: []render.Parent{
			{
				ID:         render.MakeHostID(fixture.ServerHostName),
				Label:      fixture.ServerHostName,
				TopologyID: "hosts",
			},
		},
		/*
			Connections: []render.MetadataRow{
				{
					Title:   "Connections",
					Numeric: false,
					Rank:    0,
					Rows: []render.Row{
						{Key: "Ingress packet rate", Value: "105", ValueMinor: "packets/sec"},
						{Key: "Ingress byte rate", Value: "1.0", ValueMinor: "KBps"},
						{Key: "Client", Value: "Server", Expandable: true},
						{
							Key:        fmt.Sprintf("%s:%s", fixture.UnknownClient1IP, fixture.UnknownClient1Port),
							Value:      fmt.Sprintf("%s:%s", fixture.ServerIP, fixture.ServerPort),
							Expandable: true,
						},
						{
							Key:        fmt.Sprintf("%s:%s", fixture.UnknownClient2IP, fixture.UnknownClient2Port),
							Value:      fmt.Sprintf("%s:%s", fixture.ServerIP, fixture.ServerPort),
							Expandable: true,
						},
						{
							Key:        fmt.Sprintf("%s:%s", fixture.UnknownClient3IP, fixture.UnknownClient3Port),
							Value:      fmt.Sprintf("%s:%s", fixture.ServerIP, fixture.ServerPort),
							Expandable: true,
						},
						{
							Key:        fmt.Sprintf("%s:%s", fixture.ClientIP, fixture.ClientPort54001),
							Value:      fmt.Sprintf("%s:%s", fixture.ServerIP, fixture.ServerPort),
							Expandable: true,
						},
						{
							Key:        fmt.Sprintf("%s:%s", fixture.ClientIP, fixture.ClientPort54002),
							Value:      fmt.Sprintf("%s:%s", fixture.ServerIP, fixture.ServerPort),
							Expandable: true,
						},
						{
							Key:        fmt.Sprintf("%s:%s", fixture.RandomClientIP, fixture.RandomClientPort),
							Value:      fmt.Sprintf("%s:%s", fixture.ServerIP, fixture.ServerPort),
							Expandable: true,
						},
					},
				},
			},
		*/
	}
	if !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}
}
