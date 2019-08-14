package config

import (
	"bytes"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"reflect"

	"github.com/spf13/afero"
	"github.com/spf13/viper"

	"github.com/elastic/metricbeat-tests-poc/cli/docker"
	"github.com/elastic/metricbeat-tests-poc/cli/log"
)

// OpWorkspace where the application works
var OpWorkspace string

// Op the tool's configuration, read from tool's workspace
var Op *OpConfig

const fileName = "config.yml"

// servicesDefaults initial service configuration that could be overwritten by
// users on their local configuration. This configuration will be persisted in
// the application directory as initial configuration, in the form of a YAML file
var servicesDefaults = map[string]Service{
	"apache": {
		ExposedPorts: []int{80},
		Image:        "httpd",
		Name:         "apache",
		NetworkAlias: "apache",
		Version:      "2.4",
	},
	"apm-server": {
		BuildBranch:     "master",
		BuildRepository: "elastic/apm-server",
		ExposedPorts:    []int{6060, 8200},
		Image:           "docker.elastic.co/apm/apm-server",
		Name:            "apm-server",
		NetworkAlias:    "apm-server",
		Version:         "7.2.0",
	},
	"elasticsearch": {
		BuildBranch:     "master",
		BuildRepository: "elastic/elasticsearch",
		ExposedPorts:    []int{9200},
		Env: map[string]string{
			"bootstrap.memory_lock":  "true",
			"discovery.type":         "single-node",
			"ES_JAVA_OPTS":           "-Xms512m -Xmx512m",
			"xpack.security.enabled": "true",
		},
		Image:        "docker.elastic.co/elasticsearch/elasticsearch",
		Name:         "elasticsearch",
		NetworkAlias: "elasticsearch",
		Version:      "7.2.0",
	},
	"kafka": {
		ExposedPorts: []int{9092},
		Image:        "wurstmeister/kafka",
		Name:         "kafka",
		NetworkAlias: "kafka",
		Version:      "latest",
	},
	"kibana": {
		BuildBranch:     "master",
		BuildRepository: "elastic/kibana",
		ExposedPorts:    []int{5601},
		Image:           "docker.elastic.co/kibana/kibana",
		Name:            "kibana",
		NetworkAlias:    "kibana",
		Version:         "7.2.0",
	},
	"metricbeat": {
		BuildBranch:     "master",
		BuildRepository: "elastic/beats",
		Image:           "docker.elastic.co/beats/metricbeat",
		Name:            "metricbeat",
		NetworkAlias:    "metricbeat",
		Version:         "7.2.0",
	},
	"mongodb": {
		ExposedPorts: []int{27017},
		Image:        "mongo",
		Name:         "mongodb",
		NetworkAlias: "mongodb",
		Version:      "latest",
	},
	"mysql": {
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "secret",
		},
		ExposedPorts: []int{3306},
		Image:        "mysql",
		Name:         "mysql",
		NetworkAlias: "mysql",
		Version:      "latest",
	},
	"redis": {
		ExposedPorts: []int{6379},
		Image:        "redis",
		Name:         "redis",
		NetworkAlias: "redis",
		Version:      "latest",
	},
}

// Service represents the configuration for a service
type Service struct {
	BindMounts      map[string]string `mapstructure:"BindMounts"`
	BuildBranch     string            `mapstructure:"BuildBranch"`
	BuildRepository string            `mapstructure:"BuildRepository"`
	ContainerName   string            `mapstructure:"ContainerName"`
	Daemon          bool              `mapstructure:"AsDaemon"`
	Env             map[string]string `mapstructure:"Env"`
	ExposedPorts    []int             `mapstructure:"ExposedPorts"`
	Image           string            `mapstructure:"Image"`
	Labels          map[string]string `mapstructure:"Labels"`
	Name            string            `mapstructure:"Name"`
	NetworkAlias    string            `mapstructure:"NetworkAlias"`
	Version         string            `mapstructure:"Version"`
}

// Equals checks than the Service is equals to other service
func (s Service) Equals(o Service) bool {
	return reflect.DeepEqual(s, o)
}

var stacksDefaults = map[string]Stack{
	"apm-server": {
		Name: "APM Server",
	},
	"apm-agents": {
		Name: "APM Agents",
	},
	"observability": {
		Name: "Observability",
	},
}

