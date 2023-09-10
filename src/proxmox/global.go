package proxmox

import "embed"

var (
	//go:embed test/*
	test embed.FS
)
var PROXMOX_MIN_VMID int = 100
var PROXMOX_MAX_VMID int = 1000000
var PROXMOX_MACHINE_VM = "qemu"
var PROXMOX_MACHINE_CT = "lxc"
