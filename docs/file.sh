#!/bin/bash
sudo qm create 994 --name mgmt --memory 8192 --cores 4 --net0 virtio,bridge=vmbr0 --bios ovmf --efidisk0 data:0
sudo pvesh set /pools/restricted -vms 994
sudo qm importdisk 994 /var/lib/vz/template/iso/ubuntu-noble-grub2.img data
sudo qm set 994 --scsihw virtio-scsi-pci --scsi0 data:vm-994-disk-1
sudo qm set 994 --boot order=scsi0
sudo qm set 994 --net0 virtio,bridge=vmbr0,tag=12
sudo qm start 994

# if disk becomes full, run this command after making the disk bigger under hardware
# select the disk and choose disk options, resize disk, afterwards run
sudo growpart /dev/sda 1
sudo resize2fs /dev/sda1