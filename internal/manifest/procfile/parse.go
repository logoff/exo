package procfile

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/syntax"
)

type Procfile struct {
	Processes []Process
}

type Process struct {
	Name        string
	Program     string
	Arguments   []string
	Environment map[string]string
}

func Parse(r io.Reader) (*Procfile, error) {
	var procfile Procfile
	br := bufio.NewReaderSize(r, 4096)
	lineIndex := 0
	for {
		lineIndex++
		line, isPrefix, err := br.ReadLine()
		if io.EOF == err {
			break
		}
		if isPrefix {
			return nil, fmt.Errorf("line %d is too long", lineIndex)
		}
		line = bytes.TrimSpace(line)
		if len(line) == 0 || bytes.HasPrefix(line, []byte("#")) {
			// Blank or comment.
			continue
		}
		parts := bytes.SplitN(line, []byte(":"), 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("line %d is invalid", lineIndex)
		}
		for i, part := range parts {
			parts[i] = bytes.TrimSpace(part)
		}
		name := string(parts[0]) // TODO: Validate name is alphanumeric.
		process, err := ParseCommand(bytes.NewReader(parts[1]))
		if err != nil {
			return nil, err
		}
		process.Name = name
		procfile.Processes = append(procfile.Processes, *process)
	}
	return &procfile, nil
}

func ParseCommand(r io.Reader) (*Process, error) {
	parser := syntax.NewParser(syntax.Variant(syntax.LangBash))
	name := ""
	file, err := parser.Parse(r, name)
	if err != nil {
		return nil, err
	}
	if len(file.Stmts) != 1 {
		return nil, errors.New("expected exactly one bash statement")
	}

	stmt := file.Stmts[0]
	if len(stmt.Comments) > 0 {
		return nil, errors.New("unexpected comments")
	}
	if stmt.Semicolon.IsValid() {
		return nil, fmt.Errorf("unsupported syntax at column %d", stmt.Semicolon.Col())
	}
	if stmt.Negated {
		return nil, fmt.Errorf("unsupported: command negation")
	}
	if len(stmt.Redirs) > 0 {
		return nil, errors.New("unsupported: redirection")
	}

	call, ok := stmt.Cmd.(*syntax.CallExpr)
	if !ok {
		return nil, fmt.Errorf("expected simple command, got %T", stmt.Cmd)
	}

	process := Process{
		Environment: make(map[string]string),
	}

	// Parse environment variable assignments.
	for _, assign := range call.Assigns {
		name := assign.Name.Value
		if assign.Append || assign.Naked || assign.Index != nil || assign.Array != nil {
			return nil, fmt.Errorf("unsupported assignment for %q", name)
		}
		value, err := wordToString(assign.Value)
		if err != nil {
			return nil, fmt.Errorf("parsing %q assignment: %w", name, err)
		}
		process.Environment[name] = value
	}

	// Parse program.
	if len(call.Args) < 1 {
		return nil, errors.New("expected program path")
	}
	process.Program, err = wordToString(call.Args[0])
	if err != nil {
		return nil, fmt.Errorf("parsing program path: %w", err)
	}

	// Parse arguments.
	argWords := call.Args[1:]
	process.Arguments = make([]string, len(argWords))
	for i, argWord := range argWords {
		process.Arguments[i], err = wordToString(argWord)
		if err != nil {
			return nil, fmt.Errorf("parsing argument %d: %w", i, err)
		}
	}
	return &process, nil
}

func wordToString(word *syntax.Word) (string, error) {
	strs, err := wordsToStrings([]*syntax.Word{word})
	if err != nil {
		return "", err
	}
	return strings.Join(strs, " "), nil
}

func wordsToStrings(words []*syntax.Word) ([]string, error) {
	return expand.Fields(expandConfig, words...)
}

var expandConfig = &expand.Config{
	Env: &emptyEnv{},
	CmdSubst: func(w io.Writer, cs *syntax.CmdSubst) error {
		return errors.New("unsupported: command substitution")
	},
	ProcSubst: func(ps *syntax.ProcSubst) (string, error) {
		return "", errors.New("unsupported: process substitution")
	},
	ReadDir: func(s string) ([]os.FileInfo, error) {
		return nil, errors.New("unsupported: glob patterns")
	},
	GlobStar: false,
	NullGlob: false,
	NoUnset:  true,
}

type emptyEnv struct{}

func (_ *emptyEnv) Get(name string) expand.Variable {
	return expand.Variable{}
}

func (_ *emptyEnv) Each(f func(name string, vr expand.Variable) bool) {
	// no-op.
}
