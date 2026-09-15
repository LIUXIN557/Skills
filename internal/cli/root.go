// Package cli 实现 skill 命令的 cobra 命令树，并与各内部包编排。
package cli

import (
	"fmt"
	"os"

	"github.com/liuxin/skillhub/internal/registry"
	"github.com/spf13/cobra"
)

var rootPath string

// Execute 运行根命令。
func Execute() {
	root := &cobra.Command{
		Use:   "skill",
		Short: "个人技能中控仓库（Skill Hub）",
		Long:  "从上游 git 来源维护一本技能清单，一键推送到各 AI 产品的技能目录，支持个性化补丁重放。",
	}
	root.PersistentFlags().StringVarP(&rootPath, "root", "r", ".", "中控仓库根目录（默认当前目录，必须包含 skills.yaml）")
	root.AddCommand(
		newInitCmd(),
		newListCmd(),
		newAddCmd(),
		newRMCmd(),
		newSetEnabledCmd("enable", true),
		newSetEnabledCmd("disable", false),
		newPatchCmd(),
		newPatchApplyCmd(),
		newUpdateCmd(),
		newRestoreCmd(),
		newPushCmd(),
		newServeCmd(),
	)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

// loadRegistry 加载清单（不存在则报错）。用于只读命令。
func loadRegistry() (*registry.File, error) {
	f, err := registry.Load(rootPath, false)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func fail(prefix string, err error) error {
	return fmt.Errorf("%s: %w", prefix, err)
}
