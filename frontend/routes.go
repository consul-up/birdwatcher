package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

// BackendRespWrapper is the response from the backend service.
type BackendRespWrapper struct {
	Metadata struct {
		Hostname string `json:"hostname"`
		Version  string `json:"version"`
	} `json:"metadata"`
	Response BackendResp `json:"response"`
	Error    string      `json:"error"`
}

type BackendResp struct {
	Name     string `json:"name"`
	ImageURL string `json:"imageURL"`
	Extract  string `json:"extract"`
}

type ShuffleResp struct {
	Metadata ShuffleRespMetadata `json:"metadata,omitempty"`
	Error    string              `json:"error,omitempty"`
	Response *BackendResp        `json:"response,omitempty"`
}

type ShuffleRespMetadata struct {
	BackendDuration   string `json:"backendDuration,omitempty"`
	BackendStatusCode int    `json:"backendStatusCode,omitempty"`
	BackendHostname   string `json:"backendHostname,omitempty"`
	BackendVersion    string `json:"backendVersion,omitempty"`
}

func setupRoutes(r *gin.Engine, httpClient *http.Client, backendURL string) {
	if gin.Mode() == gin.DebugMode {
		r.LoadHTMLGlob("templates/*")
	} else {
		LoadHTMLFromEmbedFS(r, templatesFS, "templates/*")
	}
	fsys, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(fmt.Sprintf("Unable to load static files: %s", err))
	}
	r.StaticFS("static", http.FS(fsys))

	// The UI.
	r.GET("/", func(c *gin.Context) {
		otelgin.HTML(c, http.StatusOK, "index.tmpl", nil)
	})

	// Admin panel.
	r.GET("/admin", func(c *gin.Context) {
		otelgin.HTML(c, http.StatusOK, "admin.tmpl", nil)
	})

	// GET /healthz.
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	// Calls backend to get a bird.
	r.GET("/shuffle", func(c *gin.Context) {
		startTime := time.Now()

		// Create backend request.
		req, err := http.NewRequestWithContext(
			c.Request.Context(), "GET",
			fmt.Sprintf("%s/bird", backendURL), nil)
		if err != nil {
			c.JSON(503, gin.H{
				"metadata": map[string]interface{}{},
				"error":    fmt.Sprintf("Unable to construct request: %s", err),
			})
			return
		}

		// Propagate query params from request.
		queryParams := req.URL.Query()
		for k, v := range c.Request.URL.Query() {
			queryParams.Set(k, v[0])
		}
		req.URL.RawQuery = queryParams.Encode()

		// Trace request (if tracing is enabled).
		tracingResponseHandler := httpClientTrace(req, "call_backend")

		// Make the request.
		resp, err := httpClient.Do(req)

		// Record the duration.
		duration := time.Since(startTime)

		if tracingResponseHandler != nil {
			tracingResponseHandler(resp, err)
		}

		// backendErrResponse is a helper func for returning an error due to
		// not getting a proper response from backend.
		backendErrResponse := func(duration time.Duration, err error) {
			log.Printf("Error calling backend: %s\n", err)
			c.JSON(200, ShuffleResp{
				Metadata: ShuffleRespMetadata{
					BackendDuration: roundDuration(duration).String(),
				},
				Error: err.Error(),
			})
		}

		// Check if the request succeeded.
		if err != nil {
			backendErrResponse(duration, fmt.Errorf("unable to call backend: %w", err))
			return
		}

		// Read the response body.
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			backendErrResponse(duration, fmt.Errorf("unable to read backend response body: %w", err))
			return
		}

		// Handle unexpected non-JSON response.
		contentType := resp.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			backendErrResponse(duration, fmt.Errorf("received status code %d from backend: %q", resp.StatusCode, string(body)))
			return
		}

		// Unmarshall response.
		var backendResp BackendRespWrapper
		err = json.Unmarshal(body, &backendResp)
		if err != nil {
			backendErrResponse(duration, fmt.Errorf("json unmarshalling response body: %s", err))
			return
		}

		// Handle JSON response that was not a 200.
		if resp.StatusCode != 200 {
			c.JSON(200, ShuffleResp{
				Metadata: ShuffleRespMetadata{
					BackendDuration:   roundDuration(duration).String(),
					BackendHostname:   backendResp.Metadata.Hostname,
					BackendVersion:    backendResp.Metadata.Version,
					BackendStatusCode: resp.StatusCode,
				},
				Error: fmt.Sprintf("received status code %d from backend: %q", resp.StatusCode, backendResp.Error),
			})
			return
		}

		// Return bird successfully.
		c.JSON(200, ShuffleResp{
			Metadata: ShuffleRespMetadata{
				BackendDuration:   roundDuration(duration).String(),
				BackendHostname:   backendResp.Metadata.Hostname,
				BackendVersion:    backendResp.Metadata.Version,
				BackendStatusCode: resp.StatusCode,
			},
			Response: &backendResp.Response,
		})
	})
}

func roundDuration(d time.Duration) time.Duration {
	if d > 1*time.Millisecond {
		return d.Round(1 * time.Millisecond)
	}
	return d.Round(1 * time.Microsecond)
}

// LoadHTMLFromEmbedFS makes it possible to use go:embed with Gin's templating
// system.
func LoadHTMLFromEmbedFS(engine *gin.Engine, embedFS embed.FS, pattern string) {
	tmpl := template.Must(template.ParseFS(embedFS, pattern))
	engine.SetHTMLTemplate(tmpl)
}
