package shared

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"github.com/DirkTheDaring/px-api-client-go"
)

func GetPxClient(node string) (PxClient, *pxapiflat.APIClient, context.Context, error) {
	var pxClient PxClient

	var apiClient *pxapiflat.APIClient
	var context context.Context
	//index, ok := nodeMap[node]

	index, ok := GlobalPxCluster.PxClientLookup[node]

	if !ok {
		return pxClient, nil, context, errors.New("node not found: " + node)
	}

	//pxClient = pxclients[index]
	pxClient = GlobalPxCluster.PxClients[index]
	apiClient = pxClient.ApiClient
	context = pxClient.Context
	return pxClient, apiClient, context, nil
}

// This function really returns the same number on subsequenct call, if there
// is no new virtual machine created
func GetClusterNextId(node string) (int64, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		return -1, err
	}

	resp, r, err := apiClient.ClusterApi.GetClusterNextid(context).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClusterApi.GetClusterNextid``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return -1, err
	}
	//fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	//fmt.Fprintf(os.Stdout, "Response from `ClusterApi.GetClusterNextid`: %v\n", resp)
	data := resp.GetData()
	n, err := strconv.ParseInt(data, 10, 64)
	return n, err
}

func GetStorageContent(node string, storage string) (*pxapiflat.GetStorageContent200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		return nil, err
	}
	resp, r, err := apiClient.NodesApi.GetStorageContent(context, node, storage).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.GetStorageContent``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `GetStorageContent`: GetStorageContent200Response
	//fmt.Fprintf(os.Stdout, "Response from `NodesApi.GetStorageContent`: %v\n", resp)
	return resp, err
}

func Upload(node string, storage string, content string, filename *os.File) (*pxapiflat.CreateVM200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		return nil, err
	}

	//checksum := "checksum_example"                   // string |  (optional)
	//checksumAlgorithm := "checksumAlgorithm_example" // string |  (optional)
	//tmpfilename := "tmpfilename_example"             // string |  (optional)

	resp, r, err := apiClient.NodesApi.UploadFile(context, node, storage).Content(content).Filename(filename).Execute()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.CreateNodesSingleStorageSingleUpload``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `CreateNodesSingleStorageSingleUpload`: CreateVM200Response
	//fmt.Fprintf(os.Stdout, "Response from `NodesApi.CreateNodesSingleStorageSingleUpload`: %v\n", resp)
	return resp, err
}

func GetStorages(node string) (*pxapiflat.GetStorages200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		return nil, err
	}

	resp, r, err := apiClient.NodesApi.GetStorages(context, node).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.GetNodesSingleStorage``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `GetNodesSingleStorage`: GetNodesSingleStorage200Response
	//fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	//fmt.Fprintf(os.Stdout, "Response from `NodesApi.GetNodesSingleStorage`: %v\n", resp)
	return resp, err
}

func GetClusterResources(node string) (*pxapiflat.GetClusterResources200Response, error) {

	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		return nil, err
	}

	resp, r, err := apiClient.ClusterApi.GetClusterResources(context).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClusterApi.GetClusterResources``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	return resp, err
}

/*
func UpdateVMConfig(node string, vmid int64, updateVMConfigRequest pxapiflat.UpdateVMConfigRequest) (*pxapiflat.CreateVM200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		return nil, err
	}

	resp, r, err := apiClient.NodesApi.UpdateVMConfig(context, node, vmid).UpdateVMConfigRequest(updateVMConfigRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.UpdateVMConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `UpdateVMConfig`: CreateVM200Response
	fmt.Fprintf(os.Stdout, "Response from `NodesApi.UpdateVMConfig`: %v\n", resp)
	return resp, err
}
*/

//********************************************************************************************************

// | **Post** /nodes/{node}/qemu | createVM
func CreateVM(node string, flatmachine pxapiflat.CreateVMRequest) (*pxapiflat.CreateVM200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		return nil, err
	}

	resp, r, err := apiClient.NodesApi.CreateVM(context, node).CreateVMRequest(flatmachine).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.CreateVM``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	return resp, err
}

// | **Post** /nodes/{node}/qemu/{vmid}/snapshot | createVMSnapshot

