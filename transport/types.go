package transport

import "io"

// Transport is the low-level NETCONF transport interface.
type Transport interface {
	io.ReadWriteCloser

	// CloseWrite closes the writer.
	CloseWrite() error

	// Error returns the transport's optional error channel. When not
	// nil it is avaiable but may only effectively provide Read() or
	// Write() support. i.e., it may return in error for the opposite
	// function call. As a general rule, server transports provide a
	// usable Write() function, while client transports can be Read().
	Error() io.ReadWriter
}

// ClientTransport is the interface for a NETCONF client transport.
//
// The transport's io.Reader is attached to the server's standard
// output, while its io.Writer is attached to the server's standard
// input channel. The Error function returns, if not nil, an io.Reader
// over the server's error output.
type ClientTransport interface {
	Transport
}

// ServerTransport is the interface for a server transport.
//
// The writer and error channels sends data to the client, while the
// reader reads input from the client.
type ServerTransport interface {
	Transport

	// Username returns the client username on this network transport.
	Username() string
}

// RFC6242Framer is the interface to RFC6242 framing management.
//
// This optional interface, found on client or server Transports using
// RFC6242 framing, allows for the framing mode to be switched from
// :base:1.0 (RFC4742) "end-of-message" mode to :base:1.1 (RFC6242)
// "chunked framing" mode. This framing mode must be used on NETCONF
// sessions where the :base:1.1 capability is negotiated early in the
// session. As the session is underway when this is discovered,
// chunked framing mode must be enabled after capability negotiation,
// prior to sending or receiving any more data on thet transport.
type RFC6242Framer interface {
	// EnableChunkedFraming enables NETCONF 1.1 chunked framing
	// encoding on the transport until it is closed. This function
	// must be called after capability negotiation prior to further
	// Read or Write calls.
	EnableChunkedFraming() error
}

// CloseWriter is the interface to bi-directional read/writers which
// can have their send side closed. This is used to indicate EOF
// cleanly to the reader on the far-end.
type CloseWriter interface {
	// CloseWrite closes the writing side of an IO stream.
	CloseWrite() error
}

// ClientUsernameProvider is a transport type that can report the
// client's username. All ServerTransport must offer this interface,
// while it is optional for ClientTransport.
type ClientUsernameProvider interface {
	// Username returns the client username on this network transport.
	Username() string
}
