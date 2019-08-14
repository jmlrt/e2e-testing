package cmd

import (
	"errors"

	"github.com/elastic/metricbeat-tests-poc/cli/config"
	"github.com/elastic/metricbeat-tests-poc/cli/log"
	"github.com/elastic/metricbeat-tests-poc/cli/services"
	"github.com/imdario/mergo"

	"github.com/spf13/cobra"
)

var versionToStop string

func init() {
	config.InitConfig()

	rootCmd.AddCommand(stopCmd)

	for k, srv := range config.AvailableServices() {
		serviceSubcommand := buildStopServiceCommand(k, srv)

		serviceSubcommand.Flags().StringVarP(&versionToStop, "version", "v", srv.Version, "Sets the image version to stop")

		stopServiceCmd.AddCommand(serviceSubcommand)
	}

	stopCmd.AddCommand(stopServiceCmd)

	for k, stack := range config.AvailableStacks() {
		stackSubcommand := buildStopStackCommand(k, stack)

		stopStackCmd.AddCommand(stackSubcommand)
	}

	stopCmd.AddCommand(stopStackCmd)
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops a Service",
	Long: `Stops a Service, stoppping the Docker container for it that exposes its internal
	configuration so that you are able to connect to it in an easy manner`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return errors.New("run requires zero or one argument representing the image tag to be run")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// NOOP
	},
}

func buildStopServiceCommand(srv string, service config.Service) *cobra.Command {
	return &cobra.Command{
		Use:   srv,
		Short: `Stops a ` + srv + ` service`,
		Long: `Stops a ` + srv + ` service, stoppping the Docker container for it that exposes its internal
		configuration so that you are able to connect to it in an easy manner`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return errors.New("run requires zero or one argument representing the image tag to be run")
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			serviceManager := services.NewServiceManager()

			s := serviceManager.Build(srv, versionToStop, true)

			serviceManager.Stop(s)
		},
	}
}

func buildStopStackCommand(key string, stack config.Stack) *cobra.Command {
	return &cobra.Command{
		Use:   key,
		Short: `Stops the ` + stack.Name + ` stack`,
		Long:  `Stops the ` + stack.Name + ` stack, stopping the Services that compound it`,
		Run: func(cmd *cobra.Command, args []string) {
			serviceManager := services.NewServiceManager()

			services := config.AvailableServices()
			if len(stack.Services) == 0 {
				log.Error("The Stack does not contain services. Please check configuration files")
			}

			for k, srv := range stack.Services {
				originalSrv := services[k]
				if !srv.Equals(originalSrv) {
					mergo.Merge(&originalSrv, srv)
				}

				originalSrv.Daemon = true
				s := serviceManager.BuildFromConfig(originalSrv)
				serviceManager.Stop(s)
			}
		},
	}
}

var stopServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Allows to stop a service, defined as subcommands",
	Long:  `Allows to stop a service, defined as subcommands, stopping the Docker containers for them.`,
	Run: func(cmd *cobra.Command, args []string) {
		// NOOP
	},
}

var stopStackCmd = &cobra.Command{
	Use:   "stack",
	Short: "Stops a Stack",
	Long:  `Stops a Stack, compounded by different services that cooperate, stoppping the Docker containers for them that expose their internal configurations`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return errors.New("run requires zero or one argument representing the image tag to be run")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// NOOP
	},
}
