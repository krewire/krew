// Package dsl is the historical import path for the `.kiw` module parser.
//
// The parser now lives in github.com/krewire/krewire/pkg/kiw. It had to move: the
// SSG loader needs to parse a .kiw file, and libraries must never import the devtool
// that sits above them in the layering (boost → kiw → pkg/* → kern). Keeping the
// parser in pkg/kiw points the dependency arrow down and lets ui and web parse
// .kiw files without a cycle.
//
// This package forwards to the canonical home so existing imports keep working. New
// code should import github.com/krewire/krewire/pkg/kiw directly.
package dsl

import kiw "github.com/krewire/krewire/pkg/kiw"

// KiwModule is the parsed result of a .kiw file.
type KiwModule = kiw.KiwModule

// StyleBlock holds a scoped style block with its attributes.
type StyleBlock = kiw.StyleBlock

// ScriptBlock holds a script block with its tier attributes.
type ScriptBlock = kiw.ScriptBlock

// ComponentCall records a parsed component invocation in a .kiw template.
type ComponentCall = kiw.ComponentCall

// DesugarTemplate translates JSX-like component tags into Go template
// component invocations.
func DesugarTemplate(src string) (string, error) { return kiw.DesugarTemplate(src) }

// ParseKiw parses a .kiw module from source.
func ParseKiw(src string) (*KiwModule, error) { return kiw.ParseKiw(src) }

// ParseKiwFile parses a .kiw module from a file path.
func ParseKiwFile(path string) (*KiwModule, error) { return kiw.ParseKiwFile(path) }
