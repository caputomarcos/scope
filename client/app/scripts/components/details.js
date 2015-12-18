import React from 'react';

import NodeDetails from './node-details';

export default class Details extends React.Component {

  render() {
    return (
      <div className="details">
        {this.props.details.toIndexedSeq().map((obj, index) => {
          const offset = 10;
          const style = {
            bottom: 48 - index * offset,
            right: 36 - index * offset,
            top: 24 + index * offset
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