func CreateVMSnapshot(node string, vmid int64, snapshotName string) (*pxapiflat.TaskStartResponse, error) {

	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		return nil, err
	}
	createVMSnapshotRequest := *pxapiflat.NewCreateVMSnapshotRequest(snapshotName)

	resp, r, err := apiClient.NodesApi.CreateVMSnapshot(context, node, vmid).CreateVMSnapshotRequest(createVMSnapshotRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.CreateVMSnapshot``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	//fmt.Fprintf(os.Stdout, "Response from `NodesApi.CreateVMSnapshot`: %v\n", resp)
	return resp, err
}

// | **Delete** /nodes/{node}/qemu/{vmid} | deleteVM
func DeleteVM(node string, vmid int64) (*pxapiflat.CreateVM200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		return nil, err
	}
	resp, r, err := apiClient.NodesApi.DeleteVM(context, node, vmid).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.DeleteVM``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `DeleteVM`: CreateVM200Response
	//fmt.Fprintf(os.Stdout, "Response from `NodesApi.DeleteVM`: %v\n", resp)
	return resp, err
}

// | **Delete** /nodes/{node}/qemu/{vmid}/snapshot/{snapname} | deleteVMSnapshot
func DeleteVMSnapshot(node string, vmid int64, snapname string) (*pxapiflat.TaskStartResponse, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		return nil, err
	}
	resp, r, err := apiClient.NodesApi.DeleteVMSnapshot(context, node, vmid, snapname).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.DeleteVMSnapshot``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `DeleteVMSnapshot`: CreateVMSnapshot200Response
	//fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	//fmt.Fprintf(os.Stdout, "Response from `NodesApi.DeleteVMSnapshot`: %v\n", resp)
	return resp, err
}

// | **Get** /nodes/{node}/qemu/{vmid}/status/current | getCurrentVMStatus
func GetCurrentVMStatus(node string, vmid int64) (*pxapiflat.GetCurrentVMStatus200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		return nil, err
	}
	resp, r, err := apiClient.NodesApi.GetCurrentVMStatus(context, node, vmid).Execute()

	if r.StatusCode == 500 {
		fmt.Fprintf(os.Stderr, "500 Full HTTP response: %v\n", r)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.GetCurrentVMStatus``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `GetCurrentVMStatus`: GetCurrentVMStatus200Response
	//fmt.Fprintf(os.Stdout, "Response from `NodesApi.GetCurrentVMStatus`: %v\n", resp)
	return resp, err
}

// | **Get** /nodes/{node}/qemu/{vmid} | getVM

func GetVM(node string, vmid int64) (*pxapiflat.GetVM200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		return nil, err
	}
	resp, r, err := apiClient.NodesApi.GetVM(context, node, vmid).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.GetVM``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetVM`: GetVM200Response
	//fmt.Fprintf(os.Stdout, "Response from `NodesApi.GetVM`: %v\n", resp)
	return resp, err
}

// | **Get** /nodes/{node}/qemu/{vmid}/config | getVMConfig

func GetVMConfig(node string, vmid int64) (*pxapiflat.GetVMConfig200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		return nil, err
	}
	resp, r, err := apiClient.NodesApi.GetVMConfig(context, node, vmid).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.GetVMConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetVMConfig`: GetVMConfig200Response
	//fmt.Fprintf(os.Stdout, "Response from `NodesApi.GetVMConfig`: %v\n", resp)
	return resp, err
}

// | **Get** /nodes/{node}/qemu/{vmid}/pending | getVMConfigPending

func GetVMConfigPending(node string, vmid int64) (*pxapiflat.GetVMConfigPending200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		return nil, err
	}
	resp, r, err := apiClient.NodesApi.GetVMConfigPending(context, node, vmid).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.GetVMConfigPending``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `GetVMConfigPending`: GetVMConfigPending200Response
	//fmt.Fprintf(os.Stdout, "Response from `NodesApi.GetVMConfigPending`: %v\n", resp)
	return resp, err
}

// | **Get** /nodes/{node}/qemu/{vmid}/snapshot/{snapname} | getVMSnapshot

