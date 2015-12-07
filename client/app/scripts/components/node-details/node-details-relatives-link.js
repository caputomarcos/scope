import React from 'react';

import { clickRelative } from '../../actions/app-actions';

export default class NodeDetailsRelativesLink extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.handleClick = this.handleClick.bind(this);
  }

  handleClick(ev) {
    ev.preventDefault();
    clickRelative(this.props.relative.id, this.props.relative.topologyId);
  }

  render() {
    const title = `View in ${this.props.relative.topologyId}: ${this.props.relative.label}`;
    return (
      <span className="node-details-relatives-link" title={title} onClick={this.handleClick}>
        {this.props.relative.label}
      </span>
    );
  }
}
