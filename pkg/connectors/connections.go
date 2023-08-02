package connectors

import (
	"crypto/tls"
	"net/http"

	"github.com/lmzuccarelli/golang-mirror-worker/pkg/api/v1alpha3"
	"github.com/lmzuccarelli/golang-mirror-worker/pkg/batch"
	clog "github.com/lmzuccarelli/golang-mirror-worker/pkg/log"
	"github.com/lmzuccarelli/golang-mirror-worker/pkg/mirror"
)

// Connections struct - all backend connections in a common object
type Connectors struct {
	Http   *http.Client
	Logger clog.PluggableLoggerInterface
	Batch  batch.BatchInterface
}

func NewClientConnections(log clog.PluggableLoggerInterface) Clients {
	// set up http object
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}

	global := &mirror.GlobalOptions{
		TlsVerify:      false,
		InsecurePolicy: true,
	}

	// setup copy options
	_, sharedOpts := mirror.SharedImageFlags()
	_, deprecatedTLSVerifyOpt := mirror.DeprecatedTLSVerifyFlags()
	_, srcOpts := mirror.ImageFlags(global, sharedOpts, deprecatedTLSVerifyOpt, "src-", "screds")
	_, destOpts := mirror.ImageDestFlags(global, sharedOpts, deprecatedTLSVerifyOpt, "dest-", "dcreds")
	_, retryOpts := mirror.RetryFlags()

	opts := mirror.CopyOptions{
		Global:              global,
		DeprecatedTLSVerify: deprecatedTLSVerifyOpt,
		SrcImage:            srcOpts,
		DestImage:           destOpts,
		RetryOpts:           retryOpts,
		Dev:                 false,
	}

	// update all dependant modules
	mc := mirror.NewMirrorCopy()
	md := mirror.NewMirrorDelete()
	m := mirror.New(mc, md)
	batch := batch.New(log, m, opts)

	return &Connectors{Http: httpClient, Logger: log, Batch: batch}
}

func (c *Connectors) Meta(info string) string {
	return info
}

func (c *Connectors) Do(req *http.Request) (*http.Response, error) {
	return c.Http.Do(req)
}

func (c *Connectors) Worker(images []v1alpha3.CopyImageSchema) error {
	return c.Batch.Worker(images)
}

func (c *Connectors) Debug(msg string, val ...interface{}) {
	c.Logger.Debug(msg, val...)
}

func (c *Connectors) Info(msg string, val ...interface{}) {
	c.Logger.Info(msg, val...)
}

func (c *Connectors) Error(msg string, val ...interface{}) {
	c.Logger.Error(msg, val...)
}
