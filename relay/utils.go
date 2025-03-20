package relay

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/log"
)

// Logs the request and returns the query args and path args
func ProcessApiRequest(logger *slog.Logger, w http.ResponseWriter, r *http.Request, requestBody any) (url.Values, map[string]string) {
	args := r.URL.Query()
	logger.Info("New request", slog.String(log.MethodKey, r.Method), slog.String(log.PathKey, r.URL.Path))
	logger.Debug("Request params:", slog.String(log.QueryKey, r.URL.RawQuery))

	if requestBody != nil {
		// Read the body
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			HandleError(w, logger, http.StatusBadRequest, fmt.Errorf("error reading request body: %w", err))
			return nil, nil
		}
		logger.Debug("Request body:", slog.String(log.BodyKey, string(bodyBytes)))

		// Deserialize the body
		err = json.Unmarshal(bodyBytes, &requestBody)
		if err != nil {
			HandleError(w, logger, http.StatusBadRequest, fmt.Errorf("error deserializing request body: %w", err))
			return nil, nil
		}
	}

	return args, mux.Vars(r)
}
