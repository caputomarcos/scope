import _ from 'lodash';
import debug from 'debug';
import Immutable from 'immutable';
import { Store } from 'flux/utils';

import AppDispatcher from '../dispatcher/app-dispatcher';
import ActionTypes from '../constants/action-types';
import { EDGE_ID_SEPARATOR } from '../constants/naming';

const makeOrderedMap = Immutable.OrderedMap;
const makeSet = Immutable.Set;
const log = debug('scope:app-store');

const error = debug('scope:error');

// Helpers

function findTopologyById(subTree, topologyId) {
  let foundTopology;

  _.each(subTree, function(topology) {
    if (_.endsWith(topology.url, topologyId)) {
      foundTopology = topology;
    }
    if (!foundTopology) {
      foundTopology = findTopologyById(topology.sub_topologies, topologyId);
    }
    if (foundTopology) {
      return false;
    }
  });

  return foundTopology;
}

function makeNode(node) {
  return {
    id: node.id,
    label_major: node.label_major,
    label_minor: node.label_minor,
    rank: node.rank,
    pseudo: node.pseudo,
    adjacency: node.adjacency
  };
}

// Initial values

let topologyOptions = makeOrderedMap(); // topologyId -> options
let adjacentNodes = makeSet();
let controlError = null;
let controlPending = false;
let currentTopology = null;
let currentTopologyId = 'containers';
let errorUrl = null;
let hostname = '...';
let version = '...';
let mouseOverEdgeId = null;
let mouseOverNodeId = null;
let nodeDetails = makeOrderedMap(); // nodeId -> details
let nodes = makeOrderedMap(); // nodeId -> node
let selectedNodeId = null;
let topologies = [];
let topologiesLoaded = false;
let topologyUrlsById = makeOrderedMap(); // topologyId -> topologyUrl
let routeSet = false;
let controlPipes = makeOrderedMap(); // pipeId -> controlPipe
let websocketClosed = true;

// adds ID field to topology (based on last part of URL path) and save urls in
// map for easy lookup
function processTopologies(topologyList) {
  _.each(topologyList, function(topology) {
    topology.id = topology.url.split('/').pop();
    topologyUrlsById = topologyUrlsById.set(topology.id, topology.url);
    processTopologies(topology.sub_topologies);
  });
  return topologyList;
}

function setTopology(topologyId) {
  currentTopologyId = topologyId;
  currentTopology = findTopologyById(topologies, topologyId);
}

function setDefaultTopologyOptions(topologyList) {
  _.each(topologyList, function(topology) {
    let defaultOptions = makeOrderedMap();
    _.each(topology.options, function(items, option) {
      _.each(items, function(item) {
        if (item.default === true) {
          defaultOptions = defaultOptions.set(option, item.value);
        }
      });
    });

    if (defaultOptions.size) {
      topologyOptions = topologyOptions.set(
        topology.id,
        defaultOptions
      );
    }

    setDefaultTopologyOptions(topology.sub_topologies);
  });
}

function popNodeDetails() {
  if (nodeDetails.size > 0) {
    const lastNodeId = nodeDetails.keySeq().last();
    // remove pipe if it belongs to the node being closed
    controlPipes = controlPipes.filter(pipe => {
      return pipe.nodeId !== lastNodeId;
    });
    nodeDetails = nodeDetails.butLast();
  }
  if (nodeDetails.size === 0) {
    selectedNodeId = null;
  }
}

// Store API

export class AppStore extends Store {

  // keep at the top
  getAppState() {
    return {
      controlPipe: this.getControlPipe(),
      nodeDetails: this.getNodeDetailsState(),
      selectedNodeId: selectedNodeId,
      topologyId: currentTopologyId,
      topologyOptions: topologyOptions.toJS() // all options
    };
  }

  getActiveTopologyOptions() {
    // options for current topology
    return topologyOptions.get(currentTopologyId);
  }

  getAdjacentNodes(nodeId) {
    adjacentNodes = adjacentNodes.clear();

    if (nodes.has(nodeId)) {
      adjacentNodes = makeSet(nodes.get(nodeId).get('adjacency'));
      // fill up set with reverse edges
      nodes.forEach(function(node, id) {
        if (node.get('adjacency') && node.get('adjacency').includes(nodeId)) {
          adjacentNodes = adjacentNodes.add(id);
        }
      });
    }

    return adjacentNodes;
  }

  getControlError() {
    return controlError;
  }

  getControlPipe() {
    return controlPipes.last();
  }

  getCurrentTopology() {
    if (!currentTopology) {
      currentTopology = setTopology(currentTopologyId);
    }
    return currentTopology;
  }

  getCurrentTopologyId() {
    return currentTopologyId;
  }

