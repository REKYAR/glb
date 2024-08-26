package main

import (
	"context"
	"fmt"
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

func (l *LoadBalancer) timeGet(url string) (*http.Response, time.Duration, error) {
	req, _ := http.NewRequest("GET", url, nil)
	var timedelta time.Duration
	var start time.Time

	trace := &httptrace.ClientTrace{
		GotFirstResponseByte: func() {
			timedelta = time.Since(start)
		},
	}
	ctx, cancel := context.WithTimeout(req.Context(), time.Duration(l.Config.HealthCheckTimeout)*time.Millisecond)
	defer cancel()

	req = req.WithContext(httptrace.WithClientTrace(ctx, trace))
	start = time.Now()
	res, err := http.DefaultTransport.RoundTrip(req)
	return res, timedelta, err
}

func (l *LoadBalancer) InitialHostCheck() {
	//check if hosts are alive
	for _, host := range l.Config.InitialAddresses {
		go func(host string) {
			//res, err := http.Get(host + l.Config.HealthCheckPath)
			res, timedelta, err := l.timeGet(host + l.Config.HealthCheckPath)
			if err != nil {
				log.Printf("Error checking host %s: %s", host, err)
				return
			}
			if res.StatusCode != 200 {
				log.Printf("Host %s unable to intialize, status %d, body %s", host, res.StatusCode, res.Body)
				l.HostStatus.Store(host, HTTP_STATUS_DOWN)
				return
			}
			if timedelta > time.Duration(l.Config.HealthCheckUnhealthyThreshold)*time.Millisecond {
				log.Printf("Host %s has high latency: %s", host, timedelta)
				l.HostStatus.Store(host, HTTP_STATUS_HIGH_LATENCY)
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
			//res, err := http.Get(host + l.Config.HealthCheckPath)
			res, timedelta, err := l.timeGet(host + l.Config.HealthCheckPath)
			if err != nil {
				log.Printf("Error checking host %s: %s", host, err)
				l.HostStatus.Store(host, HTTP_STATUS_DOWN)
				return
			}
			if res.StatusCode != 200 {
				log.Printf("Host %s non 200 response, status %d, body %s", host, res.StatusCode, res.Body)
				l.HostStatus.Store(host, HTTP_STATUS_DOWN)
				return
			}
			if timedelta > time.Duration(l.Config.HealthCheckUnhealthyThreshold)*time.Millisecond {
				log.Printf("Host %s has high latency: %s", host, timedelta)
				l.HostStatus.Store(host, HTTP_STATUS_HIGH_LATENCY)
				return
			}
			l.HostStatus.Store(host, HTTP_STATUS_HEALTHY)
		}(host)
	}
}

func (l *LoadBalancer) UpdateDownHosts() {
	for _, host := range l.Config.InitialAddresses {
		go func(host string) {
			hostStatus, _ := l.HostStatus.Load(host)
			if hostStatus != HTTP_STATUS_DOWN && hostStatus != HTTP_STATUS_UNKNOWN {
				return
			}
			res, timedelta, err := l.timeGet(host + l.Config.HealthCheckPath)
			if err != nil {
				log.Printf("Error checking host %s: %s", host, err)
				l.HostStatus.Store(host, HTTP_STATUS_DOWN)
				return
			}
			if res.StatusCode != 200 {
				log.Printf("Host %s non 200 response, status %d, body %s", host, res.StatusCode, res.Body)
				l.HostStatus.Store(host, HTTP_STATUS_DOWN)
				return
			}
			if timedelta > time.Duration(l.Config.HealthCheckUnhealthyThreshold)*time.Millisecond {
				log.Printf("Host %s has high latency: %s", host, timedelta)
				l.HostStatus.Store(host, HTTP_STATUS_HIGH_LATENCY)
				return
			}
			l.HostStatus.Store(host, HTTP_STATUS_HEALTHY)
		}(host)
	}
}

func (l *LoadBalancer) ServeHTTP() error {
	log.Printf("Starting HTTP server on %s:%d", l.Config.Host, l.Config.Port)
	//initial host scheck
	l.InitialHostCheck()
	// Schedule regular host checks for alive hosts
	go func() {
		ticker := time.NewTicker(time.Duration(l.Config.HealthCheckInterval) * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			l.UpdateAliveHosts()
		}
	}()

	// Schedule regular host checks for down hosts
	go func() {
		ticker := time.NewTicker(time.Duration(l.Config.HealthCheckDownInterval) * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			l.UpdateDownHosts()
		}
	}()

	//start http server

	rewrite := func(r *httputil.ProxyRequest) {
		r.Out.URL.Host = l.Config.InitialAddresses[0] //TODO: implement load balancing
	}

	// listener, err := net.Listen("tcp", l.Config.Host+":"+fmt.Sprintf("%d", l.Config.Port))
	// if err != nil {
	// 	return err
	// }
	// l.Config.Port = listener.Addr().(*net.TCPAddr).Port

	rpx := &httputil.ReverseProxy{
		Rewrite: rewrite,
	}

	s := &http.Server{
		Addr:    l.Config.Host + ":" + fmt.Sprintf("%d", l.Config.Port),
		Handler: rpx,
	}
	//return s.Serve(listener)
	return s.ListenAndServe()
}
