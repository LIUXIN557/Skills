#!/usr/bin/env bash
# restore-sources.sh — 从 bundles/ 恢复各来源的 git 仓库（克隆即全量 + 可 update/patch-apply）
# 用法: bash scripts-restore-sources.sh [source-id ...]   （不带参数 = 恢复全部缺失来源）
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT"

# 来源 -> 上游 remote 地址
declare -A REMOTES=(
  [superpowers]="https://github.com/obra/superpowers.git"
  [mattpocock]="https://github.com/mattpocock/skills.git"
  [ponytail]="https://github.com/DietrichGebert/ponytail.git"
  [eli5]="https://github.com/dreambigou/eli5.git"
  [go-modern-guidelines]="https://github.com/JetBrains/go-modern-guidelines.git"
  [rudder]="https://github.com/rudderlabs/rudder-agent-skills.git"
  [humanizer-zh]="https://github.com/op7418/Humanizer-zh.git"
  [humanizer]="https://github.com/blader/humanizer.git"
  [doubao-product-qa]="https://github.com/LIUXIN557/doubao-skills.git"
)

if [ $# -eq 0 ]; then ids=("${!REMOTES[@]}"); else ids=("$@"); fi

for id in "${ids[@]}"; do
  bundle="bundles/$id.bundle"
  target="sources/$id"
  if [ ! -f "$bundle" ]; then echo "跳过 $id: 无 bundle 文件 $bundle"; continue; fi
  if [ -d "$target/.git" ]; then echo "跳过 $id: git 仓库已存在"; continue; fi
  if [ -d "$target" ] && [ -n "$(ls -A "$target" 2>/dev/null)" ]; then
    echo "跳过 $id: sources/$id 非空且无 .git（可能是普通源码副本，如需恢复请先移走）"; continue
  fi
  echo "== 恢复 $id =="
  rm -rf "$target"
  git clone -q "$bundle" "$target"
  git -C "$target" remote remove origin 2>/dev/null || true
  git -C "$target" remote add origin "${REMOTES[$id]}"
  git -C "$target" config user.name "LIUXIN557"
  git -C "$target" config user.email "liuxin557@users.noreply.github.com"
  echo "   $id 已恢复（remote=origin -> ${REMOTES[$id]}）"
done
echo "完成。可选：./skill update <id> 对齐清单版本并重放补丁"
