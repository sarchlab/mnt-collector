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

// tracesCmd represents the collectTraces command
var tracesCmd = &cobra.Command{
	Use:   "traces",
	Short: "Use Nvbit to generate traces and upload the data to database & cloud.",
	Long: `Use Nvbit to generate traces and upload the data to database & cloud.
Need gpu device.
Need lib/tracer_tool.so and lib/post-traces-processing to be in the project directory`,
	Run: func(cmd *cobra.Command, args []string) {
		config.LoadConfig(collectFile, secretFile)
		if config.C.Cases == nil {
			log.Panic("Cases is not loaded from the collection settings file")
		}
		log.Info("Collection settings and secret tokens loaded.")

		config.PrepareTracesEnvirons()
		log.Info("Traces environment prepared.")

		config.LoadDevice(config.C.DeviceID)
		log.Infof("Device %s loaded.", config.C.DeviceID)
		defer config.ShutdownDevice()

		if config.C.UploadToServer {
			mntbackend.Connect()
			mntbackend.PrepareEnvID()
			log.Info("MNT backend connected.")
			aws.Connect()
			log.Info("AWS connected.")
		} else {
			log.Info("UploadToServer is set to false, no data will be uploaded to the server.")
		}

		collector.RunTraceCollection()
	},
}

func init() {
	rootCmd.AddCommand(tracesCmd)
}
