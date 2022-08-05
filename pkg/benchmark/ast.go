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
	Name          string
	FileName      string
	PackageName   string
	RootDirectory string
}

func (f Function) String() string {
	return fmt.Sprintf("%s.%s", f.PackageName, f.Name)
}

func relativePath(base, target string) string {
	rPath, err := filepath.Rel(base, target)
	if err != nil {
		panic(err)
	}
	return rPath
}

type astVisitor struct {
	RootDirectory    string
	CurrentDirectory string
	CurrentFile      string
	CurrentPackage   string
	FoundBenchmarks  []Function
}

func (a *astVisitor) Visit(node ast.Node) (w ast.Visitor) {
	if node, ok := node.(*ast.FuncDecl); ok {
		fnName := node.Name.Name
		if !strings.HasPrefix(fnName, "Benchmark") {
			return nil
		}
		a.FoundBenchmarks = append(a.FoundBenchmarks, Function{
			Name:          fnName,
			FileName:      a.CurrentFile,
			PackageName:   a.CurrentPackage,
			RootDirectory: a.RootDirectory,
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
		for astFileName, astFile := range astPkg.Files {
			bv.CurrentFile = relativePath(bv.RootDirectory, astFileName)
			ast.Walk(bv, astFile)
		}
	}
	return nil
}

func GetFunctions(rootPath string) ([]Function, error) {
	absRootPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}
	bv := &astVisitor{
		RootDirectory:   absRootPath,
		FoundBenchmarks: make([]Function, 0),
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
	return bv.FoundBenchmarks, nil
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
		if f.PackageName == search.PackageName && f.Name == search.Name && f.FileName == search.FileName {
			return f, true
		}
	}
	return Function{}, false
}

type VersionedFunctions []VersionedFunction

func (vfs VersionedFunctions) Filter(predicate func(vf VersionedFunction) bool) VersionedFunctions {
	result := make(VersionedFunctions, 0)
	for _, vf := range vfs {
		if predicate(vf) {
			result = append(result, vf)
		}
	}
	return result
}

func CombineFunctions(v1, v2 []Function) VersionedFunctions {
	result := make(VersionedFunctions, 0)
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

func CombinedFunctionsFromPaths(sourcePathV1, sourcePathV2 string) (VersionedFunctions, error) {
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
