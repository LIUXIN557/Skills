package cli

import (
	"fmt"

	"github.com/liuxin/skillhub/internal/server"
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	var port int
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "启动本地 Web 管理页面",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := loadRegistry(); err != nil {
				return fmt.Errorf("启动前先 init 并登记技能: %w", err)
			}
			return server.New(rootPath).Serve(port)
		},
	}
	cmd.Flags().IntVar(&port, "port", 8787, "监听端口")
	return cmd
}
