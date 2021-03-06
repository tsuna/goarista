// Copyright (C) 2015  Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

// Package monitor provides an embedded HTTP server to expose
// metrics for monitoring
package monitor

import (
	"expvar"
	"fmt"
	"html/template"
	"net/http"
	_ "net/http/pprof" // Go documentation recommended usage
	"strings"

	"github.com/aristanetworks/glog"
)

// Server represents a monitoring server
type Server interface {
	Run()
}

// server contains information for the monitoring server
type server struct {
	// Server name e.g. host[:port]
	serverName string
}

// NewServer creates a new server struct
func NewServer(serverName string) Server {
	return &server{
		serverName: serverName,
	}
}

func debugHandler(w http.ResponseWriter, r *http.Request) {
	indexTmpl := `<html>
	<head>
	<title>/debug</title>
	</head>
	<body>
	<p>/debug</p>
	<div><a href="/debug/vars">vars</a></div>
	<div><a href="/debug/pprof">pprof</a></div>
	</body>
	</html>
	`
	fmt.Fprintf(w, indexTmpl)
}

// Pretty prints the latency histograms
func latencyHandler(w http.ResponseWriter, r *http.Request) {
	expvar.Do(func(kv expvar.KeyValue) {
		if strings.HasSuffix(kv.Key, "Histogram") {
			template.Must(template.New("latency").Parse(
				`<html>
					<head>
						<title>/debug/latency</title>
					</head>
					<body>
						<pre>{{.}}</pre>
					</body>
				</html>
			`)).Execute(w, template.HTML(strings.Replace(kv.Value.String(), "\\n", "<br />", -1)))
		}
	})
}

// Run sets up the HTTP server and any handlers
func (s *server) Run() {
	http.HandleFunc("/debug", debugHandler)
	http.HandleFunc("/debug/latency", latencyHandler)

	// monitoring server
	err := http.ListenAndServe(s.serverName, nil)
	if err != nil {
		glog.Errorf("Could not start monitor server: %s", err)
	}
}
