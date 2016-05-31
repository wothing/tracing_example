package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/urfave/negroni"
	"github.com/gorilla/mux"
	"github.com/gorilla/context"
	"github.com/wothing/log"
	"sourcegraph.com/sourcegraph/appdash"
	
	/*
	apptracer "sourcegraph.com/sourcegraph/appdash/opentracing"
	opentracing "github.com/opentracing/opentracing-go"
	*/
)

var collector appdash.Collector

func main() {
	collector = appdash.NewRemoteCollector(":1726")
	tracemw := httptrace.Middleware(collector, &httptrace.MiddlewareConfig{
		RouteName: func(r *http.Request) string { return r.URL.Path },
		SetContextSpan: func(r *http.Request, spanID appdash.SpanID) {
			context.Set(r, "span", spanID)
		},
	})

	router := mux.NewRouter()
	router.HandleFunc("/", Home)
	router.HandleFunc("/test", Test)

	n := negroni.Classic()
	n.Use(negroni.HandlerFunc(tracemw))
	n.UseHandler(router)
	n.Run(":8699")
}

func Home(rw http.ResponseWriter, r *http.Request) {
	span := context.Get(r, "span").(appdash.SpanID)
	httpClient := &http.Client{
		Transport: &httptrace.Transport{
			Recorder: appdash.NewRecorder(span, collector),
			SetName: true,
		},
	}
	for i := 0; i < 3; i++ {
		resp, err := httpClient.Get("http://localhost:8699/test")
		if err != nil {
			log.Errorf("calling /test error: %v", err)
			continue
		}
		resp.Body.Close()
	}
	fmt.Fprintf(rw, `<p>Three API requests have been made!<p>`)
	fmt.Fprintf(rw, `<p><a href="http://localhost:8700/traces/%s" target="_">View the trace (ID:%s)</a></p>`, span.Trace, span.Trace)
}

func Test(rw http.ResponseWriter, r *http.Request) {
	time.Sleep(200 * time.Millisecond)
	fmt.Fprintf(rw, "Slept for 200ms!")
}
