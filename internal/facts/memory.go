package facts

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
)

func init() {
	Register(collectMemory)
}

func collectMemory(result *Facts) error {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		log.Printf("memory: cannot open /proc/meminfo: %v", err)
		return nil
	}
	defer f.Close()

	var totalKiB int64
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "MemTotal:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		totalKiB, _ = strconv.ParseInt(fields[1], 10, 64)
		break
	}
	if err := scanner.Err(); err != nil {
		log.Printf("memory: read error: %v", err)
		return nil
	}

	if totalKiB <= 0 {
		return nil
	}

	result.MemorySize = int(totalKiB / 1024)
	result.Memory.System.TotalBytes = totalKiB * 1024
	return nil
}
