# px


# History
This tool was created out of the need to manage lab systems:
  * infrastructure as code
  * nodes come and go. Still i would not like to unjoin and join nodes permantly. Why not a "virtual cluster"
  * if i screwed up an OS, maybe i just want to flash to os partition but leave the data partition untouched

# Design goals
* unify commands for containers and virtual machines
* make the use of fedora coreos images possible
* make the maintaince of clusters (multiple machines on different proxmox nodes) easy, so that with one command you can start, stop, shutdown and snapshot containers and virtualmachins

* Undersand the limits the proxmox api. Do not try to extend the proxmox api by things which are hard to maintain

# Environment Variables
* PX_CONFIG_FILE_PATH
* HOME

# This creates an /etc/modules-load.d/persistent-modules.conf
# with the content listed below

persistent_modules:
  - modulea
  - moduleb
