# Skill Hub 打包验证与修复报告

- 验证日期：2026-09-15
- 对象：`liuxin557/skills`（Skill Hub · 个人技能中控仓库，Go 单二进制）
- 基准：`docs/specs/2026-09-14-skill-hub-requirements.md`（需求 + grilling 技术方案）
- 验证方式：go vet / go test / go build + 端到端实测（init → add → list → push → enable/disable → patch → update → patch-apply → serve/API）

---

## 一、构建结果

| 项目 | 结果 |
|---|---|
| `go vet ./...` | 通过，无告警 |
| `go test ./...` | 全部通过（registry 4 例 / sync 3 例） |
| `go build ./cmd/skill` | 成功，单二进制 11MB（web 前端已嵌入） |

## 二、规格逐项核对

### 功能需求（需求文档第四节）

| # | 规格要求 | 结论 | 实测证据 |
|---|---|---|---|
| 1 | 技能登记：从本地/远程 git 仓库引入技能，保留目录树，记录来源地址，写入清单 | ✅ 实现 | `add` 嵌套 clone 到 `sources/<id>`（保留 .git），登记后 `skills.yaml` 记录 `source/pathInSource/enabled`；同源多技能可直接重复 `add` |
| 2 | 启用/禁用：禁用后源文件保留，推送时从产品目录移除 | ✅ 实现 | `disable` 后 `push`，各目标副本全部移除，中控 `sources/` 源文件完整保留 |
| 3 | 推送同步：启用技能展平拷贝到全部目标；删除禁用副本；不碰未登记内容 | ✅ 实现 | 多技能 × 多目标拷贝全部命中；目标中的 `orphan-dir` 未被触碰；内容 `diff -r` 一致 |
| 4 | 来源记录：清单记录上游仓库地址 | ✅ 实现 | `sources` 分区含 `id/url/branch/updateRef` |
| 5 | 上游更新：git pull 拉取最新，变更技能重新推送 | ✅ 实现 | `update` fetch + reset 到最新 commit，`updateRef` 更新 |
| 6 | 补丁记录与重放：diff 存入集中补丁库；更新后依序重放，冲突提示人工处理 | ✅ 实现 | `patch` 生成 `patches/<skill>/NNNN-标题.diff` + basestamp + 本地提交；重放自动合并非重叠修改；真冲突保留标记供手动解决 |
| 7 | 网页：本地 Go 服务，展示列表与启停，支持开关并触发同步 | ✅ 实现 | `/api/state` 返回目标/来源/技能/补丁数；`/api/skills/{id}/enable\|disable` 生效；`/api/push`（含 dryRun）可触发 |
| 8 | CLI：list / add / enable / disable / push / update / patch 等 | ✅ 实现 | 全部 10 条命令（含 rm、patch-apply、serve）实测可用 |

### 技术方案决策（第九节）

| 决策 | 结论 |
|---|---|
| YAML 四分区（targets/sources/skills/patches） | ✅ `init` 生成结构正确 |
| 技能 ID = 文件夹名；同名冲突报错 | ✅ `AddSkill` 有冲突检查 |
| 嵌套 clone 到 `sources/<来源>/`，保留 .git | ✅ |
| targets 支持可选 `subPath` | ✅ 实测 pi → `agent/` 子目录 |
| 补丁机制：diff 相对基线、basestamp、依序重放、失败即停 | ✅（冲突路径符合规格） |
| 推送：registry 为唯一事实源，dry-run/verbose，不碰未登记 | ✅ |
| 单二进制 + serve 手动启动 + 写操作文件锁 | ✅ 锁冲突时明确报错并提示 |

### 成功标准（第六节）

| 标准 | 结论 |
|---|---|
| 一次 push 全部目标与清单一致 | ✅ 内容 diff 校验通过 |
| 禁用后目标移除、中控源文件不丢 | ✅ |
| 上游更新 + 补丁重放 = 上游最新 + 个人补丁 | ✅ 非重叠修改自动重放成功（实测上游开头/中间加行场景）；重叠修改按规格交人工合并 |

## 三、修复记录（本次提交）

### 1. 【已修复】推送拷贝的文件权限为 000（严重）

- 现象：`push` 后目标目录下文件权限为 `----------`，产品无法读取。
- 根因：`internal/sync/push.go` 的 `copyDirContents` 用 `e.Type().Perm()` 取权限——`DirEntry.Type()` 只含文件类型位（普通文件为 0），`Perm()` 恒为 000。
- 修复：新增 `permOf(e)` 取 `e.Info().Mode().Perm()`，目录与文件均改用真实权限位。
- 验证：修复后推送文件权限 `0644`，`diff -r` 内容与来源一致。

### 2. 【已修复】补丁重放对上游上下文偏移误判失败

- 现象：上游在补丁 hunk 上下文之外修改（开头/中间加行）时，普通 `git apply` 无法定位 hunk，即使修改与补丁互不重叠。
- 根因：git apply 的 hunk 定位依赖行号与上下文，上游改动导致整体偏移时定位失败。
- 修复：`internal/patch/patch.go` 重放改为 **`git apply --3way` 优先**（按 blob 三方合并，不依赖行号定位）；3way 干净合并则自动完成并提交；3way 判定为真冲突时**保留冲突标记**并明确指引"手动解决后 git add + git commit"（不再回滚丢弃现场）；普通 apply 作为 3way 不可用时的兜底。
- 验证：
  - 上游开头加行 → `update` 自动合并成功：结果 = 上游最新 + 个人补丁，自动提交，工作区干净；
  - 上游在补丁插入点同位置加行（真冲突）→ 冲突标记保留 + 清晰提示，手动解决后 commit 闭环正常；
  - `patch-apply` 同样走新逻辑。

### 3. 【已修复】同一来源登记多个技能无 CLI 入口

- 现象：`skill add <url> --dir skills/xxx --source-id <已存在>` 报"来源已存在"。
- 修复：`internal/cli/commands_basic.go` 中 `add` 在来源已存在时**降级为仅登记技能**（校验技能目录确实存在于该来源克隆目录后登记）；来源不存在时保持"克隆+登记"原行为。
- 验证：连续两次 `add` 同一来源登记 checklist 与 my-todo 均成功。

## 四、结论

项目整体按规格实现，CLI/网页/推送/补丁/更新/锁等核心链路全部可用，构建干净。三处问题（权限 000、补丁重放误判、同源多技能登记）均已修复并经端到端回归验证，成功标准全部达成。
