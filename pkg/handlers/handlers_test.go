package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/lmzuccarelli/golang-mirror-worker/pkg/connectors"
	"github.com/microlib/simple"
)

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("Inject (force) readAll test error")
}

func TestHandlers(t *testing.T) {

	logger := &simple.Logger{Level: "trace"}

	t.Run("IsAlive : should pass", func(t *testing.T) {
		var STATUS int = 200
		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v2/sys/info/isalive", nil)
		connectors.NewTestConnectors("", STATUS)
		handler := http.HandlerFunc(IsAlive)
		handler.ServeHTTP(rr, req)

		body, e := io.ReadAll(rr.Body)
		if e != nil {
			t.Fatalf("Should not fail : found error %v", e)
		}
		logger.Trace(fmt.Sprintf("Response %s", string(body)))
		// ignore errors here
		if rr.Code != STATUS {
			t.Errorf(fmt.Sprintf("Handler %s returned with incorrect status code - got (%d) wanted (%d)", "IsAlive", rr.Code, STATUS))
		}
	})

	t.Run("BatchPayloadHandler : should pass", func(t *testing.T) {
		var STATUS int = 200

		requestPayload, err := os.ReadFile("../../tests/payload.json")
		if err != nil {
			panic(err)
		}
		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/batch", bytes.NewBuffer([]byte(requestPayload)))
		conn := connectors.NewTestConnectors("", STATUS)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			BatchPayloadHandler(w, r, conn)
		})

		handler.ServeHTTP(rr, req)

		body, e := io.ReadAll(rr.Body)
		if e != nil {
			t.Fatalf("Should not fail : found error %v", e)
		}
		logger.Trace(fmt.Sprintf("Response %s", string(body)))
		// ignore errors here
		if rr.Code != STATUS {
			t.Errorf(fmt.Sprintf("Handler %s returned with incorrect status code - got (%d) wanted (%d)", "SendPayloadHandler", rr.Code, STATUS))
		}
	})

	t.Run("BatchPayloadHandler : should fail (nil body)", func(t *testing.T) {
		var STATUS int = 500
		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/batch", nil)
		conn := connectors.NewTestConnectors("", STATUS)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			BatchPayloadHandler(w, r, conn)
		})

		handler.ServeHTTP(rr, req)

		body, e := io.ReadAll(rr.Body)
		if e != nil {
			t.Fatalf("Should not fail : found error %v", e)
		}
		logger.Trace(fmt.Sprintf("Response %s", string(body)))
		// ignore errors here
		if rr.Code != STATUS {
			t.Errorf(fmt.Sprintf("Handler %s returned with incorrect status code - got (%d) wanted (%d)", "SendPayloadHandler", rr.Code, STATUS))
		}
	})

	t.Run("BatchPayloadHandler : should fail (force read error)", func(t *testing.T) {
		var STATUS int = 403
		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/batch", errReader(0))
		conn := connectors.NewTestConnectors("", STATUS)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			BatchPayloadHandler(w, r, conn)
		})

		handler.ServeHTTP(rr, req)

		body, e := io.ReadAll(rr.Body)
		if e != nil {
			t.Fatalf("Should not fail : found error %v", e)
		}
		logger.Trace(fmt.Sprintf("Response %s", string(body)))
		// ignore errors here
		if rr.Code != STATUS {
			t.Errorf(fmt.Sprintf("Handler %s returned with incorrect status code - got (%d) wanted (%d)", "SendPayloadHandler", rr.Code, STATUS))
		}
	})

	t.Run("BatchPayloadHandler : should fail (forced http error)", func(t *testing.T) {
		var STATUS int = 500
		requestPayload := `{"email":"abc.xyz.com", "event":"test-this","subject": "test", "title": "Test title", "text" : "Test this out" , "jwttoken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE1OTA3NTY4MjAsInN5c3RlbSI6ImNvbnRhY3QtZm9ybSIsImN1c3RvbWVyTnVtYmVyIjoiMDAwMTE5OTQ0MTYwIiwidXNlciI6ImNkdWZmeUB0ZmQuaWUifQ.fisOWBMqnbzzcNQpqO6Cmu6DEMjroaZYgTsAeEmR36A" }`
		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/event/confirmation", bytes.NewBuffer([]byte(requestPayload)))
		conn := connectors.NewTestConnectors("", STATUS)
		conn.Meta("true")
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			BatchPayloadHandler(w, r, conn)
		})

		handler.ServeHTTP(rr, req)

		body, e := io.ReadAll(rr.Body)
		if e != nil {
			t.Fatalf("Should not fail : found error %v", e)
		}
		logger.Trace(fmt.Sprintf("Response %s", string(body)))
		// ignore errors here
		if rr.Code != STATUS {
			t.Errorf(fmt.Sprintf("Handler %s returned with incorrect status code - got (%d) wanted (%d)", "SendPayloadHandler", rr.Code, STATUS))
		}
	})

	t.Run("BatchayloadHandler : should fail (not OK response)", func(t *testing.T) {
		var STATUS int = 500
		requestPayload := `{"email":"abc.xyz.com", "event":"test-this","subject": "test", "title": "Test title", "text" : "Test this out" , "jwttoken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE1OTA3NTY4MjAsInN5c3RlbSI6ImNvbnRhY3QtZm9ybSIsImN1c3RvbWVyTnVtYmVyIjoiMDAwMTE5OTQ0MTYwIiwidXNlciI6ImNkdWZmeUB0ZmQuaWUifQ.fisOWBMqnbzzcNQpqO6Cmu6DEMjroaZYgTsAeEmR36A" }`
		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/batch", bytes.NewBuffer([]byte(requestPayload)))
		conn := connectors.NewTestConnectors("", STATUS)
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			BatchPayloadHandler(w, r, conn)
		})

		handler.ServeHTTP(rr, req)

		body, e := io.ReadAll(rr.Body)
		if e != nil {
			t.Fatalf("Should not fail : found error %v", e)
		}
		logger.Trace(fmt.Sprintf("Response %s", string(body)))
		// ignore errors here
		if rr.Code != STATUS {
			t.Errorf(fmt.Sprintf("Handler %s returned with incorrect status code - got (%d) wanted (%d)", "SendPayloadHandler", rr.Code, STATUS))
		}
	})
}
