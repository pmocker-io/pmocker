package cmd

import (
	"fmt"

	"github.com/pmocker-io/pmocker/cli/internal/instance"
	"github.com/pmocker-io/pmocker/pkg/pmocker/image"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm <name|id>",
	Short: "删除 PMSystem 实例",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		removeVolume, _ := cmd.Flags().GetBool("volumes")
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
		if err := mgr.Remove(args[0], removeVolume); err != nil {
			return err
		}
		fmt.Printf("实例 %s 已删除", args[0])
		if removeVolume {
			fmt.Print("（含数据卷）")
		}
		fmt.Println()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
	rmCmd.Flags().BoolP("volumes", "v", false, "同时删除数据卷")
}
