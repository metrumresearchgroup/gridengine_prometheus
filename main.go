package main

import (
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"time"
)

var isTest bool = false
var listenPort int = 9081

var entropy rand.Source
var random *rand.Rand

func main() {

	//Enable test mode for random data generation
	if os.Getenv("TEST") == "true" {
		isTest = true
	}

	entropy := rand.NewSource(time.Now().UnixNano())
	random := rand.New(entropy)

	log.Debug("Generating with new seed: ", random.Int())

	if os.Getenv("LISTEN_PORT") != "" {
		resconv, err := strconv.ParseInt(os.Getenv("LISTEN_PORT"), 10, 32)
		if err == nil {
			//Only set the listen port separately if it's present and we can parse it into an int
			listenPort = int(resconv)
		}

	}

	pidPtr := flag.String("pidfile", "/var/run/gridengine-exporter/gridengine-exporter.pid", "The PID that will be used to identify the service")
	flag.Parse()

	err := writePidFile(*pidPtr)

	//If we can't write the pid file, we should panic
	if err != nil {
		log.Error("An error occurred writing the pid", err)
		panic(err)
	}

	sge := newGridEngine()
	prometheus.MustRegister(sge)

	http.Handle("/metrics", promhttp.Handler())

	log.Fatal(http.ListenAndServe(":"+string(listenPort), nil))
}

func writePidFile(pidFile string) error {
	// Read in the pid file as a slice of bytes.
	if piddata, err := ioutil.ReadFile(pidFile); err == nil {
		// Convert the file contents to an integer.
		if pid, err := strconv.Atoi(string(piddata)); err == nil {
			// Look for the pid in the process list.
			if process, err := os.FindProcess(pid); err == nil {
				// Send the process a signal zero kill.
				if err := process.Signal(syscall.Signal(0)); err == nil {
					// We only get an error if the pid isn't running, or it's not ours.
					return fmt.Errorf("pid already running: %d", pid)
				}
			}
		}
	}
	// If we get here, then the pidfile didn't exist,
	// or the pid in it doesn't belong to the user running this app.
	return ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0664)
}
