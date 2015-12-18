import React from 'react';

import NodeDetails from './node-details';

export default class Details extends React.Component {

  render() {
    return (
      <div id="details">
        {this.props.details.toIndexedSeq().map(obj => {
          return (
            <div className="details-wrapper" key={obj.id}>
              <NodeDetails details={obj.details} controlError={this.props.controlError}
                controlPending={this.props.controlPending} nodes={this.props.nodes}
                nodeId={obj.id} />
            </div>
          );
        })}
      </div>
    );
  }
}
