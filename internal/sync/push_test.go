package sync

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/liuxin/skillhub/internal/registry"
)

// makeHub 构造一个临时中控仓库：root 下 skills 资源、一个目标目录、登记两个技能（一启一禁）。
func makeHub(t *testing.T) (root string, reg *registry.File, srcDir, skillDir, targetDir string) {
	t.Helper()
	base := t.TempDir()
	root = filepath.Join(base, "hub")
	srcDir = filepath.Join(root, "sources", "src1", "skills", "skA")
	skillDirRoot := filepath.Join(root, "sources", "src1", "skills")
	targetDir = filepath.Join(base, "prod-skills")

	// 来源内两个技能目录
	for _, name := range []string{"skA", "skB"} {
		d := filepath.Join(skillDirRoot, name)
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(d, "SKILL.md"), []byte("# "+name+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// 目标内放一个"未登记"目录，验证不被触碰
	if err := os.MkdirAll(filepath.Join(targetDir, "orphan"), 0o755); err != nil {
		t.Fatal(err)
	}
	// 模拟 skB 曾被推送过、现被禁用 → 应从目标删除
	if err := os.MkdirAll(filepath.Join(targetDir, "skB"), 0o755); err != nil {
		t.Fatal(err)
	}

	reg = &registry.File{
		Root: root,
		Targets: []registry.Target{
			{Name: "prod", Path: targetDir},
		},
		Sources: []registry.Source{{ID: "src1", URL: "local"}},
		Skills: []*registry.Skill{
			{ID: "skA", Source: "src1", PathInSource: "skills/skA", Enabled: true},
			{ID: "skB", Source: "src1", PathInSource: "skills/skB", Enabled: false},
		},
	}
	return root, reg, srcDir, skillDirRoot, targetDir
}

func TestRunCopiesEnabledDeletesDisabledKeepsOrphan(t *testing.T) {
	root, reg, _, _, targetDir := makeHub(t)
	_ = root

	steps, err := Run(reg, Options{DryRun: false, Out: &bytes.Buffer{}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	var copies, dels int
	for _, s := range steps {
		switch s.Action {
		case "copy":
			copies++
		case "delete":
			dels++
		}
	}
	if copies != 1 {
		t.Errorf("期望拷贝 1 个（skA），得到 %d", copies)
	}
	if dels != 1 {
		t.Errorf("期望删除 1 个（skB），得到 %d", dels)
	}
	// 断言 skA 已拷贝、skB 已删除、orphan 保留
	if _, err := os.Stat(filepath.Join(targetDir, "skA", "SKILL.md")); err != nil {
		t.Errorf("skA 应被拷贝: %v", err)
	}
	if _, err := os.Stat(filepath.Join(targetDir, "skB")); !os.IsNotExist(err) {
		t.Error("skB 应被删除")
	}
	if _, err := os.Stat(filepath.Join(targetDir, "orphan")); err != nil {
		t.Error("未登记的 orphan 不应被触碰")
	}
}

func TestDryRunDoesNotTouchFs(t *testing.T) {
	_, reg, _, _, targetDir := makeHub(t)
	if _, err := Run(reg, Options{DryRun: true, Out: &bytes.Buffer{}}); err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	if _, err := os.Stat(filepath.Join(targetDir, "skA")); !os.IsNotExist(err) {
		t.Error("dry-run 不应写入文件系统")
	}
}

func TestSubPath(t *testing.T) {
	_, reg, _, _, targetDir := makeHub(t)
	reg.Targets[0].SubPath = "nested"
	if _, err := Run(reg, Options{DryRun: false, Out: &bytes.Buffer{}}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(targetDir, "nested", "skA", "SKILL.md")); err != nil {
		t.Errorf("技能应落入 subPath: %v", err)
	}
}