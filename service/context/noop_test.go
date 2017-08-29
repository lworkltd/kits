package context

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Sirupsen/logrus"
)

func TestNoopTracer(t *testing.T) {
	tracer := &NoopTracer{}
	tracer.Finish()
	var hearder http.Header
	tracer.Inject(hearder)
	tracer.SpanId()
	tracer.TracingId()
}

func TestNoopContext(t *testing.T) {
	ctx := NewNoopContext()
	ctx.Deadline()
	ctx.Done()
	ctx.WithFields(logrus.Fields{
		"key1": "value1",
		"key2": "value2",
	}).WithFields(logrus.Fields{
		"key4": "value4",
		"key5": "value5",
	}).Info("info")
	ctx.WithError(fmt.Errorf("error2")).WithError(fmt.Errorf("error2")).Info("info")
	ctx.WithField("key", "value").WithField("key1", "value2").Info("info")
	subCtx := ctx.SubContext("")
	subCtx.TracingId()
	subCtx.SpanId()
}

func TestNoopLogger(t *testing.T) {
	var noopLogger NoopLogger
	noopLogger.WithFields(logrus.Fields{
		"key1": "value1",
		"key2": "value2",
	}).WithFields(logrus.Fields{
		"key4": "value4",
		"key5": "value5",
	}).Info("info")
	noopLogger.WithError(fmt.Errorf("error2")).WithError(fmt.Errorf("error2")).Info("info")
	noopLogger.WithField("key", "value").WithField("key1", "value2").Info("info")
}
