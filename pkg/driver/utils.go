package driver

import (
	"fmt"
	"strings"
)

// FormatDiskUUID removes any spaces and hyphens in UUID.
// Example UUID input is 42375390-71f9-43a3-a770-56803bcd7baa and output after
// format is 4237539071f943a3a77056803bcd7baa.
func FormatDiskUUID(uuid string) string {
	uuidwithNoSpace := strings.Replace(uuid, " ", "", -1)
	uuidWithNoHypens := strings.Replace(uuidwithNoSpace, "-", "", -1)
	return strings.ToLower(uuidWithNoHypens)
}

func getDeviceSource(diskUUID string) string {
	return fmt.Sprintf("/dev/disk/by-id/wwn-0x%s", diskUUID)
}
