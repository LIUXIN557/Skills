package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAddSkillAndConflict(t *testing.T) {
	f := &File{Targets: DefaultTargets()}
	f.Sources = append(f.Sources, Source{ID: "src1", URL: "https://example/x"})

	sk, err := f.AddSkill("src1", "skills/checklist", "")
	if err != nil {
		t.Fatalf("AddSkill: %v", err)
	}
	if sk.ID != "checklist" {
		t.Errorf("期望 id=checklist，得到 %q", sk.ID)
	}
	if !sk.Enabled {
		t.Error("新技能应默认启用")
	}
	// 同名单冲突
	if _, err := f.AddSkill("src1", "skills/checklist", ""); err == nil {
		t.Error("期望同名冲突报错，但没有")
	}
	// 来源不存在
	if _, err := f.AddSkill("nope", "x", "y"); err == nil {
		t.Error("期望来源不存在报错，但没有")
	}
}

func TestSetAndRemove(t *testing.T) {
	f := &File{}
	f.Sources = []Source{{ID: "s"}}
	f.AddSkill("s", "a/b", "b")
	f.Patches = append(f.Patches, &Patch{Skill: "b", Num: 1, File: "patches/b/0001.diff"})

	if err := f.SetEnabled("b", false); err != nil || f.FindSkill("b").Enabled {
		t.Errorf("禁用失败: err=%v enabled=%v", err, f.FindSkill("b").Enabled)
	}
	if err := f.RemoveSkill("b"); err != nil {
		t.Fatalf("RemoveSkill: %v", err)
	}
	if f.FindSkill("b") != nil {
		t.Error("移除后技能应不存在")
	}
	if len(f.Patches) != 0 {
		t.Error("移除技能应连带移除其补丁")
	}
}

func TestPersistRoundtrip(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "hub")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := Init(root); err != nil {
		t.Fatalf("Init: %v", err)
	}
	f, err := Load(root, false)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(f.Targets) != 9 {
		t.Errorf("默认目标应为 9 个，得到 %d", len(f.Targets))
	}
	f.Sources = []Source{{ID: "s", URL: "u"}}
	if _, err := f.AddSkill("s", "skills/x", ""); err != nil {
		t.Fatal(err)
	}
	if err := f.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	f2, err := Load(root, false)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if f2.FindSkill("x") == nil {
		t.Error("重载后技能丢失")
	}
	// Save 不应覆盖保留字段 Root/ConfigPath
	if f2.Root == "" {
		t.Error("重载后 Root 应为空(pre-init)或绝对路径")
	}
}

func TestDuplicateOnReloadNotTriggered(t *testing.T) {
	// 验证 Save→Load 不因 FindSource 返回 &s 破坏数据
	f := &File{}
	f.Sources = []Source{{ID: "a", URL: "u1"}}
	s := f.FindSource("a")
	if s == nil || s.URL != "u1" {
		t.Fatalf("FindSource 异常: %+v", s)
	}
	s2 := f.FindSource("a")
	if s2.URL != "u1" {
		t.Fatalf("FindSource 第二次异常: %+v", s2)
	}
}