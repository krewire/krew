package config

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/krewire/krewire/pkg/tui"
	"gopkg.in/yaml.v3"
)

// SchemaURL is where the canonical krewire.yaml JSON Schema is published; it
// is the schema's $id.
const SchemaURL = "https://krewire.com/schemas/krewire.json"

// SchemaHeader is the yaml-language-server modeline scaffolded at the top of
// krewire.yaml so editors validate and complete the file.
const SchemaHeader = "# yaml-language-server: $schema=" + SchemaURL

// SchemaJSON is the canonical JSON Schema (Draft-07) for krewire.yaml.
//
//go:embed krewire.schema.json
var SchemaJSON []byte

// Severity grades a validation Issue.
type Severity string

const (
	// SeverityError marks a value the tooling cannot use (wrong type, enum, range).
	SeverityError Severity = "error"
	// SeverityWarning marks an unknown or deprecated key that is ignored.
	SeverityWarning Severity = "warning"
)

// Issue is one schema violation in krewire.yaml.
type Issue struct {
	Line, Col int
	// Path is the dotted key path, e.g. build.output or gateway.routes[0].port.
	Path     string
	Severity Severity
	Message  string
}

func (i Issue) String() string {
	return fmt.Sprintf("%d:%d: %s: %s", i.Line, i.Col, i.Severity, i.Message)
}

type schemaNode struct {
	Ref                  string                 `json:"$ref"`
	Type                 any                    `json:"type"`
	Properties           map[string]*schemaNode `json:"properties"`
	AdditionalProperties json.RawMessage        `json:"additionalProperties"`
	Items                *schemaNode            `json:"items"`
	Enum                 []any                  `json:"enum"`
	Minimum              *float64               `json:"minimum"`
	Maximum              *float64               `json:"maximum"`
	Pattern              string                 `json:"pattern"`
	Deprecated           bool                   `json:"deprecated"`
	DeprecationMessage   string                 `json:"deprecationMessage"`
	Aliases              []string               `json:"x-aliases"`
	Definitions          map[string]*schemaNode `json:"definitions"`
}

var rootSchema = func() *schemaNode {
	var s schemaNode
	if err := json.Unmarshal(SchemaJSON, &s); err != nil {
		panic("config: invalid embedded krewire.schema.json: " + err.Error())
	}
	return &s
}()

func (s *schemaNode) resolve() *schemaNode {
	for s != nil && s.Ref != "" {
		s = rootSchema.Definitions[strings.TrimPrefix(s.Ref, "#/definitions/")]
	}
	return s
}

// additional returns the schema for keys outside Properties, and whether such
// keys are allowed at all.
func (s *schemaNode) additional() (*schemaNode, bool) {
	raw := strings.TrimSpace(string(s.AdditionalProperties))
	switch raw {
	case "", "true":
		return nil, true
	case "false":
		return nil, false
	}
	var sub schemaNode
	if err := json.Unmarshal(s.AdditionalProperties, &sub); err != nil {
		return nil, true
	}
	return &sub, true
}

// ValidateFile validates the krewire.yaml at path against the schema.
func ValidateFile(path string) ([]Issue, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Validate(data)
}

// Validate checks krewire.yaml content against the schema and returns the
// issues sorted by position. A YAML syntax error is returned as an error
// carrying its line.
func Validate(data []byte) ([]Issue, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("config: parse krewire.yaml: %w", err)
	}
	if len(doc.Content) == 0 {
		return nil, nil
	}
	var issues []Issue
	validateNode(doc.Content[0], rootSchema, "", &issues)
	sort.SliceStable(issues, func(i, j int) bool {
		if issues[i].Line != issues[j].Line {
			return issues[i].Line < issues[j].Line
		}
		return issues[i].Col < issues[j].Col
	})
	return issues, nil
}

// HasErrors reports whether any issue is an error.
func HasErrors(issues []Issue) bool {
	for _, i := range issues {
		if i.Severity == SeverityError {
			return true
		}
	}
	return false
}

func validateNode(n *yaml.Node, s *schemaNode, path string, out *[]Issue) {
	s = s.resolve()
	if n == nil || s == nil {
		return
	}
	if n.Kind == yaml.AliasNode {
		n = n.Alias
	}
	add := func(sev Severity, format string, args ...any) {
		*out = append(*out, Issue{Line: n.Line, Col: n.Column, Path: path, Severity: sev, Message: fmt.Sprintf(format, args...)})
	}
	if n.Kind == yaml.ScalarNode && n.Tag == "!!null" {
		return // an empty value leaves the setting unset
	}
	got := nodeType(n)
	if want := typeList(s.Type); len(want) > 0 && !typeAllowed(got, want) {
		add(SeverityError, "%s must be %s, got %s", displayPath(path), strings.Join(want, " or "), got)
		return
	}
	switch n.Kind {
	case yaml.MappingNode:
		validateMapping(n, s, path, out)
	case yaml.SequenceNode:
		for i, item := range n.Content {
			validateNode(item, s.Items, fmt.Sprintf("%s[%d]", path, i), out)
		}
	case yaml.ScalarNode:
		if len(s.Enum) > 0 && !inEnum(n.Value, s.Enum) {
			add(SeverityError, "%s: %q is not one of %s", displayPath(path), n.Value, enumList(s.Enum))
		}
		if v, err := strconv.ParseFloat(n.Value, 64); err == nil && (got == "integer" || got == "number") {
			if (s.Minimum != nil && v < *s.Minimum) || (s.Maximum != nil && v > *s.Maximum) {
				add(SeverityError, "%s: %s is out of range %s", displayPath(path), n.Value, rangeText(s))
			}
		}
		if s.Pattern != "" && got == "string" {
			if re, err := regexp.Compile(s.Pattern); err == nil && !re.MatchString(n.Value) {
				add(SeverityError, "%s: %q does not match %s", displayPath(path), n.Value, s.Pattern)
			}
		}
	}
}

