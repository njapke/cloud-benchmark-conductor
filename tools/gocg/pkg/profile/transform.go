package profile

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Tos(profileDir *os.File, profileDirPath string, outDirPath string, outConfigs []*OutConfig) error {
	dirnames, err := profileDir.Readdir(-1)
	if err != nil {
		return fmt.Errorf("could not get dirnames: %v", err)
	}

	for _, OutConfig := range outConfigs {
		var total, failed int
		for _, n := range dirnames {
			if n.IsDir() {
				continue
			}

			fn := n.Name()
			if !strings.HasSuffix(fn, ".pprof") {
				continue
			}

			inFileName := filepath.Join(profileDirPath, fn)
			outFileName := filepath.Join(outDirPath, outFileName(fn, OutConfig))
			total++

			// fmt.Printf("# start transform '%s' -> '%s'\n", inFileName, outFileName)
			err := To(inFileName, outFileName, OutConfig)
			if err != nil {
				failed++
				fmt.Fprintf(os.Stderr, "# error transform '%s' -> %s file: %v\n", inFileName, OutConfig.Type.Name(), err)
			}
			// fmt.Printf("# finished transform '%s' -> '%s'\n", inFileName, outFileName)
		}

		fmt.Fprintf(os.Stderr, "# %s transformations: %d/%d profiles\n", OutConfig.Type.Name(), (total - failed), total)
	}

	return nil
}

func To(pprofFile, outFile string, outConfig *OutConfig) error {
	cmd := exec.Command(
		"go",
		"tool",
		"pprof",
		fmt.Sprintf("-nodecount=%d", outConfig.NodeCount),
		fmt.Sprintf("-nodefraction=%f", outConfig.NodeFraction),
		fmt.Sprintf("-edgefraction=%f", outConfig.EdgeFraction),
		fmt.Sprintf("-%s",
			outConfig.Type.Name()),
		pprofFile,
	)

	out, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "execution error: %v\n", err)
		fmt.Fprintf(os.Stderr, "stdout:\n%s", string(out))
		if ee, ok := err.(*exec.ExitError); ok {
			fmt.Fprintf(os.Stderr, "stderr:\n%s", string(ee.Stderr))
		}
		return err
	}

	err = ioutil.WriteFile(outFile, out, os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not write to out file '%s': %v", outFile, err)
	}

	return nil
}

// inFileName: 0-0-0_Baseline_tsdb_BenchmarkWritePoints_NewSeries_1000000_Measurement_1_TagKey_1_TagValue_cpu
// outFileName: inFileName__node-count__node-fraction__edge-fraction.out-type
func outFileName(inFileName string, outConfig *OutConfig) string {
	return fmt.Sprintf("%s__%s", rdus(inFileName), outConfig.FileSuffix())
}

func rdus(s string) string {
	return strings.ReplaceAll(s, ".", "_")
}
