package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httptrace"
	"net/http/httputil"
	"time"
)

const (
	HTTP_STATUS_UNKNOWN      = "http_unknown"
	HTTP_STATUS_HEALTHY      = "http_healthy"
	HTTP_STATUS_HIGH_LATENCY = "http_high_latency"
	HTTP_STATUS_DOWN         = "http_down"
)

func timeGet(url string) (*http.Response, error, time.Duration) {
	req, _ := http.NewRequest("GET", url, nil)

	var res *http.Response
	var timedelta time.Duration
	var glob_err error
	glob_err = nil

	var start, connect, connect_duration, dns, dns_duration, tlsHandshake, tlsHandshake_duration, total time.Time

	trace := &httptrace.ClientTrace{
		DNSStart: func(dsi httptrace.DNSStartInfo) { dns = time.Now() },
		DNSDone: func(ddi httptrace.DNSDoneInfo) {
			dns_duration := time.Since(dns)
			if ddi.Err != nil {
				//log.Printf("URL: %s, DNS error: %s", url, ddi.Err)
				timedelta = dns_duration
				glob_err = ddi.Err
			}
		},

		TLSHandshakeStart: func() { tlsHandshake = time.Now() },
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			tlsHandshake_duration := time.Since(tlsHandshake)
			if err != nil {
				//log.Printf("URL: %s, TLS error: %s", url, err)
				if glob_err == nil {
					timedelta = tlsHandshake_duration
					glob_err = err
				}
			}
		},

		ConnectStart: func(network, addr string) { connect = time.Now() },
		ConnectDone: func(network, addr string, err error) {
			connect_duration := time.Since(connect)
			if err != nil {
				//log.Printf("URL: %s, Connect error: %s", url, err)
				if glob_err == nil {
					timedelta = connect_duration
					glob_err = err
				}
			}
		},

		GotFirstResponseByte: func() {
			total := time.Since(start)
			if glob_err == nil {
				timedelta = total
			}
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	start = time.Now()
	res, glob_err = http.DefaultTransport.RoundTrip(req)
	return res, glob_err, timedelta
}

func (l *LoadBalancer) InitialHostCheck() {
	//check if hosts are alive
	for _, host := range l.Config.InitialAddresses {
		go func(host string) {
			res, err := http.Get(host + l.Config.HealthCheckPath)
			if err != nil {
				log.Printf("Error checking host %s: %s", host, err)
				return
			}
			if res.StatusCode != 200 {
				log.Printf("Host %s unable to intialize, status %s, body %s", host, res.StatusCode, res.Body)
				return
			}
			l.HostStatus.Store(host, HTTP_STATUS_HEALTHY)
		}(host)
	}
}

func (l *LoadBalancer) UpdateAliveHosts() {
	for _, host := range l.Config.InitialAddresses {
		go func(host string) {
			hostStatus, _ := l.HostStatus.Load(host)
			if hostStatus == HTTP_STATUS_DOWN || hostStatus == HTTP_STATUS_UNKNOWN {
				return
			}
			//check status
			res, err := http.Get(host + l.Config.HealthCheckPath)
			if err != nil {
				log.Printf("Error checking host %s: %s", host, err)
				return
			}
			if res.StatusCode != 200 {
				log.Printf("Host %s unable to intialize, status %s, body %s", host, res.StatusCode, res.Body)
				return
			}
			//update host status
			l.HostStatus.Store(host, HTTP_STATUS_HEALTHY)
		}(host)
	}
}

func (l *LoadBalancer) RechekDeadHosts() {

}

func (l *LoadBalancer) ServeHTTP() {
	log.Printf("Starting HTTP server on %s:%d", l.Config.Host, l.Config.Port)
	//initial host scheck
	l.InitialHostCheck()
	//schedule regular host checks

	//start http server

	rewrite := func(r *http.Request) {
		r.URL.Scheme = "http"
		r.URL.Host
	}

	rpx := httputil.ReverseProxy{}
}
