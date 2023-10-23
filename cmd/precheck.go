package cmd

import (
	"github.com/pingcap/index_advisor/optimizer"
	"github.com/spf13/cobra"
)

func NewPreCheckCmd() *cobra.Command {
	var dsn string
	cmd := &cobra.Command{
		Use:   "pre-check",
		Short: "check what kind of index advisor mode can fit your cluster, use `index_advisor pre-check --help` to see more details",
		Long: `check what kind of index advisor mode can fit your cluster.
How it work:
1. connect to your TiDB cluster through the DSN
2. check whether you can run online-mode index advisor on your cluster`,
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := optimizer.NewTiDBWhatIfOptimizer(dsn)
			if err != nil {
				return err
			}
			reason := checkOnlineModeSupport(db)
			if reason == "" {
				cmd.Println("[pre-check] you can use online mode and offline mode on your cluster.")
			} else {
				cmd.Println("[pre-check] you can only use offline mode on your cluster.")
				cmd.Println("[pre-check] your TiDB cluster does not support Index Advisor Online Mode, reason:", reason)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&dsn, "dsn", "root:@tcp(127.0.0.1:4000)/test", "the DSN of the TiDB cluster")
	return cmd
}

// PreCheck checks whether this cluster is suitable for online-mode.
func checkOnlineModeSupport(db optimizer.WhatIfOptimizer) (reason string) {
	if !supportHypoIndex(db) {
		return "your TiDB version does not support hypothetical index feature, which is required by Index Advisor Online Mode"
	}
	if redactLogEnabled(db) {
		return "redact log is enabled, the Advisor probably cannot get the full SQL text if you use Index Advisor Online Mode"
	}
	return ""
}
