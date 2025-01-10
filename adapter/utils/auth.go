package utils

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"

	"errors"

	"github.com/goccy/go-json"
	"github.com/urfave/cli/v2"
)

const (
	AuthenticatorMetadataKey string = "authenticator"
)

// KeyedRequest is a request that contains a key used for authentication
type KeyedRequest struct {
	Key string `json:"key"`
}

// Returns the key for the request
func (r *KeyedRequest) GetKey() string {
	return r.Key
}

// Interface for requests that contain a key
type IKeyedRequest interface {
	GetKey() string
}

// Simple authenticator used for processing incoming requests
type Authenticator struct {
	key string
}

// Creates a new Authenticator instance
func NewAuthenticator(c *cli.Context) (*Authenticator, error) {
	keyFile := c.String(KeyFileFlag.Name)
	if keyFile == "" {
		return nil, fmt.Errorf("secret key file is required")
	}

	// Make sure the file exists
	_, err := os.Stat(keyFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("key file [%s] does not exist", keyFile)
		}
		return nil, fmt.Errorf("error checking key file [%s]: %w", keyFile, err)
	}

	// Read the key file
	key, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("error reading key file [%s]: %w", keyFile, err)
	}

	return &Authenticator{
		key: string(key),
	}, nil
}

// Authenticate checks if the provided key matches the stored key
func (a *Authenticator) Authenticate(key string) error {
	if key != a.key {
		return errors.New("invalid key")
	}
	return nil
}

// Handles an incoming keyed request by reading the input, parsing it, and authenticating it
func HandleKeyedRequest[RequestType IKeyedRequest](c *cli.Context) (RequestType, error) {
	var data RequestType

	// Read the input
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return data, fmt.Errorf("error reading input: %w", err)
	}

	// Parse the input
	err = json.Unmarshal([]byte(input), &data)
	if err != nil {
		return data, fmt.Errorf("error parsing input: %w", err)
	}

	// Authenticate the request
	authenticator := c.App.Metadata[AuthenticatorMetadataKey].(*Authenticator)
	err = authenticator.Authenticate(data.GetKey())
	if err != nil {
		return data, fmt.Errorf("error authenticating request: %w", err)
	}

	return data, nil
}
