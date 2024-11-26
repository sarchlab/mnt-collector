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

// profilesCmd represents the collectProfiles command
var profilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "Use Nvidia system to profile the cases and upload the data to database & cloud.",
	Long: `Use Nvidia system to profile the cases and upload the data to database & cloud.
Need gpu device.`,
	Run: func(cmd *cobra.Command, args []string) {
		config.LoadConfig(collectFile, secretFile)
		if config.C.Cases == nil {
			panic("Cases is not loaded from the collection settings file")
		}
		if config.C.RepeatTimes == 0 {
			panic("RepeatTimes is not loaded from the collection settings file")
		}
		log.Info("Collection settings and secret tokens loaded.")

		config.PrepareProfilesEnvirons()
		log.Info("Profiles environment prepared.")

		config.LoadDevice(config.C.DeviceID)
		log.Infof("Device %s loaded.", config.C.DeviceID)
		defer config.ShutdownDevice()

		if config.C.UploadToServer {
			mntbackend.Init()
			log.Info("MNT backend connected.")
			aws.Init()
			log.Info("AWS connected.")
		} else {
			log.Info("UploadToServer is set to false, no data will be uploaded to the server.")
		}

		collector.RunProfileCollection()
	},
}

func init() {
	rootCmd.AddCommand(profilesCmd)
}
