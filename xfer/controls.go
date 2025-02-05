package xfer

import (
	"fmt"
	"net/rpc"
	"sync"

	"github.com/gorilla/websocket"
)

// Request is the UI -> App -> Probe message type for control RPCs
type Request struct {
	AppID   string
	NodeID  string
	Control string
}

// Response is the Probe -> App -> UI message type for the control RPCs.
type Response struct {
	Value  interface{} `json:"value,omitempty"`
	Error  string      `json:"error,omitempty"`
	Pipe   string      `json:"pipe,omitempty"`
	RawTTY bool        `json:"raw_tty,omitempty"`
}

// Message is the unions of Request, Response and PipeIO
type Message struct {
	Request  *rpc.Request
	Response *rpc.Response
	Value    interface{}
}

// ControlHandler is interface used in the app and the probe to represent
// a control RPC.
type ControlHandler interface {
	Handle(req Request, res *Response) error
}

// ControlHandlerFunc is a adapter (ala golang's http RequestHandlerFunc)
// for ControlHandler
type ControlHandlerFunc func(Request) Response

// Handle is an adapter method to make ControlHandlers exposable via golang rpc
func (c ControlHandlerFunc) Handle(req Request, res *Response) error {
	*res = c(req)
	return nil
}

// ResponseErrorf creates a new Response with the given formatted error string.
func ResponseErrorf(format string, a ...interface{}) Response {
	return Response{
		Error: fmt.Sprintf(format, a...),
	}
}

// ResponseError creates a new Response with the given error.
func ResponseError(err error) Response {
	if err != nil {
		return Response{
			Error: err.Error(),
		}
	}
	return Response{}
}

// JSONWebsocketCodec is golang rpc compatible Server and Client Codec
// that transmits and receives RPC messages over a websocker, as JSON.
type JSONWebsocketCodec struct {
	sync.Mutex
	conn *websocket.Conn
	err  chan struct{}
}

// NewJSONWebsocketCodec makes a new JSONWebsocketCodec
func NewJSONWebsocketCodec(conn *websocket.Conn) *JSONWebsocketCodec {
	return &JSONWebsocketCodec{
		conn: conn,
		err:  make(chan struct{}),
	}
}

// WaitForReadError blocks until any read on this codec returns an error.
// This is useful to know when the server has disconnected from the client.
func (j *JSONWebsocketCodec) WaitForReadError() {
	<-j.err
}

// WriteRequest implements rpc.ClientCodec
func (j *JSONWebsocketCodec) WriteRequest(r *rpc.Request, v interface{}) error {
	j.Lock()
	defer j.Unlock()

	if err := j.conn.WriteJSON(Message{Request: r}); err != nil {
		return err
	}
	return j.conn.WriteJSON(Message{Value: v})
}

// WriteResponse implements rpc.ServerCodec
func (j *JSONWebsocketCodec) WriteResponse(r *rpc.Response, v interface{}) error {
	j.Lock()
	defer j.Unlock()

	if err := j.conn.WriteJSON(Message{Response: r}); err != nil {
		return err
	}
	return j.conn.WriteJSON(Message{Value: v})
}

func (j *JSONWebsocketCodec) readMessage(v interface{}) (*Message, error) {
	m := Message{Value: v}
	if err := j.conn.ReadJSON(&m); err != nil {
		close(j.err)
		return nil, err
	}
	return &m, nil
}

// ReadResponseHeader implements rpc.ClientCodec
func (j *JSONWebsocketCodec) ReadResponseHeader(r *rpc.Response) error {
	m, err := j.readMessage(nil)
	if err == nil {
		*r = *m.Response
	}
	return err
}

// ReadResponseBody implements rpc.ClientCodec
func (j *JSONWebsocketCodec) ReadResponseBody(v interface{}) error {
	_, err := j.readMessage(v)
	return err
}

// Close implements rpc.ClientCodec and rpc.ServerCodec
func (j *JSONWebsocketCodec) Close() error {
	return j.conn.Close()
}

// ReadRequestHeader implements rpc.ServerCodec
func (j *JSONWebsocketCodec) ReadRequestHeader(r *rpc.Request) error {
	m, err := j.readMessage(nil)
	if err == nil {
		*r = *m.Request
	}
	return err
}

// ReadRequestBody implements rpc.ServerCodec
func (j *JSONWebsocketCodec) ReadRequestBody(v interface{}) error {
	_, err := j.readMessage(v)
	return err
}
