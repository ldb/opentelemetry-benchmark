package worker

import (
	"bytes"
	"context"
	"errors"
	"go.opentelemetry.io/collector/model/otlp"
	"go.opentelemetry.io/collector/model/pdata"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// receiver is an HTTP server that accepts spans from the Openetelemetry Collector.
// It discards the spans, parses the `service.name` attribute and notifies workers about received traces.
type receiver struct {
	Host string
	um   pdata.TracesUnmarshaler
	init sync.Once
}

func (r *receiver) ReceiveTraces(notify func(int) error) (func(ctx context.Context) error, func() error) {
	r.init.Do(func() {
		r.um = otlp.NewProtobufTracesUnmarshaler()
	})

	if notify == nil {
		return nil, func() error { return errors.New("notify must not be nil") }
	}

	var handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body := bytes.Buffer{}

		_, err := body.ReadFrom(request.Body)
		if err != nil {
			log.Printf("error reading request body: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		request.Body.Close()
		tt, err := r.um.UnmarshalTraces(body.Bytes())
		if err != nil {
			log.Printf("error unmarshaling traces: %v", err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		for i := 0; i < tt.ResourceSpans().Len(); i++ {
			e := tt.ResourceSpans().At(i)
			r := e.Resource()
			v, ok := r.Attributes().Get("service.name")
			if !ok {
				log.Printf(`could not find resource attribute "service.name"`)
				continue
			}
			serviceName := v.AsString()
			if !strings.HasPrefix(serviceName, "benchd-worker") {
				continue
			}
			c := strings.Split(serviceName, ".")
			if len(c) != 3 {
				log.Printf(`malformed service.name attribute "%s"`, serviceName)
			}
			id, err := strconv.Atoi(c[2])
			if err != nil {
				log.Printf(`malformed id in service.name attribute "%s"`, serviceName)
				continue

			}
			if err := notify(id); err != nil {
				log.Printf("error notifying worker with ID %d: %v", id, err)
				continue
			}
		}
		writer.WriteHeader(http.StatusOK)
	})

	server := &http.Server{Addr: r.Host, Handler: handler}
	return server.Shutdown, server.ListenAndServe
}
