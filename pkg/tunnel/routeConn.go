package tunnel

import (
	"bytes"
	"k8s.io/klog/v2"
)

//type RouteConn struct {
//	*strings.Reader
//	*strings.Builder
//}

type RouteConn struct {
	bytes.Buffer
}

// NewRouteConn creates a new MyString instance.
func NewRouteConn() *RouteConn {
	return &RouteConn{}
}

// Close implements the Closer interface.
func (s *RouteConn) Close() error {
	return nil
}

// Peek will show the data stored in the buffer
func (s *RouteConn) Peek() error {
	// covert s.Buffer to string
	data := s.String()
	klog.Infof("[RouteConn]: the data in buffer: ", data)
	return nil
}
