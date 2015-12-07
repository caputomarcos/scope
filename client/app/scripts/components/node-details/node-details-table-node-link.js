import React from 'react';

import { clickRelative } from '../../actions/app-actions';

export default class NodeDetailsTableNodeLink extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.handleClick = this.handleClick.bind(this);
  }

  handleClick(ev) {
    ev.preventDefault();
    clickRelative(this.props.id, this.props.topologyId);
  }

  render() {
    const title = `<topologyType>: ${this.props.label}`;
    return (
      <span className="node-details-table-node-link truncate" title={title}>
        {this.props.label}
      </span>
    );
  }
}
