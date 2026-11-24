package scaffold

import (
	"strings"
	"unicode"
)

func isNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return len(s) > 0
}

// ToPascal converts strings like "user-profile", "user_profile", or "user/profile" to "UserProfile".
func ToPascal(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_' || r == ' ' || r == '.' || r == '/' || r == '\\'
	})
	var b strings.Builder
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		runes := []rune(p)
		runes[0] = unicode.ToUpper(runes[0])
		b.WriteString(string(runes))
	}
	if b.Len() == 0 {
		return "Item"
	}
	return b.String()
}

func toPascal(s string) string { return ToPascal(s) }

// ToSlug converts strings like "UserProfile" to "user-profile".
func ToSlug(s string) string {
	var b strings.Builder
	prevDash := false
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 && !prevDash {
				b.WriteRune('-')
			}
			b.WriteRune(unicode.ToLower(r))
			prevDash = false
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
			prevDash = false
		} else if !prevDash {
			b.WriteRune('-')
			prevDash = true
		}
	}
	res := strings.Trim(b.String(), "-")
	if res == "" {
		return "item"
	}
	return res
}

func toSlug(s string) string { return ToSlug(s) }

// ToTitle converts strings like "user-profile" to "User Profile".
func ToTitle(s string) string {
	slug := ToSlug(s)
	words := strings.Split(slug, "-")
	for i, w := range words {
		if len(w) > 0 {
			runes := []rune(w)
			runes[0] = unicode.ToUpper(runes[0])
			words[i] = string(runes)
		}
	}
	return strings.Join(words, " ")
}

func toTitle(s string) string { return ToTitle(s) }
