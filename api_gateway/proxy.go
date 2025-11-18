package main

import (
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

var httpClient = &http.Client{}

func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func proxyRequest(c *gin.Context, targetBase string) {
	targetURL, err := url.Parse(targetBase)
	if err != nil {
		fail(c, http.StatusInternalServerError, "CONFIG_ERROR", "Invalid target URL")
		return
	}

	// целевой URL = targetBase + оригинальный путь + query
	targetURL.Path = c.Request.URL.Path
	targetURL.RawQuery = c.Request.URL.RawQuery

	req, err := http.NewRequest(c.Request.Method, targetURL.String(), c.Request.Body)
	if err != nil {
		fail(c, http.StatusInternalServerError, "PROXY_ERROR", "Failed to create proxied request")
		return
	}

	// заголовки пользователя → сервис
	copyHeaders(req.Header, c.Request.Header)

	// гарантируем X-Request-ID
	reqID := getRequestID(c)
	req.Header.Set("X-Request-ID", reqID)

	resp, err := httpClient.Do(req)
	if err != nil {
		fail(c, http.StatusBadGateway, "UPSTREAM_ERROR", "Failed to call upstream service")
		return
	}
	defer resp.Body.Close()

	// заголовки от сервиса → клиент
	for k, vv := range resp.Header {
		for _, v := range vv {
			c.Writer.Header().Add(k, v)
		}
	}

	c.Writer.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(c.Writer, resp.Body)
}

func proxyToUsers(c *gin.Context) {
	proxyRequest(c, usersServiceURL)
}

func proxyToOrders(c *gin.Context) {
	proxyRequest(c, ordersServiceURL)
}
