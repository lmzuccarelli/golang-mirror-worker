package connectors

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/lmzuccarelli/golang-mirror-worker/pkg/api/v1alpha3"
	clog "github.com/lmzuccarelli/golang-mirror-worker/pkg/log"
)

// Mock all connections
type MockConnectors struct {
	Http   *http.Client
	Logger clog.PluggableLoggerInterface
	Flag   string
}

func (c *MockConnectors) Debug(msg string, val ...interface{}) {
	c.Logger.Debug(msg, val...)
}

func (c *MockConnectors) Info(msg string, val ...interface{}) {
	c.Logger.Info(msg, val...)
}

func (c *MockConnectors) Error(msg string, val ...interface{}) {
	c.Logger.Error(msg, val...)
}

func (c *MockConnectors) Meta(flag string) string {
	c.Flag = flag
	return flag
}

func (c *MockConnectors) Do(req *http.Request) (*http.Response, error) {
	if c.Flag == "true" {
		return nil, errors.New("forced http error")
	}
	return c.Http.Do(req)
}

func (c *MockConnectors) Worker(images []v1alpha3.CopyImageSchema) error {
	return nil
}

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewHttpTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func NewTestConnectors(file string, code int) Clients {

	// we first load the json payload to simulate a call to middleware
	// for now just ignore failures.
	var data []byte
	var err error
	if len(file) > 0 {
		data, err = os.ReadFile(file)
		if err != nil {
			fmt.Printf("file data %v\n", err)
			panic(err)
		}
	} else {
		data = []byte("")
	}
	httpclient := NewHttpTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: code,
			// Send response to be tested

			Body: io.NopCloser(bytes.NewBufferString(string(data))),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	log := clog.New("trace")

	conns := &MockConnectors{Http: httpclient, Logger: log, Flag: "false"}
	return conns
}
