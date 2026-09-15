package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/liuxin/skillhub/internal/lock"
	"github.com/liuxin/skillhub/internal/registry"
	"github.com/liuxin/skillhub/internal/sources"
	"github.com/spf13/cobra"
)

// newRestoreCmd 从 bundles/<id>.bundle 恢复来源 git 仓库。
// 克隆主仓库后 sources/ 是普通源码副本（无 .git）；需要从上游更新或重放补丁时，
// 用本命令把某来源恢复为 git 仓库（clone bundle → 修正 origin → 配置 git 身份），
// 再执行 skill update <id> 对齐清单版本并自动重放补丁。
func newRestoreCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "restore [source-id ...]",
		Short: "从 bundles/ 恢复来源 git 仓库（不指定时恢复全部缺失来源）",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := loadRegistry()
			if err != nil {
				return err
			}
			return lock.LockedDo(reg.Config.ConfigPath, func() error {
				reg2, _ := registry.Load(rootPath, false)
				ids := args
				if len(ids) == 0 {
					for _, s := range reg2.Sources {
						ids = append(ids, s.ID)
					}
				}
				restored, skipped := 0, 0
				for _, id := range ids {
					if reg2.FindSource(id) == nil {
						fmt.Printf("跳过 %s: 来源不存在\n", id)
						skipped++
						continue
					}
					bundle := filepath.Join(reg2.Root, "bundles", id+".bundle")
					if _, err := os.Stat(bundle); err != nil {
						fmt.Printf("跳过 %s: 无 bundle 文件 %s\n", id, bundle)
						skipped++
						continue
					}
					dir := sources.CloneDir(reg2.Root, id)
					if st, err := os.Stat(filepath.Join(dir, ".git")); err == nil && st.IsDir() {
						fmt.Printf("跳过 %s: git 仓库已存在\n", id)
						skipped++
						continue
					}
					// 目录存在且非空：普通源码副本，需要 force 才移走重建
					if st, err := os.Stat(dir); err == nil && st.IsDir() {
						entries, _ := os.ReadDir(dir)
						if len(entries) > 0 && !force {
							fmt.Printf("跳过 %s: %s 非空且无 .git（如需恢复请先移走，或使用 --force）\n", id, dir)
							skipped++
							continue
						}
						if err := os.RemoveAll(dir); err != nil {
							return fmt.Errorf("清理 %s 失败: %w", dir, err)
						}
					}
					src := reg2.FindSource(id)
					if err := sources.CloneBundle(bundle, dir); err != nil {
						return err
					}
					if err := sources.SetOrigin(dir, src.URL); err != nil {
						return err
					}
					if err := sources.EnsureGitIdentity(dir); err != nil {
						return err
					}
					fmt.Printf("已恢复 %s（remote origin -> %s）\n", id, src.URL)
					restored++
				}
				fmt.Printf("完成：恢复 %d 个来源，跳过 %d 个。可选 ./skill update <id> 对齐清单版本并重放补丁\n", restored, skipped)
				return nil
			})
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "非空且无 .git 的来源目录将被移走并重建")
	return cmd
}
