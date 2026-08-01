package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/pmocker-io/pmocker/cli/internal/instance"
	"github.com/spf13/cobra"
)

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "列出运行中的 PMSystem",
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		store, err := instance.InitDefaultStore()
		if err != nil {
			return err
		}
		defer store.Close()
		list, err := store.List(all)
		if err != nil {
			return err
		}
		if len(list) == 0 {
			fmt.Println("没有运行中的实例")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tIMAGE\tPORT\tSTATUS\tPID\tCREATED")
		for _, inst := range list {
			id := inst.ID
			if len(id) > 12 {
				id = id[:12]
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%d\t%s\n",
				id, inst.Name, inst.ImageRef, inst.Port,
				inst.Status, inst.PID, inst.CreatedAt.Format("2006-01-02 15:04"))
		}
		w.Flush()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(psCmd)
	psCmd.Flags().BoolP("all", "a", false, "显示所有实例（包括已停止）")
}
