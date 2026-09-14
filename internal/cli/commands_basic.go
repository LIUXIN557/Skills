package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/liuxin/skillhub/internal/lock"
	"github.com/liuxin/skillhub/internal/registry"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "初始化清单文件 skills.yaml（含默认目标产品目录）",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := registry.Init(rootPath); err != nil {
				return err
			}
			fmt.Printf("已创建清单: %s\n", registry.ConfigPathFor(rootPath))
			return nil
		},
	}
}

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出所有技能及其启用状态",
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := loadRegistry()
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "技能ID\t状态\t来源\t来源内路径\t补丁数")
			for _, sk := range reg.ListSkills() {
				patchCount := len(reg.PatchesOf(sk.ID))
				status := "禁用"
				if sk.Enabled {
					status = "启用"
				}
				src := ""
				if s := reg.FindSource(sk.Source); s != nil {
					src = s.URL
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n", sk.ID, status, src, sk.PathInSource, patchCount)
			}
			w.Flush()
			return nil
		},
	}
}

func newAddCmd() *cobra.Command {
	var sourceID, branch, pathInSource, skillID string
	cmd := &cobra.Command{
		Use:   "add <url|本地路径>",
		Short: "引入一个上游来源仓库并登记技能",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			upstream := args[0]
			reg, err := loadRegistry()
			if err != nil {
				return err
			}
			return lock.LockedDo(reg.Config.ConfigPath, func() error {
				// 重新加载以在锁内修改
				reg2, err := registry.Load(rootPath, false)
				if err != nil {
					return err
				}
				if sourceID == "" {
					sourceID = deriveSourceID(upstream)
				}
				if reg2.FindSource(sourceID) != nil {
					return fmt.Errorf("来源 %q 已存在", sourceID)
				}
				dir, err := cloneSource(reg2.Root, sourceID, upstream, branch)
				if err != nil {
					return err
				}
				reg2.Sources = append(reg2.Sources, registry.Source{
					ID: sourceID, URL: upstream, Branch: branch,
					UpdateRef: resolveHeadForReg(dir),
				})
				var added []string
				if pathInSource != "" {
					sk, err := reg2.AddSkill(sourceID, pathInSource, skillID)
					if err != nil {
						return err
					}
					added = append(added, sk.ID)
				}
				if err := reg2.Save(); err != nil {
					return err
				}
				fmt.Printf("已引入来源 %s 到 %s\n", sourceID, filepath.Join(reg2.Root, "sources", sourceID))
				for _, id := range added {
					fmt.Printf("  技能已登记: %s（来源 %s，默认启用）\n", id, sourceID)
				}
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&sourceID, "source-id", "", "来源标识（默认从 url 自动推导）")
	cmd.Flags().StringVar(&branch, "branch", "", "来源默认分支")
	cmd.Flags().StringVar(&pathInSource, "dir", "", "技能在来源内的相对路径（如 skills/checklist）；给出则同时登记技能")
	cmd.Flags().StringVar(&skillID, "skill-id", "", "技能标识（默认取 dir 的最后一段）")
	return cmd
}

func newRMCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm <skill-id>",
		Short: "移除技能登记（不影响 sources 下的源文件与产品目录副本）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := loadRegistry()
			if err != nil {
				return err
			}
			return lock.LockedDo(reg.Config.ConfigPath, func() error {
				reg2, _ := registry.Load(rootPath, false)
				if err := reg2.RemoveSkill(args[0]); err != nil {
					return err
				}
				return reg2.Save()
			})
		},
	}
}

func newSetEnabledCmd(name string, value bool) *cobra.Command {
	return &cobra.Command{
		Use:   name + " <skill-id>",
		Short: "启用/禁用技能",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := loadRegistry()
			if err != nil {
				return err
			}
			return lock.LockedDo(reg.Config.ConfigPath, func() error {
				reg2, _ := registry.Load(rootPath, false)
				if err := reg2.SetEnabled(args[0], value); err != nil {
					return err
				}
				if err := reg2.Save(); err != nil {
					return err
				}
				state := "启用"
				if !value {
					state = "禁用"
				}
				fmt.Printf("%s 已%s。运行 push 生效。\n", args[0], state)
				return nil
			})
		},
	}
}

// deriveSourceID 从 url 或本地路径推导来源标识。
func deriveSourceID(ref string) string {
	ref = strings.TrimSpace(ref)
	ref = strings.TrimSuffix(ref, ".git")
	if i := strings.LastIndexAny(ref, "/\\"); i >= 0 {
		return ref[i+1:]
	}
	return ref
}

func sortedSourceIDs(reg *registry.File) []string {
	var ids []string
	for _, s := range reg.Sources {
		ids = append(ids, s.ID)
	}
	sort.Strings(ids)
	return ids
}
