// Package message defines the transport-level message abstraction used by gqs.
//
// Message represents a user payload along with optional metadata.
// It is intentionally minimal and does not contain any delivery or state
// information (such as status, attempts, locks, etc.). Those concerns are
// handled by higher-level types (for example, job.Job) and storage
// implementations.
//
// A Message is designed to be:
//   - storage-agnostic
//   - lightweight
//   - safe to pass to user handlers
//
// The Payload field contains the opaque binary body of the message.
// The Metadata field is an optional key-value map for arbitrary structured
// data associated with the message.
//
// Message does not enforce immutability. Callers should treat Message
// instances as immutable once they are submitted to a queue to avoid
// unintended data races or side effects.
package message
