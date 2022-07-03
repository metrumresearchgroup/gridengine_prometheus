package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	go_kit_log "github.com/go-kit/log"
	"github.com/metrumresearchgroup/gridengine_prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/exporter-toolkit/web"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	ServiceName string = "gridengine_prometheus"
	viperSGEKey string = "sge."
)

var entropy rand.Source
var random *rand.Rand
var logger go_kit_log.Logger // go-kit logger is used only by exporter-toolkit
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
	promlogConfig := &promlog.Config{}
	logger = promlog.New(promlogConfig)

	if viper.GetBool("test") {
		//set the underlying gogridengine variable
		err := os.Setenv("GOGRIDENGINE_TEST","true")

		if err != nil {
			log.Fatalf("Attempting to set Gogridengine test variables failed: %s", err)
		}
	}

	if len(viper.GetString("config")) > 0 {
		err := readProvidedConfig(viper.GetString("config"))
		if err != nil {
			log.Fatalf("Attempting to open config file %s failed with error %s", viper.GetString("config"),err)
		}
	}

	if viper.GetBool("debug"){
		config := Config{}
		viper.Unmarshal(&config)
		log.Info(config)
		viper.Debug()
	}

	//Die if we don't have all the SGE configurations required.
	err := validateSGE()
	if err != nil {
		log.Fatal(err)
	}

	//Set the SGE Envs for the application
	err = setSGEEnvironmentVariables()
	if err != nil {
		log.Fatalf("Unable to set SGE environment variables. Details: %s",err)
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
	listenAddress := fmt.Sprintf(":%d", viper.GetInt("port"))
	server := &http.Server{Addr: listenAddress}
	log.Infof("Getting ready to start exporter on port %d", viper.GetInt("port"))

	log.Fatal(web.ListenAndServe(server, viper.GetString("tls_config"), logger), nil)
}

