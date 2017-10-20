package session

import (
	"context"
	"fmt"

	"github.com/andaru/opr8/transport"
)

// ID is an identifier describing a unique session in the system. The
// default value (0) does not describe a valid session.
type ID uint32

// IDGenerator is the interface to a session ID geneartor, which
// generates a non-repeating endless sequence of valid (non-zero)
// session identifiers. A random sequence of identifiers is legal as
// is an increasing or decreasing sequence.
type IDGenerator interface {
	// NextID returns the next session ID in the sequence. Session
	// managers call this function as often as they need to offer new
	// sessions.
	NextID() ID
}

// Type is a session type. It indicates whether a session is a client
// or server session. Client sessions are used by clients to interact
// with servers, while server sessions are used by servers to converse
// with clients.
type Type int

const (
	// TypeClient indicates a client session
	TypeClient Type = 1 + iota
	// TypeServer indicates a server session
	TypeServer
)

func (t Type) String() string {
	switch t {
	case TypeClient:
		return "client"
	case TypeServer:
		return "server"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", t)
	}
}

// Identifier is an interface to receiving an object's session ID.
type Identifier interface {
	// ID returns the session identifier.
	ID() ID
}

// Session is the opr8 session interface. It represents a client or
// server's view of the secure, connection oriented session used by
// clients and servers to exchange messages.
type Session interface {
	// ID returns the session identifier.
	ID() ID

	// Type returns the session type.
	Type() Type

	// Transport returns the session transport.
	Transport() transport.Transport

	// Release releases the session for closing and releases any
	// resources used by the session. The server application using
	// this session must call this function when it has finished with
	// the session. Once this is called, no sends on the session
	// transport may occur.
	//
	// For example with NETCONF, this is called when the transport
	// session ends, the session is killed or the user gracefully
	// closes the connection. RESTCONF calls this at the end of an
	// HTTP request while GNMI would call this correspondingly at the
	// end of each RPC exchange.
	Release()
}

// Server is the server session interface, extending Session for the
// interfaces necessary to interface server sessions with the session
// manager.
type Server interface {
	Session

	// Wait returns an error channel, closed when the session's
	// Release method is called.
	Wait() <-chan error
}

// Manager is the server session manager interface.
//
// Sessions received from this interface are assigned by transports to
// individual client sessions.
type Manager interface {
	// Accept is called by network servers to start a new server
	// session for the provided transport using the context. The
	// session manager tracks the session and manages session
	// resources, releasing references automatically upon completion
	// of the accepted session. Accept returns an error if the system
	// is presently unable to service a new session.
	Accept(context.Context, transport.ServerTransport) (Session, error)

	// Terminate ends a session by ID as soon as possible with the
	// provided error. Terminate returns in error if the session did
	// not exist.
	Terminate(ID, error) error
}

// Acceptor is the interface used by session preparation code and is
// registered to a session manager.
type Acceptor interface {
	// Supported returns true if this acceptor can create sessions for
	// the server transport type passed. It is the first call made by
	// the Manager in response to an Accept call when attempting to
	// find a suitable acceptor.
	Supported(transport.ServerTransport) bool

	// Accept starts and returns a server session for the supported
	// server transport and session ID, using the provided context. If
	// an error occurred while starting the transport, an error is
	// returned instead.
	Accept(context.Context, transport.ServerTransport, ID) (Server, error)
}

// Typer is the session type reporting interface.
type Typer interface {
	// Type returns the session type.
	Type() Type
}
