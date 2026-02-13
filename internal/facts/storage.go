package facts

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	Register("storage", collectStorage)
}

const blockBase = "/sys/block"

func collectStorage(result *Facts) error {
	entries, err := os.ReadDir(blockBase)
	if err != nil {
		log.Printf("storage: cannot read %s: %v", blockBase, err)
		return nil
	}
	if result.Disks == nil {
		result.Disks = make(Disks)
	}
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, "loop") || strings.HasPrefix(name, "ram") || strings.HasPrefix(name, "zram") || strings.HasPrefix(name, "dm-") {
			continue
		}
		info := DiskInfo{}
		if sizeBytes := readBlockSize(filepath.Join(blockBase, name, "size")); sizeBytes >= 0 {
			info.SizeBytes = sizeBytes
		}
		devicePath := filepath.Join(blockBase, name, "device")
		vendor := readTrimmedFile(filepath.Join(devicePath, "vendor"))
		info.Vendor = resolvePCIVendor(vendor)
		info.Model = readTrimmedFile(filepath.Join(devicePath, "model"))
		if info.Model == "" {
			info.Model = readTrimmedFile(filepath.Join(blockBase, name, "device", "model"))
		}
		result.Disks[name] = info
	}
	return nil
}
