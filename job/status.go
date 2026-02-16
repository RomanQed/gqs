package job

import "fmt"

// Status represents the current lifecycle state of a Job.
//
// The state machine is:
//
//	Pending    -> Processing
//	Processing -> Done
//	Processing -> Pending   (via Return)
//	Processing -> Dead
//
// Unknown is reserved as a zero value and may be used to indicate
// an unspecified or invalid state in filtering contexts.
type Status uint8

const (
	// Unknown represents an unspecified or invalid job state.
	// It is the zero value of Status.
	Unknown Status = iota

	// Pending indicates that the job is available for pulling.
	// A Pending job may have a future NextRunAt, delaying execution.
	Pending

	// Processing indicates that the job has been pulled and is currently
	// owned by a worker. While in this state, LockedUntil defines the
	// visibility timeout.
	Processing

	// Done indicates successful completion. The job will not be executed again
	// unless explicitly re-queued by storage logic.
	Done

	// Dead indicates that the job has permanently failed and will not
	// be retried.
	Dead
)

func statusToString(status Status) string {
	switch status {
	case Pending:
		return "Pending"
	case Processing:
		return "Processing"
	case Done:
		return "Done"
	case Dead:
		return "Dead"
	default:
		return "Unknown"
	}
}

func statusFromString(status string) (Status, error) {
	switch status {
	case "Pending":
		return Pending, nil
	case "Processing":
		return Processing, nil
	case "Done":
		return Done, nil
	case "Dead":
		return Dead, nil
	case "Unknown":
		return Unknown, nil
	default:
		return 0, fmt.Errorf("unknown status: %s", status)
	}
}

// ParseStatus converts a string representation of a status into a Status value.
//
// Recognized values are:
//
//	"Pending"
//	"Processing"
//	"Done"
//	"Dead"
//	"Unknown"
//
// An error is returned for unrecognized strings.
func ParseStatus(s string) (Status, error) {
	return statusFromString(s)
}

// MarshalText implements encoding.TextMarshaler.
//
// Status values are encoded using their canonical string names.
func (s Status) MarshalText() ([]byte, error) {
	return []byte(statusToString(s)), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
//
// The textual form must match one of the canonical status names.
func (s *Status) UnmarshalText(text []byte) error {
	status, err := statusFromString(string(text))
	if err != nil {
		return err
	}
	*s = status
	return nil
}

// String returns the canonical string representation of the status.
func (s Status) String() string {
	return statusToString(s)
}
