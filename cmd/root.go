/*
 * Copyright 2020 Netflix, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/netflix/weep/pkg/config"
	"github.com/netflix/weep/pkg/logging"
	"github.com/netflix/weep/pkg/metadata"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd = &cobra.Command{
		Use:               "weep",
		Short:             "weep helps you get the most out of ConsoleMe credentials",
		Long:              "Weep is a CLI tool that manages AWS access via ConsoleMe for local development.",
		DisableAutoGenTag: true,
		SilenceUsage:      true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// This runs before any subcommand, and cmd.CalledAs() returns the subcommand
			// that was called. We want to use this for the weep method in the instance info.
			metadata.SetWeepMethod(cmd.CalledAs())
			// Add basic metadata to ALL future logs
			metadata.AddMetadataToLogger(args)
			logging.Log.Infoln("Incoming weep command")
			if extraConfigFile != "" {
				err := config.MergeExtraConfigFile(extraConfigFile)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	cobra.OnInitialize(updateLoggingConfig)

	rootCmd.PersistentFlags().BoolVarP(&noIpRestrict, "no-ip", "n", false, "remove IP restrictions")
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.weep.yaml)")
	rootCmd.PersistentFlags().StringSliceVarP(&assumeRole, "assume-role", "A", make([]string, 0), "one or more roles to assume after retrieving credentials")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "", "log format (json or tty)")
	rootCmd.PersistentFlags().StringVar(&logFile, "log-file", viper.GetString("log_file"), "log file path")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "", "log level (debug, info, warn)")
	rootCmd.PersistentFlags().StringVarP(&region, "region", "r", viper.GetString("aws.region"), "AWS region")
	rootCmd.PersistentFlags().StringVar(&extraConfigFile, "extra-config-file", "", "extra-config-file <yaml_file>")
	if err := viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log-level")); err != nil {
		logging.LogError(err, "Error parsing")
	}
	if err := viper.BindPFlag("log_file", rootCmd.PersistentFlags().Lookup("log-file")); err != nil {
		logging.LogError(err, "Error parsing")
	}
	if err := viper.BindPFlag("log_format", rootCmd.PersistentFlags().Lookup("log-format")); err != nil {
		logging.LogError(err, "Error parsing")
	}
}

func Run(initFunctions ...func()) {
	cobra.OnInitialize(initFunctions...)
	Execute()
}

func Execute() error {
	shutdown = make(chan os.Signal, 1)
	done = make(chan int, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	if err := rootCmd.Execute(); err != nil {
		// err is already printed out by cobra's Execute
		return err
	}
	return nil
}

func AddCustomCommands(commands []*cobra.Command) {
	for _, command := range commands {
		rootCmd.AddCommand(command)
	}
}

func initConfig() {
	if err := config.InitConfig(cfgFile); err != nil {
		logging.LogError(err, "failed to initialize config")
	}
}

// updateLoggingConfig overrides the default logging settings based on the config and CLI args
func updateLoggingConfig() {
	err := logging.UpdateConfig(logLevel, logFormat, logFile)
	if err != nil {
		logging.LogError(err, "failed to configure logger")
	}
}
