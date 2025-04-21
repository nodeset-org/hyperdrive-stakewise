package testing

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"

	"github.com/goccy/go-json"
	"github.com/nodeset-org/hyperdrive-stakewise/relay"
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
)

const (
	DefaultTotal int = 100
	BatchSize    int = 10
)

var (
	// Code 503 - returned when the server is busy loading private keys or doing lookback scans
	ErrorUnavailable error = fmt.Errorf("server is currently unavailable to handle the request")

	// Code 429 - returned when there's an active request being processed already, since only one can run at a time
	ErrorTooManyRequests error = fmt.Errorf("too many requests, please try again later")
)

// Mocks the StakeWise operator for stimulating the relay
type OperatorMock struct {
	endpoint   string
	res        *swconfig.MergedResources
	startIndex int
	total      int
}

// Creates a new operator mock
func NewOperatorMock(relayEndpoint string, res *swconfig.MergedResources, startIndex int) (*OperatorMock, error) {
	endpoint, err := url.JoinPath(relayEndpoint, relay.ValidatorsPath)
	if err != nil {
		return nil, fmt.Errorf("error creating relay validators path: %w", err)
	}
	return &OperatorMock{
		endpoint:   endpoint,
		res:        res,
		startIndex: startIndex,
		total:      DefaultTotal,
	}, nil
}

// Set the start index provided in requests
func (m *OperatorMock) SetStartIndex(startIndex int) {
	m.startIndex = startIndex
}

// Set the total number of validators in requests
func (m *OperatorMock) SetTotal(total int) {
	m.total = total
}

// Simulate the relay request to create validators and return the response
func (m *OperatorMock) SubmitValidatorsRequest() (*relay.ValidatorsResponse, error) {
	request := relay.ValidatorsRequest{
		Vault:                m.res.Vault,
		ValidatorsStartIndex: m.startIndex,
		ValidatorsBatchSize:  BatchSize,
		ValidatorsTotal:      m.total,
	}

	// Serialize the request
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error serializing request: %w", err)
	}
	reader := bytes.NewReader(body)

	// Send the request
	response, err := http.Post(m.endpoint, "application/json", reader)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	// Handle the response
	switch response.StatusCode {
	case http.StatusServiceUnavailable:
		return nil, ErrorUnavailable
	case http.StatusTooManyRequests:
		return nil, ErrorTooManyRequests
	case http.StatusOK:
		// Parse the response
		var responseData relay.ValidatorsResponse
		err = json.NewDecoder(response.Body).Decode(&responseData)
		if err != nil {
			return nil, fmt.Errorf("error parsing response: %w", err)
		}
		return &responseData, nil
	default:
		return nil, fmt.Errorf("unexpected response status: %s", response.Status)
	}
}
