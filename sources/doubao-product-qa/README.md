# doubao-skills

**doubao-product-qa** —— AI Agent 的 QA 测试负责人技能（QA Test Lead Skill）。

把 PRD、原型、网页、接口、代码、测试记录和多轮上下文转成**可追踪的 QA 基线、风险用例、执行证据、Bug 与发布判断**。适用于 Web/API/App/小程序测试、回归、热修复、测试方案、Bug 复核和 QA 收口。

## 特性

- **受控流程**：`bootstrap → anchor → complete → qa_deliver` 四阶段，`qa-run.json` 是唯一事实源
- **门禁质量**：BLOCK / FIX / REPORT 三级门策略，未过 BLOCK 禁止进入下一阶段
- **发布判断**：`go / conditional_go / no_go / undetermined` 四种结论，零证据禁终审
- **多载体交付**：Markdown、豆包文档/表格/PPT、飞书载体（默认 lark_doc）
- **纯 Python 实现**：无需编译，依赖 Python 3.8+

## 目录结构

```
doubao-skills/
├── SKILL.md            # 技能入口（Agent 读取）
├── bin/                # CLI 包装器
│   ├── qa-flow         # qa_flow.py 控制器
│   └── qa-deliver      # qa_deliver.py 交付口
├── scripts/            # 24 个阶段脚本（qa_flow / qa_deliver / gate_policy ...）
├── references/         # 20 个方法论文档（覆盖模型、门策略、体裁、载体映射）
├── assets/             # 模板（bug-report / test-cases / traceability / qa-run.schema ...）
├── agents/             # 子 Agent 编排
└── tests/              # 回归测试
```

## 安装

### 作为 npm 包（CLI 方式）

```bash
npm i -g doubao-skills
# 之后可用
qa-flow --help
qa-deliver <qa-run.json>
```

### 作为 Agent 技能（推荐）

把本仓库 clone / add 进你的技能库（如 [Skill Hub](https://github.com/LIUXIN557/Skills)）：

```bash
skill add https://github.com/LIUXIN557/doubao-skills.git --source-id doubao-product-qa --dir . --skill-id doubao-product-qa
```

Agent 读取 `SKILL.md` 后按流程执行即可，CLI 仅用于阶段控制。

## 快速开始

```bash
# 1. 建立请求契约（--execute 授权真实执行）
python3 scripts/qa_flow.py bootstrap qa-results/<feature>/qa-run.json \
  --request "<目标与限制摘要>" \
  --target "<测试对象>" \
  --source "<文件/目录/URL>" \
  --execute

# 2. 锚定阶段，读取上下文卡与报告体裁卡
python3 scripts/qa_flow.py anchor qa-results/<feature>/qa-run.json --stage baseline

# 3. 填充 REQ/RM/用例/执行证据后，逐阶段过关
python3 scripts/qa_flow.py complete qa-results/<feature>/qa-run.json --stage baseline
python3 scripts/qa_flow.py complete qa-results/<feature>/qa-run.json --stage design
python3 scripts/qa_flow.py complete qa-results/<feature>/qa-run.json --stage execution

# 4. 交付（唯一交付口，生成结构化交付清单）
python3 scripts/qa_deliver.py qa-results/<feature>/qa-run.json
```

## 核心规则

- `qa-run.json` 是唯一 canonical；上屏只能通过 `qa_deliver.py`
- 只有 `DELIVERY_LOCK=CLOSED` 才算交付完成
- 正式 Bug ≥ L2 观察；严重度只用 S1–S4、优先级只用 P0–P3；P0 必进发布阻断
- 未执行就有未验证范围，`unverified` 不得为空

## License

MIT
