package tunnel

import "strings"

type RouteConn struct {
	*strings.Reader
	*strings.Builder
}

// NewRouteConn creates a new MyString instance.
func NewRouteConn(str string) *RouteConn {
	return &RouteConn{
		Reader:  strings.NewReader(str),
		Builder: &strings.Builder{},
	}
}

// Close implements the Closer interface.
func (s *RouteConn) Close() error {
	return nil
}

// Write implements the Writer interface.
func (s *RouteConn) Write(p []byte) (int, error) {
	return s.Builder.Write(p)
}

// Read implements the Reader interface.
func (s *RouteConn) Read(p []byte) (int, error) {
	return s.Reader.Read(p)
}
