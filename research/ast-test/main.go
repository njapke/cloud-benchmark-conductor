package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func split(input string, c byte) []string {
	res := make([]string, 0)
	for {
		pos := strings.IndexByte(input, c)
		if pos == -1 {
			return append(res, input)
		}
		res = append(res, input[:pos])
		input = input[pos+1:]
	}
}

func main() {
	fmt.Printf("%#v\n", split("A,B,C,D", ','))
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func onlyTestFiles(info fs.FileInfo) bool {
	return strings.HasSuffix(info.Name(), "_test.go")
}

func run() error {
	bv := &benchmarkVisitor{
		foundBenchmarks: make([]benchmarkFunction, 0),
	}
	err := filepath.WalkDir("./research", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		bv.CurrentDirectory = path
		fmt.Println("walking", path)
		return findBenchmarksInDir(bv)
	})
	if err != nil {
		return err
	}
	fmt.Println(bv.foundBenchmarks)
	return nil
}

func findBenchmarksInDir(bv *benchmarkVisitor) error {
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

type benchmarkFunction struct {
	Name        string
	Directory   string
	PackageName string
}

type benchmarkVisitor struct {
	CurrentDirectory string
	CurrentPackage   string
	foundBenchmarks  []benchmarkFunction
}

func (b *benchmarkVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch node := node.(type) {
	case *ast.FuncDecl:
		fnName := node.Name.Name
		if !strings.HasPrefix(fnName, "Benchmark") {
			return nil
		}
		b.foundBenchmarks = append(b.foundBenchmarks, benchmarkFunction{
			Name:        fnName,
			Directory:   b.CurrentDirectory,
			PackageName: b.CurrentPackage,
		})
		return nil
	}
	return b
}
