// Package skillhub 是模块根包，只负责嵌入 web 前端资源以便打进单个二进制。
package skillhub

import "embed"

// FS 暴露中控仓库根下的 web 目录资源。
//
//go:embed web/*
var FS embed.FS
