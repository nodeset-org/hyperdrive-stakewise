package relay

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/goccy/go-json"
	"github.com/rocket-pool/node-manager-core/log"
)

// Handles an error
func HandleError(w http.ResponseWriter, logger *slog.Logger, code int, err error) {
	msg := err.Error()
	bytes, err := formatError(msg)
	if err != nil {
		logger.Error("Error serializing error response", "error", err)
		writeResponse(w, logger, http.StatusInternalServerError, []byte(`{"error": "error serializing response"}`))
		return
	}
	writeResponse(w, logger, code, bytes)
}

// Handles a success
func HandleSuccess(w http.ResponseWriter, logger *slog.Logger, data any) {
	bytes, err := json.Marshal(data)
	if err != nil {
		logger.Error("Error serializing success response", "error", err)
		HandleError(w, logger, http.StatusInternalServerError, fmt.Errorf("error serializing response: %w", err))
		return
	}
	writeResponse(w, logger, http.StatusOK, bytes)
}

// Writes a response to an HTTP request back to the client and logs it
func writeResponse(w http.ResponseWriter, logger *slog.Logger, statusCode int, message []byte) {
	// Prep the log attributes
	codeMsg := fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode))
	attrs := []any{
		slog.String(log.CodeKey, codeMsg),
	}

	// Log the response
	logMsg := "Responded with:"
	switch statusCode {
	case http.StatusOK:
		logger.Info(logMsg, attrs...)
	case http.StatusInternalServerError:
		logger.Error(logMsg, attrs...)
	default:
		logger.Warn(logMsg, attrs...)
	}

	// Write it to the client
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, writeErr := w.Write(message)
	if writeErr != nil {
		logger.Error("Error writing response", "error", writeErr)
	}
}

// An error message in JSON format
type errorMessage struct {
	Error string `json:"error"`
}

// JSONifies an error for responding to requests
func formatError(message string) ([]byte, error) {
	msg := errorMessage{
		Error: message,
	}
	return json.Marshal(msg)
}
