package cmd

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/netflix/weep/pkg/logging"

	"github.com/spf13/viper"

	"github.com/kardianos/service"
	"github.com/spf13/cobra"
)

var (
	svcLogger   service.Logger
	svcConfig   *service.Config
	svcProgram  *program
	weepService service.Service
)

func init() {
	weepServiceControl.Args = cobra.MinimumNArgs(1)
	rootCmd.AddCommand(weepServiceControl)
}

var weepServiceControl = &cobra.Command{
	Use:    "service [start|stop|restart|install|uninstall|run]",
	Short:  serviceShortHelp,
	Long:   serviceLongHelp,
	RunE:   runWeepServiceControl,
	Hidden: true,
}

func runWeepServiceControl(cmd *cobra.Command, args []string) error {
	err := initService()
	if err != nil {
		logging.LogError(err, "Error initializing service")
		return err
	}
	if len(args[0]) > 0 {
		// hijack a run command and run the service
		if args[0] == "run" {
			go weepService.Run()
			<-done
			return nil
		}
		err := service.Control(weepService, args[0])
		if err != nil {
			logging.LogError(err, "Error running service control")
			return err
		}
		logging.Log.WithFields(logrus.Fields{"service": args[0]}).Infoln("Success running service")
		cmd.Printf("successfully ran service %s\n", args[0])
	}
	logging.Log.Debug("sending done signal")
	done <- 0
	return nil
}

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	var err error
	logging.Log.Info("starting weep service!")
	exitCode := 0

	flags := viper.GetStringSlice("service.args")
	err = rootCmd.ParseFlags(flags)
	if err != nil {
		logging.Log.Errorf("could not parse flags: %v", err)
	}

	args := viper.GetStringSlice("service.args")
	switch command := viper.GetString("service.command"); command {
	case "ecs_credential_provider":
		fallthrough
	case "serve":
		err = runWeepServer(nil, args)
		if err != nil {
			logging.Log.Error(err)
			exitCode = 1
		}
	default:
		logging.Log.Error("unknown command: ", command)
		exitCode = 1
	}
	logging.Log.Debug("sending done signal")
	done <- exitCode
}

func (p *program) Stop(s service.Service) error {
	// Send an interrupt to the shutdown channel so everything will clean itself up
	// This is seemingly only necessary on Windows, but it shouldn't hurt anything on other platforms.
	logging.Log.Debug("got service stop, sending interrupt")
	shutdown <- os.Interrupt

	// Wait for whatever is running to signal that it's done
	logging.Log.Debug("waiting for done signal")
	<-done
	return nil
}

func initService() error {
	var err error

	svcProgram = &program{}

	var args []string
	for _, key := range []string{"service.flags", "service.run"} {
		configArgs := viper.GetStringSlice(key)
		if len(configArgs) > 0 {
			args = append(args, configArgs...)
		}
	}
	svcConfig = &service.Config{
		Name:        "weep",
		DisplayName: "Weep",
		Description: "The ConsoleMe CLI",
		Arguments:   args,
		Option: service.KeyValue{
			"OnFailure":              "restart", // Windows only
			"OnFailureDelayDuration": "1m",      // Windows only
		},
	}

	weepService, err = service.New(svcProgram, svcConfig)
	if err != nil {
		return err
	}

	errs := make(chan error, 5)
	svcLogger, err = weepService.Logger(errs)
	if err != nil {
		return err
	}

	go func() {
		for {
			err := <-errs
			if err != nil {
				_ = svcLogger.Error(err)
			}
		}
	}()
	return nil
}
