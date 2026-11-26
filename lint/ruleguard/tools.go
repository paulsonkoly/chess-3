//go:build tools

// Package tools keeps the ruleguard DSL dependency in go.mod for golangci-lint.
package tools

import _ "github.com/quasilyte/go-ruleguard/dsl"
