/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/Octops/agones-event-broadcaster/pkg/broadcaster"
	"github.com/Octops/agones-event-broadcaster/pkg/brokers"
	"github.com/Octops/agones-event-broadcaster/pkg/brokers/pubsub"
	"github.com/Octops/agones-event-broadcaster/pkg/brokers/stdout"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/api/option"
	"k8s.io/client-go/tools/clientcmd"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	cfgFile    string
	kubeconfig string
	verbose    bool
	brokerFlag string
)

var rootCmd = &cobra.Command{
	Use:   "agones-event-broadcaster",
	Short: "Broadcast Events from Agones GameServers",
	Long:  `Broadcast Events from Agones GameServers`,
	Run: func(cmd *cobra.Command, args []string) {
		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}

		clientConf, err := clientcmd.BuildConfigFromFlags("", kubeconfig)

		var broker brokers.Broker
		var opts []option.ClientOption

		if brokerFlag == "pubsub" {
			// If the broadcaster is running within GCP, credentials don't need to be explicitly passed
			// Setting this environment variable is optional. The Service Accounts attached to the worker node should be able to perform the operation via IAM settings.
			if os.Getenv("PUBSUB_CREDENTIALS") != "" {
				opts = append(opts, option.WithCredentialsFile(os.Getenv("PUBSUB_CREDENTIALS")))
			}

			broker, err = pubsub.NewPubSubBroker(&pubsub.Config{
				ProjectID:       os.Getenv("PUBSUB_PROJECT_ID"),
				OnAddTopicID:    "gameserver.events.added",
				OnUpdateTopicID: "gameserver.events.updated",
				OnDeleteTopicID: "gameserver.events.deleted",
			}, opts...)
			if err != nil {
				logrus.WithError(err).Fatal("error creating broker")
			}
		} else {
			// Used only for debugging purpose
			broker = &stdout.StdoutBroker{}
		}

		gsBroadcaster, err := broadcaster.New(clientConf, broker)
		if err != nil {
			logrus.WithError(err).Fatal("error creating broadcaster")
		}

		if err := gsBroadcaster.Start(); err != nil {
			logrus.WithError(err).Fatal("error starting broadcaster")
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.agones-event-broadcaster.yaml)")
	rootCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "Set KUBECONFIG")
	rootCmd.Flags().StringVar(&brokerFlag, "broker", "", "The type of the broker to be used by the broadcaster")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Set log level to verbose, defaults to false")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".agones-event-broadcaster" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".agones-event-broadcaster")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
