package facts

import (
	"log"
	"path/filepath"
)

const dmiBase = "/sys/class/dmi/id"

func init() {
	Register(collectDMI)
}

func collectDMI(result *Facts) error {
	result.SerialNumber = readTrimmedFile(filepath.Join(dmiBase, "product_serial"))
	if result.SerialNumber == "" {
		result.SerialNumber = readTrimmedFile(filepath.Join(dmiBase, "board_serial"))
	}
	result.DMI.Manufacturer = readTrimmedFile(filepath.Join(dmiBase, "sys_vendor"))
	result.DMI.Product.Name = readTrimmedFile(filepath.Join(dmiBase, "product_name"))
	result.DMI.Product.UUID = readTrimmedFile(filepath.Join(dmiBase, "product_uuid"))
	if result.DMI.Manufacturer == "" && result.DMI.Product.Name == "" && result.DMI.Product.UUID == "" {
		log.Printf("dmi: no DMI data found under %s", dmiBase)
	}
	return nil
}
