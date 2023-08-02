package connectors

import (
	"net/http"

	"github.com/lmzuccarelli/golang-mirror-worker/pkg/api/v1alpha3"
)

// Client Interface - used as a receiver and can be overridden for testing
type Clients interface {
	Debug(msg string, val ...interface{})
	Info(msg string, val ...interface{})
	Error(msg string, val ...interface{})
	Worker(images []v1alpha3.CopyImageSchema) error
	Meta(string) string
	Do(req *http.Request) (*http.Response, error)
}
