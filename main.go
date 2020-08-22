package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"go.uber.org/ratelimit"
)

var fileLogDir string
var fileLogName string
var mapFiles = make(map[string]*os.File)
var opts *options
var countTotal int

func main() {
	initLogger()
	app := newApp()

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func newApp() *cli.App {
	app := cli.NewApp()
	app.Name = "Gomparator"
	app.Usage = "Compares API responses by status code and response body"
	app.Version = "1.9.2"

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "path",
			Usage: "specifies the file from which to read targets. It should contain one column only with a rel path. eg: /v1/cards?query=123",
		},
		&cli.StringSliceFlag{
			Name:  "host",
			Usage: "targeted hosts. Exactly 2 must be specified. eg: --host 'http://host1.com --host 'http://host2.com'",
		},
		&cli.StringSliceFlag{
			Name:    "header",
			Aliases: []string{"H"},
			Usage:   "headers to be used in the http call",
		},
		&cli.IntFlag{
			Name:    "ratelimit",
			Aliases: []string{"r"},
			Value:   67,
			Usage:   "operation rate limit per second",
		},
		&cli.IntFlag{
			Name:    "workers",
			Aliases: []string{"w"},
			Value:   1000,
			Usage:   "number of workers running concurrently",
		},
		&cli.BoolFlag{
			Name:  "status-code-only",
			Usage: "whether or not it only compares status code ignoring response body",
		},
		&cli.DurationFlag{
			Name:  "timeout",
			Value: DefaultTimeout,
			Usage: "request timeout",
		},
		&cli.DurationFlag{
			Name:    "duration",
			Aliases: []string{"d"},
			Value:   0,
			Usage:   "duration of the comparison [0 = forever]",
		},
		&cli.StringFlag{
			Name:  "exclude",
			Usage: "excludes a value from both json for the specified path. A path is a series of keys separated by a dot or #",
			Value: "results.#.payer_costs.#.payment_method_option_id",
		},
		&cli.StringFlag{
			Name:  "parametersCutting",
			Usage: "check is request contains the params parametrized #",
			Value: "caller_id,display_filtered,differential_pricing_id,bins,public_key",
		},
	}

	app.Action = action
	return app
}

func initLogger() {
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   true,
		FullTimestamp:   true,
	})

	log.SetOutput(os.Stdout)
}

type options struct {
	filePath          string
	hosts             []string
	headers           []string
	timeout           time.Duration
	duration          time.Duration
	workers           int
	rateLimit         int
	statusCodeOnly    bool
	maxBody           int64
	exclude           string
	parametersCutting []string
}

func action(c *cli.Context) error {
	opts = parseFlags(c)
	headers := parseHeaders(opts.headers)

	fetcher := NewHTTPClient(
		Timeout(opts.timeout),
		MaxBody(opts.maxBody))

	ctx, cancel := createContext(opts)
	defer cancel()

	file := openFile(opts)
	defer file.Close()

	logFile := createTmpFile(opts.filePath, headers["X-Caller-Scopes"])
	defer logFile.Close()

	log.Printf("created log temp file in %s", logFile.Name())
	log.SetOutput(logFile)

	lines := getTotalLinesAndSeparateParameterCuts(file, opts.parametersCutting)
	// Once we count the number of lines that will be used as total for the progress bar we reset
	// the pointer to the beginning of the file since it is much faster than closing and reopening
	_, err := file.Seek(0, 0)
	if err != nil {
		return err
	}

	bar := NewProgressBar(lines)
	bar.Start()

	reader := NewReader(file, opts.hosts)
	producer := NewProducer(opts.workers, headers,
		ratelimit.New(opts.rateLimit), fetcher)
	comparator := NewConsumer(opts.statusCodeOnly, bar, log.StandardLogger(), opts.exclude)
	p := New(reader, producer, comparator)

	p.Run(ctx)
	bar.Stop(fileLogName)
	return nil
}

