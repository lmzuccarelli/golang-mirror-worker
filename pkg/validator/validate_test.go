package validator

import (
	"fmt"
	"os"
	"testing"

	clog "github.com/lmzuccarelli/golang-mirror-worker/pkg/log"
)

func TestEnvars(t *testing.T) {
	logger := clog.New("trace")

	t.Run("ValidateEnvars : should fail", func(t *testing.T) {
		os.Setenv("NONE", "")
		err := ValidateEnvars(logger)
		if err == nil {
			t.Errorf(fmt.Sprintf("Handler %s returned with no error - got (%v) wanted (%v)", "ValidateEnvars", err, nil))
		}
	})

	t.Run("ValidateEnvars : should pass", func(t *testing.T) {
		os.Setenv("LOG_LEVEL", "info")
		os.Setenv("SERVER_PORT", "9000")
		os.Setenv("CALLBACK_URL", "http://test.com")
		err := ValidateEnvars(logger)
		if err != nil {
			t.Errorf(fmt.Sprintf("Handler %s returned with error - got (%v) wanted (%v)", "ValidateEnvars", err, nil))
		}
	})
}
