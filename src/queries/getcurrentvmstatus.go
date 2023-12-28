package queries

import (
	"fmt"
	"os"
	"px/api"
	"time"
)

func WaitForStatus(node string, vmid int64, status string, deadline int) bool {

	// FIXME implement right deadline
	for i := 0; i < deadline; i++ {
		resp, err := api.GetCurrentVMStatus(node, vmid)
		if err != nil {
			return false
		}
		data := resp.GetData()
		currentStatus := data.GetStatus()
		fmt.Fprintf(os.Stderr, "STATUS %+v\n", currentStatus)

		if currentStatus == status {
			return true
		}
		time.Sleep(time.Second)
	}
	return false
}

func getCurrentVMLock(node string, vmid int64) bool {
	resp, err := api.GetCurrentVMStatus(node, vmid)
	data := resp.GetData()
	if err != nil {
		fmt.Fprintf(os.Stderr, "getCurrentVMLock(): %v\n", err)
		return false
	}
	if data.HasLock() {
		return true
	}
	return false
}

// This leads to 500 if VM is still in process of creating
// But it works for waiting for stopping a machine
func WaitForVMUnlock(node string, vmid int64) {
	for {
		if !getCurrentVMLock(node, vmid) {
			break
		}
		time.Sleep(1 * time.Second)
	}
}
