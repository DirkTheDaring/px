package shared

/*
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
*/
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
