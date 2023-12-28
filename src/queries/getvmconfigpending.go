package queries

import (
	"fmt"
	"os"
	"px/api"
	"time"
)

func Wait(node string, vmid int64) {
	for {
		resp, err := api.GetVMConfigPending(node, vmid)
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
