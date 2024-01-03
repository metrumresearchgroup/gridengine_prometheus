package cmd

import (
	"errors"
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
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

const (
	ServiceName string = "gridengine_prometheus"
	viperSGEKey string = "sge."
)

var entropy rand.Source
var random *rand.Rand

var RootCmd = &cobra.Command{
	Use:     "gridengine_prometheus",
	Short:   "Start the exporter",
	Long:    "Start the prometheus exporter and listen for requests",
	Example: `gridengine_prometheus --pidfile /var/run/gridengine_prometheus.pid --port 9018`,
	RunE:    Start,
}

func Start(cmd *cobra.Command, args []string) error {

	entropy = rand.NewSource(time.Now().UnixNano())
	random = rand.New(entropy)

	if viper.GetBool("test") {
		//set the underlying gogridengine variable
		err := os.Setenv("GOGRIDENGINE_TEST", "true")

		if err != nil {
			log.Fatalf("Attempting to set Gogridengine test variables failed: %s", err)
		}
	}

	if len(viper.GetString("config")) > 0 {
		err := readProvidedConfig(viper.GetString("config"))
		if err != nil {
			log.Fatalf("Attempting to open config file %s failed with error %s", viper.GetString("config"), err)
		}
	}

	var config Config

	if err := viper.Unmarshal(&config); err != nil {
		return fmt.Errorf("failed to retrieve viper details: %w", err)
	}

	if config.Debug {
		viper.Debug()
	}

	//Die if we don't have all the SGE configurations required.
	err := validateSGE(config)
	if err != nil {
		return fmt.Errorf("failed to validate SGE configuration: %w", err)
	}

	//Set the SGE Envs for the application
	err = setSGEEnvironmentVariables(config)
	if err != nil {
		log.Fatalf("Unable to set SGE environment variables. Details: %s", err)
	}

	if len(config.Pidfile) > 0 {
		err = writePidFile(viper.GetString("pidfile"))
		if err != nil {
			log.Error("Unable to setup PID. Continuing without a PID File. Failure caused by: %w", err.Error())
		}
	}

	sge := gridengine_prometheus.NewGridEngine()
	prometheus.MustRegister(sge)

	http.Handle("/metrics", promhttp.Handler())

	log.Infof("Getting ready to start exporter on port %d", viper.GetInt("port"))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", viper.GetInt("port")), nil))

	return nil
}

func init() {
	pidFileIdentifier := "pidfile"
	RootCmd.PersistentFlags().String(pidFileIdentifier, "/var/run/"+ServiceName, "Location in which to store a pidfile. Most useful for SystemV daemons")
	RootCmd.PersistentFlags().Int("port", 9081, "The port on which the collector should listen")
	RootCmd.PersistentFlags().Bool("test", false, "Indicates whether the underlying gogridengine should be run in test mode")
	RootCmd.PersistentFlags().String("config", "", "Specifies a viper config to load. Should be in yaml format")
	RootCmd.PersistentFlags().Bool("debug", false, "Whether or not debug is on")

	//SGE Configurations
	RootCmd.PersistentFlags().String("sge_arch", "lx-amd64", "Identifies the architecture of the Sun Grid Engine")
	RootCmd.PersistentFlags().String("sge_cell", "default", "The SGE Cell to use")
	RootCmd.PersistentFlags().Int("sge_execd_port", 6445, "Port for the execution daemon in the grid engine")
	RootCmd.PersistentFlags().Int("sge_qmaster_port", 6445, "Port for the master scheduling daemon in the grid engine")
	RootCmd.PersistentFlags().String("sge_root", "/opt/sge", "The root location for SGE bianries")
	RootCmd.PersistentFlags().String("sge_cluster_name", "p6444", "Name of the SGE Cluster to bind to")

	_ = viper.BindPFlags(RootCmd.PersistentFlags())
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

func readProvidedConfig(path string) error {
	viper.SetConfigType("yaml")

	//Read file to get reader
	file, err := os.Open(viper.GetString("config"))

	if err != nil {
		return err
	}

	return viper.ReadConfig(file)
}

func validateSGE(config Config) error {

	if len(config.SGE.Arch) == 0 {
		return errors.New("the SGE architecture has not been provided")
	}

	if len(config.SGE.Cell) == 0 {
		return errors.New("no valid SGE cell has been configured")
	}

	if config.SGE.ExecdPort == 0 {
		return errors.New("no ExecD port has been specified for SGE binding")
	}

	if config.SGE.QmasterPort == 0 {
		return errors.New("no Qmaster port has been specified for SGE Binding")
	}

	if len(config.SGE.ClusterName) == 0 {
		return errors.New("no SGE cluster name has been provided")
	}

	return nil
}

func setSGEEnvironmentVariables(config Config) error {
	err := os.Setenv("SGE_ARCH", config.SGE.Arch)
	if err != nil {
		return err
	}

	err = os.Setenv("SGE_CELL", config.SGE.Cell)

	if err != nil {
		return err
	}

	err = os.Setenv("SGE_EXECD_PORT", strconv.Itoa(config.SGE.ExecdPort))

	if err != nil {
		return err
	}

	err = os.Setenv("SGE_QMASTER_PORT", strconv.Itoa(config.SGE.QmasterPort))

	if err != nil {
		return err
	}

	err = os.Setenv("SGE_ROOT", config.SGE.Root)

	if err != nil {
		return err
	}

	//Update Path to include SGE_ROOT BIN and any dirs matching  arch path
	path := os.Getenv("PATH")
	binPath := filepath.Join(config.SGE.Root, "bin")
	archPath := filepath.Join(binPath, config.SGE.Arch)
	err = os.Setenv("PATH", path+":"+binPath+":"+archPath)

	if err != nil {
		return err
	}

	err = os.Setenv("SGE_CLUSTER_NAME", config.SGE.ClusterName)

	if err != nil {
		return err
	}

	return nil
}

type Config struct {
	Test    bool   `yaml:"test" json:"test"`
	Port    int    `yaml:"port" josn:"port"`
	Pidfile string `yaml:"pidfile" json:"pidfile"`
	SGE     SGE    `mapstructure:"sge"`
	Debug   bool   `mapstructure:"debug" yaml:"debug"`
}

type SGE struct {
	Arch        string `yaml:"arch" json:"arch"`
	Cell        string `yaml:"cell" json:"cell"`
	ExecdPort   int    `yaml:"execd_port" json:"execd_port" mapstructure:"execd_port"`
	QmasterPort int    `yaml:"qmaster_port" json:"qmaster_port" mapstructure:"qmaster_port"`
	Root        string `yaml:"root" json:"root"`
	ClusterName string `yaml:"cluster_name" json:"cluster_name" mapstructure:"cluster_name"`
}
