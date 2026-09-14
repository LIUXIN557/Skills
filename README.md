# Skill Hub · 个人技能中控仓库

在多家 AI 产品（Trae、Codex、Grok、Qoder、Poe/Pi、豆包等）之间维护同一套技能，用一个 Go 工具统一引入、启用/禁用、打补丁并一键推送到各产品的技能目录。

## 核心概念

- **中控仓库**：本仓库存放集中清单 `skills.yaml`（唯一事实源）、来源克隆目录 `sources/` 与补丁目录 `patches/`。
- **来源**：从某个 git 仓库（远程或本地路径）嵌套克隆进来，保留上游历史，可 `update` 拉取最新。
- **技能**：来源里的一个子目录（如 `skills/checklist`），登记后默认启用。
- **补丁**：你对技能的个性化改动，以 `git diff` 相对基线固化到 `patches/<skill>/`，上游更新后自动依序重放。
- **目标**：各产品技能目录，推送时把启用技能展平拷贝到目标、删除禁用技能，不碰未登记内容。

## 快速开始

```bash
# 构建
go build ./cmd/skill

# 初始化清单（含默认目标产品目录）
skill init

# 引入一个上游来源并登记技能
skill add <git地址|本地路径> --dir skills/checklist --source-id upstream

# 查看、启停
skill list
skill enable checklist   # skill disable checklist

# 推送到所有产品（--dry-run 预览）
skill push              # 或 skill push --dry-run

# 打补丁：先修改 sources/<id> 下技能文件，再
skill patch checklist --title "加个人条目"

# 上游更新 + 重放补丁
skill update upstream

# 本地网页管理
skill serve             # 打开 http://127.0.0.1:8787
```

## CLI 命令

| 命令 | 说明 |
|---|---|
| `init` | 生成 `skills.yaml`（默认 9 个目标产品目录） |
| `list` | 列出技能与启用状态 |
| `add <url> --dir <path> [--source-id] [--branch] [--skill-id]` | 克隆来源并登记技能 |
| `rm <skill-id>` | 移除技能登记（不动源文件） |
| `enable / disable <skill-id>` | 切换启用开关 |
| `push [--dry-run] [--verbose]` | 同步到所有目标产品目录 |
| `update <source-id> [--no-replay]` | 拉取上游并重放补丁 |
| `patch <skill-id> --title <描述>` | 生成补丁（diff + 基线）并本地提交 |
| `patch-apply <skill-id>` | 重置到基线并重放该技能补丁 |
| `serve [--port]` | 启动本地网页 |

## 配置：skills.yaml（四分区）

`targets` 各产品目录（可选 `subPath`）/ `sources` 来源 / `skills` 技能条目 / `patches` 补丁。

## 设计文档

- 需求与技术方案：[docs/specs/2026-09-14-skill-hub-requirements.md](docs/specs/2026-09-14-skill-hub-requirements.md)