func init(){
	pidFileIdentifier := "pidfile"
	RootCmd.PersistentFlags().String(pidFileIdentifier,"/var/run/" + ServiceName,"Location in which to store a pidfile. Most useful for SystemV daemons")
	viper.BindPFlag(pidFileIdentifier, RootCmd.PersistentFlags().Lookup(pidFileIdentifier))

	listenPortIdentifier := "port"
	RootCmd.PersistentFlags().Int(listenPortIdentifier, 9081, "The port on which the collector should listen")
	viper.BindPFlag(listenPortIdentifier, RootCmd.PersistentFlags().Lookup(listenPortIdentifier))

	tlsConfigFile := "tls_config"
	RootCmd.PersistentFlags().String(tlsConfigFile, "", "TLS config with paths to crt and pey files to be used by https")
	viper.BindPFlag(tlsConfigFile, RootCmd.PersistentFlags().Lookup(tlsConfigFile))

	testModeIdentifier := "test"
	RootCmd.PersistentFlags().Bool(testModeIdentifier,false,"Indicates whether the underlying gogridengine should be run in test mode")
	viper.BindPFlag(testModeIdentifier,RootCmd.PersistentFlags().Lookup(testModeIdentifier))

	configFileIdentifier := "config"
	RootCmd.PersistentFlags().String(configFileIdentifier,"", "Specifies a viper config to load. Should be in yaml format")
	viper.BindPFlag(configFileIdentifier,RootCmd.PersistentFlags().Lookup(configFileIdentifier))

	debugIdentifier := "debug"
	RootCmd.PersistentFlags().Bool(debugIdentifier,false,"Whether or not debug is on")
	viper.BindPFlag(debugIdentifier, RootCmd.PersistentFlags().Lookup(debugIdentifier))

	//SGE Configurations
	sgeArchIdentifier := "sge_arch"
	RootCmd.PersistentFlags().String(sgeArchIdentifier,"lx-amd64","Identifies the architecture of the Sun Grid Engine")
	viper.BindPFlag(viperSGEKey + "arch",RootCmd.PersistentFlags().Lookup(sgeArchIdentifier))

	sgeCellIdentifier := "sge_cell"
	RootCmd.PersistentFlags().String(sgeCellIdentifier,"default","The SGE Cell to use")
	viper.BindPFlag(viperSGEKey + "cell", RootCmd.PersistentFlags().Lookup(sgeCellIdentifier))

	sgeExecDPortIdentifier := "sge_execd_port"
	RootCmd.PersistentFlags().Int(sgeExecDPortIdentifier,6445,"Port for the execution daemon in the grid engine")
	viper.BindPFlag(viperSGEKey + "execd_port", RootCmd.PersistentFlags().Lookup(sgeExecDPortIdentifier))

	sgeQmasterPortIdentifier := "sge_qmaster_port"
	RootCmd.PersistentFlags().Int(sgeQmasterPortIdentifier,6445,"Port for the master scheduling daemon in the grid engine")
	viper.BindPFlag(viperSGEKey + "qmaster_port", RootCmd.PersistentFlags().Lookup(sgeQmasterPortIdentifier))

	sgeRootIdentifier := "sge_root"
	RootCmd.PersistentFlags().String(sgeRootIdentifier,"/opt/sge", "The root location for SGE bianries")
	viper.BindPFlag(viperSGEKey + "root", RootCmd.PersistentFlags().Lookup(sgeRootIdentifier))

	sgeClusterNameIdentifier := "sge_cluster_name"
	RootCmd.PersistentFlags().String(sgeClusterNameIdentifier,"p6444","Name of the SGE Cluster to bind to")
	viper.BindPFlag(viperSGEKey + "cluster_name",RootCmd.PersistentFlags().Lookup(sgeClusterNameIdentifier))
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

//TODO Test
func readProvidedConfig(path string) error {
	viper.SetConfigType("yaml")

	//Read file to get reader
	file, err := os.Open(viper.GetString("config"))

	if err != nil {
		return err
	}

	return viper.ReadConfig(file)
}

func validateSGE() error {

	if len(viper.GetString("sge.arch")) == 0{
		return errors.New("the SGE architecture has not been provided")
	}

	if len(viper.GetString("sge.cell")) == 0 {
		return errors.New("no valid SGE cell has been configured")
	}

	if viper.GetInt("sge.execd_port") == 0 {
		return errors.New("no ExecD port has been specified for SGE binding")
	}

	if viper.GetInt("sge.qmaster_port") == 0 {
		return errors.New("no Qmaster port has been specified for SGE Binding")
	}

	if len(viper.GetString("sge.cluster_name")) == 0 {
		return errors.New("no SGE cluster name has been provided")
	}

	return nil
}

func setSGEEnvironmentVariables() error {
	err := os.Setenv("SGE_ARCH",viper.GetString("sge.arch"))
	if err != nil {
		return err
	}

	err = os.Setenv("SGE_CELL", viper.GetString("sge.cell"))

	if err != nil {
		return err
	}

	err = os.Setenv("SGE_EXECD_PORT",string(rune(viper.GetInt("sge.execd_port"))))

	if err != nil {
		return err
	}

	err = os.Setenv("SGE_QMASTER_PORT",string(rune(viper.GetInt("sge.qmaster_port"))))

	if err != nil {
		return err
	}

	err = os.Setenv("SGE_ROOT", viper.GetString("sge.root"))

	if err != nil {
		return err
	}

	//Update Path to include SGE_ROOT BIN and any dirs matching  arch path
	path := os.Getenv("PATH")
	binPath := filepath.Join(viper.GetString("sge.root"),"bin")
	archPath := filepath.Join(binPath,viper.GetString("sge.arch"))
	err = os.Setenv("PATH", path + ":" + binPath + ":" + archPath)

	if err != nil {
		return err
	}

	err = os.Setenv("SGE_CLUSTER_NAME", viper.GetString("sge.cluster_name"))

	if err != nil {
		return err
	}

	return nil
}


type Config struct {
	Test bool `yaml:"test" json:"test"`
	Port int `yaml:"port" josn:"port"`
	Pidfile string `yaml:"pidfile" json:"pidfile"`
	SGE SGE `mapstructure:"sge"`

}

type SGE struct {
	Arch string `yaml:"arch" json:"arch"`
	Cell string `yaml:"cell" json:"cell"`
	ExecdPort int `yaml:"execd_port" json:"execd_port" mapstructure:"execd_port"`
	QmasterPort int `yaml:"qmaster_port" json:"qmaster_port" mapstructure:"qmaster_port"`
	Root string `yaml:"root" json:"root"`
	ClusterName string `yaml:"cluster_name" json:"cluster_name" mapstructure:"cluster_name"`
}