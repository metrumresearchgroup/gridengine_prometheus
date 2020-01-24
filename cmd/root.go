package cmd

import (
	"fmt"
	"github.com/metrumresearchgroup/gridengine_prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"time"
)

const(
	ServiceName string = "gridengine_prometheus"
)

var entropy rand.Source
var random *rand.Rand

var RootCmd = &cobra.Command{
	Use:                        "gridengine_prometheus",
	Short:                      "Start the exporter",
	Long:                       "Start the prometheus exporter and listen for requests",
	Example:                    `gridengine_prometheus --pidfile /var/run/gridengine_prometheus.pid --port 9018`,
	Run: Start,
}

func Start( cmd *cobra.Command, args []string){

	entropy = rand.NewSource(time.Now().UnixNano())
	random = rand.New(entropy)

	if viper.GetBool("test") {
		os.Setenv("GOGRIDENGINE_TEST","true")
	}


	if len(viper.GetString("pidfile")) > 0 {
		err := writePidFile(viper.GetString("pidfile"))
		if err != nil {
			log.Error("Unable to setup PID. Continuing without a PID File")
		}
	}

	sge := gridengine_prometheus.NewGridEngine()
	prometheus.MustRegister(sge)

	http.Handle("/metrics", promhttp.Handler())

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d",viper.GetInt("port")), nil))
}

func init(){
	pidFileIdentifier := "pidfile"
	RootCmd.PersistentFlags().String(pidFileIdentifier,"/var/run/" + ServiceName,"Location in which to store a pidfile. Most useful for SystemV daemons")
	viper.BindPFlag(pidFileIdentifier,RootCmd.PersistentFlags().Lookup(pidFileIdentifier))

	listenPortIdentifier := "port"
	RootCmd.PersistentFlags().Int(listenPortIdentifier,9081,"The port on which the collector should listen")
	viper.BindPFlag(listenPortIdentifier,RootCmd.PersistentFlags().Lookup(listenPortIdentifier))

	testModeIdentifier := "test"
	RootCmd.PersistentFlags().Bool(testModeIdentifier,false,"Indicates whether the underlying gogridengine should be run in test mode")
	viper.BindPFlag(testModeIdentifier,RootCmd.PersistentFlags().Lookup(testModeIdentifier))

	configFileIdentifier := "config"
	RootCmd.PersistentFlags().String(configFileIdentifier,"", "Specifies a viper config to load. Should be in yaml format")
	viper.BindPFlag(configFileIdentifier,RootCmd.PersistentFlags().Lookup(configFileIdentifier))
}

func writePidFile(pidFile string) error {
	location := pidFile
	// Read in the pid file as a slice of bytes.
	if piddata, err := ioutil.ReadFile(location); err == nil {
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
	return ioutil.WriteFile(location, []byte(fmt.Sprintf("%d", os.Getpid())), 0664)
}
