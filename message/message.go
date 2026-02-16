package message

import (
	"github.com/google/uuid"
)

// Message represents a transport-level unit of data in gqs.
//
// It contains only the user-facing fields: a unique identifier, optional
// metadata, and an opaque payload. Message does not track delivery state
// or retry information.
//
// Id is generated automatically by NewMessage, but may also be assigned
// explicitly before pushing the message into a queue.
//
// Metadata is optional and lazily initialized. It may be nil if no metadata
// has been set.
//
// Payload contains arbitrary binary data and may be nil.
type Message struct {
	Id       uuid.UUID
	Metadata map[string]any
	Payload  []byte
}

// NewMessage creates a new Message with a randomly generated UUID.
//
// The returned Message has no metadata and no payload.
// Metadata will be allocated lazily when Set is called.
func NewMessage() *Message {
	return &Message{
		Id: uuid.New(),
	}
}

// Get returns the metadata value associated with the given key.
//
// If the key does not exist or Metadata is nil, Get returns nil.
//
// The returned value has static type any. For type-safe access,
// use the generic Get function.
func (m *Message) Get(key string) any {
	ret, ok := m.Metadata[key]
	if !ok {
		return nil
	}
	return ret
}

// Set stores the given key-value pair in the message metadata.
//
// If Metadata is nil, it is initialized automatically.
func (m *Message) Set(key string, value any) {
	if m.Metadata == nil {
		m.Metadata = make(map[string]any)
	}
	m.Metadata[key] = value
}

// Get retrieves a metadata value associated with the given key and
// attempts to cast it to type T.
//
// If the key does not exist or the stored value is not of type T,
// Get returns the zero value of T and false.
func Get[T any](m *Message, key string) (T, bool) {
	raw, ok := m.Metadata[key]
	if !ok {
		var t T
		return t, false
	}
	ret, ok := raw.(T)
	if !ok {
		var t T
		return t, false
	}
	return ret, true
}

// Set stores the given key-value pair in the message metadata
// using a type-safe generic helper.
//
// If Metadata is nil, it is initialized automatically.
func Set[T any](m *Message, key string, value T) {
	if m.Metadata == nil {
		m.Metadata = make(map[string]any)
	}
	m.Metadata[key] = value
}
