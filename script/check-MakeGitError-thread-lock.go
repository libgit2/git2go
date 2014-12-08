package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	fset = token.NewFileSet()
)

func main() {
	log.SetFlags(0)

	bpkg, err := build.ImportDir(".", 0)
	if err != nil {
		log.Fatal(err)
	}

	pkgs, err := parser.ParseDir(fset, bpkg.Dir, func(fi os.FileInfo) bool { return filepath.Ext(fi.Name()) == ".go" }, 0)
	if err != nil {
		log.Fatal(err)
	}

	for _, pkg := range pkgs {
		if err := checkPkg(pkg); err != nil {
			log.Fatal(err)
		}
	}
	if len(pkgs) == 0 {
		log.Fatal("No packages to check.")
	}
}

var ignoreViolationsInFunc = map[string]bool{
	"MakeGitError":  true,
	"MakeGitError2": true,
}

func checkPkg(pkg *ast.Package) error {
	var violations []string
	ast.Inspect(pkg, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.FuncDecl:
			var b bytes.Buffer
			if err := printer.Fprint(&b, fset, node); err != nil {
				log.Fatal(err)
			}
			src := b.String()

			if strings.Contains(src, "MakeGitError") && !strings.Contains(src, "runtime.LockOSThread()") && !strings.Contains(src, "defer runtime.UnlockOSThread()") && !ignoreViolationsInFunc[node.Name.Name] {
				pos := fset.Position(node.Pos())
				violations = append(violations, fmt.Sprintf("%s at %s:%d", node.Name.Name, pos.Filename, pos.Line))
			}
		}
		return true
	})
	if len(violations) > 0 {
		return fmt.Errorf("%d non-thread-locked calls to MakeGitError found. To fix, add the following to each func below that calls MakeGitError, before the cgo call that might produce the error:\n\n\truntime.LockOSThread()\n\tdefer runtime.UnlockOSThread()\n\n%s", len(violations), strings.Join(violations, "\n"))
	}
	return nil
}
