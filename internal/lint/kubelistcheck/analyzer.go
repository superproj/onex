// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package kubelistcheck

import (
	"flag"
	"go/ast"
	"go/types"
	"regexp"
	"sync"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	passutil "github.com/superproj/onex/pkg/util/lint/pass"
)

var (
	checkAll  = regexp.MustCompile(`^k8s.io/apimachinery/pkg/apis/meta/v1.(ListOptions|GetOptions)$`)
	checkList = regexp.MustCompile(`^k8s.io/apimachinery/pkg/apis/meta/v1.ListOptions$`)
)

type analyzer struct {
	get bool

	typesProcessCache   map[types.Type]bool
	typesProcessCacheMu sync.RWMutex
}

// NewAnalyzer returns a go/analysis-compatible analyzer.
//
//	-get arguments enable check for GetOptions
func NewAnalyzer(get bool) *analysis.Analyzer {
	a := analyzer{ //nolint:exhaustruct
		typesProcessCache: map[types.Type]bool{},
	}
	a.get = get

	return &analysis.Analyzer{ //nolint:exhaustruct
		Name:     "kubelistcheck",
		Doc:      "Checks if get/list kubernetes resources from kube-apiserver cache",
		Run:      a.run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Flags:    a.newFlagSet(),
	}
}

func (a *analyzer) newFlagSet() flag.FlagSet {
	fs := flag.NewFlagSet("kubelistcheck flags", flag.PanicOnError)

	fs.BoolVar(&a.get, "get", false, "enable check for GetOptions")

	return *fs
}

func (a *analyzer) run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector) //nolint:forcetypeassert

	nodeTypes := []ast.Node{
		(*ast.CompositeLit)(nil),
		(*ast.ReturnStmt)(nil),
	}

	insp.Preorder(nodeTypes, a.newVisitor(pass))

	return nil, nil //nolint:nilnil
}

func (a *analyzer) newVisitor(pass *analysis.Pass) func(node ast.Node) {
	var ret *ast.ReturnStmt

	return func(node ast.Node) {
		if passutil.HasNolintComment(pass, node, "kubelistcheck") {
			return
		}

		if retLit, ok := node.(*ast.ReturnStmt); ok {
			// save return statement for future (to detect error-containing returns)
			ret = retLit

			return
		}

		lit, _ := node.(*ast.CompositeLit)
		if lit.Type == nil {
			// we're not interested in non-typed literals
			return
		}

		typ := pass.TypesInfo.TypeOf(lit.Type)
		if typ == nil {
			return
		}

		if _, ok := typ.Underlying().(*types.Struct); !ok {
			// we also not interested in non-structure literals
			return
		}

		strctName := exprName(lit.Type)
		if strctName == "" {
			return
		}

		if !a.shouldProcessType(typ) {
			return
		}

		if len(lit.Elts) == 0 && ret != nil {
			if ret.End() < lit.Pos() {
				// we're outside last return statement
				ret = nil
			} else if returnContainsLiteral(ret, lit) && returnContainsError(ret, pass) {
				// we're okay with empty literals in return statements with non-nil errors, like
				// `return my.Struct{}, fmt.Errorf("non-nil error!")`
				return
			}
		}

		if !fieldSetted(lit, "ResourceVersion") {
			pass.Reportf(node.Pos(), "ResourceVersion is not setted in %s", strctName)
		}
	}
}

func (a *analyzer) shouldProcessType(typ types.Type) bool {
	a.typesProcessCacheMu.RLock()
	v, ok := a.typesProcessCache[typ]
	a.typesProcessCacheMu.RUnlock()

	if !ok {
		a.typesProcessCacheMu.Lock()
		defer a.typesProcessCacheMu.Unlock()

		reg := checkList
		if a.get {
			reg = checkAll
		}

		v = reg.MatchString(typ.String())
		a.typesProcessCache[typ] = v
	}

	return v
}

func returnContainsLiteral(ret *ast.ReturnStmt, lit *ast.CompositeLit) bool {
	for _, result := range ret.Results {
		if l, ok := result.(*ast.CompositeLit); ok {
			if lit == l {
				return true
			}
		}
	}

	return false
}

func returnContainsError(ret *ast.ReturnStmt, pass *analysis.Pass) bool {
	for _, result := range ret.Results {
		if pass.TypesInfo.TypeOf(result).String() == "error" {
			return true
		}
	}

	return false
}

func exprName(expr ast.Expr) string {
	if i, ok := expr.(*ast.Ident); ok {
		return i.Name
	}

	s, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return ""
	}

	return s.Sel.Name
}

func fieldSetted(lit *ast.CompositeLit, name string) bool {
	for _, elt := range lit.Elts {
		if k, ok := elt.(*ast.KeyValueExpr); ok {
			if ident, ok := k.Key.(*ast.Ident); ok && ident.Name == name {
				return true
			}
		}
	}

	return false
}