  getCurrentTopologyOptions() {
    return currentTopology && currentTopology.options;
  }

  getCurrentTopologyUrl() {
    return currentTopology && currentTopology.url;
  }

  getErrorUrl() {
    return errorUrl;
  }

  getHighlightedEdgeIds() {
    if (mouseOverNodeId && nodes.has(mouseOverNodeId)) {
      // all neighbour combinations because we dont know which direction exists
      const adjacency = nodes.get(mouseOverNodeId).get('adjacency');
      if (adjacency) {
        return _.flatten(
          adjacency.forEach(function(nodeId) {
            return [
              [nodeId, mouseOverNodeId].join(EDGE_ID_SEPARATOR),
              [mouseOverNodeId, nodeId].join(EDGE_ID_SEPARATOR)
            ];
          })
        );
      }
    }
    if (mouseOverEdgeId) {
      return mouseOverEdgeId;
    }
    return null;
  }

  getHighlightedNodeIds() {
    if (mouseOverNodeId) {
      const adjacency = this.getAdjacentNodes(mouseOverNodeId);
      if (adjacency.size) {
        return _.union(adjacency.toJS(), [mouseOverNodeId]);
      }
    }
    if (mouseOverEdgeId) {
      return mouseOverEdgeId.split(EDGE_ID_SEPARATOR);
    }
    return null;
  }

  getHostname() {
    return hostname;
  }

  getNodeDetails() {
    return nodeDetails;
  }

  getNodeDetailsState() {
    return nodeDetails.toIndexedSeq().map(details => {
      return {id: details.id, topologyId: details.topologyId};
    }).toJS();
  }

  getNodes() {
    return nodes;
  }

  getSelectedNodeId() {
    return selectedNodeId;
  }

  getTopologies() {
    return topologies;
  }

  getTopologyUrlsById() {
    return topologyUrlsById;
  }

  getVersion() {
    return version;
  }

  isControlPending() {
    return controlPending;
  }

  isRouteSet() {
    return routeSet;
  }

  isTopologiesLoaded() {
    return topologiesLoaded;
  }

  isTopologyEmpty() {
    return currentTopology && currentTopology.stats && currentTopology.stats.node_count === 0 && nodes.size === 0;
  }

  isWebsocketClosed() {
    return websocketClosed;
  }

