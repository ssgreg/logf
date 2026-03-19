// Context example: request-scoped fields via context.Context.
//
// Fields attached to context appear in every log entry automatically —
// no need to pass a derived logger through the call stack.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/ssgreg/logf/v2"
)

var logger *logf.Logger

func main() {
	// Setup: sync handler with ContextHandler + automatic field source.
	// The FieldSource function is called on every log entry and can
	// extract fields from context automatically (e.g., trace IDs).
	logger = logf.NewLogger().
		Level(logf.LevelInfo).
		Output(os.Stdout).
		Context(extractUserID). // enables ContextHandler + field source
		Build()

	mux := http.NewServeMux()
	mux.HandleFunc("/orders", handleOrder)

	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", requestIDMiddleware(mux))
}

// requestIDMiddleware shows two ways to add fields to context:
//
//  1. logf.With() — explicitly adds logf fields to the Bag in context.
//     The middleware knows about logf and directly creates fields.
//
//  2. context.WithValue() — stores a plain value in context using a
//     custom key. The middleware does NOT import logf — it just stores
//     its own data. The FieldSource (extractUserID) picks it up later.
//
// Both approaches produce the same result: fields appear in every log
// entry automatically. Use (1) when the middleware is logf-aware, and
// (2) when you want to decouple the middleware from the logger (e.g.,
// auth middleware stores user_id, OTel stores trace context — neither
// needs to know about logf).
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Approach 1: logf-aware — directly add fields to Bag.
		ctx := logf.With(r.Context(),
			logf.String("request_id", r.Header.Get("X-Request-ID")),
			logf.String("method", r.Method),
			logf.String("path", r.URL.Path),
		)

		// Approach 2: logf-unaware — store plain value in context.
		// extractUserID FieldSource converts it to a logf field.
		ctx = context.WithValue(ctx, userIDKey{}, r.Header.Get("X-User-ID"))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func handleOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// request_id, method, path, user_id are included automatically:
	logger.Info(ctx, "handling order")
	// → {"level":"info","msg":"handling order","request_id":"abc","method":"POST","path":"/orders","user_id":"u42"}

	// Add more context deeper in the stack:
	ctx = logf.With(ctx, logf.String("order_id", "ord-456"))

	processPayment(ctx)
	w.WriteHeader(http.StatusOK)
}

func processPayment(ctx context.Context) {
	// All accumulated fields appear without passing a logger:
	logger.Info(ctx, "processing payment", logf.Int("amount", 4999))
	// → {"level":"info","msg":"processing payment","request_id":"abc","method":"POST","path":"/orders","user_id":"u42","order_id":"ord-456","amount":4999}
}

// --- FieldSource: automatic field extraction from context ---

type userIDKey struct{}

// extractUserID is a FieldSource that reads user_id from context.
// It is called on every log entry automatically — no manual field
// passing needed.
func extractUserID(ctx context.Context) []logf.Field {
	if uid, ok := ctx.Value(userIDKey{}).(string); ok {
		return []logf.Field{logf.String("user_id", uid)}
	}
	return nil
}
