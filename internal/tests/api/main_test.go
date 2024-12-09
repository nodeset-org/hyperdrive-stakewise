package api_test

import (
	"fmt"
	"log/slog"
	"os"
	"testing"

	swtesting "github.com/nodeset-org/hyperdrive-stakewise/testing"
	"github.com/rocket-pool/node-manager-core/log"
)

// Various singleton variables used for testing
var (
	testMgr *swtesting.StakeWiseTestManager = nil
	logger  *slog.Logger                    = nil
)

// Initialize a common server used by all tests
func TestMain(m *testing.M) {
	// Create a new test manager
	var err error
	testMgr, err = swtesting.NewStakeWiseTestManager()
	if err != nil {
		fail("error creating test manager: %v", err)
	}
	logger = testMgr.GetLogger()

	// Run tests
	code := m.Run()

	// Clean up and exit
	cleanup()
	os.Exit(code)
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
	cleanup()
	os.Exit(1)
}

func cleanup() {
	if testMgr == nil {
		return
	}
	err := testMgr.Close()
	if err != nil {
		logger.Error("Error closing test manager", log.Err(err))
	}
	testMgr = nil
}