  __onDispatch(payload) {
    if (!payload.type) {
      error('Payload missing a type!', payload);
    }

    switch (payload.type) {
    case ActionTypes.CHANGE_TOPOLOGY_OPTION:
      if (topologyOptions.getIn([payload.topologyId, payload.option])
        !== payload.value) {
        nodes = nodes.clear();
      }
      topologyOptions = topologyOptions.setIn(
        [payload.topologyId, payload.option],
        payload.value
      );
      this.__emitChange();
      break;

    case ActionTypes.CLEAR_CONTROL_ERROR:
      controlError = null;
      this.__emitChange();
      break;

    case ActionTypes.CLICK_CLOSE_DETAILS:
      popNodeDetails();
      this.__emitChange();
      break;

    case ActionTypes.CLICK_CLOSE_TERMINAL:
      controlPipes = controlPipes.clear();
      this.__emitChange();
      break;

    case ActionTypes.CLICK_NODE:
      const prevSelectedNodeId = selectedNodeId;
      popNodeDetails();
      // select new node if it's not the same (in that case just delesect)
      if (prevSelectedNodeId !== payload.nodeId) {
        nodeDetails = nodeDetails.set(
          payload.nodeId,
          {id: payload.nodeId, topologyId: currentTopologyId}
        );
      }
      this.__emitChange();
      break;

    case ActionTypes.CLICK_RELATIVE:
      nodeDetails = nodeDetails.set(
        payload.nodeId,
        {id: payload.nodeId, topologyId: payload.topologyId}
      );
      this.__emitChange();
      break;

    case ActionTypes.CLICK_TOPOLOGY:
      while (nodeDetails.size > 0) {
        popNodeDetails();
      }
      if (payload.topologyId !== currentTopologyId) {
        setTopology(payload.topologyId);
        nodes = nodes.clear();
      }
      this.__emitChange();
      break;

    case ActionTypes.CLOSE_WEBSOCKET:
      websocketClosed = true;
      this.__emitChange();
      break;

    case ActionTypes.DESELECT_NODE:
      popNodeDetails();
      this.__emitChange();
      break;

    case ActionTypes.DO_CONTROL:
      controlPending = true;
      controlError = null;
      this.__emitChange();
      break;

    case ActionTypes.ENTER_EDGE:
      mouseOverEdgeId = payload.edgeId;
      this.__emitChange();
      break;

    case ActionTypes.ENTER_NODE:
      mouseOverNodeId = payload.nodeId;
      this.__emitChange();
      break;

    case ActionTypes.LEAVE_EDGE:
      mouseOverEdgeId = null;
      this.__emitChange();
      break;

    case ActionTypes.LEAVE_NODE:
      mouseOverNodeId = null;
      this.__emitChange();
      break;

    case ActionTypes.OPEN_WEBSOCKET:
      // flush nodes cache after re-connect
      nodes = nodes.clear();
      websocketClosed = false;

      this.__emitChange();
      break;

    case ActionTypes.DO_CONTROL_ERROR:
      controlPending = false;
      controlError = payload.error;
      this.__emitChange();
      break;

    case ActionTypes.DO_CONTROL_SUCCESS:
      controlPending = false;
      controlError = null;
      this.__emitChange();
      break;

    case ActionTypes.RECEIVE_CONTROL_PIPE:
      controlPipes = controlPipes.set(payload.pipeId, {
        id: payload.pipeId,
        nodeId: payload.nodeId,
        raw: payload.rawTty
      });
      this.__emitChange();
      break;

    case ActionTypes.RECEIVE_CONTROL_PIPE_STATUS:
      if (controlPipes.has(payload.pipeId)) {
        controlPipes = controlPipes.setIn([payload.pipeId, 'status'], payload.status);
        this.__emitChange();
      }
      break;

    case ActionTypes.RECEIVE_ERROR:
      errorUrl = payload.errorUrl;
      this.__emitChange();
      break;

    case ActionTypes.RECEIVE_NODE_DETAILS:
      errorUrl = null;
      // disregard if node is not selected anymore
      if (nodeDetails.has(payload.details.id)) {
        nodeDetails = nodeDetails.update(payload.details.id, obj => {
          obj.details = payload.details;
          return obj;
        });
      }
      this.__emitChange();
      break;

    case ActionTypes.RECEIVE_NODES_DELTA:
      const emptyMessage = !payload.delta.add && !payload.delta.remove
        && !payload.delta.update;

      if (!emptyMessage) {
        log('RECEIVE_NODES_DELTA',
          'remove', _.size(payload.delta.remove),
          'update', _.size(payload.delta.update),
          'add', _.size(payload.delta.add));
      }

      errorUrl = null;

      // nodes that no longer exist
      _.each(payload.delta.remove, function(nodeId) {
        // in case node disappears before mouseleave event
        if (mouseOverNodeId === nodeId) {
          mouseOverNodeId = null;
        }
        if (nodes.has(nodeId) && _.contains(mouseOverEdgeId, nodeId)) {
          mouseOverEdgeId = null;
        }
        nodes = nodes.delete(nodeId);
      });

      // update existing nodes
      _.each(payload.delta.update, function(node) {
        if (nodes.has(node.id)) {
          nodes = nodes.set(node.id, nodes.get(node.id).merge(makeNode(node)));
        }
      });

      // add new nodes
      _.each(payload.delta.add, function(node) {
        nodes = nodes.set(node.id, Immutable.fromJS(makeNode(node)));
      });

      this.__emitChange();
      break;

    case ActionTypes.RECEIVE_TOPOLOGIES:
      errorUrl = null;
      topologyUrlsById = topologyUrlsById.clear();
      topologies = processTopologies(payload.topologies);
      setTopology(currentTopologyId);
      // only set on first load, if options are not already set via route
      if (!topologiesLoaded && topologyOptions.size === 0) {
        setDefaultTopologyOptions(topologies);
      }
      topologiesLoaded = true;
      this.__emitChange();
      break;

    case ActionTypes.RECEIVE_API_DETAILS:
      errorUrl = null;
      hostname = payload.hostname;
      version = payload.version;
      this.__emitChange();
      break;

    case ActionTypes.ROUTE_TOPOLOGY:
      routeSet = true;
      if (currentTopologyId !== payload.state.topologyId) {
        nodes = nodes.clear();
      }
      setTopology(payload.state.topologyId);
      setDefaultTopologyOptions(topologies);
      selectedNodeId = payload.state.selectedNodeId;
      if (payload.state.controlPipe) {
        controlPipes = makeOrderedMap(
          [payload.state.controlPipe.pipeId, payload.state.controlPipe]
        );
      } else {
        controlPipes = controlPipes.clear();
      }
      if (payload.state.nodeDetails) {
        nodeDetails = makeOrderedMap(payload.state.nodeDetails.map(obj => [obj.id, obj]));
      } else {
        nodeDetails = nodeDetails.clear();
      }
      topologyOptions = Immutable.fromJS(payload.state.topologyOptions)
        || topologyOptions;
      this.__emitChange();
      break;

    default:
      break;

    }
  }
}

export default new AppStore(AppDispatcher);
