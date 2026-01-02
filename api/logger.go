package api

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GrpcLogger is a gRPC server interceptor for logging request metadata, response status, and execution time.
func GrpcLogger(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (response interface{}, err error) {
	startTime := time.Now()

	result, err := handler(ctx, req)
	duration := time.Since(startTime)

	statusCode := codes.Unknown
	if st, ok := status.FromError(err); ok {
		statusCode = st.Code()
	}

	logger := log.Info()
	if err != nil {
		logger = log.Error().Err(err)
	}

	logger.
		Str("protocol", "grpc").
		Str("method", info.FullMethod).
		Int("status_code", int(statusCode)).
		Str("status_text", statusCode.String()).
		Dur("duration", duration).
		Msg("received a grpc request")

	return result, err
}

type ResponseRecorder struct {
	http.ResponseWriter
	StatusCode int
	Body       []byte
}

// WriteHeader sets the status code for the HTTP response and forwards it to the underlying ResponseWriter.
func (recorder *ResponseRecorder) WriteHeader(statusCode int) {
	recorder.StatusCode = statusCode
	recorder.ResponseWriter.WriteHeader(statusCode)
}

// Write writes the provided byte slice to the response body and forwards it to the underlying ResponseWriter.
func (recorder *ResponseRecorder) Write(data []byte) (int, error) {
	recorder.Body = data
	return recorder.ResponseWriter.Write(data)
}

// HttpLogger wraps an HTTP handler to log request details such as method, path, status code, and response time.
func HttpLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		recorder := &ResponseRecorder{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}
		handler.ServeHTTP(recorder, r)
		duration := time.Since(startTime)

		logger := log.Info()
		if recorder.StatusCode >= 400 {
			logger = log.Error().Bytes("body", recorder.Body)
		}

		logger.
			Str("protocol", "http").
			Str("method", r.Method).
			Str("path", r.RequestURI).
			Int("status_code", recorder.StatusCode).
			Str("status_text", http.StatusText(recorder.StatusCode)).
			Dur("duration", duration).
			Msg("received a http request")
	})

}
