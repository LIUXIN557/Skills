// Package lock 提供跨命令的清单文件锁，防止并发写坏 skills.yaml。
package lock

import (
	"fmt"
	"os"
	"path/filepath"
)

// Guard 表示一个已持有的文件锁。
type Guard struct {
	path string
}

// Acquire 通过 O_CREATE|O_EXCL 独占创建 <config>.lock 获取锁。
// 若锁已存在则立即返回错误（简单单机场景，不做阻塞等待）。
func Acquire(configPath string) (*Guard, error) {
	lockPath := configPath + ".lock"
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		if os.IsExist(err) {
			file, _ := os.Stat(lockPath)
			modTime := "<unknown>"
			if file != nil {
				modTime = file.ModTime().Format("2006-01-02 15:04:05")
			}
			return nil, fmt.Errorf("清单锁已存在 %s（创建于 %s），可能是上次操作未正常结束，请删除后重试", lockPath, modTime)
		}
		return nil, fmt.Errorf("获取清单锁失败: %w", err)
	}
	_ = f.Close()
	return &Guard{path: lockPath}, nil
}

// Release 释放锁并删除锁文件。
func (g *Guard) Release() {
	if g == nil || g.path == "" {
		return
	}
	_ = os.Remove(g.path)
	g.path = ""
}

// LockedDo 在锁保护下执行 fn。
func LockedDo(configPath string, fn func() error) error {
	g, err := Acquire(configPath)
	if err != nil {
		return err
	}
	defer g.Release()
	return fn()
}

// LockPath 返回锁文件路径，便于错误信息中提示用户。
func LockPath(configPath string) string {
	return filepath.Dir(configPath) + string(os.PathSeparator) + filepath.Base(configPath) + ".lock"
}
