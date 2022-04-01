package main

import (
	_ "embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const (
	defaultBindAddr = "0.0.0.0:7000"
	defaultVersion  = "v1"
	envBindAddr     = "BIND_ADDR"
	envTracingURL   = "TRACING_URL"
	envVersion      = "VERSION"
)

type birdResponse struct {
	Name     string `json:"name"`
	ImageURL string `json:"imageURL"`
	Extract  string `json:"extract"`
}

func main() {
	// Decide on Gin mode.
	if os.Getenv(gin.EnvGinMode) != gin.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	if err := r.SetTrustedProxies(nil); err != nil {
		log.Fatal(err.Error())
	}

	// Figure out our bind address.
	bindAddr := defaultBindAddr
	if addr := os.Getenv(envBindAddr); addr != "" {
		bindAddr = addr
	}

	// Optionally initialize tracing.
	if zipkinURL := os.Getenv(envTracingURL); zipkinURL != "" {
		err := initTracing(zipkinURL, bindAddr, "backend")
		if err != nil {
			log.Fatalf("Error initializing tracing: %s\n", err)
		}
		logf("Tracing enabled url=%q", zipkinURL)
		r.Use(tracingMiddleware())
	}

	version := defaultVersion
	versionFromEnv := os.Getenv(envVersion)
	if versionFromEnv != "" {
		if versionFromEnv != "v1" && versionFromEnv != "v2" {
			log.Fatalf("Unsupported %s=%s; only v1 and v2 are supported\n", envVersion, versionFromEnv)
		}
		version = versionFromEnv
	}

	// Get our hostname.
	hostname, err := os.Hostname()
	if err != nil {
		hostname = err.Error()
	}

	// Get the birds.
	var birdList []birdRawData
	if version == defaultVersion {
		birdList = birds()
	} else {
		birdList = canaries()
	}

	// birdCallCount keeps track of how many times the /bird
	// endpoint has been called. We use this to ensure we return a new bird every time
	// rather than a random bird.
	// Starts at -1 because this counter is always incremented first.
	var birdCallCount int64 = -1

	// GET /bird.
	r.GET("/bird", func(c *gin.Context) {
		for k, v := range c.Request.Header {
			log.Printf("%s=%s\n", k, v[0])
		}

		// Delay handling.
		if delay := c.Query("delay"); delay != "" && delay != "0" {
			delaySec, err := strconv.ParseFloat(delay, 32)
			if err != nil {
				c.JSON(400, gin.H{
					"metadata": map[string]interface{}{
						"hostname": hostname,
						"version":  version,
					},
					"error": fmt.Sprintf("error parsing query param \"delay\": %s", err),
				})
				return
			}

			// doDelay runs in a closure just so we can call defer delaySpan.finish()
			// which makes the code cleaner.
			doDelay := func() {
				parentSpan := opentracing.SpanFromContext(c.Request.Context())
				if parentSpan != nil {
					// Add tracing span for the delay.
					delaySpan := opentracing.StartSpan("synthetic_delay",
						opentracing.ChildOf(parentSpan.Context()))
					delaySpan.SetTag("delay_seconds", delaySec)
					defer delaySpan.Finish()
				}

				time.Sleep(time.Duration(delaySec*1000) * time.Millisecond)
			}

			doDelay()
		}

		// Error handling.
		if errorRate := c.Query("error-rate"); errorRate != "" && errorRate != "0" {
			errorRateInt, err := strconv.Atoi(errorRate)
			if err != nil {
				c.JSON(400, gin.H{
					"metadata": map[string]interface{}{
						"hostname": hostname,
						"version":  version,
					},
					"error": fmt.Sprintf("error parsing query param \"error-rate\": %s", err),
				})
				return
			}

			rand.Seed(time.Now().UnixNano())
			if errorRateInt >= rand.Intn(101) {
				c.JSON(503, gin.H{
					"metadata": map[string]interface{}{
						"hostname": hostname,
						"version":  version,
					},
					"error": "randomly generated error",
				})
				return
			}
		}

		// Pick the next bird in the list.
		idx := atomic.AddInt64(&birdCallCount, 1)
		birdData := birdList[idx%int64(len(birdList))]
		birdResponse := birdResponse{
			Name:     birdData.Title,
			ImageURL: birdData.Thumbnail.Source,
			Extract:  birdData.ExtractHTML,
		}

		c.JSON(200, gin.H{
			"metadata": map[string]interface{}{
				"hostname": hostname,
				"version":  version,
			},
			"response": birdResponse,
		})
	})

	// GET /healthz.
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	// Start the server.
	logf("Starting server listen_addr=%q", bindAddr)
	r.Run(bindAddr)
}

// logF is a helper function that operates like log.Printf but always
// adds a newline.
func logf(format string, v ...interface{}) {
	format = strings.TrimSuffix(format, "\n")
	log.Printf(format+"\n", v...)
}