func GetVMSnapshot(node string, vmid int64) (*pxapiflat.GetVMSnapshots200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		return nil, err
	}
	resp, r, err := apiClient.NodesApi.GetVMSnapshots(context, node, vmid).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.GetVMSnapshots``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetVMSnapshots`: GetVMSnapshots200Response
	fmt.Fprintf(os.Stdout, "Response from `NodesApi.GetVMSnapshots`: %v\n", resp)
	return resp, err
}

// | **Get** /nodes/{node}/qemu/{vmid}/snapshot/{snapname}/config | getVMSnapshotConfig

func GetVMSnapshotConfig(node string, vmid int64, snapname string) (*pxapiflat.GetVMSnapshotConfig200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		return nil, err
	}
	resp, r, err := apiClient.NodesApi.GetVMSnapshotConfig(context, node, vmid, snapname).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.GetVMSnapshotConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetVMSnapshotConfig`: GetVMSnapshotConfig200Response
	fmt.Fprintf(os.Stdout, "Response from `NodesApi.GetVMSnapshotConfig`: %v\n", resp)
	return resp, err
}

// | **Get** /nodes/{node}/qemu/{vmid}/snapshot | getVMSnapshots

func GetVMSnapshots(node string, vmid int64) (*pxapiflat.GetVMSnapshots200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		return nil, err
	}
	resp, r, err := apiClient.NodesApi.GetVMSnapshots(context, node, vmid).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.GetVMSnapshots``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetVMSnapshots`: GetVMSnapshots200Response
	//fmt.Fprintf(os.Stdout, "Response from `NodesApi.GetVMSnapshots`: %v\n", resp)
	return resp, err
}

// | **Get** /nodes/{node}/qemu | getVMs

func GetVMs(node string) (*pxapiflat.GetVMs200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		return nil, err
	}
	resp, r, err := apiClient.NodesApi.GetVMs(context, node).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.GetVMs``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetVMs`: GetVMs200Response
	fmt.Fprintf(os.Stdout, "Response from `NodesApi.GetVMs`: %v\n", resp)
	return resp, err
}

// | **Post** /nodes/{node}/qemu/{vmid}/status/reboot | rebootVM

func RebootVM(node string, vmid int64) (*pxapiflat.CreateVM200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		return nil, err
	}
	rebootVMRequest := *pxapiflat.NewRebootVMRequest() // RebootVMRequest |  (optional)

	resp, r, err := apiClient.NodesApi.RebootVM(context, node, vmid).RebootVMRequest(rebootVMRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.RebootVM``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `RebootVM`: CreateVM200Response
	// fmt.Fprintf(os.Stdout, "Response from `NodesApi.RebootVM`: %v\n", resp)
	return resp, err
}

// | **Put** /nodes/{node}/qemu/{vmid}/resize | resizeVMDisk

func ResizeVMDisk(node string, vmid int64, disk string, size string) (*pxapiflat.TaskStartResponse, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		return nil, err
	}
	//resizeVMDiskRequest := *pxapiflat.NewResizeVMDiskRequest("virtio0", "+50G") // ResizeVMDiskRequest |  (optional)
	//strSize := "+" + strconv.FormatInt(size, 10)

	resizeVMDiskRequest := *pxapiflat.NewResizeVMDiskRequest(disk, size) // ResizeVMDiskRequest |  (optional)

	resp, r, err := apiClient.NodesApi.ResizeVMDisk(context, node, vmid).ResizeVMDiskRequest(resizeVMDiskRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.ResizeVMDisk``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `ResizeVMDisk`: CreateVM200Response
	//fmt.Fprintf(os.Stdout, "Response from `NodesApi.ResizeVMDisk`: %v\n", resp)
	return resp, err

}

// | **Post** /nodes/{node}/qemu/{vmid}/status/resume | resumeVM

