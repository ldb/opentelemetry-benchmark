package worker

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.opentelemetry.io/otel/trace"
)

type status int

const (
	success status = iota
	timeout
	stopped
)

type Logger interface {
	Println(m ...interface{})
}

type Worker struct {
	ID          int
	TraceDepth  int
	NumberSpans int
	SpanLength  time.Duration
	MaxCoolDown time.Duration

	Tracer      trace.Tracer
	FinishTrace <-chan struct{} // Manager notifies the worker on this channel that it can stop recording the current trace
	Logger      Logger

	Timeout time.Duration

	startT  time.Time
	sendT   time.Time
	finishT time.Time
}

func (w *Worker) Run(ctx context.Context) error {
	//timeoutTimer := time.NewTimer(w.Timeout)
	//defer timeoutTimer.Stop()
	for {
		select {
		case <-ctx.Done():
			w.log(stopped)
			return ctx.Err()

		case <-w.FinishTrace:
			w.finishT = time.Now()
			//	timeoutTimer.Stop()
			w.log(success)
			time.Sleep(time.Duration(rand.Int63n(w.MaxCoolDown.Milliseconds())))

		//case <-timeoutTimer.C:
		//	w.log(timeout)

		default:
			w.startT = time.Now()
			w.generateSpans()
			// flush spans
			// send spans
			w.sendT = time.Now()
			//	timeoutTimer.Reset()
		}
	}
}

// log sends a log message of the recorded timings into the (*Worker).Log channel.
func (w *Worker) log(s status) {
	w.Logger.Println(fmt.Sprintf("%d %d %d %d %d %d %d %d %d %d",
		w.ID,
		int(s),
		w.TraceDepth,
		w.NumberSpans,
		w.SpanLength.Milliseconds(),
		w.MaxCoolDown.Milliseconds(),
		w.startT.UnixMilli(),
		w.sendT.UnixMilli(),
		w.finishT.UnixMilli(),
		w.finishT.Sub(w.sendT).Milliseconds(),
	))
}

func (w *Worker) generateSpans() {
	time.Sleep(1 * time.Second)
}
