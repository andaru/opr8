/*

Package session has the Session Manager and Session interface types.

Session implementations interface with the Session Manager by
providing an Acceptor implementation, and registering it with the
manager.

An Acceptor supports one or more server transport types, as indicated
by responses to the Supported method. Positive responses to the
Supported method will be followed by a call to the Accept method,
which should be responded with by an attempt to create a session using
the provided transport and session ID. The session is started and the
Server session interface returned to the session manager.

The Session's application must respond to the context passed to
Accept's termination by calling the session's Release method, at which
time any session errors are returned to the Server Wait channel before
the channel is closed.

*/
package session
