package validator

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	clog "github.com/lmzuccarelli/golang-mirror-worker/pkg/log"
)

// checkEnvars - private function, iterates through each item and checks the required field
func checkEnvar(item string, log clog.PluggableLoggerInterface) error {
	name := strings.Split(item, ",")[0]
	required, _ := strconv.ParseBool(strings.Split(item, ",")[1])
	log.Trace("Input parameters -> name %s : required %t", name, required)
	if os.Getenv(name) == "" {
		if required {
			log.Error("%s envar is mandatory please set it", name)
			return fmt.Errorf(fmt.Sprintf("%s envar is mandatory please set it", name))
		}
		log.Error("%s envar is empty please set it", name)
	}
	return nil
}

// ValidateEnvars : public call that groups all envar validations
// These envars are set via the openshift template
func ValidateEnvars(log clog.PluggableLoggerInterface) error {
	items := []string{
		"LOG_LEVEL,true",
		"SERVER_PORT,true",
		"CALLBACK_URL,true",
	}
	for x := range items {
		if err := checkEnvar(items[x], log); err != nil {
			return err
		}
	}
	return nil
}
