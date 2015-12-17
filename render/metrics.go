package render

import (
	"encoding/json"
	"math"
	"time"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

const (
	defaultFormat  = ""
	filesizeFormat = "filesize"
	percentFormat  = "percent"
)

type MetricRow struct {
	ID     string  `json:"id"`
	Label  string  `json:"label"`
	Format string  `json:"format,omitempty"`
	Group  string  `json:"group,omitempty"`
	Value  float64 `json:"value"`
	*report.Metric
}

func (m MetricRow) MarshalJSON() ([]byte, error) {
	// TODO(paulbellamy): This is a total hack to workaround go taking the
	// MarshalJSON method from the embedded report.Metric
	samples := []report.Sample{}
	if m.Samples != nil {
		m.Samples.Reverse().ForEach(func(s interface{}) {
			samples = append(samples, s.(report.Sample))
		})
	}
	j := map[string]interface{}{
		"id":      m.ID,
		"label":   m.Label,
		"value":   m.Value,
		"samples": samples,
		"max":     m.Max,
		"min":     m.Min,
		"first":   renderTime(m.First),
		"last":    renderTime(m.Last),
	}
	if m.Format != "" {
		j["format"] = m.Format
	}
	if m.Group != "" {
		j["group"] = m.Group
	}
	return json.Marshal(j)
}

func renderTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339Nano)
}

func metricRow(id, label string, metric report.Metric, format, group string) MetricRow {
	var last float64
	if s := metric.LastSample(); s != nil {
		last = s.Value
	}
	return MetricRow{
		ID:     id,
		Label:  label,
		Format: format,
		Group:  group,
		Value:  toFixed(last, 2),
		Metric: &metric,
	}
}

// toFixed truncates decimals of float64 down to specified precision
func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(int64(num*output)) / output
}

// NodeMetrics produces a table (to be consumed directly by the UI) based on
// an origin ID, which is (optimistically) a node ID in one of our topologies.
func NodeMetrics(r report.Report, n RenderableNode) []MetricRow {
	renderers := map[string]struct {
		t report.Topology
		r func(report.Node) []MetricRow
	}{
		"process":   {r.Process, processNodeMetrics},
		"container": {r.Container, containerNodeMetrics},
		"host":      {r.Host, hostNodeMetrics},
	}
	if renderer, ok := renderers[n.SummaryTopology]; ok {
		if nmd, ok := renderer.t.Nodes[n.SummaryID]; ok {
			return renderer.r(nmd)
		}
	}
	return nil
}

func processNodeMetrics(nmd report.Node) []MetricRow {
	rows := []MetricRow{}
	for _, tuple := range []struct {
		ID, Label, fmt string
	}{
		{process.CPUUsage, "CPU Usage", percentFormat},
		{process.MemoryUsage, "Memory Usage", filesizeFormat},
	} {
		if val, ok := nmd.Metrics[tuple.ID]; ok {
			rows = append(rows, metricRow(
				tuple.ID,
				tuple.Label,
				val,
				tuple.fmt,
				"",
			))
		}
	}
	return rows
}

func containerNodeMetrics(nmd report.Node) []MetricRow {
	rows := []MetricRow{}
	if val, ok := nmd.Metrics[docker.CPUTotalUsage]; ok {
		rows = append(rows, metricRow(
			docker.CPUTotalUsage,
			"CPU Usage",
			val,
			percentFormat,
			"",
		))
	}
	if val, ok := nmd.Metrics[docker.MemoryUsage]; ok {
		rows = append(rows, metricRow(
			docker.MemoryUsage,
			"Memory Usage",
			val,
			filesizeFormat,
			"",
		))
	}
	return rows
}

func hostNodeMetrics(nmd report.Node) []MetricRow {
	// Ensure that all metrics have the same max
	maxLoad := 0.0
	for _, id := range []string{host.Load1, host.Load5, host.Load15} {
		if metric, ok := nmd.Metrics[id]; ok {
			if metric.Len() == 0 {
				continue
			}
			if metric.Max > maxLoad {
				maxLoad = metric.Max
			}
		}
	}

	rows := []MetricRow{}
	for _, tuple := range []struct{ ID, Label, fmt string }{
		{host.CPUUsage, "CPU Usage", percentFormat},
		{host.MemUsage, "Memory Usage", percentFormat},
	} {
		if val, ok := nmd.Metrics[tuple.ID]; ok {
			rows = append(rows, metricRow(tuple.ID, tuple.Label, val, tuple.fmt, ""))
		}
	}
	for _, tuple := range []struct{ ID, Label string }{
		{host.Load1, "Load (1m)"},
		{host.Load5, "Load (5m)"},
		{host.Load15, "Load (15m)"},
	} {
		if val, ok := nmd.Metrics[tuple.ID]; ok {
			val.Max = maxLoad
			rows = append(rows, metricRow(tuple.ID, tuple.Label, val, defaultFormat, "load"))
		}
	}
	return rows
}
