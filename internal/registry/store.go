package registry

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ConfigFileName 为清单文件名。
const ConfigFileName = "skills.yaml"

// ConfigPathFor 返回 root 下的清单文件路径。
func ConfigPathFor(root string) string {
	return filepath.Join(root, ConfigFileName)
}

// Load 读取并解析清单文件。若文件不存在且 create 为真，则初始化一份默认清单。
func Load(root string, create bool) (*File, error) {
	path := ConfigPathFor(root)
	f := &File{Root: root, Config: Config{ConfigPath: path}}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) && create {
			if err := Init(root); err != nil {
				return nil, err
			}
			return Load(root, false)
		}
		return nil, fmt.Errorf("读取清单 %s 失败: %w", path, err)
	}
	if err := yaml.Unmarshal(data, f); err != nil {
		return nil, fmt.Errorf("解析清单 %s 失败: %w", path, err)
	}
	f.Root = root
	f.Config.ConfigPath = path
	// 规范 root 为绝对路径
	if abs, err := filepath.Abs(root); err == nil {
		f.Root = abs
	}
	return f, nil
}

// Init 创建一份带默认目标的初始清单（不覆盖已存在文件）。
func Init(root string) error {
	path := ConfigPathFor(root)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("清单已存在: %s", path)
	}
	f := &File{
		Targets: DefaultTargets(),
	}
	return SaveFile(path, f)
}

// Save 序列化并写出清单文件。
func (f *File) Save() error {
	return SaveFile(f.Config.ConfigPath, f)
}

// SaveFile 将 File 写入 path。
func SaveFile(path string, f *File) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(f)
	if err != nil {
		return fmt.Errorf("序列化清单失败: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("写出清单失败: %w", err)
	}
	return nil
}
