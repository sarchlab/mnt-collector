/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var collectFile string
var secretFile string

var rootCmd = &cobra.Command{
	Use:   "mnt-collector",
	Short: "Data Collector for the MGPUSim NVIDIA Trace Project.",
	Long: `Data Collector for the MGPUSim NVIDIA Trace Project.
[Commands]
traces
profiles
simulations
[Flags]
-collect to specify the collection settings file
-secret	to specify the secret tokens file`,
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
	rootCmd.PersistentFlags().StringVar(&collectFile, "collect", "etc/collects.yaml", "yaml file that store collection settings")
	rootCmd.PersistentFlags().StringVar(&secretFile, "secret", "etc/secrets.yaml", "yaml file that store secret tokens")

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
