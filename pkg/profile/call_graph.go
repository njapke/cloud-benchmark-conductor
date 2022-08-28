package profile

import (
	"fmt"
	"sync"

	"github.com/christophwitzko/master-thesis/pkg/logger"
	"github.com/google/pprof/driver"
)

// global mutex to synchronize pprof calls
var pprofMutex = sync.Mutex{}

func ToCallGraph(log *logger.Logger, logPrefix, inputFile, outputFile string) error {
	pprofMutex.Lock()
	defer pprofMutex.Unlock()
	log.Infof("%s generating callgraph from profile %s", logPrefix, inputFile)
	fs := &mockFlagSet{
		flags: make(map[string]*flagSetFlag),
		parseHook: func(m *mockFlagSet) []string {
			setFlag[bool](m, "dot", true)
			setFlag[string](m, "output", outputFile)
			setFlag[int](m, "nodecount", 100000)
			setFlag[float64](m, "nodefraction", 0)
			setFlag[float64](m, "edgefraction", 0)
			return []string{inputFile}
		},
	}
	err := driver.PProf(&driver.Options{
		UI:      &logUI{prefix: logPrefix, log: log},
		Flagset: fs,
	})
	if err != nil {
		return fmt.Errorf("could not generate call graph: %w", err)
	}
	return nil
}

type logUI struct {
	prefix string
	log    *logger.Logger
}

func (ui *logUI) ReadLine(prompt string) (string, error) { return "", fmt.Errorf("not implemented") }

func (ui *logUI) Print(args ...interface{}) {
	ui.log.Info(append([]interface{}{ui.prefix + " "}, args...)...)
}

func (ui *logUI) PrintErr(args ...interface{}) {
	ui.Print(args...)
}

func (ui *logUI) IsTerminal() bool { return false }

func (ui *logUI) WantBrowser() bool { return false }

func (ui *logUI) SetAutoComplete(func(string) string) {}

type flagSetFlag struct {
	Name  string
	Value any
}

type mockFlagSet struct {
	flags     map[string]*flagSetFlag
	parseHook func(m *mockFlagSet) []string
}

func addFlag[T any](fs *mockFlagSet, name string, def T) *T {
	fs.flags[name] = &flagSetFlag{
		Name:  name,
		Value: &def,
	}
	return &def
}

func setFlag[T any](fs *mockFlagSet, name string, value T) {
	valPtr := fs.flags[name].Value.(*T)
	*valPtr = value
}

func (m *mockFlagSet) Bool(name string, def bool, usage string) *bool {
	return addFlag[bool](m, name, def)
}

func (m *mockFlagSet) Int(name string, def int, usage string) *int {
	return addFlag[int](m, name, def)
}

func (m *mockFlagSet) Float64(name string, def float64, usage string) *float64 {
	return addFlag[float64](m, name, def)
}

func (m *mockFlagSet) String(name, def, usage string) *string {
	return addFlag[string](m, name, def)
}

func (m *mockFlagSet) StringList(name, def, usage string) *[]*string {
	return addFlag[[]*string](m, name, []*string{})
}

func (m *mockFlagSet) ExtraUsage() string {
	return ""
}

func (m *mockFlagSet) AddExtraUsage(eu string) {}

func (m *mockFlagSet) Parse(usage func()) []string {
	return m.parseHook(m)
}
