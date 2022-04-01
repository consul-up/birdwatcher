package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	reporterhttp "github.com/openzipkin/zipkin-go/reporter/http"
)

// initTracing initializes distributed tracing and set the global tracer.
func initTracing(zipkinCollectorURL string, bindAddr string, serviceName string) error {
	reporter := reporterhttp.NewReporter(fmt.Sprintf("%s/api/v2/spans", zipkinCollectorURL))
	endpoint, err := zipkin.NewEndpoint(serviceName, bindAddr)
	if err != nil {
		return err
	}

	zipkinTracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint), zipkin.WithSharedSpans(false))
	if err != nil {
		return err
	}

	tracer := zipkinot.Wrap(zipkinTracer)
	opentracing.SetGlobalTracer(tracer)
	return nil
}

// tracingMiddleware adds spans for request handling.
func tracingMiddleware() func(*gin.Context) {
	return func(c *gin.Context) {
		wireContext, err := opentracing.GlobalTracer().Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(c.Request.Header),
		)

		if err != nil && err != opentracing.ErrSpanContextNotFound {
			logf("error obtaining context: %s", err)
			c.Next()
			return
		}

		// Create the span referring to the RPC client if available.
		// If wireContext == nil, a root span will be created.
		serverSpan := opentracing.StartSpan(
			c.FullPath(),
			ext.RPCServerOption(wireContext))
		serverSpan.SetTag("http.host", c.Request.Host)
		ext.HTTPMethod.Set(serverSpan, c.Request.Method)
		ext.HTTPUrl.Set(serverSpan, c.Request.URL.String())
		serverSpan.SetTag("http.user_agent", c.Request.UserAgent())
		ext.PeerAddress.Set(serverSpan, c.Request.RemoteAddr)

		// Set into context.
		c.Request = c.Request.WithContext(opentracing.ContextWithSpan(c.Request.Context(), serverSpan))

		// Call next middleware.
		c.Next()

		// Record response in span.
		status := c.Writer.Status()
		ext.HTTPStatusCode.Set(serverSpan, uint16(status))
		if status < 200 || status > 299 {
			ext.Error.Set(serverSpan, true)
		}
		if len(c.Errors) > 0 {
			serverSpan.LogFields(otlog.String("gin.errors", c.Errors.String()))
		}
		serverSpan.Finish()
	}
}
