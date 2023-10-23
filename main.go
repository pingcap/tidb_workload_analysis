package main

import (
	"github.com/pingcap/index_advisor/cmd"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "TiDB-index-advisor",
		Short: "TiDB index advisor",
		Long:  `TiDB index advisor recommends you the best indexes for your workload`,
	}
)

func init() {
	cobra.OnInitialize()
	rootCmd.AddCommand(cmd.NewAdviseOnlineCmd())
	rootCmd.AddCommand(cmd.NewAdviseOfflineCmd())
	rootCmd.AddCommand(cmd.NewPreCheckCmd())
	rootCmd.AddCommand(cmd.NewEvaluateCmd())
	rootCmd.AddCommand(cmd.NewWorkloadExportCmd())
}

func main() {
	rootCmd.Execute()
}
