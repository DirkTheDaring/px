# px


# History
This tool was created out of the need to manage lab systems:
  * infrastructure as code
  * nodes come and go. Still i would not like to unjoin and join nodes permantly. Why not a "virtual cluster"
  * if i screwed up an OS, maybe i just want to flash to os partition but leave the data partition untouched
  * there was no way to create an ignition file for fedora coreos 

# Design goals
* unify commands for containers and virtual machines
* make the use of fedora coreos images possible, by supporting creating ignition file (usining butane)
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

# Setup ignition
edit the file on each node /etc/pve/storage.cfg
add the following entry (iso is important to add!)

dir: ignition
        path /etc/pve/ignition
        content iso,vztmpl,snippets
        prune-backups keep-all=1
        shared 1

mkdir -p /etc/pve/ignition (will be synced over clusters)

# Setup images dir (required)

# Add images to the content line
dir: local
        path /var/lib/vz
        content backup,vztmpl,iso,images

