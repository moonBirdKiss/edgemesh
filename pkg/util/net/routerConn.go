package util

import (
	"bytes"
	"k8s.io/klog/v2"
)

type RouteConnection struct {
	bytes.Buffer
}

// NewRouteConn creates a new MyString instance.
func NewRouteConnection() *RouteConnection {
	return &RouteConnection{}
}

// Close implements the Closer interface.
func (s *RouteConnection) Close() error {
	return nil
}

// Peek will show the data stored in the buffer
func (s *RouteConnection) Peek() error {
	// covert s.Buffer to string
	data := s.String()
	klog.Infof("[RouteConn]: the data in buffer: ", data)
	return nil
}

func (s *RouteConnection) GetBuffer() bytes.Buffer {
	return s.Buffer
}
