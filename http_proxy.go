package socks

import (
	"bytes"
	"strings"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// HTTPProxy is an HTTP Handler that serve CONNECT method and
// route request to proxy server by Router.
type HTTPProxy struct {
	scheme  string
	forward Dialer
	*httputil.ReverseProxy
}

// NewHTTPProxy constructs one HTTPProxy
func NewHTTPProxy(scheme string, forward Dialer, transport http.RoundTripper) *HTTPProxy {
	return &HTTPProxy{
		scheme:  scheme,
		forward: forward,
		ReverseProxy: &httputil.ReverseProxy{
			Director:  director,
			Transport: transport,
		},
	}
}

func director(request *http.Request) {
	u, err := url.Parse(request.RequestURI)
	if err != nil {
		return
	}

	request.RequestURI = u.RequestURI()
	valueConnection := request.Header.Get("Proxy-Connection")
	if valueConnection != "" {
		request.Header.Del("Connection")
		request.Header.Del("Proxy-Connection")
		request.Header.Add("Connection", valueConnection)
	}
}

// ServeHTTPTunnel serve incoming request with CONNECT method, then route data to proxy server
func (h *HTTPProxy) ServeHTTPTunnel(response http.ResponseWriter, request *http.Request) {
	var conn net.Conn
	if hj, ok := response.(http.Hijacker); ok {
		var err error
		if conn, _, err = hj.Hijack(); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(response, "Hijacker failed", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	dest, err := h.forward.Dial("tcp", request.Host)
	if err != nil {
		fmt.Fprintf(conn, "HTTP/1.0 500 NewRemoteSocks failed, err:%s\r\n\r\n", err)
		return
	}
	defer dest.Close()

	if request.Body != nil {
		if _, err = io.Copy(dest, request.Body); err != nil {
			fmt.Fprintf(conn, "%d %s", http.StatusBadGateway, err.Error())
			return
		}
	}

	fmt.Fprintf(conn, "HTTP/1.0 200 Connection established\r\n\r\n")

	go func() {
		defer conn.Close()
		defer dest.Close()
		io.Copy(dest, conn)
	}()
	io.Copy(conn, dest)
}


// ServeHTTPTunnel serve incoming request with CONNECT method, then route data to proxy server
func (h *HTTPProxy) xInternalConnect(req *http.Request, response http.ResponseWriter) {
	var conn net.Conn
	if hj, ok := response.(http.Hijacker); ok {
		var err error
		if conn, _, err = hj.Hijack(); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(response, "Hijacker failed", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	host := req.Host+":80"
	if strings.EqualFold(req.URL.Scheme,"https") {
		host = req.Host+":443"
	}
	
	dest, err := h.forward.Dial("tcp", host)
	if err != nil {
		fmt.Fprintf(conn, "HTTP/1.0 500 NewRemoteSocks failed, err:%s\r\n\r\n", err)
		return
	}
	defer dest.Close()

	sendRequest, err :=  httputil.DumpRequestOut(req, true)
	if err != nil {
		fmt.Fprintf(conn, "HTTP/1.0 500 Dump Request failed, err:%s\r\n\r\n", err)
		return
	}
	println(string(sendRequest))
	if _, err = io.Copy(dest, bytes.NewReader(sendRequest)); err != nil {
		fmt.Fprintf(conn, "%d %s", http.StatusBadGateway, err.Error())
		return
	}

	go func() {
		defer conn.Close()
		defer dest.Close()
		io.Copy(dest, conn)
	}()
	io.Copy(conn, dest)
}

// ServeHTTP implements HTTP Handler
func (h *HTTPProxy) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	request.URL.Scheme = h.scheme
	request.URL.Host = request.Host
	
	if request.Method == "GET" && strings.EqualFold(request.Header.Get("Connection"), "Upgrade") && strings.EqualFold(request.Header.Get("Upgrade"), "websocket") {
		h.xInternalConnect(request, response)
	}

	if request.Method == "CONNECT" {
		h.ServeHTTPTunnel(response, request)
	} else {
		h.ReverseProxy.ServeHTTP(response, request)
	}
}
