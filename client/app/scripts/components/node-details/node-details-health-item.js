import React from 'react';

import Sparkline from '../sparkline';
import { formatMetric } from '../../utils/string-utils';

export default class NodeDetailsHealth extends React.Component {
  render() {
    return (
      <div className="node-details-health-item">
      <div className="node-details-health-item-value">{formatMetric(this.props.item.value, this.props.item)}</div>
        <div className="node-details-health-item-sparkline">
          <Sparkline data={this.props.item.samples} min={0} max={this.props.item.max}
            first={this.props.item.first} last={this.props.item.last} interpolate="none" />
        </div>
        <div className="node-details-health-item-label">{this.props.item.label}</div>
      </div>
    );
  }
}
