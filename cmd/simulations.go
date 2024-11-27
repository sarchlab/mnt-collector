/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/labstack/gommon/log"
	"github.com/sarchlab/mnt-collector/aws"
	"github.com/sarchlab/mnt-collector/collector"
	"github.com/sarchlab/mnt-collector/config"
	"github.com/sarchlab/mnt-collector/mntbackend"
	"github.com/spf13/cobra"
)

// simulationsCmd represents the collectSimulations command
var simulationsCmd = &cobra.Command{
	Use:   "simulations",
	Short: "Use the given simulator to run traces and upload the data to database.",
	Long: `Use the given simulator to run traces and upload the data to database.
`,
	Run: func(cmd *cobra.Command, args []string) {
		config.LoadConfig(collectFile, secretFile)
		if config.C.TraceIDs == nil {
			log.Panic("TraceIDs is not loaded from the collection settings file")
		}
		if config.C.Experiment.Runfile == "" {
			log.Panic("Runfile is not loaded from the collection settings file")
		}
		log.Info("Collection settings and secret tokens loaded.")

		aws.Connect()
		log.Info("AWS connected.")
		mntbackend.Connect()
		if config.C.UploadToServer {
			mntbackend.PrepareExpID()
			log.Info("Simulations environment prepared.")
		} else {
			log.Info("UploadToServer is set to false, no data will be uploaded to the server.")
		}

		collector.RunSimulationCollection()
	},
}

func init() {
	rootCmd.AddCommand(simulationsCmd)
}
