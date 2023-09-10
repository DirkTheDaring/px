package shared

func CreateVirtualmachine() {
	/*
		fmt.Fprintf(os.Stderr, "Virtualmachine: %v\n", cmd.createVirtualmachineOptions.Node)
		if !query.In(GlobalPxCluster.Nodes, cmd.createVirtualmachineOptions.Node) {
			fmt.Fprintf(os.Stderr, "node does not exist: %s\n", cmd.createVirtualmachineOptions.Node)
			os.Exit(1)
		}

		data := map[string]interface{}{}
		data["node"] = cmd.createVirtualmachineOptions.Node
		data["vmid"] = cmd.createVirtualmachineOptions.Vmid

		result := cattles.CreateCattle("vm", "small", data)
		if cmd.createVirtualmachineOptions.Dump {
			proxmox.DumpJson(result)
		}
		aliases := GlobalPxCluster.GetAliasOnNode(cmd.createVirtualmachineOptions.Node)
		storageNames := GlobalPxCluster.GetStorageNamesOnNode(cmd.createVirtualmachineOptions.Node)
		cattles.ProcessStorage(result, aliases, storageNames)
		os.Exit(0)
	*/
}