func validateMapping(n *yaml.Node, s *schemaNode, path string, out *[]Issue) {
	extra, allowExtra := s.additional()
	for i := 0; i+1 < len(n.Content); i += 2 {
		k, v := n.Content[i], n.Content[i+1]
		child := joinPath(path, k.Value)
		if ps, ok := s.Properties[k.Value]; ok {
			if r := ps.resolve(); ps.Deprecated || (r != nil && r.Deprecated) {
				msg := ps.DeprecationMessage
				if msg == "" && r != nil {
					msg = r.DeprecationMessage
				}
				*out = append(*out, Issue{Line: k.Line, Col: k.Column, Path: child, Severity: SeverityWarning, Message: fmt.Sprintf("%s is deprecated. %s", child, msg)})
			}
			validateNode(v, ps, child, out)
			continue
		}
		if extra != nil {
			validateNode(v, extra, child, out)
			continue
		}
		if allowExtra {
			continue
		}
		msg := fmt.Sprintf("unknown key %q", k.Value)
		if path != "" {
			msg += " in " + path
		}
		if sug := suggestKey(k.Value, s, path); sug != "" {
			msg += fmt.Sprintf("; did you mean %q?", sug)
		}
		*out = append(*out, Issue{Line: k.Line, Col: k.Column, Path: child, Severity: SeverityWarning, Message: msg + " (ignored)"})
	}
}

// suggestKey proposes the key an unknown key most likely meant: a sibling
// whose x-aliases list it, a sibling within a small edit distance, or the
// same key one level down (port -> dev.port).
func suggestKey(key string, s *schemaNode, path string) string {
	names := make([]string, 0, len(s.Properties))
	for name, ps := range s.Properties {
		names = append(names, name)
		for _, a := range ps.Aliases {
			if strings.EqualFold(a, key) {
				return joinPath(path, name)
			}
		}
	}
	sort.Strings(names)
	norm := strings.NewReplacer("-", "_").Replace(strings.ToLower(key))
	for _, name := range names {
		if strings.ReplaceAll(norm, "_", "") == strings.ReplaceAll(name, "_", "") {
			return joinPath(path, name)
		}
	}
	if sug := tui.Suggest(key, names); sug != "" {
		return joinPath(path, sug)
	}
	for _, name := range names {
		if child := s.Properties[name].resolve(); child != nil {
			if _, ok := child.Properties[key]; ok {
				return joinPath(joinPath(path, name), key)
			}
		}
	}
	return ""
}

func nodeType(n *yaml.Node) string {
	switch n.Kind {
	case yaml.MappingNode:
		return "object"
	case yaml.SequenceNode:
		return "array"
	}
	switch n.Tag {
	case "!!int":
		return "integer"
	case "!!float":
		return "number"
	case "!!bool":
		return "boolean"
	}
	return "string"
}

func typeList(t any) []string {
	switch v := t.(type) {
	case string:
		return []string{v}
	case []any:
		var out []string
		for _, x := range v {
			if s, ok := x.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func typeAllowed(got string, want []string) bool {
	for _, w := range want {
		if w == got || (w == "number" && got == "integer") {
			return true
		}
	}
	return false
}

func inEnum(v string, enum []any) bool {
	for _, e := range enum {
		if fmt.Sprint(e) == v {
			return true
		}
	}
	return false
}

func enumList(enum []any) string {
	parts := make([]string, len(enum))
	for i, e := range enum {
		parts[i] = fmt.Sprint(e)
	}
	return strings.Join(parts, ", ")
}

func rangeText(s *schemaNode) string {
	lo, hi := "-inf", "+inf"
	if s.Minimum != nil {
		lo = strconv.FormatFloat(*s.Minimum, 'f', -1, 64)
	}
	if s.Maximum != nil {
		hi = strconv.FormatFloat(*s.Maximum, 'f', -1, 64)
	}
	return lo + ".." + hi
}

func joinPath(path, key string) string {
	if path == "" {
		return key
	}
	return path + "." + key
}

func displayPath(path string) string {
	if path == "" {
		return "krewire.yaml"
	}
	return path
}
