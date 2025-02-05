package report

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/weaveworks/scope/common/mtime"
)

// Controls describe the control tags within the Nodes
type Controls map[string]Control

// A Control basically describes an RPC
type Control struct {
	ID    string `json:"id"`
	Human string `json:"human"`
	Icon  string `json:"icon"` // from https://fortawesome.github.io/Font-Awesome/cheatsheet/ please
}

// Merge merges other with cs, returning a fresh Controls.
func (cs Controls) Merge(other Controls) Controls {
	result := cs.Copy()
	for k, v := range other {
		result[k] = v
	}
	return result
}

// Copy produces a copy of cs.
func (cs Controls) Copy() Controls {
	result := Controls{}
	for k, v := range cs {
		result[k] = v
	}
	return result
}

// AddControl returns a fresh Controls, c added to cs.
func (cs Controls) AddControl(c Control) {
	cs[c.ID] = c
}

// NodeControls represent the individual controls that are valid for a given
// node at a given point in time.  Its is immutable. A zero-value for Timestamp
// indicated this NodeControls is 'not set'.
type NodeControls struct {
	Timestamp time.Time
	Controls  StringSet
}

// MakeNodeControls makes a new NodeControls
func MakeNodeControls() NodeControls {
	return NodeControls{
		Controls: MakeStringSet(),
	}
}

// Copy is a noop, as NodeControls is immutable
func (nc NodeControls) Copy() NodeControls {
	return nc
}

// Merge returns the newest of the two NodeControls; it does not take the union
// of the valid Controls.
func (nc NodeControls) Merge(other NodeControls) NodeControls {
	if nc.Timestamp.Before(other.Timestamp) {
		return other
	}
	return nc
}

// Add the new control IDs to this NodeControls, producing a fresh NodeControls.
func (nc NodeControls) Add(ids ...string) NodeControls {
	return NodeControls{
		Timestamp: mtime.Now(),
		Controls:  nc.Controls.Add(ids...),
	}
}

// WireNodeControls is the intermediate type for json encoding.
type WireNodeControls struct {
	Timestamp string    `json:"timestamp,omitempty"`
	Controls  StringSet `json:"controls,omitempty"`
}

// MarshalJSON implements json.Marshaller
func (nc NodeControls) MarshalJSON() ([]byte, error) {
	buf := bytes.Buffer{}
	err := json.NewEncoder(&buf).Encode(WireNodeControls{
		Timestamp: renderTime(nc.Timestamp),
		Controls:  nc.Controls,
	})
	return buf.Bytes(), err
}

// UnmarshalJSON implements json.Unmarshaler
func (nc *NodeControls) UnmarshalJSON(input []byte) error {
	in := WireNodeControls{}
	if err := json.NewDecoder(bytes.NewBuffer(input)).Decode(&in); err != nil {
		return err
	}
	*nc = NodeControls{
		Timestamp: parseTime(in.Timestamp),
		Controls:  in.Controls,
	}
	return nil
}
