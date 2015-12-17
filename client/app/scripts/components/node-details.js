import _ from 'lodash';
import React from 'react';

import NodeDetailsControls from './node-details/node-details-controls';
import NodeDetailsHealth from './node-details/node-details-health';
import NodeDetailsInfo from './node-details/node-details-info';
import NodeDetailsRelatives from './node-details/node-details-relatives';
import NodeDetailsTable from './node-details/node-details-table';
import { brightenColor, getNodeColorDark } from '../utils/color-utils';
import { resetDocumentTitle, setDocumentTitle } from '../utils/title-utils';

export default class NodeDetails extends React.Component {
  componentDidMount() {
    this.updateTitle();
  }

  componentWillUnmount() {
    resetDocumentTitle();
  }

  renderLoading() {
    const node = this.props.nodes.get(this.props.nodeId);
    const nodeColor = getNodeColorDark(node.get('rank'), node.get('label_major'));
    const styles = {
      header: {
        'backgroundColor': nodeColor
      }
    };

    return (
      <div className="node-details">
        <div className="node-details-header" style={styles.header}>
          <div className="node-details-header-wrapper">
            <h2 className="node-details-header-label truncate">
              {node.get('label_major')}
            </h2>
            <div className="node-details-header-label-minor truncate">
              {node.get('label_minor')}
            </div>
          </div>
        </div>
        <div className="node-details-content">
          <div className="node-details-content-loading">
            <span className="fa fa-circle-o-notch fa-spin" />
          </div>
        </div>
      </div>
    );
  }

  renderNotAvailable() {
    return (
      <div className="node-details">
        <div className="node-details-header node-details-header-notavailable">
          <div className="node-details-header-wrapper">
            <h2 className="node-details-header-label">
              n/a
            </h2>
            <div className="node-details-header-label-minor truncate">
              {this.props.nodeId}
            </div>
          </div>
        </div>
        <div className="node-details-content">
          <p className="node-details-content-info">
            This node is not visible to Scope anymore.
            The node will re-appear if it communicates again.
          </p>
        </div>
      </div>
    );
  }

  renderTable(table) {
    const key = _.snakeCase(table.title);
    return <NodeDetailsTable title={table.title} key={key} rows={table.rows} isNumeric={table.numeric} />;
  }

  render() {
    const details = this.props.details;
    const nodeExists = this.props.nodes && this.props.nodes.has(this.props.nodeId);

    if (details) {
      return this.renderDetails();
    }

    if (!nodeExists) {
      return this.renderNotAvailable();
    }

    return this.renderLoading();
  }

  renderDetails() {
    const details = this.props.details;
    const nodeColor = getNodeColorDark(details.rank, details.label_major);
    const styles = {
      controls: {
        'backgroundColor': brightenColor(nodeColor)
      },
      header: {
        'backgroundColor': nodeColor
      }
    };

    let summary;
    if (details.metadata || details.metrics) {
      summary = (
        <div className="node-details-content-section">
          <div className="node-details-content-section-header">Status</div>
          <NodeDetailsHealth metrics={details.metrics} />
          <NodeDetailsInfo metadata={details.metadata} />
        </div>
      );
    }

    return (
      <div className="node-details">
        <div className="node-details-header" style={styles.header}>
          <div className="node-details-header-wrapper">
            <h2 className="node-details-header-label truncate" title={details.label}>
              {details.label}
            </h2>
            <div className="node-details-header-relatives">
              <NodeDetailsRelatives relatives={details.parents} />
            </div>
          </div>
        </div>

        {details.controls && details.controls.length > 0 && <div className="node-details-controls-wrapper" style={styles.controls}>
          <NodeDetailsControls controls={details.controls}
            pending={this.props.controlPending} error={this.props.controlError} />
        </div>}

        <div className="node-details-content">
          {summary && summary}
          {details.children && details.children.map(children => {
            return (
              <div className="node-details-content-section" key={children.topologyId}>
                <NodeDetailsTable nodes={children.nodes} label={children.label} />
              </div>
            );
          })}
        </div>
      </div>
    );
  }

  componentDidUpdate() {
    this.updateTitle();
  }

  updateTitle() {
    setDocumentTitle(this.props.details && this.props.details.label_major);
  }
}
