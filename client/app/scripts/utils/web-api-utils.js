import debug from 'debug';
import reqwest from 'reqwest';

import { clearControlError, closeWebsocket, openWebsocket, receiveError,
  receiveApiDetails, receiveNodesDelta, receiveNodeDetails, receiveControlError,
  receiveControlPipe, receiveControlPipeStatus, receiveControlSuccess,
  receiveTopologies } from '../actions/app-actions';

const wsProto = location.protocol === 'https:' ? 'wss' : 'ws';
const wsUrl = wsProto + '://' + location.host + location.pathname.replace(/\/$/, '');
const log = debug('scope:web-api-utils');

const apiTimerInterval = 10000;
const reconnectTimerInterval = 5000;
const topologyTimerInterval = apiTimerInterval;
const updateFrequency = '5s';

let socket;
let reconnectTimer = 0;
let currentUrl = null;
let currentOptions = null;
let topologyTimer = 0;
let apiDetailsTimer = 0;
let controlErrorTimer = 0;

function buildOptionsQuery(options) {
  if (options) {
    return options.reduce(function(query, value, param) {
      return `${query}&${param}=${value}`;
    }, '');
  }
  return '';
}

export function basePath(urlPath) {
  //
  // "/scope/terminal.html" -> "/scope"
  // "/scope/" -> "/scope"
  // "/scope" -> "/scope"
  // "/" -> ""
  //
  const parts = urlPath.split('/');
  // if the last item has a "." in it, e.g. foo.html...
  if (parts[parts.length - 1].indexOf('.') !== -1) {
    return parts.slice(0, -1).join('/');
  }
  return parts.join('/').replace(/\/$/, '');
}

function createWebsocket(topologyUrl, optionsQuery) {
  if (socket) {
    socket.onclose = null;
    socket.onerror = null;
    socket.close();
  }

  socket = new WebSocket(wsUrl + topologyUrl
    + '/ws?t=' + updateFrequency + '&' + optionsQuery);

  socket.onopen = function() {
    openWebsocket();
  };

  socket.onclose = function() {
    clearTimeout(reconnectTimer);
    socket = null;
    closeWebsocket();
    log('Closed websocket to ' + topologyUrl);

    reconnectTimer = setTimeout(function() {
      createWebsocket(topologyUrl, optionsQuery);
    }, reconnectTimerInterval);
  };

  socket.onerror = function() {
    log('Error in websocket to ' + topologyUrl);
    receiveError(currentUrl);
  };

  socket.onmessage = function(event) {
    const msg = JSON.parse(event.data);
    receiveNodesDelta(msg);
  };
}

/* keep URLs relative */

export function getTopologies(options) {
  clearTimeout(topologyTimer);
  const optionsQuery = buildOptionsQuery(options);
  const url = `api/topology?${optionsQuery}`;
  reqwest({
    url: url,
    success: function(res) {
      receiveTopologies(res);
      topologyTimer = setTimeout(function() {
        getTopologies(options);
      }, topologyTimerInterval / 2);
    },
    error: function(err) {
      log('Error in topology request: ' + err);
      receiveError(url);
      topologyTimer = setTimeout(function() {
        getTopologies(options);
      }, topologyTimerInterval / 2);
    }
  });
}

export function getNodesDelta(topologyUrl, options) {
  const optionsQuery = buildOptionsQuery(options);

  // only recreate websocket if url changed
  if (topologyUrl && (topologyUrl !== currentUrl || currentOptions !== optionsQuery)) {
    createWebsocket(topologyUrl, optionsQuery);
    currentUrl = topologyUrl;
    currentOptions = optionsQuery;
  }
}

export function getNodeDetails(topologyUrl, nodeMap) {
  if (topologyUrl) {
    nodeMap.forEach((_, nodeId) => {
      const url = [topologyUrl, '/', encodeURIComponent(nodeId)]
        .join('').substr(1);
      reqwest({
        url: url,
        success: function(res) {
          receiveNodeDetails(res.node);
        },
        error: function(err) {
          log('Error in node details request: ' + err.responseText);
          // dont treat missing node as error
          if (err.status !== 404) {
            receiveError(topologyUrl);
          }
        }
      });
    });
  }
}

export function getApiDetails() {
  clearTimeout(apiDetailsTimer);
  const url = 'api';
  reqwest({
    url: url,
    success: function(res) {
      receiveApiDetails(res);
      apiDetailsTimer = setTimeout(getApiDetails, apiTimerInterval);
    },
    error: function(err) {
      log('Error in api details request: ' + err);
      receiveError(url);
      apiDetailsTimer = setTimeout(getApiDetails, apiTimerInterval / 2);
    }
  });
}

export function doControl(probeId, nodeId, control) {
  clearTimeout(controlErrorTimer);
  const url = `api/control/${encodeURIComponent(probeId)}/`
    + `${encodeURIComponent(nodeId)}/${control}`;
  reqwest({
    method: 'POST',
    url: url,
    success: function(res) {
      receiveControlSuccess();
      if (res && res.pipe) {
        receiveControlPipe(res.pipe, nodeId, res.raw_tty, true);
      }
    },
    error: function(err) {
      receiveControlError(err.response);
      controlErrorTimer = setTimeout(function() {
        clearControlError();
      }, 10000);
    }
  });
}

export function deletePipe(pipeId) {
  const url = `api/pipe/${encodeURIComponent(pipeId)}`;
  reqwest({
    method: 'DELETE',
    url: url,
    success: function() {
      log('Closed the pipe!');
    },
    error: function(err) {
      log('Error closing pipe:' + err);
      receiveError(url);
    }
  });
}

export function getPipeStatus(pipeId) {
  const url = `/api/pipe/${encodeURIComponent(pipeId)}`;
  reqwest({
    method: 'GET',
    url: url,
    success: function(res) {
      log('ERROR: expected responses: [400, 404]. Got:', res);
    },
    error: function(err) {
      const status = {
        400: 'PIPE_ALIVE',
        404: 'PIPE_DELETED'
      }[err.status];

      if (!status) {
        log('Unexpected pipe status:', err.status);
        return;
      }

      receiveControlPipeStatus(pipeId, status);
    }
  });
}
