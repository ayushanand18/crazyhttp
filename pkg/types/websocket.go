package types

type WebSocketHandshakeFunc func()

// WebsocketStreamChunk represents a single chunk of data in a streaming HTTP response,
// such as Server-Sent Events (SSE) or other streaming protocols.
//
// Fields
//
//	Id:   A unique identifier for the chunk, typically used to track ordering
//	      or support reconnection/resume logic.
//	MessageType: The type of message being sent (e.g., text, binary, close).
//	Data: The raw byte payload of the chunk to be sent to the client.
type WebsocketStreamChunk struct {
	Id          uint32
	MessageType WebsocketMessageType
	Data        []byte
}

// WebsocketMessageType represents the type of a WebSocket frame as defined by
// the WebSocket protocol specification (RFC 6455). It is used to indicate how
// the payload of a WebSocket message should be interpreted.
//
// # Constants
//
// WebsocketUnspecifiedMessage: An unspecified message type, typically unused
//
//	WebsocketTextMessage:   A UTF-8 encoded text message.
//	WebsocketBinaryMessage: A binary data message.
//	WebsocketCloseMessage:  A control message to close the WebSocket connection.
//	WebsocketPingMessage:   A control message to check if the peer is alive.
//	WebsocketPongMessage:   A control message sent in response to a ping.
type WebsocketMessageType int

const (
	WebsocketUnspecifiedMessage WebsocketMessageType = 0
	WebsocketTextMessage        WebsocketMessageType = 1
	WebsocketBinaryMessage      WebsocketMessageType = 2
	WebsocketCloseMessage       WebsocketMessageType = 8
	WebsocketPingMessage        WebsocketMessageType = 9
	WebsocketPongMessage        WebsocketMessageType = 10
)

func (w WebsocketMessageType) ToInt() int {
	return int(w)
}
