package dir

import (
	"fmt"
	"strconv"
	"strings"

	"bitbucket.org/sealuzh/gocg/pkg/profile"
)

// func ParseMicroBenchmark(filename string, noNameParsing string) (benchmark string, outConfig *profile.OutConfig, err error) {
// 	name, config, err := ParseFileNameConfig(filename)

// 	// benchmark
// 	if name != noNameParsing {
// 		// 0-0-0_Baseline_tsdb_BenchmarkWritePoints_NewSeries_5000_Measurement_1_TagKey_1_TagValue_cpu_pprof
// 		benchSplitted := strings.Split(name, "_")
// 		lbs := len(benchSplitted)
// 		if lbs < 5 {
// 			return "", nil, fmt.Errorf("could not parse microbenchmark in micro string '%s': expected min %d elements, got %d", filename, 5, lbs)
// 		}

// 		oldName := name

// 		name = strings.Join(benchSplitted[2:lbs-2], "_")
// 		benchIdx := strings.Index(name, "Benchmark")
// 		if benchIdx != 0 {
// 			fmt.Println(benchIdx, name, oldName)
// 			name = fmt.Sprintf(
// 				"%s/%s",
// 				strings.Join(strings.Split(name[:benchIdx-1], "_"), "/"),
// 				name[benchIdx:],
// 			)
// 		}
// 	}

// 	return name, config, nil
// }

func ParseFileNameConfig(s string) (name string, outConfig *profile.OutConfig, err error) {
	// name__100000__0_00000__0_00000.dot
	splitted := strings.Split(s, "__")
	ls := len(splitted)
	if ls < 4 {
		fmt.Println(s)
		fmt.Println(splitted)
		return "", nil, fmt.Errorf("could not parse micro string '%s': expected %d elements, got %d", s, 4, ls)
	}

	// node count
	nc, err := strconv.Atoi(splitted[ls-3])
	if err != nil {
		return "", nil, fmt.Errorf("could not parse node count in micro string: %v", err)
	}

	// node fraction
	nf, err := strconv.ParseFloat(strings.ReplaceAll(splitted[ls-2], "_", "."), 64)
	if err != nil {
		return "", nil, fmt.Errorf("could not parse node fraction in micro string: %v", err)
	}

	// edge fraction
	lastStr := splitted[ls-1]
	dotIdx := strings.LastIndex(lastStr, ".")

	efStr := strings.ReplaceAll(lastStr[:dotIdx], "_", ".")
	ef, err := strconv.ParseFloat(efStr, 64)
	if err != nil {
		return "", nil, fmt.Errorf("could not parse edge fraction in micro string: %v", err)
	}

	// out type
	t, err := profile.OutTypeFrom(lastStr[dotIdx+1:])
	if err != nil {
		return "", nil, fmt.Errorf("could not parse out type in micro string: %v", err)
	}

	return strings.Join(splitted[:ls-3], "__"), &profile.OutConfig{
		Type:         t,
		NodeCount:    nc,
		NodeFraction: nf,
		EdgeFraction: ef,
	}, nil
}
