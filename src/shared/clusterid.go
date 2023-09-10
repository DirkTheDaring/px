package shared

func GenerateClusterId(node string) (int64, error) {

	if len(GlobalPxCluster.PxClients) > 1 {
		pxClient := GlobalPxCluster.GetPxClient(node)
		formula := 8*100000 + pxClient.OrigIndex*1000

		offset := 0
		for offset = 0; offset < 1000; offset++ {
			_, found := GlobalPxCluster.UniqueMachines[formula+offset]
			if !found {
				break
			}
		}
		vmid64 := int64(formula + offset)
		return vmid64, nil
	}

	vmid64, err := GetClusterNextId(node)
	if err != nil {
		return 0, err
	}
	return vmid64, nil
}
