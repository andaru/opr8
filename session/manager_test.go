package session

import (
	"context"
	"io"
	"testing"

	"math"

	"github.com/andaru/opr8/transport"
	"github.com/pkg/errors"
)

type testSessionBase struct {
	id      ID
	t       Type
	chError chan error
}

type testSessionFoo struct{ testSessionBase }
type testSessionBar struct{ testSessionBase }

func newTestSessionFoo(id ID) testSessionFoo {
	return testSessionFoo{testSessionBase{id: id, t: TypeServer, chError: make(chan error, 1)}}
}

func newTestSessionBar(id ID) testSessionBar {
	return testSessionBar{testSessionBase{id: id, t: TypeServer, chError: make(chan error, 1)}}
}

func (s testSessionBase) ID() ID                         { return s.id }
func (s testSessionBase) Type() Type                     { return s.t }
func (s testSessionBase) Transport() transport.Transport { return nil }
func (s testSessionBase) Wait() <-chan error             { return s.chError }
func (s testSessionBase) Release() {
	select {
	case <-s.chError:
	default:
		return
	}
	close(s.chError)
}

type testTransport struct{}
type testTransportFoo struct{ testTransport }
type testTransportBar struct{ testTransport }

func (t testTransport) Read([]byte) (int, error)  { return 0, nil }
func (t testTransport) Write([]byte) (int, error) { return 0, nil }
func (t testTransport) CloseWrite() error         { return nil }
func (t testTransport) Close() error              { return nil }
func (t testTransport) Error() io.ReadWriter      { return nil }
func (t testTransport) Username() string          { return "" }

type testAcceptor struct{}

type testAcceptorFoo struct{ testAcceptor }
type testAcceptorBar struct{ testAcceptor }

func (a testAcceptorFoo) Accept(_ context.Context, using transport.ServerTransport, id ID) (Server, error) {
	return newTestSessionFoo(id), nil
}

func (a testAcceptorFoo) Supported(tt transport.ServerTransport) bool {
	switch tt.(type) {
	case *testTransportFoo:
		return true
	default:
		return false
	}
}

func (a testAcceptorBar) Accept(_ context.Context, using transport.ServerTransport, id ID) (Server, error) {
	return newTestSessionBar(id), nil
}

func (a testAcceptorBar) Supported(tt transport.ServerTransport) bool {
	switch tt.(type) {
	case *testTransportBar:
		return true
	default:
		return false
	}
}

type testIDGen struct{ id ID }

func (gen *testIDGen) NextID() ID {
	gen.id++
	if gen.id == 0 {
		gen.id++
	}
	return gen.id
}

func TestManager(t *testing.T) {
	for _, tt := range []struct {
		name             string
		acceptors        []Acceptor
		transport        transport.ServerTransport
		sessionTypeCheck func(Session) error
		wantErr          bool
	}{
		{
			acceptors: []Acceptor{&testAcceptorFoo{}},
			transport: &testTransportFoo{},
			sessionTypeCheck: func(s Session) error {
				if _, ok := s.(testSessionFoo); !ok {
					return errors.Errorf("want testSessionFoo, got %T", s)
				}
				return nil
			},
		},

		{
			acceptors: []Acceptor{&testAcceptorBar{}},
			transport: &testTransportBar{},
			sessionTypeCheck: func(s Session) error {
				if _, ok := s.(testSessionBar); !ok {
					return errors.Errorf("want testSessionBar, got %T", s)
				}
				return nil
			},
		},

		{
			acceptors: []Acceptor{&testAcceptorFoo{}, &testAcceptorBar{}},
			transport: &testTransportFoo{},
			sessionTypeCheck: func(s Session) error {
				if _, ok := s.(testSessionFoo); !ok {
					return errors.Errorf("want testSessionFoo, got %T", s)
				}
				return nil
			},
		},

		{
			acceptors: []Acceptor{&testAcceptorFoo{}, &testAcceptorBar{}},
			transport: &testTransportBar{},
			sessionTypeCheck: func(s Session) error {
				if _, ok := s.(testSessionBar); !ok {
					return errors.Errorf("want testSessionBar, got %T", s)
				}
				return nil
			},
		},

		{
			acceptors: []Acceptor{&testAcceptorBar{}, &testAcceptorFoo{}},
			transport: &testTransportBar{},
			sessionTypeCheck: func(s Session) error {
				if _, ok := s.(testSessionBar); !ok {
					return errors.Errorf("want testSessionBar, got %T", s)
				}
				return nil
			},
		},

		{
			acceptors:        []Acceptor{},
			transport:        &testTransportFoo{},
			sessionTypeCheck: func(s Session) error { return nil },
			wantErr:          true,
		},
		{
			acceptors:        []Acceptor{},
			transport:        &testTransportBar{},
			sessionTypeCheck: func(s Session) error { return nil },
			wantErr:          true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(WithAcceptor(tt.acceptors...), WithIDSource(&genIncrement{id: math.MaxUint32}))
			session, err := m.Accept(context.Background(), tt.transport)
			if session != nil {
				m.Terminate(session.ID(), nil)
				session.Release()
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.Accept() unexpected error = %v", err)
			}
			err = tt.sessionTypeCheck(session)
			if err != nil {
				t.Errorf("sessionTypeCheck error = %v", err)
			}
		})
	}
}
