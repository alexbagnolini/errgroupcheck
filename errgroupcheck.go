package errgroupcheck

import (
	"flag"
	"fmt"
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
	RequireWait bool
}

func DefaultSettings() *Settings {
	return &Settings{
		RequireWait: true,
	}
}

// NewAnalyzer creates a new errgroupcheck analyzer.
func NewAnalyzer(settings *Settings) *analysis.Analyzer {
	if settings == nil {
		settings = DefaultSettings()
	}

	return &analysis.Analyzer{
		Name:  "errgroupcheck",
		Doc:   "Checks that each errgroup has Wait called at least once",
		Flags: flags(settings),
		Run: func(p *analysis.Pass) (any, error) {
			return Run(p, settings)
		},
	}
}

func flags(settings *Settings) flag.FlagSet {
	flags := flag.NewFlagSet("", flag.ExitOnError)
	flags.BoolVar(&settings.RequireWait, "require-wait", settings.RequireWait, "Check that each errgroup has Wait called at least once")
	return *flags
}

func Run(pass *analysis.Pass, settings *Settings) (interface{}, error) {
	if !settings.RequireWait {
		return nil, nil
	}

	for _, file := range pass.Files {
		filename := pass.Fset.Position(file.Pos()).Filename
		if filename == "" || filename[len(filename)-3:] != ".go" {
			continue
		}

		runFile(pass, file)
	}

	return nil, nil
}

type errgroupVar struct {
	waitCalled bool
	ident      *ast.Ident
}

type Scope struct {
	vars map[string]*errgroupVar
}

type ScopeStack struct {
	stack []*Scope
}

func NewScopeStack() *ScopeStack {
	stack := &ScopeStack{stack: []*Scope{}}
	stack.Push()
	return stack
}

func (s *ScopeStack) Push() {
	s.stack = append(s.stack, &Scope{vars: make(map[string]*errgroupVar)})
}

func (s *ScopeStack) Pop() *Scope {
	scope := s.Current()
	s.stack = s.stack[:len(s.stack)-1]
	return scope
}

func (s *ScopeStack) Current() *Scope {
	return s.stack[len(s.stack)-1]
}

func (s *ScopeStack) FindVar(name string) *errgroupVar {
	for i := len(s.stack) - 1; i >= 0; i-- {
		if v, ok := s.stack[i].vars[name]; ok {
			return v
		}
	}
	return nil
}

func (s *ScopeStack) AddVar(name string, v *errgroupVar) {
	s.Current().vars[name] = v
}

func runFile(pass *analysis.Pass, file *ast.File) {
	scopes := NewScopeStack()

	var inspectNode func(node ast.Node) bool
	inspectScoped := func(node ast.Node) {
		// New function scope
		scopes.Push()
		ast.Inspect(node, inspectNode)
		scope := scopes.Pop()

		for varName, varData := range scope.vars {
			if !varData.waitCalled {
				pass.Reportf(varData.ident.Pos(), fmt.Sprintf("errgroup '%s' does not have Wait called", varName))
			}
		}
	}

	inspectNode = func(node ast.Node) bool {
		switch stmt := node.(type) {
		case *ast.FuncDecl:
			if stmt.Body != nil {
				inspectScoped(stmt.Body) // Handle function declaration scope
			}
			return false // Stop inspecting this branch, as it has been handled

		case *ast.FuncLit:
			inspectScoped(stmt.Body) // Handle function literal scope
			return false

		case *ast.AssignStmt:
			for _, rhs := range stmt.Rhs {
				switch expr := rhs.(type) {
				case *ast.CompositeLit:
					// Check if this is a composition of errgroup.Group
					if sel, ok := expr.Type.(*ast.SelectorExpr); ok {
						if ident, ok := sel.X.(*ast.Ident); ok && sel.Sel.Name == "Group" && ident.Name == "errgroup" {
							// This is an errgroup.Group initialization
							for _, lhs := range stmt.Lhs {
								if varIdent, ok := lhs.(*ast.Ident); ok {
									scopes.AddVar(varIdent.Name, &errgroupVar{
										waitCalled: false,
										ident:      varIdent,
									})
								}
							}
						}
					}
				case *ast.CallExpr:
					if sel, ok := expr.Fun.(*ast.SelectorExpr); ok {
						if ident, ok := sel.X.(*ast.Ident); ok && sel.Sel.Name == "WithContext" && ident.Name == "errgroup" {
							for _, lhs := range stmt.Lhs {
								if varIdent, ok := lhs.(*ast.Ident); ok {
									scopes.AddVar(varIdent.Name, &errgroupVar{
										waitCalled: false,
										ident:      varIdent,
									})

									// First variable is the errgroup, second is the context
									break
								}
							}
						}
					}
				}
			}
		case *ast.CallExpr:
			// Check for Wait calls on errgroup variables
			if sel, ok := stmt.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok && sel.Sel.Name == "Wait" {
					if errgroupVar := scopes.FindVar(ident.Name); errgroupVar != nil {
						errgroupVar.waitCalled = true
					}
				}
			}
		}
		return true
	}

	ast.Inspect(file, inspectNode)
}