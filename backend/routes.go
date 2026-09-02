package main

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"et-interview/internal/demo"
)

func Routes(w Wires) *gin.Engine {
	r := gin.Default()

	r.GET("/_/health", handleHealth)

	// Throwaway; see internal/demo. Your endpoints go here too.
	r.GET("/api/demo", handleDemo(w.Demo))

	fileServer := http.FileServer(http.Dir(w.WebRoot))
	r.NoRoute(func(ctx *gin.Context) {
		// Don't serve static for API paths that fell through
		if strings.HasPrefix(ctx.Request.URL.Path, "/api/") {
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		ctx.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		fileServer.ServeHTTP(ctx.Writer, ctx.Request)
	})

	return r
}

func handleHealth(ctx *gin.Context) {
	ctx.String(http.StatusOK, "OK")
}

// handleDemo reads the seeded values through the store. The handler's job is
// the request and the response; the query lives in internal/demo.
func handleDemo(store demo.Store) gin.HandlerFunc {
	type demoResponse struct {
		Values []float64 `json:"values"`
	}

	return func(ctx *gin.Context) {
		values, err := store.Values(ctx.Request.Context())
		if err != nil {
			// Log the cause, return something a client can act on.
			slog.Error("failed to read demo values", slog.String("error", err.Error()))
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not read demo values",
			})
			return
		}

		ctx.JSON(http.StatusOK, demoResponse{Values: values})
	}
}
