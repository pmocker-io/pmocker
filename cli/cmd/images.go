package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/pmocker-io/pmocker/pkg/pmocker/image"
	"github.com/spf13/cobra"
)

var imagesCmd = &cobra.Command{
	Use:   "images",
	Short: "列出本地镜像缓存",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := image.DefaultStoreDir()
		if err != nil {
			return err
		}
		store := image.NewStore(dir)
		images, err := store.ListImages()
		if err != nil {
			return err
		}
		if len(images) == 0 {
			fmt.Println("没有本地镜像")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tTAG\tDIGEST\tSIZE\tMETHOD\tVERSION")
		for _, img := range images {
			digest := img.Digest
			if len(digest) > 19 {
				digest = digest[:19] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%dB\t%s\t%s\n",
				img.Name, img.Tag, digest, img.Size, img.Config.Methodology, img.Config.Version)
		}
		w.Flush()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(imagesCmd)
}
