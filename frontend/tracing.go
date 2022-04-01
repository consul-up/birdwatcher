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
	"net/http"
	"strings"
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

// httpClientTrace traces a request. It returns a function that should be called
// on the result of httpClient.Get().
func httpClientTrace(req *http.Request, opName string) func(response *http.Response, err error) {
	parentSpan := opentracing.SpanFromContext(req.Context())
	if parentSpan == nil {
		return nil
	}

	clientSpan := opentracing.StartSpan(opName,
		opentracing.ChildOf(parentSpan.Context()))
	ext.SpanKindRPCClient.Set(clientSpan)
	ext.HTTPUrl.Set(clientSpan, req.URL.String())
	ext.HTTPMethod.Set(clientSpan, req.Method)

	// Transmit the span's TraceContext as HTTP headers on our
	// outbound request.
	opentracing.GlobalTracer().Inject(
		clientSpan.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header))

	return func(response *http.Response, err error) {
		if err != nil {
			ext.Error.Set(clientSpan, true)
			clientSpan.LogFields(otlog.Error(err))
			return
		}
		ext.HTTPStatusCode.Set(clientSpan, uint16(response.StatusCode))
		if response.StatusCode < 200 || response.StatusCode > 299 {
			ext.Error.Set(clientSpan, true)
		}
		clientSpan.Finish()
	}
}

// tracingMiddleware adds spans for request handling.
func tracingMiddleware() func(*gin.Context) {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/static") {
			c.Next()
			return
		}

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