func ResumeVM(node string, vmid int64) (*pxapiflat.CreateVM200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		return nil, err
	}
	resumeVMRequest := *pxapiflat.NewResumeVMRequest() // ResumeVMRequest |  (optional)

	resp, r, err := apiClient.NodesApi.ResumeVM(context, node, vmid).ResumeVMRequest(resumeVMRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.ResumeVM``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ResumeVM`: CreateVM200Response
	fmt.Fprintf(os.Stdout, "Response from `NodesApi.ResumeVM`: %v\n", resp)
	return resp, err
}

// | **Post** /nodes/{node}/qemu/{vmid}/snapshot/{snapname}/rollback | rollbackVMSnapshot

func RollbackVMSnapshot(node string, vmid int64, snapname string) (*pxapiflat.TaskStartResponse, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		return nil, err
	}
	rollbackVMSnapshotRequest := *pxapiflat.NewRollbackVMSnapshotRequest() // RollbackVMSnapshotRequest |  (optional)

	resp, r, err := apiClient.NodesApi.RollbackVMSnapshot(context, node, vmid, snapname).RollbackVMSnapshotRequest(rollbackVMSnapshotRequest).Execute()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.RollbackVMSnapshot``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `RollbackVMSnapshot`: CreateVMSnapshot200Response
	fmt.Fprintf(os.Stdout, "Response from `NodesApi.RollbackVMSnapshot`: %v\n", resp)
	return resp, err
}

// | **Post** /nodes/{node}/qemu/{vmid}/status/shutdown | shutdownVM

func ShutdownVM(node string, vmid int64) (*pxapiflat.CreateVM200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		return nil, err
	}
	shutdownVMRequest := *pxapiflat.NewShutdownVMRequest() // ShutdownVMRequest |  (optional)

	//configuration := openapiclient.NewConfiguration()
	//apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodesApi.ShutdownVM(context, node, vmid).ShutdownVMRequest(shutdownVMRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.ShutdownVM``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `ShutdownVM`: CreateVM200Response
	//fmt.Fprintf(os.Stdout, "Response from `NodesApi.ShutdownVM`: %v\n", resp)
	return resp, err
}

// | **Post** /nodes/{node}/qemu/{vmid}/status/start | startVM

func StartVM(node string, vmid int64) (*pxapiflat.CreateVM200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		return nil, err
	}
	startVMRequest := *pxapiflat.NewStartVMRequest() // StartVMRequest |  (optional)

	resp, r, err := apiClient.NodesApi.StartVM(context, node, vmid).StartVMRequest(startVMRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.StartVM``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `StartVM`: CreateVM200Response
	//fmt.Fprintf(os.Stdout, "Response from `NodesApi.StartVM`: %v\n", resp)
	return resp, err
}

// | **Post** /nodes/{node}/qemu/{vmid}/status/stop | stopVM

func StopVM(node string, vmid int64) (*pxapiflat.CreateVM200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		return nil, err
	}

	stopVMRequest := *pxapiflat.NewStopVMRequest() // StopVMRequest |  (optional)

	resp, r, err := apiClient.NodesApi.StopVM(context, node, vmid).StopVMRequest(stopVMRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.StopVM``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `StopVM`: CreateVM200Response
	//fmt.Fprintf(os.Stdout, "Response from `NodesApi.StopVM`: %v\n", resp)
	return resp, err
}

// | **Post** /nodes/{node}/qemu/{vmid}/status/suspend | suspendVM

func SuspendVM(node string, vmid int64, suspendVMRequest *pxapiflat.SuspendVMRequest) error {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return err
	}

	if suspendVMRequest == nil {
		suspendVMRequest = pxapiflat.NewSuspendVMRequest()
	}

	resp, r, err := apiClient.NodesApi.SuspendVM(context, node, vmid).SuspendVMRequest(*suspendVMRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.SuspendVM``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return err
	}
	// response from `SuspendVM`: CreateVM200Response
	fmt.Fprintf(os.Stdout, "Response from `NodesApi.SuspendVM`: %v\n", resp)
	return nil
}

// | **Post** /nodes/{node}/qemu/{vmid}/config | updateVMConfig

func UpdateVMConfig(node string, vmid int64, updateVMConfigRequest *pxapiflat.UpdateVMConfigRequest) (*pxapiflat.CreateVM200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return nil, err
	}

	if updateVMConfigRequest == nil {
		updateVMConfigRequest = pxapiflat.NewUpdateVMConfigRequest()
	}

	resp, r, err := apiClient.NodesApi.UpdateVMConfig(context, node, vmid).UpdateVMConfigRequest(*updateVMConfigRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.UpdateVMConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	fmt.Fprintf(os.Stdout, "Response from `NodesApi.UpdateVMConfig`: %s\n", *resp.Data)
	return resp, err
}

// | **Put** /nodes/{node}/qemu/{vmid}/config | updateVMConfigSync

func UpdateVMConfigSync(node string, vmid int64, updateVMConfigSyncRequest *pxapiflat.UpdateVMConfigSyncRequest) error {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return err
	}
	if updateVMConfigSyncRequest == nil {
		updateVMConfigSyncRequest = pxapiflat.NewUpdateVMConfigSyncRequest()
	}

	resp, r, err := apiClient.NodesApi.UpdateVMConfigSync(context, node, vmid).UpdateVMConfigSyncRequest(*updateVMConfigSyncRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.UpdateVMConfigSync``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return err
	}
	// response from `UpdateVMConfigSync`: CreateVM200Response
	fmt.Fprintf(os.Stdout, "Response from `NodesApi.UpdateVMConfigSync`: %v\n", resp)
	return nil
}

// | **Put** /nodes/{node}/qemu/{vmid}/snapshot/{snapname}/config | updateVMSnapshotConfig

func UpdateVMSnapshotConfig(node string, vmid int64, snapname string, updateVMSnapshotConfigRequest *pxapiflat.UpdateVMSnapshotConfigRequest) error {

	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return err
	}
	if updateVMSnapshotConfigRequest == nil {
		updateVMSnapshotConfigRequest = pxapiflat.NewUpdateVMSnapshotConfigRequest()
	}

	resp, r, err := apiClient.NodesApi.UpdateVMSnapshotConfig(context, node, vmid, snapname).UpdateVMSnapshotConfigRequest(*updateVMSnapshotConfigRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.UpdateVMSnapshotConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return err
	}
	// response from `UpdateVMSnapshotConfig`: CreateVM200Response
	fmt.Fprintf(os.Stdout, "Response from `NodesApi.UpdateVMSnapshotConfig`: %v\n", resp)
	return nil
}

// | **Post** /nodes/{node}/lxc | createContainer

func CreateContainer(node string, createContainerRequest pxapiflat.CreateContainerRequest) (*pxapiflat.CreateVM200Response, error) {

	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return nil, err
	}
	resp, r, err := apiClient.NodesApi.CreateContainer(context, node).CreateContainerRequest(createContainerRequest).Execute()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.CreateContainer``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	//fmt.Fprintf(os.Stdout, "Response from `NodesApi.CreateContainer`: %v\n", resp)
	return resp, err
}

// | **Post** /nodes/{node}/lxc/{vmid}/snapshot | createContainerSnapshot

func CreateContainerSnapshot(node string, vmid int64, snapshotName string) (*pxapiflat.TaskStartResponse, error) {

	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return nil, err
	}

	createContainerSnapshotRequest := *pxapiflat.NewCreateContainerSnapshotRequest(snapshotName)

	resp, r, err := apiClient.NodesApi.CreateContainerSnapshot(context, node, vmid).CreateContainerSnapshotRequest(createContainerSnapshotRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.CreateContainerSnapshot``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `CreateContainerSnapshot`: CreateVMSnapshot200Response
	//fmt.Fprintf(os.Stdout, "Response from `NodesApi.CreateContainerSnapshot`: %v\n", resp)
	return resp, err
}

// | **Delete** /nodes/{node}/lxc/{vmid} | deleteContainer
func DeleteContainer(node string, vmid int64) (*pxapiflat.CreateVM200Response, error) {

	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return nil, err
	}
	resp, r, err := apiClient.NodesApi.DeleteContainer(context, node, vmid).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.DeleteContainer``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `DeleteContainer`: CreateVM200Response
	//fmt.Fprintf(os.Stdout, "Response from `NodesApi.DeleteContainer`: %v\n", resp)
	return resp, err
}

// FIXME wrong parameter ordering - correct, but needs fixing in the openapi file
// | **Delete** /nodes/{node}/lxc/{vmid}/snapshot/{snapname} | deleteContainerSnapshot

func DeleteContainerSnapshot(node string, vmid int64, snapname string) (*pxapiflat.TaskStartResponse, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return nil, err
	}
	resp, r, err := apiClient.NodesApi.DeleteContainerSnapshot(context, node, vmid, snapname).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.DeleteContainerSnapshot``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `DeleteContainerSnapshot`: CreateVMSnapshot200Response
	//fmt.Fprintf(os.Stdout, "Response from `NodesApi.DeleteContainerSnapshot`: %v\n", resp)
	return resp, err
}

// | **Get** /nodes/{node}/lxc/{vmid} | getContainer

func GetContainer(node string, vmid int64) (*pxapiflat.GetVM200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return nil, err
	}
	resp, r, err := apiClient.NodesApi.GetContainer(context, node, vmid).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.GetContainer``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `GetContainer`: GetVM200Response
	fmt.Fprintf(os.Stdout, "Response from `NodesApi.GetContainer`: %v\n", resp)
	return resp, err
}

// | **Get** /nodes/{node}/lxc/{vmid}/config | getContainerConfig

func GetContainerConfig(node string, vmid int64) (*pxapiflat.GetContainerConfig200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return nil, err
	}

	resp, r, err := apiClient.NodesApi.GetContainerConfig(context, node, vmid).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.GetContainerConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `GetContainerConfig`: GetContainerConfig200Response
	fmt.Fprintf(os.Stdout, "Response from `NodesApi.GetContainerConfig`: %v\n", resp)
	return resp, err
}

// | **Get** /nodes/{node}/lxc/{vmid}/pending | getContainerConfigPending

func GetContainerConfigPending(node string, vmid int64) (*pxapiflat.GetContainerConfigPending200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return nil, err
	}
	resp, r, err := apiClient.NodesApi.GetContainerConfigPending(context, node, vmid).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.GetContainerConfigPending``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `GetContainerConfigPending`: GetContainerConfigPending200Response
	fmt.Fprintf(os.Stdout, "Response from `NodesApi.GetContainerConfigPending`: %v\n", resp)
	return resp, err
}

// FIXME parameter sequence needs to be changed in openapi
// | **Get** /nodes/{node}/lxc/{vmid}/snapshot/{snapname} | getContainerSnapshot

func GetContainerSnapshot(node string, vmid int64, snapname string) (*pxapiflat.GetVMSnapshot200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return nil, err
	}
	resp, r, err := apiClient.NodesApi.GetContainerSnapshot(context, node, vmid, snapname).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.GetContainerSnapshot``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `GetContainerSnapshot`: GetVMSnapshot200Response
	fmt.Fprintf(os.Stdout, "Response from `NodesApi.GetContainerSnapshot`: %v\n", resp)
	return resp, err
}

// | **Get** /nodes/{node}/lxc/{vmid}/snapshot/{snapname}/config | getContainerSnapshotConfig

func GetContainerSnapshotConfig(node string, vmid int64, snapname string) (*pxapiflat.GetVMSnapshotConfig200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return nil, err
	}
	resp, r, err := apiClient.NodesApi.GetContainerSnapshotConfig(context, node, vmid, snapname).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.GetContainerSnapshotConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `GetContainerSnapshotConfig`: GetVMSnapshotConfig200Response
	fmt.Fprintf(os.Stdout, "Response from `NodesApi.GetContainerSnapshotConfig`: %v\n", resp)
	return resp, err
}

// | **Get** /nodes/{node}/lxc/{vmid}/snapshot | getContainerSnapshots

func GetContainerSnapshots(node string, vmid int64) (*pxapiflat.GetContainerSnapshots200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return nil, err
	}

	resp, r, err := apiClient.NodesApi.GetContainerSnapshots(context, node, vmid).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.GetContainerSnapshots``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `GetContainerSnapshots`: GetContainerSnapshots200Response
	//fmt.Fprintf(os.Stdout, "Response from `NodesApi.GetContainerSnapshots`: %v\n", resp)
	return resp, err
}

// | **Get** /nodes/{node}/lxc/{vmid}/status | getContainerStatus

func GetContainerStatus(node string, vmid int64) (*pxapiflat.GetVM200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return nil, err
	}

	resp, r, err := apiClient.NodesApi.GetContainerStatus(context, node, vmid).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.GetContainerStatus``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `GetContainerStatus`: GetVM200Response
	fmt.Fprintf(os.Stdout, "Response from `NodesApi.GetContainerStatus`: %v\n", resp)
	return resp, err
}

// | **Get** /nodes/{node}/lxc | getContainers

func GetContainers(node string) (*pxapiflat.GetContainers200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return nil, err
	}

	resp, r, err := apiClient.NodesApi.GetContainers(context, node).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.GetContainers``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `GetContainers`: GetContainers200Response
	fmt.Fprintf(os.Stdout, "Response from `NodesApi.GetContainers`: %v\n", resp)
	return resp, err
}

// | **Get** /nodes/{node}/lxc/{vmid}/status/current | getCurrentContainerStatus

func GetCurrentContainerStatus(node string, vmid int64) (*pxapiflat.GetCurrentContainerStatus200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return nil, err
	}
	resp, r, err := apiClient.NodesApi.GetCurrentContainerStatus(context, node, vmid).Execute()

	// Not found, we use this function also check if a Container exists, so no error log here
	if r.StatusCode == 500 {
		return nil, err
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.GetCurrentContainerStatus``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `GetCurrentContainerStatus`: GetCurrentContainerStatus200Response
	// fmt.Fprintf(os.Stdout, "Response from `NodesApi.GetCurrentContainerStatus`: %v\n", resp)
	return resp, err
}

// | **Post** /nodes/{node}/lxc/{vmid}/status/reboot | rebootContainer
func RebootContainer(node string, vmid int64) (*pxapiflat.CreateVM200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return nil, err
	}
	rebootContainerRequest := *pxapiflat.NewRebootContainerRequest() // RebootContainerRequest |  (optional)
	resp, r, err := apiClient.NodesApi.RebootContainer(context, node, vmid).RebootContainerRequest(rebootContainerRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.RebootContainer``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `RebootContainer`: CreateVM200Response
	fmt.Fprintf(os.Stdout, "Response from `NodesApi.RebootContainer`: %v\n", resp)
	return resp, err
}

// | **Put** /nodes/{node}/lxc/{vmid}/resize | resizeContainerDisk

func ResizeContainerDisk(node string, vmid int64, resizeContainerDiskRequest pxapiflat.ResizeContainerDiskRequest) (*pxapiflat.TaskStartResponse, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return nil, err
	}
	//resizeContainerDiskRequest := *pxapiflat.NewResizeContainerDiskRequest("Disk_example", "Size_example") // ResizeContainerDiskRequest |  (optional)

	resp, r, err := apiClient.NodesApi.ResizeContainerDisk(context, node, vmid).ResizeContainerDiskRequest(resizeContainerDiskRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.ResizeContainerDisk``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}

	// response from `ResizeContainerDisk`: CreateVMSnapshot200Response
	fmt.Fprintf(os.Stdout, "Response from `NodesApi.ResizeContainerDisk`: %v\n", resp)
	return resp, err
}

// | **Post** /nodes/{node}/lxc/{vmid}/status/resume | resumeContainer

func ResumeContainer(node string, vmid int64) (*pxapiflat.CreateVM200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return nil, err
	}

	//body := map[string]interface{}{ ... } // map[string]interface{} |  (optional)
	//resp, r, err := apiClient.NodesApi.ResumeContainer(context, node, vmid).Body(body).Execute()
	resp, r, err := apiClient.NodesApi.ResumeContainer(context, node, vmid).Execute()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.ResumeContainer``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `ResumeContainer`: CreateVM200Response
	fmt.Fprintf(os.Stdout, "Response from `NodesApi.ResumeContainer`: %v\n", resp)
	return resp, err
}

// | **Post** /nodes/{node}/lxc/{vmid}/snapshot/{snapname}/rollback | rollbackContainerSnapshot

func RollbackContainerSnapshot(node string, vmid int64, snapname string) (*pxapiflat.TaskStartResponse, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return nil, err
	}

	//body := map[string]interface{}{ ... } // map[string]interface{} |  (optional)
	//resp, r, err := apiClient.NodesApi.RollbackContainerSnapshot(context, node, snapname, vmid).Body(body).Execute()
	resp, r, err := apiClient.NodesApi.RollbackContainerSnapshot(context, node, vmid, snapname).Execute()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.RollbackContainerSnapshot``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `RollbackContainerSnapshot`: CreateVMSnapshot200Response
	//fmt.Fprintf(os.Stdout, "Response from `NodesApi.RollbackContainerSnapshot`: %v\n", resp)
	return resp, err
}

// | **Post** /nodes/{node}/lxc/{vmid}/status/shutdown | shutdownContainer

func ShutdownContainer(node string, vmid int64) (*pxapiflat.CreateVM200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return nil, err
	}

	shutdownContainerRequest := *pxapiflat.NewShutdownContainerRequest() // ShutdownContainerRequest |  (optional)

	resp, r, err := apiClient.NodesApi.ShutdownContainer(context, node, vmid).ShutdownContainerRequest(shutdownContainerRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.ShutdownContainer``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `ShutdownContainer`: CreateVM200Response
	//fmt.Fprintf(os.Stdout, "Response from `NodesApi.ShutdownContainer`: %v\n", resp)
	return resp, err
}

// | **Post** /nodes/{node}/lxc/{vmid}/status/start | startContainer

func StartContainer(node string, vmid int64) (*pxapiflat.CreateVM200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return nil, err
	}

	//startContainerRequest := *pxapiflat.NewStartContainerRequest() // StartContainerRequest |  (optional)
	//resp, r, err := apiClient.NodesApi.StartContainer(context.Background(), node, vmid).StartContainerRequest(startContainerRequest).Execute()
	resp, r, err := apiClient.NodesApi.StartContainer(context, node, vmid).Execute()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.StartContainer``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `StartContainer`: CreateVM200Response
	//fmt.Fprintf(os.Stdout, "Response from `NodesApi.StartContainer`: %v\n", resp)
	return resp, err
}

// | **Post** /nodes/{node}/lxc/{vmid}/status/stop | stopContainer

func StopContainer(node string, vmid int64) (*pxapiflat.CreateVM200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return nil, err
	}

	//stopContainerRequest := *openapiclient.NewStopContainerRequest() // StopContainerRequest |  (optional)
	//resp, r, err := apiClient.NodesApi.StopContainer(context.Background(), node, vmid).StopContainerRequest(stopContainerRequest).Execute()
	resp, r, err := apiClient.NodesApi.StopContainer(context, node, vmid).Execute()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.StopContainer``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `StopContainer`: CreateVM200Response
	//fmt.Fprintf(os.Stdout, "Response from `NodesApi.StopContainer`: %v\n", resp)
	return resp, err
}

// | **Post** /nodes/{node}/lxc/{vmid}/status/suspend | suspendContainer

func SuspendContainer(node string, vmid int64) (*pxapiflat.CreateVM200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return nil, err
	}
	//body := map[string]interface{}{ ... } // map[string]interface{} |  (optional)
	//resp, r, err := apiClient.NodesApi.SuspendContainer(context.Background(), node, vmid).Body(body).Execute()
	resp, r, err := apiClient.NodesApi.SuspendContainer(context, node, vmid).Execute()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.SuspendContainer``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `SuspendContainer`: CreateVM200Response
	fmt.Fprintf(os.Stdout, "Response from `NodesApi.SuspendContainer`: %v\n", resp)
	return resp, nil
}

//| **Put** /nodes/{node}/lxc/{vmid}/config | updateContainerConfigSync

func UpdateContainerConfigSync(node string, vmid int64, updateContainerConfigSyncRequest pxapiflat.UpdateContainerConfigSyncRequest) (*pxapiflat.CreateVM200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return nil, err
	}
	//updateContainerConfigSyncRequest := *openapiclient.NewUpdateContainerConfigSyncRequest() // UpdateContainerConfigSyncRequest |  (optional)
	resp, r, err := apiClient.NodesApi.UpdateContainerConfigSync(context, node, vmid).UpdateContainerConfigSyncRequest(updateContainerConfigSyncRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.UpdateContainerConfigSync``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `UpdateContainerConfigSync`: CreateVM200Response
	fmt.Fprintf(os.Stdout, "Response from `NodesApi.UpdateContainerConfigSync`: %v\n", resp)
	return resp, err
}

// | **Put** /nodes/{node}/lxc/{vmid}/snapshot/{snapname}/config | updateContainerSnapshotConfig

func UpdateContainerSnapshotConfig(node string, vmid int64, snapname string, updateContainerSnapshotConfigRequest pxapiflat.UpdateContainerSnapshotConfigRequest) (*pxapiflat.CreateVM200Response, error) {

	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return nil, err
	}

	//updateContainerSnapshotConfigRequest := *openapiclient.NewUpdateContainerSnapshotConfigRequest() // UpdateContainerSnapshotConfigRequest |  (optional)

	//configuration := openapiclient.NewConfiguration()
	//apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodesApi.UpdateContainerSnapshotConfig(context, node, vmid, snapname).UpdateContainerSnapshotConfigRequest(updateContainerSnapshotConfigRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.UpdateContainerSnapshotConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `UpdateContainerSnapshotConfig`: CreateVM200Response
	fmt.Fprintf(os.Stdout, "Response from `NodesApi.UpdateContainerSnapshotConfig`: %v\n", resp)
	return resp, err
}

func GetNodeTaskStatus(node string, upid string) (*pxapiflat.GetNodeTaskStatus200Response, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
		return nil, err
	}
	resp, r, err := apiClient.NodesApi.GetNodeTaskStatus(context, node, upid).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.GetNodeTaskStatus``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	// response from `GetNodeTaskStatus`: GetNodeTaskStatus200Response
	fmt.Fprintf(os.Stdout, "Response from `NodesApi.GetNodeTaskStatus`: %v\n", resp)
	return resp, err
}
