package types

// StreamChunk represents a single chunk of data in a streaming HTTP response,
// such as Server-Sent Events (SSE) or other streaming protocols.
//
// Fields
//
//	Id:   A unique identifier for the chunk, typically used to track ordering
//	      or support reconnection/resume logic.
//	Data: The raw byte payload of the chunk to be sent to the client.
type StreamChunk struct {
	Id   uint32
	Data []byte
}

