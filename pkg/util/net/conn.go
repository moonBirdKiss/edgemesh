package util

import (
	"bytes"
	"fmt"
	"github.com/libp2p/go-libp2p/core/network"
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
	go RouteCopyBytes("[RouteConn]: from out to in", in, out, &wg)
	go RouteCopyBytes("[RouteConn]: from in to out", out, in, &wg)
	wg.Wait()
}

func RouteCopyBytes(direction string, dest, src io.ReadWriteCloser, wg *sync.WaitGroup) {
	defer wg.Done()
	klog.Info("[RouteCopyBytes]: Copying remote address bytes")
	n, err := io.Copy(dest, src)
	if err != nil {
		if !IsClosedError(err) && !IsStreamResetError(err) {
			klog.ErrorS(err, "I/O error occurred")
		}
	}
	klog.Info("[RouteCopyBytes]: Copied remote address bytes.", "bytes: ", n, " direction: ", direction)
	if err = dest.Close(); err != nil && !IsClosedError(err) {
		klog.ErrorS(err, "[RouteCopyBytes]: dest close failed")
	}
	if err = src.Close(); err != nil && !IsClosedError(err) {
		klog.ErrorS(err, "[RouteCopyBytes]: src close failed")
	}
}

// RouteCopyStreamBak needs go-routine to execute the io.Copy
func RouteCopyStreamBak(dest, src io.ReadWriteCloser) {
	klog.Info("[RouteCopyStream]: Copying remote address bytes")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		_, err := io.Copy(dest, src)
		if err != nil {
			if !IsClosedError(err) && !IsStreamResetError(err) {
				klog.ErrorS(err, "I/O error occurred")
			}
		}
		wg.Done()
	}()

	//klog.Info("[RouteCopyStream]: Copied remote address bytes.", "bytes: ", n)

	msgStr := `HTTP/1.0 200 OK
Content-Type: text/plain
Content-Length: 16

Hello, my friend
`
	cnt, err := src.Write([]byte(msgStr))
	if err != nil {
		klog.Errorf("[RouteStream]: Write data error: %v", err)
		return
	}
	klog.Infof("[RouteStream]: Successfully writ %d to src", cnt)

	wg.Wait()
	if err = src.Close(); err != nil && !IsClosedError(err) {
		klog.ErrorS(err, "src close failed")
	}
	klog.Infof("[RouteStream]: Successfully close src")

}

func RouteCopyStream(dest *RouteConnection, src network.Stream) {

	//n, err := io.Copy(dest, src)
	klog.Info("[RouteCopyStream]: Start copying remote bytes")
	dstBuffer := dest.GetBuffer()
	n, err := copyBuffer(&dstBuffer, src, nil, "RouteCopyStream")

	if err != nil {
		klog.Infof("[RouteCopyStream]: Failed to copy data: %v", err)
	}

	// n, err := io.Copy(&dstBuffer, src)
	klog.Info("[RouteCopyStream]: Copying remote address bytes: ", n)

	msgStr := `HTTP/1.0 200 OK
Content-Type: text/plain
Content-Length: 16

Hello, my friend
`
	cnt, err := src.Write([]byte(msgStr))
	if err != nil {
		klog.Errorf("[RouteStream]: Write data error: %v", err)
		return
	}
	klog.Infof("[RouteStream]: Successfully writ %d to src", cnt)

	if err = src.Close(); err != nil && !IsClosedError(err) {
		klog.ErrorS(err, "src close failed")
	}
	klog.Infof("[RouteStream]: Successfully close src")

}

// ProxyConn proxies data bi-directionally between in and out.
func ProxyConn(in, out net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)
	//klog.Info("Creating proxy between remote and local addresses", "inRemoteAddress", in.RemoteAddr(),
	//	"inLocalAddress", in.LocalAddr(), "outLocalAddress", out.LocalAddr(), "outRemoteAddress", out.RemoteAddr())
	go copyBytes("from backend", in, out, &wg)
	go copyBytes("to backend", out, in, &wg)
	wg.Wait()
}

func copyBytes(direction string, dest, src net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	//klog.V(4).InfoS("Copying remote address bytes", "direction", direction, "sourceRemoteAddress", src.RemoteAddr(), "destinationRemoteAddress", dest.RemoteAddr())
	klog.Infof("[copyBytes]: %s", direction)
	n, err := copyBuffer(dest, src, nil, direction)
	if err != nil {
		if !IsClosedError(err) && !IsStreamResetError(err) {
			klog.ErrorS(err, "I/O error occurred")
		}
	}
	klog.Infof("[copyBytes]: %s, and the total bytes: %d", direction, n)
	//klog.V(4).InfoS("Copied remote address bytes", "bytes", n, "direction", direction, "sourceRemoteAddress", src.RemoteAddr(), "destinationRemoteAddress", dest.RemoteAddr())
	if err = dest.Close(); err != nil && !IsClosedError(err) {
		klog.ErrorS(err, "dest close failed")
	}
	if err = src.Close(); err != nil && !IsClosedError(err) {
		klog.ErrorS(err, "src close failed")
	}
	klog.Infof("[copyBytes]: %s successfully", direction)
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

// copyBuffer is the actual implementation of Copy and CopyBuffer.
// if buf is nil, one is allocated.
func copyBuffer(dst io.Writer, src io.Reader, buf []byte, direction string) (written int64, err error) {
	if buf == nil {
		// todo: here we use 1MB as the buffer
		size := 1024 * 1024
		if l, ok := src.(*io.LimitedReader); ok && int64(size) > l.N {
			if l.N < 1 {
				size = 1
			} else {
				size = int(l.N)
			}
		}
		buf = make([]byte, size)
	}

	klog.Infof("[copyBuffer]: %s start to copy buffer", direction)
	for i := 0; ; i++ {
		klog.Infof("[copyBuffer]: %s start to read, index: %d", direction, i)
		nr, er := src.Read(buf)
		klog.Infof("[copyBuffer]: %s end to read, index: %d, size: %d ", direction, i, nr)
		klog.Infof("[copyBuffer]: the data: \n %s", buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = fmt.Errorf("errInvalidWrite")
					klog.Infoln(ew)
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				klog.Infoln(ew)
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				klog.Infoln(ew)
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			klog.Infoln(er)
			break
		}
		klog.Infof("[copyBuffer] index: %d finished!", i)
	}
	klog.Infof("[copyBuffer]: %s end to copy buffer, err: %v", direction, err)
	return written, err
}
