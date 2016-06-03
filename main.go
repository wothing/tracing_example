package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/appdash"

	opentracing "github.com/opentracing/opentracing-go"
	basictracer "github.com/opentracing/basictracer-go"
	apptracer "sourcegraph.com/sourcegraph/appdash/opentracing"
)

var collector appdash.Collector

func main() {
	collector := appdash.NewRemoteCollector(":1726")
	tracer := apptracer.NewTracer(collector)
	opentracing.InitGlobalTracer(tracer)

	router := mux.NewRouter()
	router.HandleFunc("/", Home)

	n := negroni.Classic()
	n.UseHandler(router)
	n.Run(":8699")
}

func Home(rw http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	span, ctx := opentracing.StartSpanFromContext(ctx, "HOME")
	defer span.Finish()

	span.SetTag("Request.Host", r.Host)
	span.SetTag("Reqeust.Address", r.RemoteAddr)

	span.SetBaggageItem("User", "rod")

	for i := 0; i < 3; i++ {
		Test(ctx)
	}
	fmt.Fprintf(rw, `<p>Three API requests have been made!<p>`)
	fmt.Fprintf(rw, `<p><a href="http://localhost:8700/traces" target="_">View the trace</a></p>`)
}

func Test(ctx context.Context) int {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Test")
	defer span.Finish()

	span.LogEvent("before sleeping")
	span.LogEvent(fmt.Sprintf("%016x", span.(basictracer.Span).Context().TraceID))
	time.Sleep(200 * time.Millisecond)
	span.LogEvent("after sleeping")

	return 1
}