func createContext(opts *options) (context.Context, context.CancelFunc) {
	var ctx context.Context
	var cancel context.CancelFunc
	t := opts.duration
	if t == 0 {
		ctx, cancel = context.WithCancel(context.Background())
	} else {
		// The request has a timeout, so create a context that is
		// canceled automatically when the timeout expires.
		ctx, cancel = context.WithTimeout(context.Background(), t)
	}
	return ctx, cancel
}

func openFile(opts *options) *os.File {
	file, err := os.Open(opts.filePath)
	if err != nil {
		log.Fatal(err)
	}
	return file
}

func createTmpFile(filePath string, callerScope string) *os.File {
	now := time.Now()
	dir, file := filepath.Split(filePath)
	chanelScope := "no-caller-scope"
	if callerScope != "" {
		chanelScope = callerScope
	}

	fileLogDir = dir + now.Format("20060102150405")
	os.Mkdir(fileLogDir, os.FileMode(int(0777)))

	fileLogDir = fileLogDir + "/" + chanelScope
	os.Mkdir(fileLogDir, os.FileMode(int(0777)))

	fileLogName = fmt.Sprintf("%s/%s.error", fileLogDir, strings.TrimSuffix(file, filepath.Ext(file)))

	logFileGeneral, err := os.Create(fileLogName)
	if err != nil {
		log.Fatal(err)
	}
	return logFileGeneral
}

func getTotalLinesAndSeparateParameterCuts(reader io.Reader, parametersCutting []string) int {
	scanner := bufio.NewScanner(reader)
	// Set the split function for the scanning operation.
	scanner.Split(bufio.ScanLines)

	for _, parameter := range parametersCutting {
		dir := fileLogDir + "/" + parameter
		createFile(dir, parameter, "src")
		createFile(dir, parameter, "error")
	}

	dir := fileLogDir + "/total.src"
	fileTotal, _ := os.Create(dir)

	// Count the lines.
	count := 0
	for scanner.Scan() {
		line := scanner.Text()
		if line != "request_uri" {
			for _, parameter := range parametersCutting {
				if strings.Contains(line, parameter) {
					file := mapFiles[parameter+"-src"]
					w := bufio.NewWriter(file)
					fmt.Fprintln(w, line)
					w.Flush()
					//mapFiles[parameter] = file
				}
			}
			w := bufio.NewWriter(fileTotal)
			fmt.Fprintln(w, line)
			w.Flush()
			count++
		}
	}
	countTotal = count
	return count
}

func createFile(dir string, parameter string, extension string) {
	dir = dir + "." + extension
	file, err := os.Create(dir)
	if err == nil {
		mapFiles[parameter+"-"+extension] = file
	}
}

func parseFlags(c *cli.Context) *options {
	opts := &options{}

	if opts.hosts = c.StringSlice("host"); len(opts.hosts) != 2 {
		log.Fatal("invalid number of hosts provided")
	}

	opts.filePath = c.String("path")
	opts.headers = c.StringSlice("header")
	opts.timeout = c.Duration("timeout")
	opts.duration = c.Duration("duration")
	opts.workers = c.Int("workers")
	opts.rateLimit = c.Int("ratelimit")
	opts.statusCodeOnly = c.Bool("status-code-only")
	opts.parametersCutting = strings.Split(c.String("parametersCutting"), ",")
	if opts.statusCodeOnly {
		opts.maxBody = 0
	} else {
		opts.maxBody = DefaultMaxBody
	}
	opts.exclude = c.String("exclude")
	return opts
}

func parseHeaders(h []string) map[string]string {
	result := make(map[string]string, len(h))

	for _, header := range h {
		if header == "" {
			continue
		}

		h := strings.Split(header, ":")
		if len(h) != 2 {
			log.Fatal("invalid header")
		}

		result[h[0]] = h[1]
	}

	return result
}
