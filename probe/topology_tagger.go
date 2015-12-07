package probe

import (
	"github.com/weaveworks/scope/report"
)

const (
	// Topology is the Node key for the origin topology.
	Topology = "topology"

	// ID is the key where this can be looked up in the origin topology.
	ID = "id"

	// Rank is the priority this node should have when merging, bigger is higher priority
	Rank = "rank"
)

type topologyTagger struct{}

// NewTopologyTagger tags each node with the topology that it comes from. It's
// kind of a proof-of-concept tagger, useful primarily for debugging.
func NewTopologyTagger() Tagger {
	return &topologyTagger{}
}

func (topologyTagger) Name() string { return "Topology" }

// Tag implements Tagger
func (topologyTagger) Tag(r report.Report) (report.Report, error) {
	for rank, t := range []struct {
		val string
		*report.Topology
	}{
		{"endpoint", &(r.Endpoint)},
		{"address", &(r.Address)},
		{"process", &(r.Process)},
		{"container", &(r.Container)},
		{"container_image", &(r.ContainerImage)},
		{"pod", &(r.Pod)},
		{"service", &(r.Service)},
		{"host", &(r.Host)},
		{"overlay", &(r.Overlay)},
	} {
		for id, node := range t.Topology.Nodes {
			metadata := map[string]string{ID: id, Topology: t.val}
			t.Topology.AddNode(id, node.WithMetadata(metadata).WithRank(rank))
		}
	}
	return r, nil
}
