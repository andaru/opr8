package session

import (
	"context"
	"sync"

	"github.com/andaru/opr8/transport"
	"github.com/pkg/errors"
)

// ManagerOption is a constructor option for Manager.
type ManagerOption func(*manager)

// NewManager returns a new session manager configured with supplied
// options.
func NewManager(options ...ManagerOption) Manager {
	mgr := &manager{sessions: map[ID]Server{}, idgen: &genIncrement{}}
	for _, option := range options {
		option(mgr)
	}
	return mgr
}

// WithAcceptor is a Manager option which uses the provided Acceptor when
// accepting a connection.
func WithAcceptor(acc ...Acceptor) ManagerOption {
	return func(m *manager) { m.acc = append(m.acc, acc...) }
}

// WithIDSource is a Manager option which sets the manager's session
// ID source to the provided IDGenerator.
func WithIDSource(gen IDGenerator) ManagerOption {
	return func(m *manager) { m.idgen = gen }
}

// genIncrement is an incrementing valid session ID generator, values
// starting at its "id" field plus one. It is the default session ID
// source for Manager created with NewManager.
type genIncrement struct{ id ID }

func (gen *genIncrement) NextID() ID {
	gen.id++
	if gen.id == 0 {
		gen.id++
	}
	return gen.id
}

// manager is the session manager. It implements Manager.
type manager struct {
	sync.Mutex
	acc      []Acceptor
	sessions map[ID]Server
	idgen    IDGenerator
}

const (
	maxIDtries = 16
)

// Accept accepts a session using the specified transport. If no
// session can be accepted by the system, an error is
// returned. Otherwise, the session is started and returned to the
// caller. It is safe to call from concurrent goroutines.
func (mgr *manager) Accept(ctx context.Context, using transport.ServerTransport) (session Session, err error) {
	// find an acceptor for this transport type
	var acceptor Acceptor
	for _, acc := range mgr.acc {
		if acc.Supported(using) {
			acceptor = acc
			break
		}
	}
	if acceptor == nil {
		return nil, errors.Errorf("failed to create session using transport %T", using)
	}

	var serverSession Server
	var id ID

	// critical section
	{
		mgr.Lock()
		defer mgr.Unlock()

		// get a unique, valid session ID
		for i := 0; i < maxIDtries; i++ {
			if id = mgr.idgen.NextID(); id != 0 && mgr.sessions[id] == nil {
				break
			}
		}
		if id == 0 {
			return nil, errors.Errorf("failed to get unique session ID after %d tries", maxIDtries)
		}

		// accept the real session from the backend
		serverSession, err = acceptor.Accept(ctx, using, id)
		if err != nil {
			return nil, errors.Wrap(err, "failed to accept transport")
		}

		// ensure the session ID we asked for is the one we got back
		if serverSession.ID() != id {
			return nil, errors.Errorf("unexpected session ID %v, wanted %v", serverSession.ID(), id)
		}

		// track the session
		mgr.sessions[id] = serverSession
	}
	// end critical section

	go func() {
		// delete the session when it ends
		<-serverSession.Wait()
		mgr.Lock()
		delete(mgr.sessions, id)
		mgr.Unlock()
	}()

	// return the session, ready for application use
	return serverSession, nil
}

func (mgr *manager) Terminate(id ID, with error) error {
	mgr.Lock()
	defer mgr.Unlock()
	if s, ok := mgr.sessions[id]; ok {
		// release session resources and cease tracking it
		s.Release()
		return nil
	}
	return errors.Errorf("session %v does not exist", id)
}

var _ Manager = &manager{}
