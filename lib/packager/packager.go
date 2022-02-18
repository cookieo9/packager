package packager

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"io"
	"regexp"

	"golang.org/x/tools/go/packages"
)

func LoadPackage(pkg string) (*types.Package, error) {
	cfg := &packages.Config{
		Mode:       packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax | packages.NeedDeps | packages.NeedImports,
		BuildFlags: []string{"-tags", "packager"},
	}

	pkgs, err := packages.Load(cfg, pkg)
	if err != nil {
		return nil, err
	}

	return pkgs[0].Types, nil
}

type Processor struct {
	Local        string
	Allow, Block string
}

func (p *Processor) Process(out io.Writer, pkg *types.Package) (err error) {
	var allowRE, blockRE *regexp.Regexp
	if allowRE, err = regexp.Compile(p.Allow); err != nil {
		return fmt.Errorf("can't parse allow regexp %q: %w", p.Allow, err)
	}
	if blockRE, err = regexp.Compile(p.Block); err != nil {
		return fmt.Errorf("can't parse block regexp %q: %w", p.Block, err)
	}

	useSymbol := func(name string) bool {
		return allowRE.MatchString(name) && !blockRE.MatchString(name)
	}

	obj := pkg.Scope().Lookup(p.Local)
	if obj == nil {
		return fmt.Errorf("can't find local symbol %q", (p.Local))
	}

	otype := obj.Type()
	if _, ok := otype.(*types.Pointer); !ok {
		otype = types.NewPointer(otype)
	}

	methods := types.NewMethodSet(otype)
	qf := qf(pkg)

	fmt.Fprintln(out, "package", pkg.Name())
	fmt.Fprintln(out)

	for i := 0; i < methods.Len(); i++ {
		method := methods.At(i)
		name := method.Obj().Name()
		if !useSymbol(name) || !method.Obj().Exported() {
			continue
		}
		sig := method.Type().(*types.Signature)
		sigStr := types.TypeString(sig, qf)[4:]
		fmt.Fprintf(out,
			"// %s is package exported version of %q called on a package level value.\n",
			name, types.ObjectString(method.Obj(), qf),
		)

		fmt.Fprintln(out, "func", name, sigStr, "{")
		format.Node(out, token.NewFileSet(), delegateCall(p.Local, method))
		fmt.Fprintln(out, "}")
		fmt.Fprintln(out)
	}

	return nil
}

func delegateCall(local string, method *types.Selection) ast.Node {
	sig := method.Type().(*types.Signature)
	args := []ast.Expr{}
	for i := 0; i < sig.Params().Len(); i++ {
		args = append(args, ast.NewIdent(sig.Params().At(i).Name()))
	}
	call := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent(local),
			Sel: ast.NewIdent(method.Obj().Name()),
		},
		Args: args,
	}
	if sig.Results().Len() > 0 {
		return &ast.ReturnStmt{Results: []ast.Expr{call}}
	}
	return call
}

func qf(me *types.Package) types.Qualifier {
	if me == nil {
		return nil
	}
	return func(other *types.Package) string {
		if me == other {
			return ""
		}
		return other.Name()
	}
}
