package logf

// Encoder defines the interface to create your own log format.
type Encoder interface {
	Encode(*Buffer, Entry) error
}
