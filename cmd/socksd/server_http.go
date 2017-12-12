package main

import (
	"net/http"

	"github.com/ssoor/socks"
)

func HTTPServe(addr string, router socks.Dialer, tran *HTTPTransport) (error) {
	handler := socks.NewHTTPProxyHandler("http", router, tran)

	return  http.ListenAndServe(addr, handler)
}