import React from 'react';

import NodeDetailsHealthItem from './node-details-health-item';

export default class NodeDetailsHealth extends React.Component {
  render() {
    // TODO implement grouping metrics
    return (
      <div className="node-details-health">
        {this.props.metrics && this.props.metrics.slice(0, 3).map(item => {
          return <NodeDetailsHealthItem key={item.id} item={item} />;
        })}
      </div>
    );
  }
}
