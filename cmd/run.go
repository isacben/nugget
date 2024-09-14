/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"nugget/pkg/request"

	"github.com/spf13/cobra"
)

var JsonFlag bool
var Header bool
var Quiet bool

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run requests in file.",
	Long:  `Run the API requests defined in a yaml file. This command will execute all the requests listed in a signle file.`,
	Run: func(cmd *cobra.Command, args []string) {
		fileName := args[0]
		request.Execute(fileName, JsonFlag, Header, Quiet)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.
	runCmd.PersistentFlags().BoolVar(&JsonFlag, "json", false, "indent json output")
	runCmd.PersistentFlags().BoolVarP(&Header, "header", "H", false, "print response headers")
	runCmd.PersistentFlags().BoolVarP(&Quiet, "quiet", "q", false, "display less output")

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
