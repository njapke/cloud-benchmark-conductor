package benchmark

import (
	"fmt"
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

func (f Function) String() string {
	return fmt.Sprintf("%s.%s", f.PackageName, f.Name)
}

func (f Function) RelativeDirectory(wd string) string {
	relativeDirectory, _ := filepath.Rel(wd, f.Directory)
	return relativeDirectory
}

type astVisitor struct {
	CurrentDirectory string
	CurrentPackage   string
	foundBenchmarks  []Function
}

func (a *astVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch node := node.(type) {
	case *ast.FuncDecl:
		fnName := node.Name.Name
		if !strings.HasPrefix(fnName, "Benchmark") {
			return nil
		}
		a.foundBenchmarks = append(a.foundBenchmarks, Function{
			Name:        fnName,
			Directory:   a.CurrentDirectory,
			PackageName: a.CurrentPackage,
		})
		return nil
	}
	return a
}

func filterOnlyTestFiles(info fs.FileInfo) bool {
	return strings.HasSuffix(info.Name(), "_test.go")
}

func findBenchmarksInDir(bv *astVisitor) error {
	fileSet := token.NewFileSet()
	pkg, err := parser.ParseDir(fileSet, bv.CurrentDirectory, filterOnlyTestFiles, parser.AllErrors)
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
	bv := &astVisitor{
		foundBenchmarks: make([]Function, 0),
	}
	absRootPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}
	err = filepath.WalkDir(absRootPath, func(path string, d fs.DirEntry, err error) error {
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

type VersionedFunction struct {
	V1, V2 Function
}

func (vf VersionedFunction) String() string {
	// the PackageName and Name are identical for both versions
	return vf.V1.String()
}

func findFunction(fns []Function, search Function) (Function, bool) {
	for _, f := range fns {
		if f.PackageName == search.PackageName && f.Name == search.Name {
			return f, true
		}
	}
	return Function{}, false
}

func CombineFunctions(v1, v2 []Function) []VersionedFunction {
	result := make([]VersionedFunction, 0)
	for _, functionV1 := range v1 {
		functionV2, ok := findFunction(v2, functionV1)
		if !ok {
			continue
		}
		result = append(result, VersionedFunction{
			V1: functionV1,
			V2: functionV2,
		})
	}
	return result
}

func CombinedFunctionsFromPaths(sourcePathV1, sourcePathV2 string) ([]VersionedFunction, error) {
	functionsV1, err := GetFunctions(sourcePathV1)
	if err != nil {
		return nil, err
	}

	functionsV2, err := GetFunctions(sourcePathV2)
	if err != nil {
		return nil, err
	}

	return CombineFunctions(functionsV1, functionsV2), nil
}
