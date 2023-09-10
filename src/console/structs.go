package console

type DumpSystem struct {
}
type DumpNodes struct {
}
type DumpMachines struct {
	Match string
}
type DumpCmd struct {
	System   DumpSystem   `cmd`
	Nodes    DumpNodes    `cmd`
	Machines DumpMachines `cmd`
}

type LoginCmd struct {
}
type LsCmd struct {
	Clusters LsClustersCmd `cmd:"" help:"clusers"`
}
type LsClustersCmd struct {
}
type StatusCmd struct {
	Match string
}
type ListStorageCmd struct {
	Match string
}
type MatchStorageCmd struct {
	Match string
}
type LatestStorageCmd struct {
	Match string
}
type ContentStorageCmd struct {
	Match string
}
type StorageCmd struct {
	List    ListStorageCmd    `cmd:"list"`
	Match   MatchStorageCmd   `cmd:"match"`
	Latest  LatestStorageCmd  `cmd:"latest"`
	Content ContentStorageCmd `cmd:"content"`
}
type VirtualMachineCreateCmd struct {
	Vmid   string `default:"0"`
	Node   string `required:""`
	Cattle string ``
	Dump   bool
}
type ContainerCreateCmd struct {
	Vmid   int    `default:"0"`
	Node   string `required:""`
	Cattle string ``
	Dump   bool
}
type DefaultCreateCmd struct {
	Filename string `short:"f" required:""`
}

type CreateCmd struct {
	Virtualmachine VirtualMachineCreateCmd `cmd:"list" aliases:"vm"`
	Container      ContainerCreateCmd      `cmd:"match" aliases:"ct"`
	Default        DefaultCreateCmd        `cmd:"default" default:"withargs" hidden:""`
}

type CommandLineInterface struct {
	Dump    DumpCmd    `cmd:"" help:"dump system configuration."`
	Cluster string     `cmd:"" help:"clustername" default:"0"`
	Login   LoginCmd   `cmd:"" help:"login"`
	Ls      LsCmd      `cmd:"" help:"ls objects"`
	Status  StatusCmd  `cmd:"status"`
	Storage StorageCmd `cmd:"storage"`
	Create  CreateCmd  `cmd:"create"`
}
