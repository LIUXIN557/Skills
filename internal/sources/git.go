// Package sources 负责管理上游来源仓库：嵌套 clone 引入、git pull 更新。
// 直接调用本机 git 可执行文件，复用已安装的 git。
package sources

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CloneDir 返回来源在中控仓库下的本地目录（嵌套 clone，保留 .git）。
func CloneDir(root, sourceID string) string {
	return filepath.Join(root, "sources", sourceID)
}

// Clone 将 upstream 克隆到 sources/<sourceID>。branch 为空则用上游默认分支。
func Clone(root, sourceID, upstream, branch string) (string, error) {
	dir := CloneDir(root, sourceID)
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		return "", fmt.Errorf("来源已存在: %s", dir)
	}
	args := []string{"clone"}
	if branch != "" {
		args = append(args, "-b", branch)
	}
	args = append(args, "--", upstream, dir)
	if out, err := runGit("", args...); err != nil {
		return "", fmt.Errorf("clone %s 失败: %v\n%s", upstream, err, out)
	}
	return dir, nil
}

// FetchLatest 拉取上游最新到本地引用，不动工作区，返回 origin 目标的 resolved ref（fetch 后的 commit）。
func FetchLatest(dir string) (string, error) {
	// 先确保 origin 存在
	if err := ensureRemote(dir); err != nil {
		return "", err
	}
	if out, err := runGit(dir, "fetch", "origin"); err != nil {
		return "", fmt.Errorf("git fetch origin 失败: %v\n%s", err, out)
	}
	rev, err := revParse(dir, "origin")
	if err != nil {
		return "", err
	}
	return rev, nil
}

// ensureRemote 在无 remote 的仓库上添加 origin 指向当前 HEAD 所在提交的远端。
// 本地仓库（user 用本地路径作来源）可能没有 remote，这里退化为读取本地最新提交。
func ensureRemote(dir string) error {
	out, err := runGit(dir, "remote")
	if err != nil {
		return err
	}
	for _, r := range strings.Fields(out) {
		if r == "origin" {
			return nil
		}
	}
	return nil // 无 origin：调用方用 HEAD 兜底
}

// RevParse 解析某 ref 的 commit。
func RevParse(dir, ref string) (string, error) {
	return revParse(dir, ref)
}

func revParse(dir, ref string) (string, error) {
	out, err := runGit(dir, "rev-parse", ref+"^{commit}")
	if err != nil {
		return "", fmt.Errorf("解析 ref %q 失败: %v", ref, err)
	}
	return strings.TrimSpace(out), nil
}

// ResolveFind 返回来源当前应被跟踪的 ref：
// 优先 origin/HEAD，其次 origin/默认分支，再次当前 HEAD。供 update/patch 作为基线来源。
func ResolveFind(dir string) (string, error) {
	if rev, err := RevParse(dir, "origin/HEAD"); err == nil {
		return rev, nil
	}
	if rev, err := RevParse(dir, "HEAD"); err == nil {
		return rev, nil
	}
	return "", fmt.Errorf("无法解析 %s 的上游引用", dir)
}

func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return stderr.String(), err
	}
	return stdout.String(), nil
}
