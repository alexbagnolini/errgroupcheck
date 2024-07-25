package errgroupcheck

import (
	"flag"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

// MessageType describes what should happen to fix the warning.
type MessageType uint8

// List of MessageTypes.
const (
	MessageTypeAdd MessageType = iota + 1
)

// RunningMode describes the mode the linter is run in. This can be either
// native or golangci-lint.
type RunningMode uint8

const (
	RunningModeNative RunningMode = iota
	RunningModeGolangCI
)

// Message contains a message and diagnostic information.
type Message struct {
	Diagnostic  token.Pos
	FixStart    token.Pos
	FixEnd      token.Pos
	LineNumbers []int
	MessageType MessageType
	Message     string
}

// Settings contains settings for edge-cases.
type Settings struct {
	Mode        RunningMode
	RequireWait bool
}

// NewAnalyzer creates a new errgroupcheck analyzer.
func NewAnalyzer(settings *Settings) *analysis.Analyzer {
	if settings == nil {
		settings = &Settings{RequireWait: true}
	}

	return &analysis.Analyzer{
		Name:  "errgroupcheck",
		Doc:   "Checks that each errgroup has Wait called at least once",
		Flags: flags(settings),
		Run: func(p *analysis.Pass) (any, error) {
			return Run(p, settings), nil
		},
	}
}

func flags(settings *Settings) flag.FlagSet {
	flags := flag.NewFlagSet("", flag.ExitOnError)
	flags.BoolVar(&settings.RequireWait, "require-wait", settings.RequireWait, "Check that each errgroup has Wait called at least once")
	return *flags
}

func Run(pass *analysis.Pass, settings *Settings) []Message {
	if !settings.RequireWait {
		return nil
	}

	messages := []Message{}

	for _, file := range pass.Files {
		filename := pass.Fset.Position(file.Pos()).Filename
		if filename == "" || filename[len(filename)-3:] != ".go" {
			continue
		}

		fileMessages := runFile(file, pass.Fset)

		if settings.Mode == RunningModeGolangCI {
			messages = append(messages, fileMessages...)
			continue
		}

		for _, message := range fileMessages {
			pass.Report(analysis.Diagnostic{
				Pos:      message.Diagnostic,
				Category: "errgroupcheck",
				Message:  message.Message,
				SuggestedFixes: []analysis.SuggestedFix{
					{
						TextEdits: []analysis.TextEdit{
							{
								Pos:     message.FixStart,
								End:     message.FixEnd,
								NewText: []byte("errgroup.Wait()"),
							},
						},
					},
				},
			})
		}
	}

	return messages
}

func runFile(file *ast.File, fset *token.FileSet) []Message {
	var messages []Message
	var errGroupVars = map[string]bool{}

	inspect := func(node ast.Node) bool {
		switch stmt := node.(type) {
		case *ast.AssignStmt:
			for _, rhs := range stmt.Rhs {
				if call, ok := rhs.(*ast.CallExpr); ok {
					if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
						if ident, ok := sel.X.(*ast.Ident); ok && (sel.Sel.Name == "WithContext" && ident.Name == "errgroup" || sel.Sel.Name == "Group" && ident.Name == "errgroup") {
							for _, lhs := range stmt.Lhs {
								if varIdent, ok := lhs.(*ast.Ident); ok {
									errGroupVars[varIdent.Name] = false
								}
							}
						}
					}
				}
			}
		case *ast.CallExpr:
			if sel, ok := stmt.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok && sel.Sel.Name == "Wait" {
					if _, exists := errGroupVars[ident.Name]; exists {
						errGroupVars[ident.Name] = true
					}
				}
			}
		}
		return true
	}

	ast.Inspect(file, inspect)

	for varName, hasWait := range errGroupVars {
		if !hasWait {
			for _, decl := range file.Decls {
				if funcDecl, ok := decl.(*ast.FuncDecl); ok {
					if funcDecl.Body != nil {
						for _, stmt := range funcDecl.Body.List {
							ast.Inspect(stmt, func(n ast.Node) bool {
								if ident, ok := n.(*ast.Ident); ok && ident.Name == varName {
									messages = append(messages, Message{
										Diagnostic:  ident.Pos(),
										FixStart:    ident.Pos(),
										FixEnd:      ident.End(),
										LineNumbers: []int{posLine(fset, ident.Pos())},
										MessageType: MessageTypeAdd,
										Message:     "errgroup does not have Wait called",
									})
								}
								return true
							})
						}
					}
				}
			}
		}
	}

	return messages
}

func posLine(fset *token.FileSet, pos token.Pos) int {
	return fset.PositionFor(pos, false).Line
}
