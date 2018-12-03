# building a new base VM image using packer

## Why

The existing mobiledgex-16.04 base VM image has a number of issues.

* it was manually constructed, its creation steps and details are therefore not documented
* It uses too much disk space in the base
* it does not use all of the disk space allocated to the instantiated VM using the base image beause it does not resize the partition and file system
* related to above, it does not fit itself to different openstack nova flavor types. No matter what flavor types the same amount of disk is used, until you manually lvextend it.

The new way of using `packer` has a few advantages

* this is `code as infrastructure` so we can see how it was created in code and reproduce it and version control it
* it uses cloud-img and cloud-init features so it auto resizes the partition and filesystem as per given flavor
* it uses much less disk space for base

## Step by step

First, install `packer` from hashicorp. https://www.packer.io/docs/install/index.html

Second, make sure you have a working openstack cluster. You can chooose from hamburg, bonn, berlin, munich, ...
Accessing these cloudlets can be enabled by sourcing in environment variables.  Checkout github.com/mobiledgex/bob-priv/mobiledgex.env.tar and untar in your home directory which will create ~/.mobiledgex/ directory containing various environment variable files.
To use hamburg cloudlet source ~/.mobiledgex/hamburg.env

To check out what images are currently available do:

``` 
openstack image list
openstack image list
+--------------------------------------+------------------------------+--------+
| ID                                   | Name                         | Status |
+--------------------------------------+------------------------------+--------+
| a095f3b7-dc42-40c1-8d5f-8b67cc5c7e2a | CentOS-7-x86_64-GenericCloud | active |
| 29261993-89af-405c-bd6d-b2336f9cac4d | Debian-9.5.2-stretch         | active |
| 1b581cac-1fab-4529-b419-2a53ad6f7b36 | cirros-0.4.0                 | active |
| ddbb5dd0-c6d2-4c05-8771-136a61111c93 | mobiledgex                   | active |
| 34f6bab6-db1e-43f6-a102-17324c24f7bc | mobiledgex-16.04-2           | active |
| 6d50ccb3-10df-4d88-9c22-c1ec92f313e9 | mobiledgex-16.04-3           | active |
| 3ccda010-e10e-4844-907d-2fec04ee0201 | mobiledgex-16.04-3           | active |
| 6a7888da-2112-4660-8d5a-8c06ab845d52 | xenial-server                | active |
+--------------------------------------+------------------------------+--------+
```

You need to make sure that xenial-server or the equivalent image is present.  This image is based on xenial-server-cloudimg-amd64-disk1.img which can be downloaded from https://cloud-images.ubuntu.com/xenial/current/

Once you have  xenial-server-cloudimg-amd64-disk1.img downloaed you can create instance of this in openstack glance by doing:

```
openstack image create --file  xenial-server-cloudimg-amd64-disk1.img --disk-format qcow2 xenial-server
```

Note the UUID of xenial-server image returned by `openstack image list`.  In the above example, it is 6a7888da-2112-4660-8d5a-8c06ab845d52.
You will use this UUID to fill in `source_image` field inside `packer_template.mobiledgex.json`.  Unfortunately packer does not
know how to properly use symbolic name so you have to use UUID and get it from the cloudlet's glance service and
manually fill it in.
The same is true of `networks` field of `packer_template.mobiledgex.json`. You need to get the network UUID by doing:

```
openstack network list
+--------------------------------------+-------------------------+--------------------------------------+
| ID                                   | Name                    | Subnets                              |
+--------------------------------------+-------------------------+--------------------------------------+
| 4b4ab7f2-df0f-4fb8-a827-3f99b49d60f4 | waf-m2m                 | c070f5f9-44ec-4250-bcfb-0017573e8995 |
| 604b4b54-3aa7-488b-a381-f872366d1b91 | mex-k8s-net-1           |                                      |
| 9405454b-59e4-4aef-a6df-c2a81305c5bf | waf-external            | 19eab730-715e-4dae-a34e-77338b931839 |
| f99dfbb9-161c-4929-b0de-2eb00e765725 | external-network-shared | 391a345c-25e2-4388-9c66-1d7cc2213cc3 |
+--------------------------------------+-------------------------+--------------------------------------+
```

In the above case we will use UUID for `external-network-shared` as value for `networks` field inside `packer_template.mobiledgex.json`.


Finally,

```
PACKER_LOG=1 packer build packer_template.mobiledgex.json
```

This will produce a lot of debug log output. It takes a while. Finally, when it is finished you have created a new image and instantiated it in cloudlet's glance database. The name of the new image is `image_name` in  `packer_template.mobiledgex.json`.


## mexpacker

The `mexpacker.go` can be run as

```
go run mexpacker.go
```

Make sure you have OS_ and MEX_ environment variable set correctly before running. The `mexpacker` is a utility that installs `xenial-server` image into glance and then retrieves the ID which and long network ID will be used to create the json for packer to run. It automates the above manual steps.
