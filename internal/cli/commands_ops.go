package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/liuxin/skillhub/internal/lock"
	"github.com/liuxin/skillhub/internal/patch"
	"github.com/liuxin/skillhub/internal/registry"
	"github.com/liuxin/skillhub/internal/sources"
	"github.com/liuxin/skillhub/internal/sync"
	"github.com/spf13/cobra"
)

// newPatchCmd 生成补丁：把来源仓库工作区改动相对当前 HEAD 导出为 diff 并本地提交。
func newPatchCmd() *cobra.Command {
	var title string
	cmd := &cobra.Command{
		Use:   "patch <skill-id>",
		Short: "记录当前来源仓库工作区的改动为补丁（导出 diff 并本地提交）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(title) == "" {
				return fmt.Errorf("请用 --title 说明补丁内容")
			}
			reg, err := loadRegistry()
			if err != nil {
				return err
			}
			return lock.LockedDo(reg.Config.ConfigPath, func() error {
				reg2, _ := registry.Load(rootPath, false)
				sk, src, err := reg2.SkillWithSource(args[0])
				if err != nil {
					return err
				}
				repo := mustSourceRepo(reg2.Root, src.ID)
				if err := requireSourceGitRepo(reg2.Root, src.ID); err != nil {
					return err
				}
				base, err := patch.Head(repo)
				if err != nil {
					return err
				}
				diffTxt, err := patch.Diff(repo, base, sk.PathInSource)
				if err != nil {
					return err
				}
				if strings.TrimSpace(diffTxt) == "" {
					return fmt.Errorf("技能 %s 的来源工作区没有改动可记录", sk.ID)
				}
				num := nextPatchNum(reg2, sk.ID)
				fname := patchFileName(num, title)
				rel := "patches" + "/" + sk.ID + "/" + fname
				dst := abs(reg2.Root, rel)
				if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
					return err
				}
				if err := os.WriteFile(dst, []byte(diffTxt), 0o644); err != nil {
					return err
				}
				reg2.Patches = append(reg2.Patches, &registry.Patch{
					Skill: sk.ID, Num: num, Title: title, Base: base, File: rel,
				})
				cm := fmt.Sprintf("skillhub(%s): %s", sk.ID, title)
				if err := patch.CommitPath(repo, sk.PathInSource, cm); err != nil {
					return fmt.Errorf("提交工作区改动失败: %w", err)
				}
				if err := reg2.Save(); err != nil {
					return err
				}
				fmt.Printf("补丁已记录: %s （基线 %s）\n", rel, shortHash(base))
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "补丁描述（必填）")
	return cmd
}

// newPatchApplyCmd 对单个技能：重置到基线后依序重放该技能补丁。
func newPatchApplyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "patch-apply <skill-id>",
		Short: "重置来源到基线并依序重放该技能的补丁（update 后重建）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := loadRegistry()
			if err != nil {
				return err
			}
			return lock.LockedDo(reg.Config.ConfigPath, func() error {
				reg2, _ := registry.Load(rootPath, false)
				sk, src, err := reg2.SkillWithSource(args[0])
				if err != nil {
					return err
				}
				repo := mustSourceRepo(reg2.Root, src.ID)
				if err := requireSourceGitRepo(reg2.Root, src.ID); err != nil {
					return err
				}
				clean, err := patch.IsClean(repo)
				if err != nil {
					return err
				}
				if !clean {
					return fmt.Errorf("来源 %s 工作区不干净，请先执行 patch 固化改动", src.ID)
				}
				base, err := sources.ResolveFind(repo)
				if err != nil {
					return err
				}
				if err := patch.ResetHard(repo, base); err != nil {
					return err
				}
				files := patchFilesOf(reg2, sk.ID)
				if len(files) == 0 {
					fmt.Printf("技能 %s 无补丁可重放（已重置到基线）\n", sk.ID)
					return nil
				}
				if err := patch.ApplyFiles(repo, files, fmt.Sprintf("skillhub: 重放 %s 的补丁", sk.ID)); err != nil {
					return err
				}
				fmt.Printf("已重放 %d 个补丁: %s\n", len(files), sk.ID)
				return nil
			})
		},
	}
}

// newUpdateCmd 更新某来源：fetch 上游、重置到最新、依序重放补丁。
func newUpdateCmd() *cobra.Command {
	var noReplay bool
	cmd := &cobra.Command{
		Use:   "update <source-id>",
		Short: "从上游拉取最新并重放该来源下所有技能的补丁",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := loadRegistry()
			if err != nil {
				return err
			}
			return lock.LockedDo(reg.Config.ConfigPath, func() error {
				reg2, _ := registry.Load(rootPath, false)
				src := reg2.FindSource(args[0])
				if src == nil {
					return fmt.Errorf("来源 %q 不存在", args[0])
				}
				repo := mustSourceRepo(reg2.Root, src.ID)
				if err := requireSourceGitRepo(reg2.Root, src.ID); err != nil {
					return err
				}
				clean, err := patch.IsClean(repo)
				if err != nil {
					return err
				}
				if !clean {
					return fmt.Errorf("来源 %s 工作区不干净（有未固化改动），请先执行 patch", src.ID)
				}
				newRef, err := sources.FetchLatest(repo)
				if err != nil {
					return err
				}
				if newRef == src.UpdateRef && src.UpdateRef != "" {
					fmt.Printf("来源 %s 已是最新（%s）\n", src.ID, shortHash(newRef))
					return nil
				}
				fmt.Printf("来源 %s: %s -> %s\n", src.ID, shortHash(src.UpdateRef), shortHash(newRef))
				if err := patch.ResetHard(repo, newRef); err != nil {
					return err
				}
				src.UpdateRef = newRef
				if !noReplay {
					files := collectSourcePatchFiles(reg2, src.ID)
					if len(files) > 0 {
						if err := patch.ApplyFiles(repo, files, "skillhub: 上游更新后重放补丁"); err != nil {
							// 保留 UpdateRef 更新，但对重放失败给出明确提示
							_ = reg2.Save()
							return err
						}
						fmt.Printf("已重放 %d 个补丁\n", len(files))
					}
				}
				return reg2.Save()
			})
		},
	}
	cmd.Flags().BoolVar(&noReplay, "no-replay", false, "更新后不自动重放补丁")
	return cmd
}

// newPushCmd 推送到所有产品技能目录。
func newPushCmd() *cobra.Command {
	var dryRun, verbose bool
	cmd := &cobra.Command{
		Use:   "push",
		Short: "把启用技能同步到所有目标产品目录，并移除已禁用技能",
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := loadRegistry()
			if err != nil {
				return err
			}
			opts := sync.Options{DryRun: dryRun, Verbose: verbose, Out: os.Stdout}
			executed, err := sync.Run(reg, opts)
			if err != nil {
				return err
			}
			var copies, dels int
			for _, s := range executed {
				switch s.Action {
				case "copy":
					copies++
				case "delete":
					dels++
				}
			}
			mode := ""
			if dryRun {
				mode = "（预览，未实际执行）"
			}
			fmt.Printf("完成%s: 拷贝 %d 个，删除 %d 个\n", mode, copies, dels)
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "仅预览，不实际写入")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "冗余输出")
	return cmd
}

func abs(root, rel string) string {
	return root + string(os.PathSeparator) + strings.ReplaceAll(rel, "/", string(os.PathSeparator))
}

func shortHash(rev string) string {
	if len(rev) > 12 {
		return rev[:12]
	}
	return rev
}
