import React from 'react';

import NodeDetails from './node-details';

export default class Details extends React.Component {

  render() {
    const details = this.props.details.toIndexedSeq();
    const pile = details.slice(0, -2);
    const openDetails = details.slice(-2);
    return (
      <div className="details">
        {pile.map((obj, index) => {
          const offset = (index + 1) * 8;
          const style = {
            right: 36 - offset,
            top: 24 - offset
          };
          return (
            <div className="details-wrapper" style={style} key={obj.id}>
              <div className="details-wrapper-row">
                <NodeDetails details={obj.details} controlError={this.props.controlError}
                  controlPending={this.props.controlPending} nodes={this.props.nodes}
                  nodeId={obj.id} />
              </div>
              <div className="details-wrapper-row details-wrapper-row-dummy" />
            </div>
          );
        })}
        <div className="details-wrapper">
          {openDetails.map((obj) => {
            return (
              <div className="details-wrapper-row" key={obj.id}>
                <NodeDetails details={obj.details} controlError={this.props.controlError}
                  controlPending={this.props.controlPending} nodes={this.props.nodes}
                  nodeId={obj.id} />
              </div>
            );
          })}
        </div>
      </div>
    );
  }
}
