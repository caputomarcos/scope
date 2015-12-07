import React from 'react';

import NodeDetailsTableNodeLink from './node-details-table-node-link';
import { formatMetric } from '../../utils/string-utils';

export default class NodeDetailsTable extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.DEFAULT_LIMIT = 5;
    this.state = {
      limit: this.DEFAULT_LIMIT
    };
    this.handleLimitClick = this.handleLimitClick.bind(this);
  }

  handleLimitClick(ev) {
    ev.preventDefault();
    const limit = this.state.limit ? 0 : this.DEFAULT_LIMIT;
    this.setState({limit: limit});
  }

  renderHeaders() {
    if (this.props.nodes && this.props.nodes.length > 0) {
      const headers = this.props.nodes[0].metrics && this.props.nodes[0].metrics.map(metric => {
        const { id, label } = metric;
        return { id, label };
      });

      return (
        <div className="node-details-table-header-wrapper">
          <div className="node-details-table-header" key="label">{this.props.label}</div>
          {headers && headers.map(header => {
            return <div className="node-details-table-header" key={header.id}>{header.label}</div>;
          })}
        </div>
      );
    }
    return '';
  }

  render() {
    const headers = this.renderHeaders();
    let nodes = this.props.nodes;
    const limited = nodes && this.state.limit > 0 && nodes.length > this.state.limit;
    const showLimitAction = nodes && (limited || (this.state.limit === 0 && nodes.length > this.DEFAULT_LIMIT));
    const limitActionText = limited ? 'Show more' : 'Show less';
    if (nodes && limited) {
      nodes = nodes.slice(0, this.state.limit);
    }

    return (
      <div className="node-details-table">
        {headers}
        {nodes && nodes.map(function(node) {
          return (
            <div className="node-details-table-node" key={node.id}>
              <div className="node-details-table-node-label">
                <NodeDetailsTableNodeLink label={node.label} topologyId={node.topologyId} id={node.id} />
              </div>
              {node.metrics && node.metrics.map(metric => {
                return (
                  <div className="node-details-table-node-value" key={metric.id}>
                    {formatMetric(metric.value, metric)}
                  </div>
                );
              })}
            </div>
          );
        })}
        {showLimitAction && <div className="node-details-table-more" onClick={this.handleLimitClick}>{limitActionText}</div>}
      </div>
    );
  }
}
