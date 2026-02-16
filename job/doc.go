// Package job defines the stateful representation of a message within the
// gqs queue lifecycle.
//
// A Job extends message.Message with delivery and scheduling metadata.
// It represents a message as stored and managed by a queue implementation.
//
// Unlike message.Message, Job contains state-machine fields such as Status,
// Attempts, lock information, and scheduling timestamps. These fields are
// maintained by the queue storage and worker logic.
//
// Job values are typically returned by Pull operations and passed back to
// the storage layer for state transitions (Complete, Return, Kill, etc.).
//
// Job is not intended to be constructed manually by user code.
// Its fields reflect the authoritative state stored by the queue backend.
package job
