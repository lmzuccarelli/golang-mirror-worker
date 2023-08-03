package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/lmzuccarelli/golang-mirror-worker/pkg/api/v1alpha3"
	"github.com/lmzuccarelli/golang-mirror-worker/pkg/connectors"
)

const (
	CONTENTTYPE     string = "Content-Type"
	APPLICATIONJSON string = "application/json"
)

// BatchPayloadHandler - api function handler that receives images to mirror
func BatchPayloadHandler(w http.ResponseWriter, r *http.Request, conn connectors.Clients) {
	var images []v1alpha3.CopyImageSchema

	addHeaders(w, r)

	// add a semaphore so that we don't execute a new batch while we are still processing
	// we first read to the semaphore

	if _, err := os.Stat("semaphore.txt"); errors.Is(err, os.ErrNotExist) {
		e := os.WriteFile("semaphore.txt", []byte("executing"), 0755)
		if e != nil {
			conn.Error("semaphore.txt write error %v", e)
		}
	} else {
		host, _ := os.Hostname()
		msg := "BatchPayloadHandler status on host %s : busy"
		b := responseFormat(http.StatusOK, w, msg, host)
		fmt.Fprintf(w, "%s", string(b))
		return
	}

	if r.Body == nil {
		r.Body = io.NopCloser(bytes.NewBufferString(""))
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		msg := "BatchPayloadHandler error :  %v"
		b := responseFormat(http.StatusInternalServerError, w, msg, err)
		fmt.Fprintf(w, "%s", string(b))
		return
	}

	// unmarshal result from master backend
	errs := json.Unmarshal(body, &images)
	if errs != nil {
		msg := "BatchPayloadHandler could not unmarshal input data to schema %v"
		b := responseFormat(http.StatusInternalServerError, w, msg, errs)
		fmt.Fprintf(w, "%s", string(b))
		return
	}

	// once we have the payload
	// execute the batch
	host, _ := os.Hostname()
	msg := "BatchPayloadHandler starting image mirroring on host " + host + " - status will be posted when completed"
	response := &v1alpha3.Response{Name: "golang-mirror-worker", StatusCode: "200", Status: "OK", Message: msg}
	w.WriteHeader(http.StatusOK)
	b, _ := json.MarshalIndent(response, "", "	")
	fmt.Fprintf(w, "%s", string(b))

	// execute the worker in a go routine
	go executeWorker(images, conn)

}

func IsAlive(w http.ResponseWriter, r *http.Request) {
	host, _ := os.Hostname()
	fmt.Fprintf(w, "%s", "{ \"version\" : \"v1.0.0.\" , \"name\": \"golang-mirror-worker"+host+"\"}")
}

// headers (with cors) utility
func addHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(CONTENTTYPE, APPLICATIONJSON)
	// use this for cors
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

// responsFormat - utility function
func responseFormat(code int, w http.ResponseWriter, msg string, val ...interface{}) []byte {
	var b []byte
	status := "ERROR"
	if code == http.StatusOK {
		status = "OK"
	}
	response := &v1alpha3.Response{Name: "golang-mirror-worker", StatusCode: strconv.Itoa(code), Status: status, Message: fmt.Sprintf(msg, val...)}
	w.WriteHeader(code)
	b, _ = json.MarshalIndent(response, "", "	")
	return b
}

// makePostRequest - private utility function for POST
func makePostRequest(generic *v1alpha3.GenericSchema, msg string, con connectors.Clients) ([]byte, error) {
	var b []byte
	req, _ := http.NewRequest("POST", generic.Url, bytes.NewBuffer([]byte("{\"id\":\"1\",\"message\":\""+msg+"\"}")))
	resp, err := con.Do(req)
	if err != nil {
		return b, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return []byte("ok"), nil
	}
	return []byte("ko"), fmt.Errorf(strconv.Itoa(resp.StatusCode))
}

func executeWorker(images []v1alpha3.CopyImageSchema, conn connectors.Clients) {
	updatedList := checkImageOnDisk(images, conn)
	e := conn.Worker(updatedList)
	var result string
	host, _ := os.Hostname()
	if e != nil {
		result = "BatchPayloadHandler status on host " + host + " : FAIL"
	} else {
		result = "BatchPayloadHandler status on host " + host + " : PASS"
	}
	// now make the call to get all data
	generic := &v1alpha3.GenericSchema{Url: os.Getenv("CALLBACK_URL")}

	res, err := makePostRequest(generic, result, conn)
	if err != nil {
		conn.Error("BatchPayloadHandler %v", err)
	}
	os.Remove("semaphore.txt")
	conn.Info("BatchPayloadHandler %s", string(res))
}

func checkImageOnDisk(images []v1alpha3.CopyImageSchema, conn connectors.Clients) []v1alpha3.CopyImageSchema {
	var res []v1alpha3.CopyImageSchema
	for index := range images {
		conn.Debug("checking for directory %s", images[index].Destination)
		i := strings.LastIndex(images[index].Destination, "/")
		// ignore dir:// at beginning of the string
		dir := images[index].Destination[5:i]
		if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				conn.Error("creating directory %s", dir)
			}
			res = append(res, images[index])
		} else {
			// check if we aleady have the image on disk
			if _, err := os.Stat(images[index].Destination[5:]); errors.Is(err, os.ErrNotExist) {
				res = append(res, images[index])
			} else {
				conn.Info("directory exists %s", images[index].Destination)
			}
		}
	}
	return res
}
