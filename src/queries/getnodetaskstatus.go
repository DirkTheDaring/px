package queries

import (
	"px/api"
	"time"
)

func WaitForUPID(node string, upid string) {
	for {
		status, _ := getNodeTaskStatus2(node, upid)
		if status != "running" {
			return
		}
		//fmt.Fprintf(os.Stderr, "STATUS %+v\n", data)
		time.Sleep(5 * time.Second)
	}
}
func getNodeTaskStatus2(node string, upid string) (string, error) {

	resp, err := api.GetNodeTaskStatus(node, upid)
	if err != nil {
		return "", err
	}
	data := resp.GetData()
	status := data.GetStatus()

	return status, nil

}
