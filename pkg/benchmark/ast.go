package benchmark

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
)

type Function struct {
	Name        string
	Directory   string
	PackageName string
}

type Visitor struct {
	CurrentDirectory string
	CurrentPackage   string
	foundBenchmarks  []Function
}

func (v *Visitor) Visit(node ast.Node) (w ast.Visitor) {
	switch node := node.(type) {
	case *ast.FuncDecl:
		fnName := node.Name.Name
		if !strings.HasPrefix(fnName, "Benchmark") {
			return nil
		}
		v.foundBenchmarks = append(v.foundBenchmarks, Function{
			Name:        fnName,
			Directory:   v.CurrentDirectory,
			PackageName: v.CurrentPackage,
		})
		return nil
	}
	return v
}

func onlyTestFiles(info fs.FileInfo) bool {
	return strings.HasSuffix(info.Name(), "_test.go")
}

func findBenchmarksInDir(bv *Visitor) error {
	fileSet := token.NewFileSet()
	pkg, err := parser.ParseDir(fileSet, bv.CurrentDirectory, onlyTestFiles, parser.AllErrors)
	if err != nil {
		return err
	}
	for pkgName, astPkg := range pkg {
		bv.CurrentPackage = pkgName
		for _, astFile := range astPkg.Files {
			ast.Walk(bv, astFile)
		}
	}
	return nil
}

func GetFunctions(rootPath string) ([]Function, error) {
	bv := &Visitor{
		foundBenchmarks: make([]Function, 0),
	}
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		bv.CurrentDirectory = path
		return findBenchmarksInDir(bv)
	})
	if err != nil {
		return nil, err
	}
	return bv.foundBenchmarks, nil
}
