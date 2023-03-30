package util

import (
	"bytes"
	"fmt"
	"io"
	"k8s.io/klog/v2"
	"net"
	"net/http"
	"sync"
	"time"
)

// HttpRequestToBytes transforms http.Request to bytes
func HttpRequestToBytes(req *http.Request) ([]byte, error) {
	if req == nil {
		return nil, fmt.Errorf("http request nil")
	}
	buf := new(bytes.Buffer)
	err := req.Write(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func RouteConn(in, out io.ReadWriteCloser) {
	var wg sync.WaitGroup
	wg.Add(2)
	go RouteCopyBytes("from out to in", in, out, &wg)
	go RouteCopyBytes("from in to out", out, in, &wg)
	wg.Wait()
}

func RouteCopyBytes(direction string, dest, src io.ReadWriteCloser, wg *sync.WaitGroup) {
	defer wg.Done()
	klog.Info("[route]: Copying remote address bytes")
	n, err := io.Copy(dest, src)
	if err != nil {
		if !IsClosedError(err) && !IsStreamResetError(err) {
			klog.ErrorS(err, "I/O error occurred")
		}
	}
	klog.Info("Copied remote address bytes.", "bytes: ", n, " direction: ", direction)
	if err = dest.Close(); err != nil && !IsClosedError(err) {
		klog.ErrorS(err, "dest close failed")
	}
	if err = src.Close(); err != nil && !IsClosedError(err) {
		klog.ErrorS(err, "src close failed")
	}
}

func RouteCopyStream(dest, src io.ReadWriteCloser) {
	klog.Info("[RouteCopyStream]: Copying remote address bytes")
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		n, err := io.Copy(dest, src)
		if err != nil {
			if !IsClosedError(err) && !IsStreamResetError(err) {
				klog.ErrorS(err, "I/O error occurred")
			}
		}
		klog.Info("[RouteCopyStream]: Copied remote address bytes.", "bytes: ", n)
		wg.Done()
	}()

	// before the src is closed, we should send the correct msg back to the client
	// 返回一个报文，表示成功收到了
	msgStr := `HTTP/1.0 200 OK
Content-Type: text/plain
Content-Length: 16

Hello, my friend
`
	_, err := src.Write([]byte(msgStr))
	if err != nil {
		klog.Errorf("[RouteStream]: Write data error: %v", err)
		return
	}

	wg.Wait()
	if err = src.Close(); err != nil && !IsClosedError(err) {
		klog.ErrorS(err, "src close failed")
	}
}

// ProxyConn proxies data bi-directionally between in and out.
func ProxyConn(in, out net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)
	klog.Info("Creating proxy between remote and local addresses", "inRemoteAddress", in.RemoteAddr(),
		"inLocalAddress", in.LocalAddr(), "outLocalAddress", out.LocalAddr(), "outRemoteAddress", out.RemoteAddr())
	go copyBytes("from backend", in, out, &wg)
	go copyBytes("to backend", out, in, &wg)
	wg.Wait()
}

func copyBytes(direction string, dest, src net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	klog.V(4).InfoS("Copying remote address bytes", "direction", direction, "sourceRemoteAddress", src.RemoteAddr(), "destinationRemoteAddress", dest.RemoteAddr())
	n, err := io.Copy(dest, src)
	if err != nil {
		if !IsClosedError(err) && !IsStreamResetError(err) {
			klog.ErrorS(err, "I/O error occurred")
		}
	}
	klog.V(4).InfoS("Copied remote address bytes", "bytes", n, "direction", direction, "sourceRemoteAddress", src.RemoteAddr(), "destinationRemoteAddress", dest.RemoteAddr())
	if err = dest.Close(); err != nil && !IsClosedError(err) {
		klog.ErrorS(err, "dest close failed")
	}
	if err = src.Close(); err != nil && !IsClosedError(err) {
		klog.ErrorS(err, "src close failed")
	}
}

func ProxyConnUDP(inConn net.Conn, udpConn *net.UDPConn) {
	var buffer [4096]byte
	for {
		n, err := inConn.Read(buffer[0:])
		if err != nil {
			if e, ok := err.(net.Error); ok {
				if e.Temporary() {
					klog.V(1).ErrorS(err, "ReadFrom had a temporary failure")
					continue
				}
			}
			if !IsClosedError(err) && !IsStreamResetError(err) {
				klog.ErrorS(err, "ReadFrom failed, exiting")
			}
			break
		}
		go copyDatagram(udpConn, inConn)
		_, err = udpConn.Write(buffer[0:n])
		if err != nil {
			if !IsTimeoutError(err) {
				klog.ErrorS(err, "Write failed")
			}
			continue
		}
		err = udpConn.SetDeadline(time.Now().Add(time.Second))
		if err != nil {
			klog.ErrorS(err, "SetDeadline failed")
			continue
		}
	}
}

func copyDatagram(udpConn *net.UDPConn, outConn net.Conn) {
	defer udpConn.Close()
	var buffer [4096]byte
	for {
		n, _, err := udpConn.ReadFromUDP(buffer[0:])
		if err != nil {
			if !IsTimeoutError(err) && !IsEOFError(err) {
				klog.ErrorS(err, "Read failed")
			}
			break
		}
		err = udpConn.SetDeadline(time.Now().Add(time.Second))
		if err != nil {
			klog.ErrorS(err, "SetDeadline failed")
			break
		}
		_, err = outConn.Write(buffer[0:n])
		if err != nil {
			if !IsTimeoutError(err) {
				klog.ErrorS(err, "WriteTo failed")
			}
			break
		}
	}
}
