package registry

import (
	"fmt"
	"sort"
	"strings"
)

// AddSkill 登记一个新技能。sourceID 来源必须已存在；pathInSource 为来源内相对路径。
// id 缺省时取 pathInSource 的最后一层目录名。
func (f *File) AddSkill(sourceID, pathInSource, id string) (*Skill, error) {
	src := f.FindSource(sourceID)
	if src == nil {
		return nil, fmt.Errorf("来源 %q 不存在，请先 add 来源", sourceID)
	}
	if pathInSource == "" {
		return nil, fmt.Errorf("pathInSource 不能为空")
	}
	pathInSource = strings.TrimSuffix(strings.TrimSpace(pathInSource), "/")
	pathInSource = strings.TrimSuffix(pathInSource, "\\")
	if id == "" {
		parts := strings.FieldsFunc(pathInSource, func(r rune) bool {
			return r == '/' || r == '\\'
		})
		if len(parts) == 0 {
			return nil, fmt.Errorf("无法从 pathInSource 推导技能名，请显式指定 --id")
		}
		id = parts[len(parts)-1]
	}
	if f.FindSkill(id) != nil {
		return nil, fmt.Errorf("技能 %q 已存在（同名单冲突），请指定不同 --id 或先移除旧技能", id)
	}
	sk := &Skill{ID: id, Source: sourceID, PathInSource: pathInSource, Enabled: true}
	f.Skills = append(f.Skills, sk)
	return sk, nil
}

// RemoveSkill 移除技能登记（不影响 sources 下的源文件）。
func (f *File) RemoveSkill(id string) error {
	for i, s := range f.Skills {
		if s.ID == id {
			f.Skills = append(f.Skills[:i], f.Skills[i+1:]...)
			// 连带移除该技能的补丁记录
			f.removePatchesOf(id)
			return nil
		}
	}
	return fmt.Errorf("技能 %q 未登记", id)
}

func (f *File) removePatchesOf(skillID string) {
	var kept []*Patch
	for _, p := range f.Patches {
		if p.Skill != skillID {
			kept = append(kept, p)
		}
	}
	f.Patches = kept
}

// SetEnabled 设置技能启用状态。
func (f *File) SetEnabled(id string, enabled bool) error {
	sk := f.FindSkill(id)
	if sk == nil {
		return fmt.Errorf("技能 %q 未登记", id)
	}
	sk.Enabled = enabled
	return nil
}

// ListSkills 返回排序后的技能列表。
func (f *File) ListSkills() []*Skill {
	out := make([]*Skill, 0, len(f.Skills))
	out = append(out, f.Skills...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
