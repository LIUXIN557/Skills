// Package registry 定义集中清单（skills.yaml）的数据模型，它是整个工具的唯一事实源。
package registry

import (
	"fmt"
	"sort"
)

// File 对应 skills.yaml 顶层结构：四分区 targets / sources / skills / patches。
type File struct {
	// Root 为中控仓库根目录（绝对路径），物理上由运行时提供，运行时注入，
	// 不在 YAML 中持久化用户可能移动的路径，但保留字段以记录初始位置。
	Root    string   `yaml:"-" json:"-"`
	Config  Config   `yaml:"config" json:"config"`
	Targets []Target `yaml:"targets" json:"targets"`
	Sources []Source `yaml:"sources" json:"sources"`
	Skills  []*Skill `yaml:"skills" json:"skills"`
	Patches []*Patch `yaml:"patches" json:"patches"`
}

// Config 存放全局配置。
type Config struct {
	// ConfigPath 为 skills.yaml 的绝对路径（运行时注入）。
	ConfigPath string `yaml:"-" json:"-"`
}

// Target 描述一个产品技能目录（推送目标）。
type Target struct {
	// Name 为产品名，用于识别（如 trae-cn）。
	Name string `yaml:"name" json:"name"`
	// Path 为产品技能目录绝对路径。
	Path string `yaml:"path" json:"path"`
	// SubPath 可选：技能落入该目标时追加到 Path 的子目录（默认落到根）。
	SubPath string `yaml:"subPath,omitempty" json:"subPath,omitempty"`
}

// Source 描述一个上游来源仓库。
type Source struct {
	// ID 为来源唯一标识（如 owner/repo），用于定位 sources/<ID>。
	ID string `yaml:"id" json:"id"`
	// URL 为上游 git 地址或本地路径。
	URL string `yaml:"url" json:"url"`
	// Branch 为默认跟踪分支，留空则用 origin HEAD。
	Branch string `yaml:"branch,omitempty" json:"branch,omitempty"`
	// UpdateRef 为最近一次同步到的上游 commit（basestamp 基线的来源）。
	UpdateRef string `yaml:"updateRef,omitempty" json:"updateRef,omitempty"`
}

// Skill 描述一个已登记技能。
type Skill struct {
	// ID 为技能唯一标识（技能文件夹名）。
	ID string `yaml:"id" json:"id"`
	// Source 为所属来源 ID。
	Source string `yaml:"source" json:"source"`
	// PathInSource 为该技能在来源仓库内的相对路径。
	PathInSource string `yaml:"pathInSource" json:"pathInSource"`
	// Enabled 为全局启用开关。
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// Patch 描述对某技能的一个个性化补丁。
type Patch struct {
	// Skill 为补丁所属技能 ID。
	Skill string `yaml:"skill" json:"skill"`
	// Num 为序号，用于排序重放（0001、0002...）。
	Num int `yaml:"num" json:"num"`
	// Title 为补丁描述。
	Title string `yaml:"title" json:"title"`
	// Base 为产生该补丁时的上游基线 commit。
	Base string `yaml:"base" json:"base"`
	// File 为该补丁 diff 文件（相对于中控仓库根，如 patches/<skill>/0001-x.diff）。
	File string `yaml:"file" json:"file"`
}

// DefaultTargets 返回默认的 9 个产品技能目录。
func DefaultTargets() []Target {
	base := `C:\Users\Administrator`
	return []Target{
		{Name: "acecode", Path: base + `\.acecode\skills`},
		{Name: "agents", Path: base + `\.agents\skills`},
		{Name: "codex", Path: base + `\.codex\skills`},
		{Name: "grok", Path: base + `\.grok\skills`},
		{Name: "pi", Path: base + `\.pi\agent\skills`},
		{Name: "qoder", Path: base + `\.qoder\skills`},
		{Name: "trae", Path: base + `\.trae\skills`},
		{Name: "trae-cn", Path: base + `\.trae-cn\skills`},
		{Name: "doubao", Path: base + `\DoubaoWork\skills`},
	}
}

// FindSkill 按 ID 查找技能。
func (f *File) FindSkill(id string) *Skill {
	for _, s := range f.Skills {
		if s.ID == id {
			return s
		}
	}
	return nil
}

// FindSource 按 ID 查找来源。
func (f *File) FindSource(id string) *Source {
	for _, s := range f.Sources {
		if s.ID == id {
			return &s
		}
	}
	return nil
}

// FindTarget 按 Name 查找目标。
func (f *File) FindTarget(name string) *Target {
	for _, t := range f.Targets {
		if t.Name == name {
			return &t
		}
	}
	return nil
}

// SkillWithSource 返回技能连同其来源；来源缺失时报错。
func (f *File) SkillWithSource(id string) (*Skill, *Source, error) {
	sk := f.FindSkill(id)
	if sk == nil {
		return nil, nil, fmt.Errorf("技能 %q 未登记", id)
	}
	src := f.FindSource(sk.Source)
	if src == nil {
		return nil, nil, fmt.Errorf("技能 %q 的来源 %q 不存在", id, sk.Source)
	}
	return sk, src, nil
}

// PatchesOf 返回某技能按序号排序的补丁。
func (f *File) PatchesOf(skillID string) []*Patch {
	var out []*Patch
	for _, p := range f.Patches {
		if p.Skill == skillID {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Num < out[j].Num })
	return out
}
