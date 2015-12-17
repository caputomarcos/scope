package expected

import (
	"fmt"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/fixture"
)

// Exported for testing.
var (
	uncontainedServerID  = render.MakePseudoNodeID(render.UncontainedID, fixture.ServerHostName)
	unknownPseudoNode1ID = render.MakePseudoNodeID("10.10.10.10", fixture.ServerIP, "80")
	unknownPseudoNode2ID = render.MakePseudoNodeID("10.10.10.11", fixture.ServerIP, "80")
	unknownPseudoNode1   = func(adjacent string) render.RenderableNode {
		return render.RenderableNode{
			ID:         unknownPseudoNode1ID,
			LabelMajor: "10.10.10.10",
			Pseudo:     true,
			Node:       report.MakeNode().WithAdjacent(adjacent),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(70),
				EgressByteCount:   newu64(700),
			},
			Children: render.Children{
				fixture.UnknownClient1NodeID: fixture.Report.Endpoint.Nodes[fixture.UnknownClient1NodeID],
				fixture.UnknownClient2NodeID: fixture.Report.Endpoint.Nodes[fixture.UnknownClient2NodeID],
			},
		}
	}
	unknownPseudoNode2 = func(adjacent string) render.RenderableNode {
		return render.RenderableNode{
			ID:         unknownPseudoNode2ID,
			LabelMajor: "10.10.10.11",
			Pseudo:     true,
			Node:       report.MakeNode().WithAdjacent(adjacent),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(50),
				EgressByteCount:   newu64(500),
			},
			Children: render.Children{
				fixture.UnknownClient3NodeID: fixture.Report.Endpoint.Nodes[fixture.UnknownClient3NodeID],
			},
		}
	}
	theInternetNode = func(adjacent string) render.RenderableNode {
		return render.RenderableNode{
			ID:         render.TheInternetID,
			LabelMajor: render.TheInternetMajor,
			Pseudo:     true,
			Node:       report.MakeNode().WithAdjacent(adjacent),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(60),
				EgressByteCount:   newu64(600),
			},
			Children: render.Children{
				fixture.RandomClientNodeID:   fixture.Report.Endpoint.Nodes[fixture.RandomClientNodeID],
				fixture.GoogleEndpointNodeID: fixture.Report.Endpoint.Nodes[fixture.GoogleEndpointNodeID],
			},
		}
	}
	ClientProcess1ID      = render.MakeProcessID(fixture.ClientHostID, fixture.Client1PID)
	ClientProcess2ID      = render.MakeProcessID(fixture.ClientHostID, fixture.Client2PID)
	ServerProcessID       = render.MakeProcessID(fixture.ServerHostID, fixture.ServerPID)
	nonContainerProcessID = render.MakeProcessID(fixture.ServerHostID, fixture.NonContainerPID)

	RenderedProcesses = (render.RenderableNodes{
		ClientProcess1ID: {
			ID:         ClientProcess1ID,
			LabelMajor: fixture.Client1Comm,
			LabelMinor: fmt.Sprintf("%s (%s)", fixture.ClientHostID, fixture.Client1PID),
			Rank:       fixture.Client1Comm,
			Pseudo:     false,
			Children: render.Children{
				fixture.Client54001NodeID: fixture.Report.Endpoint.Nodes[fixture.Client54001NodeID],
			},
			Node: report.MakeNode().WithAdjacent(ServerProcessID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(10),
				EgressByteCount:   newu64(100),
			},
		},
		ClientProcess2ID: {
			ID:         ClientProcess2ID,
			LabelMajor: fixture.Client2Comm,
			LabelMinor: fmt.Sprintf("%s (%s)", fixture.ClientHostID, fixture.Client2PID),
			Rank:       fixture.Client2Comm,
			Pseudo:     false,
			Children: render.Children{
				fixture.Client54002NodeID: fixture.Report.Endpoint.Nodes[fixture.Client54002NodeID],
			},
			Node: report.MakeNode().WithAdjacent(ServerProcessID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(20),
				EgressByteCount:   newu64(200),
			},
		},
		ServerProcessID: {
			ID:         ServerProcessID,
			LabelMajor: "apache",
			LabelMinor: fmt.Sprintf("%s (%s)", fixture.ServerHostID, fixture.ServerPID),
			Rank:       fixture.ServerComm,
			Pseudo:     false,
			Children: render.Children{
				fixture.Server80NodeID: fixture.Report.Endpoint.Nodes[fixture.Server80NodeID],
			},
			Node: report.MakeNode(),
			EdgeMetadata: report.EdgeMetadata{
				IngressPacketCount: newu64(210),
				IngressByteCount:   newu64(2100),
			},
		},
		nonContainerProcessID: {
			ID:         nonContainerProcessID,
			LabelMajor: fixture.NonContainerComm,
			LabelMinor: fmt.Sprintf("%s (%s)", fixture.ServerHostID, fixture.NonContainerPID),
			Rank:       fixture.NonContainerComm,
			Pseudo:     false,
			Children: render.Children{
				fixture.NonContainerProcessNodeID: fixture.Report.Process.Nodes[fixture.NonContainerProcessNodeID],
				fixture.NonContainerNodeID:        fixture.Report.Endpoint.Nodes[fixture.NonContainerNodeID],
			},
			Node:         report.MakeNode().WithAdjacent(render.TheInternetID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		unknownPseudoNode1ID: unknownPseudoNode1(ServerProcessID),
		unknownPseudoNode2ID: unknownPseudoNode2(ServerProcessID),
		render.TheInternetID: theInternetNode(ServerProcessID),
	}).Prune()

	RenderedProcessNames = (render.RenderableNodes{
		"curl": {
			ID:         "curl",
			LabelMajor: "curl",
			LabelMinor: "2 processes",
			Rank:       "curl",
			Pseudo:     false,
			Children: render.Children{
				fixture.Client54001NodeID:    fixture.Report.Endpoint.Nodes[fixture.Client54001NodeID],
				fixture.Client54002NodeID:    fixture.Report.Endpoint.Nodes[fixture.Client54002NodeID],
				fixture.ClientProcess1NodeID: fixture.Report.Process.Nodes[fixture.ClientProcess1NodeID],
				fixture.ClientProcess2NodeID: fixture.Report.Process.Nodes[fixture.ClientProcess2NodeID],
			},
			Node: report.MakeNode().WithAdjacent("apache"),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(30),
				EgressByteCount:   newu64(300),
			},
		},
		"apache": {
			ID:         "apache",
			LabelMajor: "apache",
			LabelMinor: "1 process",
			Rank:       "apache",
			Pseudo:     false,
			Children: render.Children{
				fixture.Server80NodeID:      fixture.Report.Endpoint.Nodes[fixture.Server80NodeID],
				fixture.ServerProcessNodeID: fixture.Report.Process.Nodes[fixture.ServerProcessNodeID],
			},
			Node: report.MakeNode(),
			EdgeMetadata: report.EdgeMetadata{
				IngressPacketCount: newu64(210),
				IngressByteCount:   newu64(2100),
			},
		},
		fixture.NonContainerComm: {
			ID:         fixture.NonContainerComm,
			LabelMajor: fixture.NonContainerComm,
			LabelMinor: "1 process",
			Rank:       fixture.NonContainerComm,
			Pseudo:     false,
			Children: render.Children{
				fixture.NonContainerProcessNodeID: fixture.Report.Process.Nodes[fixture.NonContainerProcessNodeID],
			},
			Node:         report.MakeNode().WithAdjacent(render.TheInternetID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		unknownPseudoNode1ID: unknownPseudoNode1("apache"),
		unknownPseudoNode2ID: unknownPseudoNode2("apache"),
		render.TheInternetID: theInternetNode("apache"),
	}).Prune()

	RenderedContainers = (render.RenderableNodes{
		fixture.ClientContainerID: {
			ID:         fixture.ClientContainerID,
			LabelMajor: "client",
			LabelMinor: fixture.ClientHostName,
			Rank:       fixture.ClientContainerImageName,
			Pseudo:     false,
			Children: render.Children{
				fixture.Client54001NodeID:    fixture.Report.Endpoint.Nodes[fixture.Client54001NodeID],
				fixture.Client54002NodeID:    fixture.Report.Endpoint.Nodes[fixture.Client54002NodeID],
				fixture.ClientProcess1NodeID: fixture.Report.Process.Nodes[fixture.ClientProcess1NodeID],
				fixture.ClientProcess2NodeID: fixture.Report.Process.Nodes[fixture.ClientProcess2NodeID],
			},
			Node: report.MakeNode().WithAdjacent(fixture.ServerContainerID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(30),
				EgressByteCount:   newu64(300),
			},
			ControlNode: fixture.ClientContainerNodeID,
		},
		fixture.ServerContainerID: {
			ID:         fixture.ServerContainerID,
			LabelMajor: "server",
			LabelMinor: fixture.ServerHostName,
			Rank:       fixture.ServerContainerImageName,
			Pseudo:     false,
			Children: render.Children{
				fixture.Server80NodeID:      fixture.Report.Endpoint.Nodes[fixture.Server80NodeID],
				fixture.ServerProcessNodeID: fixture.Report.Process.Nodes[fixture.ServerProcessNodeID],
			},
			Node: report.MakeNode(),
			EdgeMetadata: report.EdgeMetadata{
				IngressPacketCount: newu64(210),
				IngressByteCount:   newu64(2100),
			},
			ControlNode: fixture.ServerContainerNodeID,
		},
		uncontainedServerID: {
			ID:         uncontainedServerID,
			LabelMajor: render.UncontainedMajor,
			LabelMinor: fixture.ServerHostName,
			Rank:       "",
			Pseudo:     true,
			Children: render.Children{
				fixture.NonContainerProcessNodeID: fixture.Report.Process.Nodes[fixture.NonContainerProcessNodeID],
				fixture.NonContainerNodeID:        fixture.Report.Endpoint.Nodes[fixture.NonContainerNodeID],
			},
			Node:         report.MakeNode().WithAdjacent(render.TheInternetID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		render.TheInternetID: theInternetNode(fixture.ServerContainerID),
	}).Prune()

	RenderedContainerImages = (render.RenderableNodes{
		fixture.ClientContainerImageName: {
			ID:         fixture.ClientContainerImageName,
			LabelMajor: fixture.ClientContainerImageName,
			LabelMinor: "1 container",
			Rank:       fixture.ClientContainerImageName,
			Pseudo:     false,
			Children: render.Children{
				fixture.Client54001NodeID:     fixture.Report.Endpoint.Nodes[fixture.Client54001NodeID],
				fixture.Client54002NodeID:     fixture.Report.Endpoint.Nodes[fixture.Client54002NodeID],
				fixture.ClientProcess1NodeID:  fixture.Report.Process.Nodes[fixture.ClientProcess1NodeID],
				fixture.ClientProcess2NodeID:  fixture.Report.Process.Nodes[fixture.ClientProcess2NodeID],
				fixture.ClientContainerNodeID: fixture.Report.Container.Nodes[fixture.ClientContainerNodeID],
			},
			Node: report.MakeNode().WithAdjacent(fixture.ServerContainerImageName),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(30),
				EgressByteCount:   newu64(300),
			},
		},
		fixture.ServerContainerImageName: {
			ID:         fixture.ServerContainerImageName,
			LabelMajor: fixture.ServerContainerImageName,
			LabelMinor: "1 container",
			Rank:       fixture.ServerContainerImageName,
			Pseudo:     false,
			Children: render.Children{
				fixture.Server80NodeID:        fixture.Report.Endpoint.Nodes[fixture.Server80NodeID],
				fixture.ServerProcessNodeID:   fixture.Report.Process.Nodes[fixture.ServerProcessNodeID],
				fixture.ServerContainerNodeID: fixture.Report.Container.Nodes[fixture.ServerContainerNodeID],
			},
			Node: report.MakeNode(),
			EdgeMetadata: report.EdgeMetadata{
				IngressPacketCount: newu64(210),
				IngressByteCount:   newu64(2100),
			},
		},
		uncontainedServerID: {
			ID:         uncontainedServerID,
			LabelMajor: render.UncontainedMajor,
			LabelMinor: fixture.ServerHostName,
			Rank:       "",
			Pseudo:     true,
			Children: render.Children{
				fixture.NonContainerNodeID:        fixture.Report.Endpoint.Nodes[fixture.NonContainerNodeID],
				fixture.NonContainerProcessNodeID: fixture.Report.Process.Nodes[fixture.NonContainerProcessNodeID],
			},
			Node:         report.MakeNode().WithAdjacent(render.TheInternetID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		render.TheInternetID: theInternetNode(fixture.ServerContainerImageName),
	}).Prune()

	ServerHostRenderedID = render.MakeHostID(fixture.ServerHostID)
	ClientHostRenderedID = render.MakeHostID(fixture.ClientHostID)
	pseudoHostID1        = render.MakePseudoNodeID(fixture.UnknownClient1IP, fixture.ServerIP)
	pseudoHostID2        = render.MakePseudoNodeID(fixture.UnknownClient3IP, fixture.ServerIP)

	RenderedHosts = (render.RenderableNodes{
		ServerHostRenderedID: {
			ID:         ServerHostRenderedID,
			LabelMajor: "server",       // before first .
			LabelMinor: "hostname.com", // after first .
			Rank:       "hostname.com",
			Pseudo:     false,
			Children:   render.Children{},
			Node:       report.MakeNode(),
			EdgeMetadata: report.EdgeMetadata{
				MaxConnCountTCP: newu64(3),
			},
		},
		ClientHostRenderedID: {
			ID:         ClientHostRenderedID,
			LabelMajor: "client",       // before first .
			LabelMinor: "hostname.com", // after first .
			Rank:       "hostname.com",
			Pseudo:     false,
			Children: render.Children{
				fixture.ClientAddressNodeID: fixture.Report.Address.Nodes[fixture.ClientAddressNodeID],
			},
			Node: report.MakeNode().WithAdjacent(ServerHostRenderedID),
			EdgeMetadata: report.EdgeMetadata{
				MaxConnCountTCP: newu64(3),
			},
		},
		pseudoHostID1: {
			ID:           pseudoHostID1,
			LabelMajor:   fixture.UnknownClient1IP,
			Pseudo:       true,
			Node:         report.MakeNode().WithAdjacent(ServerHostRenderedID),
			EdgeMetadata: report.EdgeMetadata{},
			Children: render.Children{
				fixture.UnknownAddress1NodeID: fixture.Report.Address.Nodes[fixture.UnknownAddress1NodeID],
				fixture.UnknownAddress2NodeID: fixture.Report.Address.Nodes[fixture.UnknownAddress2NodeID],
			},
		},
		pseudoHostID2: {
			ID:           pseudoHostID2,
			LabelMajor:   fixture.UnknownClient3IP,
			Pseudo:       true,
			Node:         report.MakeNode().WithAdjacent(ServerHostRenderedID),
			EdgeMetadata: report.EdgeMetadata{},
			Children: render.Children{
				fixture.UnknownAddress3NodeID: fixture.Report.Address.Nodes[fixture.UnknownAddress3NodeID],
			},
		},
		render.TheInternetID: {
			ID:           render.TheInternetID,
			LabelMajor:   render.TheInternetMajor,
			Pseudo:       true,
			Node:         report.MakeNode().WithAdjacent(ServerHostRenderedID),
			EdgeMetadata: report.EdgeMetadata{},
			Children: render.Children{
				fixture.RandomAddressNodeID: fixture.Report.Address.Nodes[fixture.RandomAddressNodeID],
			},
		},
	}).Prune()

	RenderedPods = (render.RenderableNodes{
		"ping/pong-a": {
			ID:         "ping/pong-a",
			LabelMajor: "pong-a",
			LabelMinor: "1 container",
			Rank:       "ping/pong-a",
			Pseudo:     false,
			Children: render.Children{
				fixture.Client54001NodeID:          fixture.Report.Endpoint.Nodes[fixture.Client54001NodeID],
				fixture.Client54002NodeID:          fixture.Report.Endpoint.Nodes[fixture.Client54002NodeID],
				fixture.ClientProcess1NodeID:       fixture.Report.Process.Nodes[fixture.ClientProcess1NodeID],
				fixture.ClientProcess2NodeID:       fixture.Report.Process.Nodes[fixture.ClientProcess2NodeID],
				fixture.ClientContainerNodeID:      fixture.Report.Container.Nodes[fixture.ClientContainerNodeID],
				fixture.ClientContainerImageNodeID: fixture.Report.ContainerImage.Nodes[fixture.ClientContainerImageNodeID],
				fixture.ClientPodNodeID:            fixture.Report.Pod.Nodes[fixture.ClientPodNodeID],
			},
			Node: report.MakeNode().WithAdjacent("ping/pong-b"),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(30),
				EgressByteCount:   newu64(300),
			},
		},
		"ping/pong-b": {
			ID:         "ping/pong-b",
			LabelMajor: "pong-b",
			LabelMinor: "1 container",
			Rank:       "ping/pong-b",
			Pseudo:     false,
			Children: render.Children{
				fixture.Server80NodeID:             fixture.Report.Endpoint.Nodes[fixture.Server80NodeID],
				fixture.ServerProcessNodeID:        fixture.Report.Process.Nodes[fixture.ServerProcessNodeID],
				fixture.ServerContainerNodeID:      fixture.Report.Container.Nodes[fixture.ServerContainerNodeID],
				fixture.ServerContainerImageNodeID: fixture.Report.ContainerImage.Nodes[fixture.ServerContainerImageNodeID],
				fixture.ServerPodNodeID:            fixture.Report.Pod.Nodes[fixture.ServerPodNodeID],
			},
			Node: report.MakeNode(),
			EdgeMetadata: report.EdgeMetadata{
				IngressPacketCount: newu64(210),
				IngressByteCount:   newu64(2100),
			},
		},
		uncontainedServerID: {
			ID:         uncontainedServerID,
			LabelMajor: render.UncontainedMajor,
			LabelMinor: fixture.ServerHostName,
			Rank:       "",
			Pseudo:     true,
			Children: render.Children{
				fixture.NonContainerNodeID:        fixture.Report.Endpoint.Nodes[fixture.NonContainerNodeID],
				fixture.NonContainerProcessNodeID: fixture.Report.Process.Nodes[fixture.NonContainerProcessNodeID],
			},
			Node:         report.MakeNode().WithAdjacent(render.TheInternetID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		render.TheInternetID: {
			ID:         render.TheInternetID,
			LabelMajor: render.TheInternetMajor,
			Pseudo:     true,
			Node:       report.MakeNode().WithAdjacent("ping/pong-b"),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(60),
				EgressByteCount:   newu64(600),
			},
			Children: render.Children{
				fixture.RandomClientNodeID:   fixture.Report.Endpoint.Nodes[fixture.RandomClientNodeID],
				fixture.GoogleEndpointNodeID: fixture.Report.Endpoint.Nodes[fixture.GoogleEndpointNodeID],
			},
		},
	}).Prune()

	RenderedPodServices = (render.RenderableNodes{
		"ping/pongservice": {
			ID:         fixture.ServiceID,
			LabelMajor: "pongservice",
			LabelMinor: "2 pods",
			Rank:       fixture.ServiceID,
			Pseudo:     false,
			Children: render.Children{
				fixture.Client54001NodeID:          fixture.Report.Endpoint.Nodes[fixture.Client54001NodeID],
				fixture.Client54002NodeID:          fixture.Report.Endpoint.Nodes[fixture.Client54002NodeID],
				fixture.ClientProcess1NodeID:       fixture.Report.Process.Nodes[fixture.ClientProcess1NodeID],
				fixture.ClientProcess2NodeID:       fixture.Report.Process.Nodes[fixture.ClientProcess2NodeID],
				fixture.ClientContainerNodeID:      fixture.Report.Container.Nodes[fixture.ClientContainerNodeID],
				fixture.ClientContainerImageNodeID: fixture.Report.ContainerImage.Nodes[fixture.ClientContainerImageNodeID],
				fixture.ClientPodNodeID:            fixture.Report.Pod.Nodes[fixture.ClientPodNodeID],
				fixture.Server80NodeID:             fixture.Report.Endpoint.Nodes[fixture.Server80NodeID],
				fixture.ServiceNodeID:              fixture.Report.Endpoint.Nodes[fixture.ServiceNodeID],
				fixture.ServerProcessNodeID:        fixture.Report.Process.Nodes[fixture.ServerProcessNodeID],
				fixture.ServerContainerNodeID:      fixture.Report.Container.Nodes[fixture.ServerContainerNodeID],
				fixture.ServerContainerImageNodeID: fixture.Report.ContainerImage.Nodes[fixture.ServerContainerImageNodeID],
				fixture.ServerPodNodeID:            fixture.Report.Pod.Nodes[fixture.ServerPodNodeID],
			},
			Node: report.MakeNode().WithAdjacent(fixture.ServiceID), // ?? Shouldn't be adjacent to itself?
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount:  newu64(30),
				EgressByteCount:    newu64(300),
				IngressPacketCount: newu64(210),
				IngressByteCount:   newu64(2100),
			},
		},
		uncontainedServerID: {
			ID:         uncontainedServerID,
			LabelMajor: render.UncontainedMajor,
			LabelMinor: fixture.ServerHostName,
			Rank:       "",
			Pseudo:     true,
			Children: render.Children{
				fixture.NonContainerNodeID:        fixture.Report.Endpoint.Nodes[fixture.NonContainerNodeID],
				fixture.NonContainerProcessNodeID: fixture.Report.Process.Nodes[fixture.NonContainerProcessNodeID],
			},
			Node:         report.MakeNode().WithAdjacent(render.TheInternetID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		render.TheInternetID: {
			ID:         render.TheInternetID,
			LabelMajor: render.TheInternetMajor,
			Pseudo:     true,
			Node:       report.MakeNode().WithAdjacent(fixture.ServiceID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(60),
				EgressByteCount:   newu64(600),
			},
			Children: render.Children{
				fixture.RandomClientNodeID:   fixture.Report.Endpoint.Nodes[fixture.RandomClientNodeID],
				fixture.GoogleEndpointNodeID: fixture.Report.Endpoint.Nodes[fixture.GoogleEndpointNodeID],
			},
		},
	}).Prune()
)

func newu64(value uint64) *uint64 { return &value }