// Stack represents the configuration for a stack, which is an aggregation of services
type Stack struct {
	Name     string             `mapstructure:"Name"`
	Services map[string]Service `mapstructure:"Services"`
}

// checkInstalledSoftware checks that the required software is present
func checkInstalledSoftware() {
	log.Info("Validating required tools...")
	binaries := []string{
		"docker",
	}

	for _, binary := range binaries {
		which(binary)
	}
}

// Init creates this tool workspace under user's home, in a hidden directory named ".op"
func Init() {
	checkInstalledSoftware()

	InitConfig()

	docker.GetDevNetwork()
}

// InitConfig initialises configuration
func InitConfig() {
	if Op != nil {
		return
	}

	usr, _ := user.Current()

	w := filepath.Join(usr.HomeDir, ".op")

	newConfig(w)

	OpWorkspace = w
}

// OpConfig tool configuration
type OpConfig struct {
	Services map[string]Service `mapstructure:"services"`
	Stacks   map[string]Stack   `mapstructure:"stacks"`
}

// AvailableServices return the services in the configuration file
func AvailableServices() map[string]Service {
	return Op.Services
}

// AvailableStacks return the stacks in the configuration file
func AvailableStacks() map[string]Stack {
	return Op.Stacks
}

// GetServiceConfig configuration of a service
func GetServiceConfig(service string) (Service, bool) {
	return Op.GetServiceConfig(service)
}

// GetServiceConfig configuration of a service
func (c *OpConfig) GetServiceConfig(service string) (Service, bool) {
	srv, exists := c.Services[service]

	return srv, exists
}

// GetStackConfig configuration of a stack
func GetStackConfig(stack string) (Stack, bool) {
	return Op.GetStackConfig(stack)
}

// GetStackConfig configuration of a stack
func (c *OpConfig) GetStackConfig(stack string) (Stack, bool) {
	stck, exists := c.Stacks[stack]

	return stck, exists
}

// newConfig returns a new configuration
func newConfig(workspace string) {
	if Op != nil {
		return
	}

	opConfig, err := readConfig(workspace)
	log.CheckIfErrorMessage(err, "Error when reading config.")

	Op = &opConfig
}

func checkConfigFile(workspace string) {
	found, err := exists(workspace)
	if found && err == nil {
		return
	}
	err = os.MkdirAll(workspace, 0755)
	log.CheckIfErrorMessage(err, "Cannot create workdir for 'op' at "+workspace)
	log.Success("'op' workdir created at " + workspace)

	log.Info("Creating %s with default values in %s.", fileName, workspace)

	configFilePath := filepath.Join(workspace, fileName)

	os.Create(configFilePath)

	v := viper.New()
	v.SetDefault("services", servicesDefaults)
	v.SetDefault("stacks", stacksDefaults)

	v.SetConfigType("yaml")
	v.SetConfigName("config")
	v.AddConfigPath(workspace)

	err = v.WriteConfig()
	log.CheckIfErrorMessage(err, `Cannot save default configuration file at `+configFilePath)
	log.Success("Config file initialised with default values")
}

func checkServices(cfg OpConfig) {
	for k := range servicesDefaults {
		if _, exists := cfg.Services[k]; !exists {
			s := Service{}
			viper.UnmarshalKey("services."+k, &s)
			cfg.Services[k] = s
		}
	}
}

func checkStacks(cfg OpConfig) {
	for k := range stacksDefaults {
		if _, exists := cfg.Stacks[k]; !exists {
			s := Stack{}
			viper.UnmarshalKey("stacks."+k, &s)
			cfg.Stacks[k] = s
		}
	}
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func readConfig(workspace string) (OpConfig, error) {
	checkConfigFile(workspace)

	viper.SetConfigType("yml")
	viper.SetConfigName("config")
	viper.AddConfigPath(workspace)
	viper.ReadInConfig()

	// read user-defined configuration at execution path
	fs := afero.NewOsFs()
	userFile, _ := afero.ReadFile(fs, fileName)
	if userFile != nil {
		viper.MergeConfig(bytes.NewReader(userFile))
	}

	cfg := OpConfig{}
	viper.Unmarshal(&cfg)

	checkServices(cfg)
	checkStacks(cfg)

	return cfg, nil
}

// which checks if software is installed
func which(software string) {
	path, err := exec.LookPath(software)
	log.CheckIfError(err)
	log.Success("%s is present at %s", software, path)
}
