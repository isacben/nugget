/*
Copyright © 2024 Isaac Benitez
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/isacben/nugget/runner"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var RawFlag bool
var Header bool
var Quiet bool
var ParserFlag bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "nugget",
	Short: "Run API requests.",
	Long: `nugget is a CLI application to test APIs.
This application lets you chain API requests defined in configuration files.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Please provide a file name.")
			os.Exit(1)
		}
		fileName := args[0]
		runner.Execute(fileName, RawFlag, Header, Quiet, ParserFlag)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.nugget.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.PersistentFlags().BoolVar(&RawFlag, "raw", false, "raw output")
	rootCmd.PersistentFlags().BoolVarP(&Header, "header", "H", false, "print response headers")
	rootCmd.PersistentFlags().BoolVarP(&Quiet, "quiet", "q", false, "display less output")
	rootCmd.PersistentFlags().BoolVarP(&ParserFlag, "parser", "p", false, "use parser (experimental)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".nugget" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".nugget")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
