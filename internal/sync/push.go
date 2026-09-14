// Package sync 实现把启用技能推送到各产品技能目录的同步逻辑。
// 唯一事实源 = registry；覆盖启用、删除禁用、dry-run、subPath，绝不触碰未登记内容。
package sync

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/liuxin/skillhub/internal/registry"
	"github.com/liuxin/skillhub/internal/sources"
)

// Options 控制同步行为。
type Options struct {
	DryRun  bool
	Verbose bool
	Out     io.Writer
}

// PlanStep 描述一次将要/已执行的文件操作。
type PlanStep struct {
	Action string // copy / delete / noop
	Target string
	Detail string
}

// Plan 执行一次推送计划（不涉及文件系统写入），返回计划步骤。
func Plan(reg *registry.File) ([]PlanStep, error) {
	var steps []PlanStep

	for _, t := range reg.Targets {
		root := t.Path
		if t.SubPath != "" {
			root = filepath.Join(root, t.SubPath)
		}
		// 1) 覆盖启用技能
		for _, sk := range reg.ListSkills() {
			if !sk.Enabled {
				continue
			}
			if reg.FindSource(sk.Source) == nil {
				steps = append(steps, PlanStep{Action: "skip", Target: sk.ID, Detail: "来源不存在"})
				continue
			}
			// 计算来源内技能本地绝对路径
			srcDir := sources.CloneDir(reg.Root, sk.Source)
			from := filepath.Join(srcDir, sk.PathInSource)
			if st, err := os.Stat(from); err != nil || !st.IsDir() {
				steps = append(steps, PlanStep{Action: "skip", Target: from, Detail: "来源内目录不存在"})
				continue
			}
			to := filepath.Join(root, sk.ID)
			steps = append(steps, PlanStep{Action: "copy", Target: to, Detail: fmt.Sprintf("from %s", from)})
		}
		// 2) 删除已登记但禁用的技能在目标下的副本
		for _, sk := range reg.ListSkills() {
			if sk.Enabled {
				continue
			}
			to := filepath.Join(root, sk.ID)
			if _, err := os.Stat(to); err == nil {
				steps = append(steps, PlanStep{Action: "delete", Target: to, Detail: "disabled skill"})
			}
		}
	}
	return steps, nil
}

// Run 执行计划。dryRun 时仅打印。返回实际执行步骤数。
func Run(reg *registry.File, opts Options) ([]PlanStep, error) {
	steps, err := Plan(reg)
	if err != nil {
		return nil, err
	}
	var executed []PlanStep
	for _, s := range steps {
		line := actionLabel(s.Action) + " " + s.Target
		if s.Detail != "" {
			line += "  (" + s.Detail + ")"
		}
		fmt.Fprintln(opts.Out, line)
		if s.Action == "skip" {
			continue
		}
		executed = append(executed, s)
		if opts.DryRun {
			continue
		}
		switch s.Action {
		case "copy":
			if err := CopyDir(s.Target, sDetailSource(s)); err != nil {
				return executed, fmt.Errorf("复制 %s 失败: %w", s.Target, err)
			}
		case "delete":
			if err := os.RemoveAll(s.Target); err != nil {
				return executed, fmt.Errorf("删除 %s 失败: %w", s.Target, err)
			}
		}
	}
	return executed, nil
}

func sDetailSource(s PlanStep) string {
	// detail 格式为 "from <path>"，取出源路径
	const p = "from "
	if strings.HasPrefix(s.Detail, p) {
		return strings.TrimPrefix(s.Detail, p)
	}
	return ""
}

func actionLabel(a string) string {
	switch a {
	case "copy":
		return "[拷贝]"
	case "delete":
		return "[删除]"
	case "skip":
		return "[跳过]"
	}
	return "[" + a + "]"
}

// CopyDir 将源目录递归复制到目标目录（先删旧目标再复制，实现覆盖同步）。
func CopyDir(dst, src string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s 不是目录", src)
	}
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	if err := os.MkdirAll(dst, info.Mode().Perm()); err != nil {
		return err
	}
	return copyDirContents(dst, src)
}

func copyDirContents(dst, src string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())
		if e.IsDir() {
			if err := os.MkdirAll(dstPath, e.Type().Perm()); err != nil {
				return err
			}
			if err := copyDirContents(dstPath, srcPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(dstPath, srcPath, e.Type().Perm()); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(dst, src string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
