# Skill Hub 实施计划

## Context（背景）

个人在多家 AI 产品（Trae、Codex、Grok、Qoder、Poe/Pi、豆包等）间维护同一套技能，靠手工复制/开关易失控。需求与技术方案已通过 brainstorming + grilling 双阶段确认并落档：
- 需求文档：`docs/specs/2026-09-14-skill-hub-requirements.md`
- 已敲定技术方案见该文档「九、技术方案」节

本仓库当前为空（仅 README 与 specs）。本计划从零搭建一个 **Go 单二进制** 工具 `skill`，实现：集中清单（YAML 四分区）作唯一事实源，从上游 git 来源嵌套 clone 引入技能，支持个性化补丁（git diff 相对基线 + 依序重放），一键推送到 9 个产品技能目录（覆盖启用 / 删除禁用 / dry-run），并提供本地网页管理与 CLI 等价命令。本机 Go 1.26.2、git 2.52 已就绪。

## 项目结构

```
c:\dev\skills\
├── go.mod                        # module github.com/liuxin/skillhub
├── cmd\skill\main.go             # 入口：cobra 根命令
├── internal\
│   ├── registry\                 # 清单模型与读写（唯一事实源）
│   │   ├── model.go              #   YAML schema（Root/Config, Targets, Sources, Skills, Patches）
│   │   ├── store.go              #   加载/保存 skills.yaml + 文件锁
│   │   └── skills.go             #   增删/开关/查询 skill 条目
│   ├── sources\git.go            #   嵌套 clone（add）/ git pull（update）
│   ├── patch\                    #   补丁生成与应用
│   │   ├── create.go             #   git diff 相对基线 → patches/<skill>/NNN-描述.diff，记录 basestamp
│   │   └── replay.go             #   干净工作区依序 git apply，失败即停交人工
│   ├── sync\push.go              #   覆盖启用 + 删除禁用 + dry-run/verbose + subPath
│   ├── server\server.go          #   本地 web（net/http + go:embed + JSON API）
│   └── lock\lock.go              #   跨命令清单文件锁
├── web\index.html                # 网页（go:embed 打包进二进制）
└── docs\specs\...                 # 已存在
```

## 关键设计决策

- **依赖最小化**：CLI 用 `github.com/spf13/cobra`，YAML 用 `gopkg.in/yaml.v3`。git 操作直接调本机 `git`（`os/exec`，避免 go-git 重型依赖），复用已安装 git。
- **清单路径**：命令统一 `-r/--root`（默认当前目录，即 c:\dev\skills），root 下必须有 `skills.yaml`。
- **文件锁**：`skills.yaml.lock` 用 `O_CREATE|O_EXCL` 独占创建，防止并发写坏清单；用完删除。
- **目录树/展平**：`skills` 条目含 `source` + `path_in_source`（来源内相对路径），id=技能文件夹名；推送时把该子目录整体拷贝展平到目标。target 支持可选 `subPath`。
- **补丁流**：用户改的是来源嵌套仓库工作区 → `patch` 用 `git diff basestamp..工作区` 生成 diff，`basestamp`=上次同步的上游 commit，记入 `patches`；`update` = fetch 上游 + `git apply` 依序重放 + 更新 basestamp，全成功才算完成。
- **推送流**（`sync/push.go`）：遍历启用 skills → 拷贝覆盖；遍历已登记但禁用的 skills → 删目标对应目录；`--dry-run` 只打印计划；绝不触碰未登记内容。

## CLI 命令面

| 命令 | 作用 |
|---|---|
| `skill list` | 列出技能与启用状态 |
| `skill add <source> --dir <path_in_source> [--id 名]` | clone 上游到 `sources/`，登记技能，默认启用 |
| `skill rm <id>` | 移除登记（不动 sources） |
| `skill enable/disable <id>` | 切换开关 |
| `skill push [--dry-run] [--verbose]` | 同步到所有 target |
| `skill update <source> [--dry-run]` | fetch 上游 + 重放补丁 |
| `skill patch <id> --title 描述` | 生成补丁 diff 并记录 basestamp |
| `skill patch-apply <id>` | 干净工作区重放补丁 |
| `skill serve [--port]` | 启动本地网页需由用户授权，非守护 |

## 实施顺序

1. **骨架**：`go mod init`、cobra 根命令、空清单文件初始化、`lock` 包。
2. **registry 模型 + 读写 + CRUD**：`model.go`/`store.go`/`skills.go`，实现 `list/add/rm/enable/disable` 与 `skills.yaml` 增删改。含同名冲突校验。
3. **sources**：`git.go` 的 clone/pull，`add` 接入。
4. **patch**：`create.go`（git diff 相对基线 + basestamp 记录）、`replay.go`（git apply 依序 + 冲突停止）。
5. **sync/push**：覆盖启用 + 删除禁用 + dry-run + verbose + `subPath`。
6. **server + web**：net/http JSON API + `web/index.html`（技能列表 + 开关 + 触发 push + 查看来源/补丁），`go:embed` 打包。

## 验证方式

- **构建**：`go build ./...` 与 `go vet ./...` 零报错。
- **单元测试**：`registry`（CRUD/冲突）与 `sync/push`（覆盖/删除/dry-run，对着临时沙盒目录）各若干用例，`go test ./...` 通过。
- **端到端手测**：建一个临时沙盒 hub（含 `skills.yaml` + 一个本地假来源仓库），依次跑 `add → list → enable/disable → push --dry-run → push`，核对目标目录出现/移除对应技能、未登记内容原样保留；再改来源文件后 `patch` 生成 diff，模拟上游提交后 `update` 验证补丁被重放。
- **网页**：`skill serve` 起本地服务，浏览器验证列表/开关/触发推送。

## 不在本次范围（YAGNI）

不做文件监听自动同步、不做逐产品开关、不做目标完整镜像、不做多用户/多机、网页内不做增删/更新（走 CLI）。