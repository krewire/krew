package memory

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/krewire/krewire/pkg/ai"
)

// sessionIDPattern constrains session identifiers to filesystem-safe names so
// the JSONL path can never escape the sessions directory.
var sessionIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// SessionRecord is one JSONL transcript line: a message with its append time.
type SessionRecord struct {
	Time    int64      `json:"time"` // Unix nanoseconds, UTC
	Message ai.Message `json:"message"`
}

// SessionInfo summarises a stored transcript for session listings.
type SessionInfo struct {
	ID       string    `json:"id"`
	ModTime  time.Time `json:"mod_time"`
	Messages int       `json:"messages"`
	Tokens   int       `json:"tokens"`
}

// DefaultSessionsDir resolves ~/.krew/sessions, the canonical transcript
// storage location.
func DefaultSessionsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".krew", "sessions"), nil
}

// SessionLog appends conversation turns to a JSON Lines transcript file.
type SessionLog struct {
	path string
	file *os.File
}

// OpenSessionLog opens (creating when needed) the transcript for sessionID
// under dir in append mode. Callers must Close the log; every Append is
// flushed to disk immediately.
func OpenSessionLog(dir, sessionID string) (*SessionLog, error) {
	path, err := sessionFilePath(dir, sessionID)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create sessions directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open session transcript: %w", err)
	}
	return &SessionLog{path: path, file: file}, nil
}

// Append writes one JSONL line per message and flushes both to disk.
func (l *SessionLog) Append(msgs ...ai.Message) error {
	if l == nil || l.file == nil {
		return fmt.Errorf("session log is closed")
	}
	now := time.Now().UTC().UnixNano()
	for _, msg := range msgs {
		line, err := json.Marshal(SessionRecord{Time: now, Message: msg})
		if err != nil {
			return fmt.Errorf("encode session record: %w", err)
		}
		if _, err := l.file.Write(append(line, '\n')); err != nil {
			return fmt.Errorf("write session record: %w", err)
		}
	}
	return l.file.Sync()
}

// Path returns the transcript file location.
func (l *SessionLog) Path() string {
	if l == nil {
		return ""
	}
	return l.path
}

// Close flushes and releases the transcript file. Append after Close fails.
func (l *SessionLog) Close() error {
	if l == nil || l.file == nil {
		return nil
	}
	err := l.file.Close()
	l.file = nil
	return err
}

// LoadSession replays a stored transcript as conversation messages in FIFO
// order. Well-formed records are expected; a corrupt line aborts loading.
func LoadSession(dir, sessionID string) ([]ai.Message, error) {
	path, err := sessionFilePath(dir, sessionID)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("session %q not found", sessionID)
		}
		return nil, fmt.Errorf("open session transcript: %w", err)
	}
	defer file.Close()

	var msgs []ai.Message
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 1<<20), 16*1<<20) // tolerate large tool outputs
	line := 0
	for scanner.Scan() {
		line++
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		var record SessionRecord
		if err := json.Unmarshal([]byte(text), &record); err != nil {
			return nil, fmt.Errorf("corrupt session record at line %d: %w", line, err)
		}
		msgs = append(msgs, record.Message)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read session transcript: %w", err)
	}
	return msgs, nil
}

// ListSessions summarises every stored transcript under dir, newest first.
func ListSessions(dir string) ([]SessionInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("list sessions directory: %w", err)
	}

	infos := make([]SessionInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".jsonl" {
			continue
		}
		msgs, err := LoadSession(dir, strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())))
		if err != nil {
			continue // unreadable transcripts are skipped, not fatal
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		infos = append(infos, SessionInfo{
			ID:       strings.TrimSuffix(entry.Name(), ".jsonl"),
			ModTime:  info.ModTime().UTC(),
			Messages: len(msgs),
			Tokens:   tokensFor(msgs),
		})
	}
	sort.Slice(infos, func(i, j int) bool { return infos[i].ModTime.After(infos[j].ModTime) })
	return infos, nil
}

func sessionFilePath(dir, sessionID string) (string, error) {
	if dir == "" {
		return "", fmt.Errorf("sessions directory is required")
	}
	if !sessionIDPattern.MatchString(sessionID) {
		return "", fmt.Errorf("invalid session id %q: use letters, digits, dots, dashes, or underscores", sessionID)
	}
	return filepath.Join(dir, sessionID+".jsonl"), nil
}

// sessionNamePattern matches the slug summary derived from a user prompt.
var sessionNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

// dashRun collapses runs of separator characters emitted while slugging a
// user prompt into a session name.
var dashRun = regexp.MustCompile(`-{2,}`)

// maxSessionNameLen bounds auto-derived session names derived from prompts.
const maxSessionNameLen = 48

// SessionName derives a filesystem-safe transcript name from the first user
// prompt: lowercased slug words joined with dashes, capped in length.
func SessionName(prompt string) string {
	clean := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			return r
		case r >= 'A' && r <= 'Z':
			return r + ('a' - 'A')
		default:
			return '-'
		}
	}, prompt)
	clean = dashRun.ReplaceAllString(clean, "-")
	clean = strings.Trim(clean, "-")
	if clean == "" {
		return "session"
	}
	if len(clean) > maxSessionNameLen {
		if cut := strings.LastIndex(clean[:maxSessionNameLen], "-"); cut > 0 {
			clean = clean[:cut]
		} else {
			clean = clean[:maxSessionNameLen]
		}
		clean = strings.Trim(clean, "-")
	}
	if !sessionNamePattern.MatchString(clean) {
		return "session"
	}
	return clean
}

// StartSession derives a unique transcript name from the first user prompt
// (slug summary with -2, -3… suffixes on collision) and opens its log. The
// caller must Close the returned log.
func StartSession(dir, prompt string) (*SessionLog, string, error) {
	base := SessionName(prompt)
	name := base
	for n := 2; ; n++ {
		if _, err := os.Stat(filepath.Join(dir, name+".jsonl")); err != nil && !os.IsNotExist(err) {
			return nil, "", fmt.Errorf("probe session file: %w", err)
		} else if err == nil {
			name = fmt.Sprintf("%s-%d", base, n)
			continue
		}
		break
	}
	log, err := OpenSessionLog(dir, name)
	if err != nil {
		return nil, "", err
	}
	return log, name, nil
}
