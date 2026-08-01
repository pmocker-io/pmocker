package cmd

import (
	"fmt"

	"github.com/pmocker-io/pmocker/cli/internal/instance"
	"github.com/pmocker-io/pmocker/pkg/pmocker/image"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop <name|id>",
	Short: "停止 PMSystem 实例",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := instance.InitDefaultStore()
		if err != nil {
			return err
		}
		defer store.Close()
		vols, err := instance.InitDefaultVolumes()
		if err != nil {
			return err
		}
		imgDir, _ := image.DefaultStoreDir()
		imgStore := image.NewStore(imgDir)
		mgr := instance.NewManager(store, vols, imgStore, "")
		if err := mgr.Stop(args[0]); err != nil {
			return err
		}
		fmt.Printf("实例 %s 已停止\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
