package queries

import (
	"fmt"
	"px/api"
	"strconv"
)

func ResizeVMDisk(node string, vmid int, storageDrive string, deltaSize int64) error {
	res, err := api.ResizeVMDisk(node, int64(vmid), storageDrive, "+"+strconv.FormatInt(deltaSize, 10))
	if err != nil {
		return fmt.Errorf("failed to resize disk: %v", err)
	}
	WaitForUPID(node, res.GetData())
	return nil
}
