package cli

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/liuxin/skillhub/internal/registry"
	"github.com/liuxin/skillhub/internal/sources"
)

// cloneSource 克隆来源仓库，返回本地目录。
func cloneSource(root, sourceID, upstream, branch string) (string, error) {
	dir, err := sources.Clone(root, sourceID, upstream, branch)
	if err != nil {
		return "", err
	}
	return dir, nil
}

// resolveHeadForReg 返回来源克隆后的 HEAD commit，作为 UpdateRef 基线。
func resolveHeadForReg(dir string) string {
	if rev, err := sources.RevParse(dir, "HEAD"); err == nil {
		return rev
	}
	return ""
}

var safeTitleRe = regexp.MustCompile(`[^A-Za-z0-9_\p{Han}-]+`)

// patchFileName 生成补丁文件名：NNNN-标题.diff。
func patchFileName(num int, title string) string {
	t := strings.TrimSpace(title)
	t = safeTitleRe.ReplaceAllString(t, "-")
	t = strings.Trim(t, "-")
	if t == "" {
		t = "misc"
	}
	return fmt.Sprintf("%04d-%s.diff", num, t)
}

// nextPatchNum 返回某技能下一个补丁序号。
func nextPatchNum(reg *registry.File, skillID string) int {
	ps := reg.PatchesOf(skillID)
	if len(ps) == 0 {
		return 1
	}
	return ps[len(ps)-1].Num + 1
}

// patchAbsPath 返回某补丁记录对应的绝对路径。
func patchAbsPath(root string, p *registry.Patch) string {
	return filepath.Join(root, p.File)
}

// patchFilesOf 返回某技能补丁的绝对路径列表（按序号）。
func patchFilesOf(reg *registry.File, skillID string) []string {
	var files []string
	for _, p := range reg.PatchesOf(skillID) {
		files = append(files, patchAbsPath(reg.Root, p))
	}
	return files
}

// collectSourcePatchFiles 收集某来源下所有技能的所有补丁文件（用于 update 后整体重放）。
func collectSourcePatchFiles(reg *registry.File, sourceID string) []string {
	var files []string
	for _, sk := range reg.ListSkills() {
		if sk.Source == sourceID {
			files = append(files, patchFilesOf(reg, sk.ID)...)
		}
	}
	return files
}

func mustSourceRepo(root, sourceID string) string {
	return sources.CloneDir(root, sourceID)
}
