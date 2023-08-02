package batch

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/lmzuccarelli/golang-mirror-worker/pkg/api/v1alpha3"
	clog "github.com/lmzuccarelli/golang-mirror-worker/pkg/log"
	"github.com/lmzuccarelli/golang-mirror-worker/pkg/mirror"
)

const (
	BATCH_SIZE int    = 8
	logFile    string = "logs/worker-{batch}.log"
)

type BatchInterface interface {
	Worker(images []v1alpha3.CopyImageSchema) error
}

func New(log clog.PluggableLoggerInterface,
	mirror mirror.MirrorInterface,
	opts mirror.CopyOptions,
) BatchInterface {
	return &Batch{Log: log, Mirror: mirror, Opts: opts}
}

type Batch struct {
	Log    clog.PluggableLoggerInterface
	Mirror mirror.MirrorInterface
	Opts   mirror.CopyOptions
}

type BatchSchema struct {
	Writer     io.Writer
	Items      int
	Count      int
	BatchSize  int
	BatchIndex int
	Remainder  int
}

// Worker - the main batch processor
func (o *Batch) Worker(images []v1alpha3.CopyImageSchema) error {

	var errArray []error
	var wg sync.WaitGroup
	var err error

	var b *BatchSchema
	imgs := len(images)
	if imgs < BATCH_SIZE {
		b = &BatchSchema{Items: imgs, Count: 1, BatchSize: imgs, BatchIndex: 0, Remainder: 0}
	} else {
		b = &BatchSchema{Items: imgs, Count: (imgs / BATCH_SIZE), BatchSize: BATCH_SIZE, Remainder: (imgs % BATCH_SIZE)}
	}

	o.Log.Info("images to mirror %d ", b.Items)
	o.Log.Info("batch count %d ", b.Count)
	o.Log.Info("batch concurrency %d ", BATCH_SIZE)
	o.Log.Info("batch size %d ", b.BatchSize)
	o.Log.Info("remainder size %d ", b.Remainder)

	f := make([]*os.File, b.Count)
	// prepare batching
	wg.Add(b.BatchSize)
	for i := 0; i < b.Count; i++ {
		// create a log file for each batch
		f[i], err = os.Create(strings.Replace(logFile, "{batch}", strconv.Itoa(i), -1))
		if err != nil {
			o.Log.Error("[Worker] %v", err)
		}
		writer := bufio.NewWriter(f[i])
		o.Log.Info(fmt.Sprintf("starting batch %d ", i))
		for x := 0; x < b.BatchSize; x++ {
			index := (i * b.BatchSize) + x
			o.Log.Debug("source %s ", images[index].Source)
			o.Log.Debug("destination %s ", images[index].Destination)
			go func(src, dest string, opts *mirror.CopyOptions, writer bufio.Writer) {
				defer wg.Done()
				err := o.Mirror.Run(src, dest, "copy", opts, writer)
				if err != nil {
					errArray = append(errArray, err)
				}
			}(images[index].Source, images[index].Destination, &o.Opts, *writer)
		}
		wg.Wait()
		// rather than use defer Close we intentianally close the log files
		for _, f := range f {
			f.Close()
		}
		o.Log.Info("completed batch %d", i)
		if b.Count > 1 {
			wg.Add(BATCH_SIZE)
		}
		if len(errArray) > 0 {
			for _, err := range errArray {
				o.Log.Error("[Worker] errArray %v", err)
			}
			return fmt.Errorf("[Worker] error in batch - refer to console logs")
		}
	}
	if b.Remainder > 0 {
		// one level of simple recursion
		i := b.Count * BATCH_SIZE
		o.Log.Info("executing remainder [batch size of 1]")
		err := o.Worker(images[i:])
		if err != nil {
			return err
		}
		// output the logs to console
		if !o.Opts.Global.Quiet {
			consoleLogFromFile(o.Log, o.Opts.Global.LogDir)
		}
		o.Log.Info("[Worker] successfully completed all batches")
	}
	return nil
}

// consoleLogFromFile
func consoleLogFromFile(log clog.PluggableLoggerInterface, directory string) {
	dir, _ := os.ReadDir(directory)
	for _, f := range dir {
		if strings.Contains(f.Name(), "worker") {
			log.Debug("[batch] %s ", f.Name())
			data, _ := os.ReadFile("logs/" + f.Name())
			lines := strings.Split(string(data), "\n")
			for _, s := range lines {
				if len(s) > 0 {
					// clean the line
					log.Debug("%s ", strings.ToLower(s))
				}
			}
		}
	}
}
