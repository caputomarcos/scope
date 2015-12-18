import React from 'react';

import NodeDetails from './node-details';

export default class Details extends React.Component {

  render() {
    return (
      <div className="details">
        {this.props.details.toIndexedSeq().map((obj, index) => {
          const offset = (420 + 12) * index;
          const style = {
            bottom: 48,
            right: 36 + offset,
            top: 24
          };

          return (
            <div id="details" style={style} key={obj.id}>
              <div className="details-wrapper">
                <NodeDetails details={obj.details} controlError={this.props.controlError}
                  controlPending={this.props.controlPending} nodes={this.props.nodes}
                  nodeId={obj.id} />
              </div>
            </div>
          );
        })};
      </div>
    );
  }
}
