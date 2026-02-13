package facts

import (
	"bufio"
	"log"
	"os"
	"runtime"
	"strings"
)

func init() {
	Register("cpu", collectCPU)
}

func collectCPU(result *Facts) error {
	result.Architecture = mapArch(runtime.GOARCH)
	result.ProcessorCount = runtime.NumCPU()

	f, err := os.Open("/proc/cpuinfo")
	if err != nil {
		log.Printf("cpu: cannot open /proc/cpuinfo: %v", err)
		return nil
	}
	defer f.Close()

	var modelNames []string
	seenModel := make(map[string]bool)
	var physicalIDs map[string]struct{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "model name") {
			if _, after, ok := strings.Cut(line, ":"); ok {
				model := strings.TrimSpace(after)
				if model != "" && !seenModel[model] {
					seenModel[model] = true
					modelNames = append(modelNames, model)
				}
			}
		}
		if strings.HasPrefix(line, "physical id") {
			if _, after, ok := strings.Cut(line, ":"); ok {
				id := strings.TrimSpace(after)
				if physicalIDs == nil {
					physicalIDs = make(map[string]struct{})
				}
				physicalIDs[id] = struct{}{}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("cpu: read error: %v", err)
		return nil
	}

	result.Processors.Count = result.ProcessorCount
	result.Processors.PhysicalCount = len(physicalIDs)
	if result.Processors.PhysicalCount == 0 {
		result.Processors.PhysicalCount = 1
	}
	result.Processors.Models = modelNames
	return nil
}
