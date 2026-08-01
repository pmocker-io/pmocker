package cmd

import (
	"fmt"

	"github.com/pmocker-io/pmocker/cli/internal/builder"
	"github.com/pmocker-io/pmocker/cli/internal/instance"
	"github.com/pmocker-io/pmocker/pkg/pmocker/image"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start <name|id>",
	Short: "启动已停止的 PMSystem 实例",
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

		pmockerHome, _ := instance.DefaultPMockerHome()
	b := builder.NewBuilder(pmockerHome, findGVAServerDir(), findGVAWebDir())
		if err := b.Ensure(); err != nil {
			return err
		}

		mgr := instance.NewManager(store, vols, imgStore, b.BinPath())
		if err := mgr.Start(args[0]); err != nil {
			return err
		}
		fmt.Printf("实例 %s 已启动\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
