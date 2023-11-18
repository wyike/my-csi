package vsphere

import (
	"fmt"
	"os"
	"path"
	"strings"
)

const (
	dmiDir     = "/sys/class/dmi"
	UUIDPrefix = "VMware-"
)

// GetSystemUUID returns the UUID used to identify node vm
func GetSystemUUID() (string, error) {
	idb, err := os.ReadFile(path.Join(dmiDir, "id", "product_serial"))
	if err != nil {
		return "", err
	}
	uuidFromFile := string(idb[:])
	//strip leading and trailing white space and new line char
	uuid := strings.TrimSpace(uuidFromFile)
	fmt.Printf("product_serial in string: %s\n", uuid)
	// check the uuid starts with "VMware-"
	if !strings.HasPrefix(uuid, UUIDPrefix) {
		return "", fmt.Errorf("failed to match Prefix, UUID read from the file is %s",
			uuidFromFile)
	}
	// Strip the prefix and while spaces and -
	uuid = strings.Replace(uuid[len(UUIDPrefix):], " ", "", -1)
	uuid = strings.Replace(uuid, "-", "", -1)
	if len(uuid) != 32 {
		return "", fmt.Errorf("length check failed, UUID read from the file is %v", uuidFromFile)
	}
	// need to add dashes, e.g. "564d395e-d807-e18a-cb25-b79f65eb2b9f"
	uuid = fmt.Sprintf("%s-%s-%s-%s-%s", uuid[0:8], uuid[8:12], uuid[12:16], uuid[16:20], uuid[20:32])
	fmt.Printf("UUID is %s\n", uuid)
	return uuid, nil
}
