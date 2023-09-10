package shared

import (
	"fmt"
	"os"
	"time"
)

func WaitForStatus(node string, vmid int64, status string, deadline int) bool {

	// FIXME implement right deadline
	for i := 0; i < deadline; i++ {
		resp, err := GetCurrentVMStatus(node, vmid)
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

func WaitForUPID(node string, upid string) {
	for true {
		resp, err := GetNodeTaskStatus(node, upid)
		if err != nil {
			return
		}
		data := resp.GetData()
		status := data.GetStatus() 
		if status != "running" {
			return
		}
		//fmt.Fprintf(os.Stderr, "STATUS %+v\n", data)
		time.Sleep(5 * time.Second)
	}
}
/*
// does not work for waiting on stop
func Wait(node string, vmid int64) {
	for true {
		resp, err := GetVMConfigPending(node, vmid)
		if err != nil {
			break
		}
		data := resp.GetData()

		for _, item := range data {
			if item.GetKey() == "lock" {
				fmt.Fprintf(os.Stderr, "lock: %v\n", item.GetValue())
				goto found
			}
		}
		break
	found:
		time.Sleep(1 * time.Second)
	}
}
*/
// does not work for waiting on stop
func Wait(node string, vmid int64) {
	for true {
		resp, err := GetVMConfigPending(node, vmid)
		if err != nil {
			break
		}
		data := resp.GetData()

		for _, item := range data {
		   //fmt.Fprintf(os.Stderr, "ITEM  %+v\n", item)
		   if item.GetVMConfigPending200ResponseDataInnerOneOf != nil {
			   stringItem := item.GetVMConfigPending200ResponseDataInnerOneOf
			   //fmt.Fprintf(os.Stderr, "%v\n", stringItem.GetValue())
			   if stringItem.GetKey() == "lock" {
				   fmt.Fprintf(os.Stderr, "(string) lock: %v\n", stringItem.GetValue())
				   goto found
			   }

		   }
		   if item.GetVMConfigPending200ResponseDataInnerOneOf1 != nil {
			   intItem := item.GetVMConfigPending200ResponseDataInnerOneOf1
			   //fmt.Fprintf(os.Stderr, "%v\n", intItem.GetValue())
			   if intItem.GetKey() == "lock" {
				   fmt.Fprintf(os.Stderr, "(int) lock: %v\n", intItem.GetValue())
				   goto found
			   }

		   }
			   
		}
		break
	found:
		time.Sleep(1 * time.Second)

	}
}


func getCurrentVMLock(node string, vmid int64) bool {
	resp, err := GetCurrentVMStatus(node, vmid)
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
	for true {
		if !getCurrentVMLock(node, vmid) {
			break
		}
		time.Sleep(1 * time.Second)
	}
}

// This leads to 500 if VM is still in process of creating
// But it works for waiting for stopping a machine
func WaitForCTUnlock(node string, vmid int64) {
	for true {
		if !getCurrentContainerLock(node, vmid) {
			break
		}
		time.Sleep(1 * time.Second)
	}
}

func getCurrentContainerLock(node string, vmid int64) bool {
	resp, err := GetCurrentContainerStatus(node, vmid)
	data := resp.GetData()
	if err != nil {
		fmt.Fprintf(os.Stderr, "getCurrentContainerLock(): %v\n", err)
		return false
	}
	if data.HasLock() {
		return true
	}
	return false
}

// This leads to 500 if VM is still in process of creating
// But it works for waiting for stopping a machine
func WaitForContainerUnlock(node string, vmid int64) {
	for true {
		if !getCurrentContainerLock(node, vmid) {
			break
		}
		//fmt.Fprintf(os.Stderr, "WaitForContainerUnlock(): %v\n", vmid)
		time.Sleep(1 * time.Second)
	}
}
