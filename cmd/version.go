package cmd

import "github.com/spf13/cobra"

var Version string = "develop"

var versionCmd = &cobra.Command{
	Use:                        "version",
	Short:                      "Display binary version",
	Long:                       "Display binary version",
	Example:                    `gridengine_prometheus version`,
	Run: version,
}

func version(cmd *cobra.Command, args []string) {
	println(Version)
}

func init(){
	RootCmd.AddCommand(versionCmd)
}