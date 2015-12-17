package render_test

import (
	"testing"
)

func TestNodeMetrics(t *testing.T) {
	t.Error("pending")
	/*
		if _, ok := render.NodeSummary(fixture.Report, "not-found", false, false); ok {
			t.Errorf("unknown origin ID gave unexpected success")
		}
		for originID, want := range map[string]render.Table{
			fixture.ServerProcessNodeID: {
				Title:   fmt.Sprintf(`Process "apache" (%s)`, fixture.ServerPID),
				Numeric: false,
				Rank:    2,
				Rows:    []render.Row{},
			},
			fixture.ServerHostNodeID: {
				Title:   fmt.Sprintf("Host %q", fixture.ServerHostName),
				Numeric: false,
				Rank:    1,
				Rows: []render.Row{
					{Key: "Load (1m)", ValueMajor: "0.01", Metric: &fixture.LoadMetric, ValueType: "sparkline"},
					{Key: "Load (5m)", ValueMajor: "0.01", Metric: &fixture.LoadMetric, ValueType: "sparkline"},
					{Key: "Load (15m)", ValueMajor: "0.01", Metric: &fixture.LoadMetric, ValueType: "sparkline"},
					{Key: "Operating system", ValueMajor: "Linux"},
				},
			},
		} {
			have, ok := render.OriginTable(fixture.Report, originID, false, false)
			if !ok {
				t.Errorf("%q: not OK", originID)
				continue
			}
			if !reflect.DeepEqual(want, have) {
				t.Errorf("%q: %s", originID, test.Diff(want, have))
			}
		}

		// Test host/container tags
		for originID, want := range map[string]render.Table{
			fixture.ServerProcessNodeID: {
				Title:   fmt.Sprintf(`Process "apache" (%s)`, fixture.ServerPID),
				Numeric: false,
				Rank:    2,
				Rows: []render.Row{
					{Key: "Host", ValueMajor: fixture.ServerHostID},
					{Key: "Container ID", ValueMajor: fixture.ServerContainerID},
				},
			},
			fixture.ServerContainerNodeID: {
				Title:   `Container "server"`,
				Numeric: false,
				Rank:    3,
				Rows: []render.Row{
					{Key: "Host", ValueMajor: fixture.ServerHostID},
					{Key: "State", ValueMajor: "running"},
					{Key: "ID", ValueMajor: fixture.ServerContainerID},
					{Key: "Image ID", ValueMajor: fixture.ServerContainerImageID},
					{Key: fmt.Sprintf(`Label %q`, render.AmazonECSContainerNameLabel), ValueMajor: `server`},
					{Key: `Label "foo1"`, ValueMajor: `bar1`},
					{Key: `Label "foo2"`, ValueMajor: `bar2`},
					{Key: `Label "io.kubernetes.pod.name"`, ValueMajor: "ping/pong-b"},
				},
			},
		} {
			have, ok := render.OriginTable(fixture.Report, originID, true, true)
			if !ok {
				t.Errorf("%q: not OK", originID)
				continue
			}
			if !reflect.DeepEqual(want, have) {
				t.Errorf("%q: %s", originID, test.Diff(want, have))
			}
		}
	*/
}
