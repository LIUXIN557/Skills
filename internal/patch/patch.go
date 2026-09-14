// Package patch 实现个性化补丁的生成（git diff 相对基线）与应用（git apply 依序重放）。
// 补丁模型：来源仓库工作区 = 上游 + 已应用补丁（本地提交）。
// - patch create：把工作区未提交改动相对当前 HEAD 导出为 diff 文件；改动会本地提交，使推送包含补丁。
// - patch apply  ：重置到基线后依序 git apply 已记录的补丁文件并提交，供 update 后重建最终技能状态。
package patch

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Head 返回仓库当前 HEAD commit。
func Head(repoDir string) (string, error) {
	out, err := runGit(repoDir, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// Diff 返回 relativeTo 到工作区在 path 上的差异。path 为空则输出完整差异。
func Diff(repoDir, relativeTo, path string) (string, error) {
	args := []string{"diff"}
	if relativeTo != "" {
		args = append(args, relativeTo)
	}
	args = append(args, "--", path)
	out, err := runGit(repoDir, args...)
	if err != nil {
		return "", err
	}
	return out, nil
}

// IsClean 报告仓库工作区是否干净（无未提交改动）。
func IsClean(repoDir string) (bool, error) {
	out, err := runGit(repoDir, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) == "", nil
}

// CommitPath 对 path 的改动执行 git add + commit。path 使用来源仓库内相对路径。
func CommitPath(repoDir, path, message string) error {
	if _, err := runGit(repoDir, "add", "-A", "--", path); err != nil {
		return err
	}
	if _, err := runGit(repoDir, "commit", "-m", message); err != nil {
		return fmt.Errorf("提交失败（可能无改动）: %w", err)
	}
	return nil
}

// errPatchConflict 标记三方合并产生冲突（与普通失败区分：冲突保留现场，失败回滚）。
var errPatchConflict = errors.New("补丁三方合并产生冲突")

// ApplyFiles 依序对干净仓库执行 git apply（路径已相对仓库根），并整体提交一次。
// 用于 update 后重建最终技能状态；任一补丁失败即停止返回错误。
// 冲突时保留冲突标记供用户手动解决（不自动回滚）；其余失败回滚到干净基线。
func ApplyFiles(repoDir string, patchFiles []string, commitMsg string) error {
	clean, err := IsClean(repoDir)
	if err != nil {
		return err
	}
	if !clean {
		return fmt.Errorf("仓库 %s 工作区不干净，请先提交或执行 patch 固化改动", repoDir)
	}
	for i, f := range patchFiles {
		if _, err := os.Stat(f); err != nil {
			return fmt.Errorf("补丁文件 %s 不存在", f)
		}
		if err := applyOne(repoDir, f); err != nil {
			if errors.Is(err, errPatchConflict) {
				return fmt.Errorf("重放补丁[%d/%d] %s 产生三方合并冲突，冲突标记已保留在工作区，请手动解决后执行 git add + git commit 完成合并", i+1, len(patchFiles), f)
			}
			// 失败后回滚到干净基线，便于用户手动重试
			_, _ = runGit(repoDir, "reset", "--hard", "HEAD")
			return fmt.Errorf("重放补丁[%d/%d] %s 失败: %v（已回滚到干净基线，请手动处理）", i+1, len(patchFiles), f, err)
		}
	}
	if len(patchFiles) > 0 {
		if _, err := runGit(repoDir, "commit", "-m", commitMsg); err != nil {
			return fmt.Errorf("提交补丁结果失败: %w", err)
		}
	}
	return nil
}

// applyOne 应用单个补丁。优先 git apply --3way（按 blob 三方合并，对上游偏移
// 导致普通 apply 定位失败更鲁棒）；3way 产生冲突时保留冲突标记返回 errPatchConflict；
// 3way 不可用（如补丁缺 index 行）时回退普通 apply，容忍 CRLF/空白差异。
func applyOne(repoDir, patchFile string) error {
	out, err := runGit(repoDir, "apply", "--3way", "--index", patchFile)
	if err == nil {
		return nil
	}
	if strings.Contains(out, "with conflicts") {
		return fmt.Errorf("%w: %s", errPatchConflict, strings.TrimSpace(out))
	}
	if _, err := runGit(repoDir, "apply", "--index", "--whitespace=nowarn", patchFile); err == nil {
		return nil
	}
	// 最后兜底：仅 --check 看是否可干净应用，把确切错误返回给用户
	out, err = runGit(repoDir, "apply", "--check", patchFile)
	if err != nil {
		return fmt.Errorf("%v\n%s", err, out)
	}
	_, err = runGit(repoDir, "apply", "--index", patchFile)
	return err
}

// ResetHard 将仓库工作区与当前分支硬重置到 ref（丢弃本地提交，补丁已记录在 diff 文件中）。
func ResetHard(repoDir, ref string) error {
	if _, err := runGit(repoDir, "reset", "--hard", ref); err != nil {
		return fmt.Errorf("重置到 %s 失败: %w", ref, err)
	}
	return nil
}

func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return stderr.String(), err
	}
	return stdout.String(), nil
}
