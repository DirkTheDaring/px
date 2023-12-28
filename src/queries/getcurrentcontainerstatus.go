package queries

import (
	"fmt"
	"os"
	"px/api"
	"time"
)

const sleepDuration = 1 * time.Second // Constant for sleep duration

// getCurrentContainerLock checks if the container is currently locked.
func getCurrentContainerLock(node string, vmid int64) bool {
	resp, err := api.GetCurrentContainerStatus(node, vmid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "getCurrentContainerLock(): %v\n", err)
		return false
	}
	data := resp.GetData()
	return data.HasLock()
}

// WaitForContainerUnlock waits until the container is unlocked.
func WaitForContainerUnlock(node string, vmid int64) {
	for getCurrentContainerLock(node, vmid) {
		fmt.Fprintf(os.Stderr, "Waiting for container %v to unlock...\n", vmid)
		time.Sleep(sleepDuration)
	}
}
