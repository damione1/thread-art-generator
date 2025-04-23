package interceptors

import (
	"context"
	"net/http"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/rs/zerolog/log"
)

// ConnectLogger creates a Connect middleware for logging requests
func ConnectLogger() connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			startTime := time.Now()

			// Get the procedure name from the request
			procedure := req.Spec().Procedure

			result, err := next(ctx, req)
			duration := time.Since(startTime)

			// Determine status code based on error
			statusCode := http.StatusOK
			if connectErr, ok := err.(*connect.Error); ok {
				// Map connect error code to HTTP status
				switch connectErr.Code() {
				case connect.CodeInvalidArgument, connect.CodeOutOfRange, connect.CodeFailedPrecondition:
					statusCode = http.StatusBadRequest
				case connect.CodeUnauthenticated:
					statusCode = http.StatusUnauthorized
				case connect.CodePermissionDenied:
					statusCode = http.StatusForbidden
				case connect.CodeNotFound:
					statusCode = http.StatusNotFound
				case connect.CodeAborted, connect.CodeAlreadyExists:
					statusCode = http.StatusConflict
				case connect.CodeResourceExhausted:
					statusCode = http.StatusTooManyRequests
				case connect.CodeCanceled:
					statusCode = 499 // Client Closed Request
				case connect.CodeUnknown, connect.CodeInternal, connect.CodeDataLoss:
					statusCode = http.StatusInternalServerError
				case connect.CodeUnimplemented:
					statusCode = http.StatusNotImplemented
				case connect.CodeUnavailable:
					statusCode = http.StatusServiceUnavailable
				case connect.CodeDeadlineExceeded:
					statusCode = http.StatusGatewayTimeout
				}
			} else if err != nil {
				statusCode = http.StatusInternalServerError
			}

			logger := log.Info()
			if err != nil {
				logger = log.Error().Err(err)
			}

			logger.
				Str("protocol", "connect").
				Str("method", procedure).
				Int("status_code", statusCode).
				Str("status_text", http.StatusText(statusCode)).
				Dur("duration", duration).
				Msg("received a Connect request")

			return result, err
		}
	}
	return interceptor
}

type ResponseRecorder struct {
	http.ResponseWriter
	StatusCode int
	Body       []byte
}

func (rec *ResponseRecorder) WriteHeader(statusCode int) {
	rec.StatusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

func (rec *ResponseRecorder) Write(body []byte) (int, error) {
	rec.Body = body
	return rec.ResponseWriter.Write(body)
}

func HttpLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		startTime := time.Now()
		rec := &ResponseRecorder{
			ResponseWriter: res,
			StatusCode:     http.StatusOK,
		}
		handler.ServeHTTP(rec, req)
		duration := time.Since(startTime)

		logger := log.Info()
		if rec.StatusCode != http.StatusOK {
			logger = log.Error().Bytes("body", rec.Body)
		}

		logger.Str("protocol", "http").
			Str("method", req.Method).
			Str("path", req.RequestURI).
			Int("status_code", rec.StatusCode).
			Str("status_text", http.StatusText(rec.StatusCode)).
			Dur("duration", duration).
			Msg("received a HTTP request")
	})
}
