## In this demo
* Three servers with Ubuntu 18.04 and Single port ConnectX5 NIC (firmware-version: 16.23.1020 (MT_0000000010)
```
# kubectl version
Client Version: version.Info{Major:"1", Minor:"11", GitVersion:"v1.11.3", GitCommit:"a4529464e4629c21224b3d52edfe0ea91b072862", GitTreeState:"clean", BuildDate:"2018-09-09T18:02:47Z", GoVersion:"go1.10.3", Compiler:"gc", Platform:"linux/amd64"}
Server Version: version.Info{Major:"1", Minor:"11", GitVersion:"v1.11.7", GitCommit:"65ecaf0671341311ce6aea0edab46ee69f65d59e", GitTreeState:"clean", BuildDate:"2019-01-24T19:22:45Z", GoVersion:"go1.10.7", Compiler:"gc", Platform:"linux/amd64"}
```
* Servers are running  ovs-vswitchd and ovsdb-server pre k8s. (2.10)



### Nvidia OVN-kubernetes with SRIOV support is here:
https://github.com/shahar-klein/nv-ovn-k8s/tree/mlnx-sriov


### All the configuration files and scripts involved can be found here:
https://github.com/shahar-klein/nv-ovn-k8s/tree/mlnx-sriov/demo


### Building the image:
```
# cd dist/images/
# make ubuntu
now you can tag and push - for example:
# docker tag ovn-kube-u shaharklein/ovn-kube-u-demo:latest
# docker push shaharklein/ovn-kube-u-demo:latest
```



### Basic setup: One k8s master(nd-sjc3a-c18-cpu-06) 
### and two workers(nd-sjc3a-c18-cpu-07 and nd-sjc3a-c18-cpu-10)
### No workload on the master
```
# kubeadm init --pod-network-cidr 10.244.0.0/16
# kubectl get nodes
NAME                  STATUS     ROLES     AGE       VERSION
nd-sjc3a-c18-cpu-06   NotReady   master    2m        v1.11.3
 
 
Join two worker nodes
 
 
# kubectl get nodes -o wide
NAME                  STATUS     ROLES     AGE       VERSION   INTERNAL-IP   EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION      CONTAINER-RUNTIME
nd-sjc3a-c18-cpu-06   NotReady   master    3m        v1.11.3   10.0.2.18     <none>        Ubuntu 18.04.1 LTS   4.15.0-45-generic   docker://18.6.1
nd-sjc3a-c18-cpu-07   NotReady   <none>    17s       v1.11.3   10.0.2.19     <none>        Ubuntu 18.04.1 LTS   4.15.0-45-generic   docker://18.6.1
nd-sjc3a-c18-cpu-10   NotReady   <none>    4s        v1.11.2   10.0.2.22     <none>        Ubuntu 18.04.1 LTS   4.15.0-33-generic   docker://18.6.1
```

### SRIOV:
### Use the sriov device plugin
```
# kubectl create -f rdma-sriov-node-config.yaml
# kubectl create -f rdma-device-plugin.yaml
```

### For each node set switchdev mode:
```
# bash set_switchdev_mode.sh enp94s0
```


## Run the k8s ovn daemonset
* In this demo the master ip is 10.0.2.18. You'll need to set the master ip according to your setup in ovnkube-config.yaml 
* In this demo the PF name is enp94s0. You'll need to set the PF name according to your setup in the "ovn" section in ovnkube-master.yaml
```
# kubectl label node nd-sjc3a-c18-cpu-06 node-role.kubernetes.io/master=true --overwrite
# kubectl label node nd-sjc3a-c18-cpu-07 role=node
# kubectl label node nd-sjc3a-c18-cpu-10 role=node
# kubectl apply -f ovn-namespace.yaml
# kubectl apply -f ovn-policy.yaml
# kubectl apply -f ovnkube-config.yaml
# kubectl create -f ovnkube-master.yaml
# kubectl create -f ovnkube.yaml
# kubectl -n ovn-kubernetes get pod -o wide
NAME                   READY     STATUS    RESTARTS   AGE       IP          NODE                  NOMINATED NODE
ovnkube-gfrcp          2/2       Running   0          2m        10.0.2.22   nd-sjc3a-c18-cpu-10   <none>
ovnkube-master-w82lj   4/4       Running   0          4m        10.0.2.18   nd-sjc3a-c18-cpu-06   <none>
ovnkube-wlw6l          2/2       Running   0          2m        10.0.2.19   nd-sjc3a-c18-cpu-07   <none>
```


## Run a service based iperf server and client
```
# kubectl create -f iperf-demo-service.yaml
# kubectl get svc iperf-server-service
NAME                   TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)     AGE
iperf-server-service   ClusterIP   10.109.152.153   <none>        10005/TCP   1h
 
# kubectl create -f iperf-client.yaml
# kubectl get pods
NAME                                       READY     STATUS    RESTARTS   AGE
iperf-server-deployment-6667f8d5dc-wqrmd   1/1       Running   0          1h
iperfk8s-client                            1/1       Running   1          1h
# kubectl exec -it iperfk8s-client -- iperf -c 10.109.152.153 -p 10005 -i 1 -t 3000
```

## For ct/nat rules run south-north traffic
```
# kubectl exec -it iperfk8s-client -- ping 8.8.8.8

```






### Ovn, Before running workload
```
# ovn-nbctl show
switch 1b7fe604-79e0-418e-8512-da719855b5c9 (nd-sjc3a-c18-cpu-10-ovn)
    port kube-system_coredns-78fcdf6894-tcpbx-ovn
        addresses: ["dynamic"]
    port kube-system_coredns-78fcdf6894-vvjvh-ovn
        addresses: ["dynamic"]
    port stor-nd-sjc3a-c18-cpu-10-ovn
        type: router
        addresses: ["00:00:00:E0:D1:90"]
        router-port: rtos-nd-sjc3a-c18-cpu-10-ovn
    port k8s-nd-sjc3a-c18-cpu-10-ovn
        addresses: ["22:18:fc:dd:b2:dd 10.244.2.2"]
switch 7972fcd1-9f0b-4edb-bcc3-7ab0a118f447 (ext_nd-sjc3a-c18-cpu-10-ovn)
    port brenp94s0_nd-sjc3a-c18-cpu-10-ovn
        addresses: ["unknown"]
    port etor-GR_nd-sjc3a-c18-cpu-10-ovn
        type: router
        addresses: ["50:6b:4b:c3:95:44"]
        router-port: rtoe-GR_nd-sjc3a-c18-cpu-10-ovn
switch 1b1548da-dc71-42a3-86c9-11cee877feff (ext_nd-sjc3a-c18-cpu-07-ovn)
    port etor-GR_nd-sjc3a-c18-cpu-07-ovn
        type: router
        addresses: ["50:6b:4b:c3:98:78"]
        router-port: rtoe-GR_nd-sjc3a-c18-cpu-07-ovn
    port brenp94s0_nd-sjc3a-c18-cpu-07-ovn
        addresses: ["unknown"]
switch 92b9ccde-4d1a-45ae-b805-b62bcffbe24c (nd-sjc3a-c18-cpu-07-ovn)
    port stor-nd-sjc3a-c18-cpu-07-ovn
        type: router
        addresses: ["00:00:00:8E:CE:30"]
        router-port: rtos-nd-sjc3a-c18-cpu-07-ovn
    port k8s-nd-sjc3a-c18-cpu-07-ovn
        addresses: ["66:5b:ec:01:88:83 10.244.1.2"]
switch a2e3f59f-ee6e-4ec2-a6d8-ae6010153cb5 (join)
    port jtor-GR_nd-sjc3a-c18-cpu-07-ovn
        type: router
        addresses: ["00:00:00:98:13:44"]
        router-port: rtoj-GR_nd-sjc3a-c18-cpu-07-ovn
    port jtor-ovn_cluster_router
        type: router
        addresses: ["00:00:00:FB:74:C8"]
        router-port: rtoj-ovn_cluster_router
    port jtor-GR_nd-sjc3a-c18-cpu-10-ovn
        type: router
        addresses: ["00:00:00:69:D9:80"]
        router-port: rtoj-GR_nd-sjc3a-c18-cpu-10-ovn
router dbc148cc-7208-4d99-ad8b-7b572a0ae465 (ovn_cluster_router)
    port rtoj-ovn_cluster_router
        mac: "00:00:00:FB:74:C8"
        networks: ["100.64.1.1/24"]
    port rtos-nd-sjc3a-c18-cpu-07-ovn
        mac: "00:00:00:8E:CE:30"
        networks: ["10.244.1.1/24"]
    port rtos-nd-sjc3a-c18-cpu-10-ovn
        mac: "00:00:00:E0:D1:90"
        networks: ["10.244.2.1/24"]
router 3cf30c79-e744-4172-9b83-b34c1c1489e9 (GR_nd-sjc3a-c18-cpu-07-ovn)
    port rtoj-GR_nd-sjc3a-c18-cpu-07-ovn
        mac: "00:00:00:98:13:44"
        networks: ["100.64.1.2/24"]
    port rtoe-GR_nd-sjc3a-c18-cpu-07-ovn
        mac: "50:6b:4b:c3:98:78"
        networks: ["10.0.2.19/20"]
    nat 7d6ae2bb-017f-4a21-acb3-6adcfeadacc4
        external ip: "10.0.2.19"
        logical ip: "10.244.0.0/16"
        type: "snat"
router 0fc0cf07-762b-4f84-8d61-fd7cb678b775 (GR_nd-sjc3a-c18-cpu-10-ovn)
    port rtoe-GR_nd-sjc3a-c18-cpu-10-ovn
        mac: "50:6b:4b:c3:95:44"
        networks: ["10.0.2.22/20"]
    port rtoj-GR_nd-sjc3a-c18-cpu-10-ovn
        mac: "00:00:00:69:D9:80"
        networks: ["100.64.1.3/24"]
    nat edfbc2f3-1193-451c-86a7-e6049454fe8e
        external ip: "10.0.2.22"
        logical ip: "10.244.0.0/16"
        type: "snat"
 
# ovn-sbctl show
Chassis "fb20aa1b-9ae3-4888-a24a-1a6f8d62cb19"
    hostname: "nd-sjc3a-c18-cpu-07"
    Encap geneve
        ip: "10.0.2.19"
        options: {csum="true"}
    Port_Binding "brenp94s0_nd-sjc3a-c18-cpu-07-ovn"
    Port_Binding "rtoe-GR_nd-sjc3a-c18-cpu-07-ovn"
    Port_Binding "rtoj-GR_nd-sjc3a-c18-cpu-07-ovn"
    Port_Binding "jtor-GR_nd-sjc3a-c18-cpu-07-ovn"
    Port_Binding "etor-GR_nd-sjc3a-c18-cpu-07-ovn"
    Port_Binding "k8s-nd-sjc3a-c18-cpu-07-ovn"
Chassis "9e8fa038-dbce-4d5d-b09c-debb7c090d36"
    hostname: "nd-sjc3a-c18-cpu-10"
    Encap geneve
        ip: "10.0.2.22"
        options: {csum="true"}
    Port_Binding "k8s-nd-sjc3a-c18-cpu-10-ovn"
    Port_Binding "brenp94s0_nd-sjc3a-c18-cpu-10-ovn"
    Port_Binding "jtor-GR_nd-sjc3a-c18-cpu-10-ovn"
    Port_Binding "rtoe-GR_nd-sjc3a-c18-cpu-10-ovn"
    Port_Binding "rtoj-GR_nd-sjc3a-c18-cpu-10-ovn"
    Port_Binding "kube-system_coredns-78fcdf6894-vvjvh-ovn"
    Port_Binding "kube-system_coredns-78fcdf6894-tcpbx-ovn"
    Port_Binding "etor-GR_nd-sjc3a-c18-cpu-10-ovn"

```

### Worker node before running any workload
```
# ovs-vsctl show
e73d2967-3e54-4164-9ac6-e05c1fa892b2
    Bridge "brenp94s0"
        fail_mode: standalone
        Port "brenp94s0"
            Interface "brenp94s0"
                type: internal
        Port "k8s-patch-brenp94s0-br-int"
            Interface "k8s-patch-brenp94s0-br-int"
                type: patch
                options: {peer="k8s-patch-br-int-brenp94s0"}
        Port "enp94s0"
            Interface "enp94s0"
    Bridge br-int
        fail_mode: secure
        Port "k8s-nd-sjc3a-c1"
            Interface "k8s-nd-sjc3a-c1"
                type: internal
        Port "ovn-9e8fa0-0"
            Interface "ovn-9e8fa0-0"
                type: geneve
                options: {csum="true", key=flow, remote_ip="10.0.2.22"}
        Port "k8s-patch-br-int-brenp94s0"
            Interface "k8s-patch-br-int-brenp94s0"
                type: patch
                options: {peer="k8s-patch-brenp94s0-br-int"}
        Port br-int
            Interface br-int
                type: internal
    ovs_version: "2.10.1_nv_6bf19aa6e"
```

### Ovn after scheduling the iperf pods
```
# ovn-nbctl show
switch 7fd45f74-858c-4577-b9e9-e4ab7ee7b2d1 (ext_nd-sjc3a-c18-cpu-07-ovn)
    port brenp94s0_nd-sjc3a-c18-cpu-07-ovn
        addresses: ["unknown"]
    port etor-GR_nd-sjc3a-c18-cpu-07-ovn
        type: router
        addresses: ["50:6b:4b:c3:98:78"]
        router-port: rtoe-GR_nd-sjc3a-c18-cpu-07-ovn
switch b5e8ac4d-c27e-447d-867b-f0dbd72b586b (nd-sjc3a-c18-cpu-10-ovn)
    port stor-nd-sjc3a-c18-cpu-10-ovn
        type: router
        addresses: ["00:00:00:ED:82:CB"]
        router-port: rtos-nd-sjc3a-c18-cpu-10-ovn
    port default_iperf-server-deployment-6667f8d5dc-wqrmd-ovn
        addresses: ["dynamic"]
    port k8s-nd-sjc3a-c18-cpu-10-ovn
        addresses: ["1a:31:a3:74:82:9c 10.244.2.2"]
switch 642f2195-8b8b-465e-9e13-ae22b7e2ab2b (ext_nd-sjc3a-c18-cpu-10-ovn)
    port etor-GR_nd-sjc3a-c18-cpu-10-ovn
        type: router
        addresses: ["50:6b:4b:c3:95:44"]
        router-port: rtoe-GR_nd-sjc3a-c18-cpu-10-ovn
    port brenp94s0_nd-sjc3a-c18-cpu-10-ovn
        addresses: ["unknown"]
switch 60978412-ce58-47c5-a429-e49f2c097164 (nd-sjc3a-c18-cpu-07-ovn)
    port default_iperfk8s-client-ovn
        addresses: ["dynamic"]
    port stor-nd-sjc3a-c18-cpu-07-ovn
        type: router
        addresses: ["00:00:00:55:8D:0A"]
        router-port: rtos-nd-sjc3a-c18-cpu-07-ovn
    port k8s-nd-sjc3a-c18-cpu-07-ovn
        addresses: ["1e:7e:f5:db:38:a9 10.244.1.2"]
switch 831c6a62-b141-462c-b5f8-515f299c56bd (join)
    port jtor-GR_nd-sjc3a-c18-cpu-07-ovn
        type: router
        addresses: ["00:00:00:9A:BB:5C"]
        router-port: rtoj-GR_nd-sjc3a-c18-cpu-07-ovn
    port jtor-GR_nd-sjc3a-c18-cpu-10-ovn
        type: router
        addresses: ["00:00:00:40:AA:2B"]
        router-port: rtoj-GR_nd-sjc3a-c18-cpu-10-ovn
    port jtor-ovn_cluster_router
        type: router
        addresses: ["00:00:00:D9:F9:C3"]
        router-port: rtoj-ovn_cluster_router
router 1540a78c-6ee5-44e6-9f18-d325aa6aee4c (GR_nd-sjc3a-c18-cpu-10-ovn)
    port rtoj-GR_nd-sjc3a-c18-cpu-10-ovn
        mac: "00:00:00:40:AA:2B"
        networks: ["100.64.1.2/24"]
    port rtoe-GR_nd-sjc3a-c18-cpu-10-ovn
        mac: "50:6b:4b:c3:95:44"
        networks: ["10.0.2.22/20"]
    nat 73bdb7b0-c6ee-4b63-972f-abea53156075
        external ip: "10.0.2.22"
        logical ip: "10.244.0.0/16"
        type: "snat"
router 45c00a34-46ff-47ef-b361-c78a3d6c948e (GR_nd-sjc3a-c18-cpu-07-ovn)
    port rtoe-GR_nd-sjc3a-c18-cpu-07-ovn
        mac: "50:6b:4b:c3:98:78"
        networks: ["10.0.2.19/20"]
    port rtoj-GR_nd-sjc3a-c18-cpu-07-ovn
        mac: "00:00:00:9A:BB:5C"
        networks: ["100.64.1.3/24"]
    nat 5633161a-92f9-4fc8-8bfd-63ce5f33e7d0
        external ip: "10.0.2.19"
        logical ip: "10.244.0.0/16"
        type: "snat"
router 2af9d896-c1ba-4e3a-a019-9d99305946e4 (ovn_cluster_router)
    port rtos-nd-sjc3a-c18-cpu-10-ovn
        mac: "00:00:00:ED:82:CB"
        networks: ["10.244.2.1/24"]
    port rtoj-ovn_cluster_router
        mac: "00:00:00:D9:F9:C3"
        networks: ["100.64.1.1/24"]
    port rtos-nd-sjc3a-c18-cpu-07-ovn
        mac: "00:00:00:55:8D:0A"
        networks: ["10.244.1.1/24"]
```

### ovs on the worker node after scheduling the iperf pods
```
# ovs-vsctl show
e73d2967-3e54-4164-9ac6-e05c1fa892b2
    Bridge "brenp94s0"
        fail_mode: standalone
        Port "brenp94s0"
            Interface "brenp94s0"
                type: internal
        Port "k8s-patch-brenp94s0-br-int"
            Interface "k8s-patch-brenp94s0-br-int"
                type: patch
                options: {peer="k8s-patch-br-int-brenp94s0"}
        Port "enp94s0"
            Interface "enp94s0"
    Bridge br-int
        fail_mode: secure
        Port "k8s-patch-br-int-brenp94s0"
            Interface "k8s-patch-br-int-brenp94s0"
                type: patch
                options: {peer="k8s-patch-brenp94s0-br-int"}
        Port br-int
            Interface br-int
                type: internal
        Port "k8s-nd-sjc3a-c1"
            Interface "k8s-nd-sjc3a-c1"
                type: internal
        Port "ovn-9e8fa0-0"
            Interface "ovn-9e8fa0-0"
                type: geneve
                options: {csum="true", key=flow, remote_ip="10.0.2.22"}
        Port "enp94s0_0"
            Interface "enp94s0_0"
    ovs_version: "2.10.1_nv_6bf19aa6e"
```

### Node dpctl south-north traffic
```
root@nd-sjc3a-c18-cpu-07:~# ovs-dpctl show
system@ovs-system:
  lookups: hit:120418607 missed:17972 lost:0
  flows: 75
  masks: hit:1148614645 total:24 hit/pkt:9.54
  port 0: ovs-system (internal)
  port 1: br-int (internal)
  port 2: genev_sys_6081 (geneve: packet_type=ptap)
  port 3: k8s-nd-sjc3a-c1 (internal)
  port 4: brenp94s0 (internal)
  port 5: enp94s0
  port 6: enp94s0_0
root@nd-sjc3a-c18-cpu-07:~# ovs-dpctl dump-flows | grep "in_port(6)" 
recirc_id(0),in_port(6),eth(src=00:00:00:00:00:00/01:00:00:00:00:00),eth_type(0x0800),ipv4(dst=8.0.0.0/254.0.0.0,proto=1,frag=no),icmp(type=8/0xf8), packets:33, bytes:3234, used:0.532s, actions:ct(zone=15),recirc(0xc69)
recirc_id(0xc6a),in_port(6),ct_state(-new+est-rel-rpl-inv+trk),ct_label(0/0x1),eth(),eth_type(0x0800),ipv4(src=10.244.1.3,frag=no), packets:52, bytes:5096, used:0.533s, actions:ct(zone=15,nat),recirc(0xc72)
recirc_id(0xc69),in_port(6),ct_state(-new+est-rel-rpl-inv+trk),ct_label(0/0x1),eth(),eth_type(0x0800),ipv4(frag=no), packets:52, bytes:5096, used:0.533s, actions:ct(zone=15,nat),recirc(0xc71)
recirc_id(0xc6b),in_port(6),ct_state(+new-est-rel-rpl-inv+trk),ct_label(0/0x1),eth(src=00:00:00:d9:f9:c3,dst=00:00:00:9a:bb:5c),eth_type(0x0800),ipv4(src=10.244.0.0/255.255.0.0,dst=8.0.0.0/254.0.0.0,ttl=63,frag=no), packets:53, bytes:5194, used:0.533s, actions:set(eth(src=50:6b:4b:c3:98:78,dst=00:aa:aa:aa:aa:aa)),set(ipv4(src=10.244.0.0/255.255.0.0,dst=8.0.0.0/254.0.0.0,ttl=62)),ct(commit,zone=11,nat(src=10.0.2.19)),recirc(0xc6c)
recirc_id(0xc6c),in_port(6),ct_state(-new+est-rel-rpl-inv+trk),ct_label(0/0x1),eth(src=00:00:00:00:00:00/01:00:00:00:00:00,dst=00:aa:aa:aa:aa:aa),eth_type(0x0800),ipv4(frag=no), packets:52, bytes:5096, used:0.533s, actions:ct_clear,ct(commit,zone=64000),5
recirc_id(0xc72),in_port(6),ct_state(-new+est-rel-rpl-inv+trk),ct_label(0/0x1),eth(src=0a:00:00:00:00:04,dst=00:00:00:55:8d:0a),eth_type(0x0800),ipv4(src=10.244.1.2/255.255.255.254,dst=8.0.0.0/254.0.0.0,ttl=64,frag=no), packets:52, bytes:5096, used:0.533s, actions:ct_clear,ct_clear,ct_clear,set(eth(src=00:00:00:d9:f9:c3,dst=00:00:00:9a:bb:5c)),set(ipv4(src=10.244.1.2/255.255.255.254,dst=8.0.0.0/254.0.0.0,ttl=63)),ct(zone=10,nat),recirc(0xc6b)
recirc_id(0),in_port(6),ct_state(-new-est-rel-rpl-inv-trk),ct_label(0/0x1),eth(src=0a:00:00:00:00:04,dst=00:00:00:55:8d:0a),eth_type(0x0806),arp(sip=10.244.1.3,tip=10.244.1.1,op=1/0xff,sha=0a:00:00:00:00:04,tha=00:00:00:00:00:00), packets:0, bytes:0, used:never, actions:userspace(pid=2327846706,slow_path(action))
recirc_id(0xc71),in_port(6),eth(dst=00:00:00:55:8d:0a),eth_type(0x0800),ipv4(proto=1,frag=no),icmp(type=8/0xf8), packets:33, bytes:3234, used:0.533s, actions:ct(zone=15),recirc(0xc6a)
Node dpctl east-west traffic (client pod side)
root@nd-sjc3a-c18-cpu-07:~# ovs-dpctl show
system@ovs-system:
  lookups: hit:124939600 missed:18258 lost:0
  flows: 69
  masks: hit:1193876163 total:18 hit/pkt:9.55
  port 0: ovs-system (internal)
  port 1: br-int (internal)
  port 2: genev_sys_6081 (geneve: packet_type=ptap)
  port 3: k8s-nd-sjc3a-c1 (internal)
  port 4: brenp94s0 (internal)
  port 5: enp94s0
  port 6: enp94s0_0
root@nd-sjc3a-c18-cpu-07:~# ovs-dpctl dump-flows | grep "in_port(6)" 
recirc_id(0),in_port(6),eth(src=00:00:00:00:00:00/01:00:00:00:00:00),eth_type(0x0800),ipv4(dst=10.109.152.153,proto=6,frag=no),tcp_flags(psh|ack), packets:7898, bytes:339306020, used:0.001s, flags:P., actions:ct(zone=15),recirc(0xca9)
recirc_id(0xcb0),in_port(6),ct_state(-new+est-rel-rpl-inv+trk),ct_label(0/0x1),eth(src=0a:00:00:00:00:04,dst=00:00:00:55:8d:0a),eth_type(0x0800),ipv4(src=10.244.1.2/255.255.255.254,dst=10.244.2.3,tos=0/0x3,ttl=64,frag=no), packets:244064, bytes:10879855904, used:0.001s, flags:P., actions:ct_clear,ct_clear,set(tunnel(tun_id=0x3,dst=10.0.2.22,ttl=64,tp_dst=6081,geneve({class=0x102,type=0x80,len=4,0x10003}),flags(df|csum|key))),set(eth(src=00:00:00:ed:82:cb,dst=0a:00:00:00:00:03)),set(ipv4(src=10.244.1.2/255.255.255.254,dst=10.244.2.3,ttl=63)),2
recirc_id(0),in_port(6),eth(src=00:00:00:00:00:00/01:00:00:00:00:00),eth_type(0x0800),ipv4(dst=10.109.152.153,proto=6,frag=no),tcp_flags(ack), packets:236150, bytes:10540604096, used:0.001s, flags:., actions:ct(zone=15),recirc(0xca9)
recirc_id(0xcaf),in_port(6),eth(dst=00:00:00:55:8d:0a),eth_type(0x0800),ipv4(proto=6,frag=no),tcp_flags(ack), packets:236166, bytes:10540614588, used:0.001s, flags:., actions:ct(zone=15),recirc(0xcac)
recirc_id(0xca9),in_port(6),ct_state(-new+est-rel-rpl-inv+trk),ct_label(0/0x1),eth(),eth_type(0x0800),ipv4(frag=no), packets:244057, bytes:10879910710, used:0.001s, flags:P., actions:ct(zone=15,nat),recirc(0xcaf)
recirc_id(0xcaf),in_port(6),eth(dst=00:00:00:55:8d:0a),eth_type(0x0800),ipv4(proto=6,frag=no),tcp_flags(psh|ack), packets:7898, bytes:339306020, used:0.001s, flags:P., actions:ct(zone=15),recirc(0xcac)
recirc_id(0xcac),in_port(6),ct_state(-new+est-rel-rpl-inv+trk),ct_label(0/0x1),eth(),eth_type(0x0800),ipv4(src=10.244.1.3,frag=no), packets:244066, bytes:10879923436, used:0.001s, flags:P., actions:ct(zone=15,nat),recirc(0xcb0)
```

### Node dpctl. east-west tcp from iperf

```
# ovs-dpctl dump-flows | grep "in_port(6)"
recirc_id(0x47),in_port(6),eth(dst=0a:6c:a2:17:be:39),eth_type(0x0800),ipv4(proto=6,frag=no),tcp_flags(syn|ack), packets:5793, bytes:428682, used:6.078s, flags:S., actions:ct(zone=1),recirc(0x80)
recirc_id(0x2d),in_port(6),ct_state(-new+est-rel-rpl-inv+trk),ct_label(0/0x1),eth(),eth_type(0x0800),ipv4(frag=no), packets:43901, bytes:3186968, used:0.098s, flags:PR., actions:ct(zone=15,nat),recirc(0x47)
recirc_id(0x81),in_port(6),eth(dst=00:00:00:00:00:00/01:00:00:00:00:00),eth_type(0x0800),ipv4(frag=no), packets:28969, bytes:2752076, used:6.078s, flags:SFP., actions:3
recirc_id(0x48),in_port(6),ct_state(-new+est-rel-rpl-inv+trk),ct_label(0/0x1),eth(src=0a:00:00:00:00:02,dst=00:00:00:ee:c6:f9),eth_type(0x0800),ipv4(src=10.244.1.4/255.255.255.252,dst=10.0.2.18,ttl=64,frag=no), packets:43901, bytes:3186968, used:0.097s, flags:PR., actions:ct_clear,ct_clear,ct_clear,set(eth(src=00:00:00:77:0c:46,dst=00:00:00:90:9c:e1)),set(ipv4(src=10.244.1.4/255.255.255.252,dst=10.0.2.18,ttl=63)),ct(zone=14,nat),recirc(0x32)
recirc_id(0),in_port(6),eth(src=00:00:00:00:00:00/01:00:00:00:00:00),eth_type(0x0800),ipv4(dst=10.128.0.0/255.128.0.0,proto=6,frag=no),tcp_flags(fin|ack), packets:40, bytes:2640, used:6.077s, flags:F., actions:ct(zone=15),recirc(0x2d)
recirc_id(0x47),in_port(6),eth(dst=00:00:00:ee:c6:f9),eth_type(0x0800),ipv4(proto=6,frag=no),tcp_flags(psh|ack), packets:12, bytes:1350, used:0.097s, flags:P., actions:ct(zone=15),recirc(0x31)
recirc_id(0),in_port(6),eth(src=00:00:00:00:00:00/01:00:00:00:00:00),eth_type(0x0800),ipv4(dst=10.128.0.0/255.128.0.0,proto=6,frag=no),tcp_flags(syn|ack), packets:5793, bytes:428682, used:6.077s, flags:S., actions:ct(zone=15),recirc(0x2d)
recirc_id(0x2d),in_port(6),ct_state(-new+est-rel+rpl-inv+trk),ct_label(0/0x1),eth(),eth_type(0x0800),ipv4(frag=no), packets:28969, bytes:2752076, used:6.077s, flags:SFP., actions:ct(zone=15,nat),recirc(0x47)
recirc_id(0x47),in_port(6),eth(dst=00:00:00:ee:c6:f9),eth_type(0x0800),ipv4(proto=6,frag=no),tcp_flags(ack), packets:37763, bytes:2492370, used:0.097s, flags:., actions:ct(zone=15),recirc(0x31)
recirc_id(0x31),in_port(6),ct_state(-new+est-rel-rpl-inv+trk),ct_label(0/0x1),eth(),eth_type(0x0800),ipv4(src=10.244.1.4/255.255.255.252,frag=no), packets:43902, bytes:3187034, used:0.097s, flags:PR., actions:ct(zone=15,nat),recirc(0x48)
recirc_id(0),in_port(6),eth(src=00:00:00:00:00:00/01:00:00:00:00:00),eth_type(0x0800),ipv4(dst=10.128.0.0/255.128.0.0,proto=6,frag=no),tcp_flags(ack), packets:1042, bytes:68772, used:6.077s, flags:., actions:ct(zone=15),recirc(0x2d)
recirc_id(0),in_port(6),eth(src=00:00:00:00:00:00/01:00:00:00:00:00),eth_type(0x0800),ipv4(dst=10.128.0.0/255.128.0.0,proto=6,frag=no),tcp_flags(psh|ack), packets:520, bytes:105560, used:6.077s, flags:P., actions:ct(zone=15),recirc(0x2d)
recirc_id(0x47),in_port(6),eth(dst=0a:6c:a2:17:be:39),eth_type(0x0800),ipv4(proto=6,frag=no),tcp_flags(psh|ack), packets:62, bytes:12586, used:6.077s, flags:P., actions:ct(zone=1),recirc(0x80)
recirc_id(0x80),in_port(6),ct_state(-new+est-rel+rpl-inv+trk),ct_label(0/0x1),eth(),eth_type(0x0800),ipv4(frag=no), packets:28969, bytes:2752076, used:6.077s, flags:SFP., actions:ct(zone=1,nat),recirc(0x81)
recirc_id(0),in_port(6),eth(src=00:00:00:00:00:00/01:00:00:00:00:00),eth_type(0x0800),ipv4(dst=10.96.0.1,proto=6,frag=no),tcp_flags(psh|ack), packets:12, bytes:1350, used:0.098s, flags:P., actions:ct(zone=15),recirc(0x2d)
recirc_id(0x47),in_port(6),eth(dst=0a:6c:a2:17:be:39),eth_type(0x0800),ipv4(proto=6,frag=no),tcp_flags(ack), packets:1535, bytes:101310, used:6.078s, flags:., actions:ct(zone=1),recirc(0x80)
recirc_id(0x47),in_port(6),eth(dst=0a:6c:a2:17:be:39),eth_type(0x0800),ipv4(proto=6,frag=no),tcp_flags(fin|ack), packets:62, bytes:4092, used:6.078s, flags:F., actions:ct(zone=1),recirc(0x80)
recirc_id(0),in_port(6),eth(src=00:00:00:00:00:00/01:00:00:00:00:00),eth_type(0x0800),ipv4(dst=10.96.0.1,proto=6,frag=no),tcp_flags(ack), packets:37763, bytes:2492370, used:0.098s, flags:., actions:ct(zone=15),recirc(0x2d)
recirc_id(0x32),in_port(6),ct_state(+new-est-rel-rpl-inv+trk),ct_label(0/0x1),eth(src=00:00:00:77:0c:46,dst=00:00:00:90:9c:e1),eth_type(0x0800),ipv4(src=10.244.0.0/255.255.0.0,dst=10.0.2.18,ttl=63,frag=no), packets:43899, bytes:3186900, used:0.098s, flags:SP., actions:set(eth(src=50:6b:4b:c3:98:78,dst=50:6b:4b:c3:98:6c)),set(ipv4(src=10.244.0.0/255.255.0.0,dst=10.0.2.18,ttl=62)),ct(commit,zone=13,nat(src=10.0.2.19)),recirc(0x43)
recirc_id(0x43),in_port(6),ct_state(-new+est-rel-rpl-inv+trk),ct_label(0/0x1),eth(src=00:00:00:00:00:00/01:00:00:00:00:00,dst=50:6b:4b:c3:98:6c),eth_type(0x0800),ipv4(frag=no), packets:43901, bytes:3186968, used:0.098s, flags:PR., actions:ct_clear,ct(commit,zone=64000),5
```

### open flow rules. east-west - iperf
```
# ovs-ofctl dump-flows br-int
 cookie=0x0, duration=58078.883s, table=0, n_packets=1230310, n_bytes=81255320, priority=100,in_port="ovn-9e8fa0-0" actions=move:NXM_NX_TUN_ID[0..23]->OXM_OF_METADATA[0..23],move:NXM_NX_TUN_METADATA0[16..30]->NXM_NX_REG14[0..14],move:NXM_NX_TUN_METADATA0[0..15]->NXM_NX_REG15[0..15],resubmit(,33)
 cookie=0x0, duration=58078.068s, table=0, n_packets=58088, n_bytes=5295792, priority=100,in_port="k8s-nd-sjc3a-c1" actions=load:0x1->NXM_NX_REG13[],load:0x5->NXM_NX_REG11[],load:0x6->NXM_NX_REG12[],load:0x3->OXM_OF_METADATA[],load:0x2->NXM_NX_REG14[],resubmit(,8)
 cookie=0x0, duration=58077.302s, table=0, n_packets=1061160, n_bytes=121069768, priority=100,in_port="k8s-patch-br-in" actions=load:0xa->NXM_NX_REG13[],load:0xc->NXM_NX_REG11[],load:0xb->NXM_NX_REG12[],load:0x6->OXM_OF_METADATA[],load:0x1->NXM_NX_REG14[],resubmit(,8)
 cookie=0x0, duration=58072.092s, table=0, n_packets=72967, n_bytes=5946823, priority=100,in_port="enp94s0_0" actions=load:0xf->NXM_NX_REG13[],load:0x5->NXM_NX_REG11[],load:0x6->NXM_NX_REG12[],load:0x3->OXM_OF_METADATA[],load:0x4->NXM_NX_REG14[],resubmit(,8)
 cookie=0x0, duration=58071.185s, table=0, n_packets=73198, n_bytes=5962204, priority=100,in_port="enp94s0_1" actions=load:0x10->NXM_NX_REG13[],load:0x5->NXM_NX_REG11[],load:0x6->NXM_NX_REG12[],load:0x3->OXM_OF_METADATA[],load:0x3->NXM_NX_REG14[],resubmit(,8)
 cookie=0x0, duration=58033.567s, table=0, n_packets=1470242, n_bytes=61488096764, priority=100,in_port="enp94s0_2" actions=load:0x11->NXM_NX_REG13[],load:0x5->NXM_NX_REG11[],load:0x6->NXM_NX_REG12[],load:0x3->OXM_OF_METADATA[],load:0x5->NXM_NX_REG14[],resubmit(,8)
 cookie=0x8557eb28, duration=58078.068s, table=8, n_packets=0, n_bytes=0, priority=100,metadata=0x3,vlan_tci=0x1000/0x1000 actions=drop
 cookie=0x3911bddc, duration=58078.055s, table=8, n_packets=0, n_bytes=0, priority=100,metadata=0x2,vlan_tci=0x1000/0x1000 actions=drop
 cookie=0x18b2424d, duration=58078.055s, table=8, n_packets=0, n_bytes=0, priority=100,metadata=0x1,vlan_tci=0x1000/0x1000 actions=drop
 cookie=0xce6dec5a, duration=58077.780s, table=8, n_packets=0, n_bytes=0, priority=100,metadata=0x4,vlan_tci=0x1000/0x1000 actions=drop
 cookie=0xb78a7a9c, duration=58077.301s, table=8, n_packets=439134, n_bytes=26348040, priority=100,metadata=0x6,vlan_tci=0x1000/0x1000 actions=drop
 cookie=0xa72dc35f, duration=58077.301s, table=8, n_packets=0, n_bytes=0, priority=100,metadata=0x5,vlan_tci=0x1000/0x1000 actions=drop
 cookie=0xf2037c37, duration=58078.066s, table=8, n_packets=0, n_bytes=0, priority=100,metadata=0x3,dl_src=01:00:00:00:00:00/01:00:00:00:00:00 actions=drop
 cookie=0xef29a420, duration=58078.066s, table=8, n_packets=0, n_bytes=0, priority=100,metadata=0x2,dl_src=01:00:00:00:00:00/01:00:00:00:00:00 actions=drop
 cookie=0x18b2424d, duration=58077.845s, table=8, n_packets=0, n_bytes=0, priority=100,metadata=0x1,dl_src=01:00:00:00:00:00/01:00:00:00:00:00 actions=drop
 cookie=0x8099f91, duration=58077.780s, table=8, n_packets=0, n_bytes=0, priority=100,metadata=0x4,dl_src=01:00:00:00:00:00/01:00:00:00:00:00 actions=drop
 cookie=0x5e176c0d, duration=58077.301s, table=8, n_packets=0, n_bytes=0, priority=100,metadata=0x6,dl_src=01:00:00:00:00:00/01:00:00:00:00:00 actions=drop
 cookie=0xa72dc35f, duration=58077.300s, table=8, n_packets=0, n_bytes=0, priority=100,metadata=0x5,dl_src=01:00:00:00:00:00/01:00:00:00:00:00 actions=drop
 cookie=0xea8e33f2, duration=58078.068s, table=8, n_packets=88274, n_bytes=6409735, priority=50,reg14=0x1,metadata=0x2 actions=resubmit(,9)
 cookie=0xc26fe3bb, duration=58078.066s, table=8, n_packets=114824, n_bytes=63806778, priority=50,reg14=0x1,metadata=0x3 actions=resubmit(,9)
 cookie=0x8da2d8ec, duration=58078.066s, table=8, n_packets=58088, n_bytes=5295792, priority=50,reg14=0x2,metadata=0x3 actions=resubmit(,9)
 cookie=0x7d97c0de, duration=58077.785s, table=8, n_packets=1470133, n_bytes=61488064222, priority=50,reg14=0x1,metadata=0x4 actions=resubmit(,9)
 cookie=0x7bf869dc, duration=58077.302s, table=8, n_packets=114816, n_bytes=63806298, priority=50,reg14=0x2,metadata=0x2 actions=resubmit(,9)
 cookie=0xbbb75259, duration=58077.300s, table=8, n_packets=88274, n_bytes=6409543, priority=50,reg14=0x2,metadata=0x6 actions=resubmit(,9)
 cookie=0xa7facb5, duration=58077.300s, table=8, n_packets=622026, n_bytes=94721728, priority=50,reg14=0x1,metadata=0x6 actions=resubmit(,9)
 cookie=0xb1664416, duration=58072.092s, table=8, n_packets=72967, n_bytes=5946823, priority=50,reg14=0x4,metadata=0x3 actions=resubmit(,9)
 cookie=0x36b0216b, duration=58071.185s, table=8, n_packets=73198, n_bytes=5962204, priority=50,reg14=0x3,metadata=0x3 actions=resubmit(,9)
 cookie=0x3236e539, duration=58033.567s, table=8, n_packets=1470242, n_bytes=61488096764, priority=50,reg14=0x5,metadata=0x3 actions=resubmit(,9)
 cookie=0x33f4f94c, duration=58078.068s, table=8, n_packets=114816, n_bytes=63806298, priority=50,reg14=0x1,metadata=0x1,dl_dst=00:00:00:77:0c:46 actions=resubmit(,9)
 cookie=0x9eb7f379, duration=58077.814s, table=8, n_packets=1558412, n_bytes=61494474257, priority=50,reg14=0x2,metadata=0x1,dl_dst=00:00:00:ee:c6:f9 actions=resubmit(,9)
 cookie=0xc853654b, duration=58077.814s, table=8, n_packets=0, n_bytes=0, priority=50,reg14=0x3,metadata=0x1,dl_dst=00:00:00:37:69:8b actions=resubmit(,9)
 cookie=0xe08183f5, duration=58077.301s, table=8, n_packets=88274, n_bytes=6409735, priority=50,reg14=0x1,metadata=0x5,dl_dst=00:00:00:90:9c:e1 actions=resubmit(,9)
 cookie=0xbee0ff83, duration=58077.300s, table=8, n_packets=114822, n_bytes=63806658, priority=50,reg14=0x2,metadata=0x5,dl_dst=50:6b:4b:c3:98:78 actions=resubmit(,9)
 cookie=0xd44cc774, duration=58077.846s, table=8, n_packets=0, n_bytes=0, priority=50,reg14=0x1,metadata=0x1,dl_dst=01:00:00:00:00:00/01:00:00:00:00:00 actions=resubmit(,9)
 cookie=0xd397b601, duration=58077.846s, table=8, n_packets=4, n_bytes=270, priority=50,reg14=0x2,metadata=0x1,dl_dst=01:00:00:00:00:00/01:00:00:00:00:00 actions=resubmit(,9)
 cookie=0x7d680e7, duration=58077.814s, table=8, n_packets=0, n_bytes=0, priority=50,reg14=0x3,metadata=0x1,dl_dst=01:00:00:00:00:00/01:00:00:00:00:00 actions=resubmit(,9)
 cookie=0x62ba6345, duration=58077.301s, table=8, n_packets=492124, n_bytes=29635876, priority=50,reg14=0x2,metadata=0x5,dl_dst=01:00:00:00:00:00/01:00:00:00:00:00 actions=resubmit(,9)
 cookie=0x261b2850, duration=58077.301s, table=8, n_packets=0, n_bytes=0, priority=50,reg14=0x1,metadata=0x5,dl_dst=01:00:00:00:00:00/01:00:00:00:00:00 actions=resubmit(,9)
 cookie=0x1fc2d04a, duration=58078.068s, table=9, n_packets=0, n_bytes=0, priority=100,ip,metadata=0x1,nw_src=255.255.255.255 actions=drop
 cookie=0x98c6081d, duration=58077.300s, table=9, n_packets=0, n_bytes=0, priority=100,ip,metadata=0x5,nw_src=255.255.255.255 actions=drop
 cookie=0x1fc2d04a, duration=58078.066s, table=9, n_packets=0, n_bytes=0, priority=100,ip,metadata=0x1,nw_dst=0.0.0.0/8 actions=drop
 cookie=0x1fc2d04a, duration=58077.846s, table=9, n_packets=0, n_bytes=0, priority=100,ip,metadata=0x1,nw_dst=127.0.0.0/8 actions=drop
 cookie=0x98c6081d, duration=58077.301s, table=9, n_packets=0, n_bytes=0, priority=100,ip,metadata=0x5,nw_dst=0.0.0.0/8 actions=drop
 cookie=0x98c6081d, duration=58077.302s, table=9, n_packets=0, n_bytes=0, priority=100,ip,metadata=0x5,nw_dst=127.0.0.0/8 actions=drop
 cookie=0x66e033db, duration=58078.067s, table=9, n_packets=0, n_bytes=0, priority=100,ip,reg9=0/0x2,metadata=0x1,nw_src=10.244.1.255 actions=drop
 cookie=0x4f706a24, duration=58078.067s, table=9, n_packets=0, n_bytes=0, priority=100,ip,reg9=0/0x2,metadata=0x1,nw_src=100.64.1.255 actions=drop
 cookie=0x4f706a24, duration=58077.847s, table=9, n_packets=0, n_bytes=0, priority=100,ip,reg9=0/0x2,metadata=0x1,nw_src=100.64.1.1 actions=drop
 cookie=0x66e033db, duration=58077.846s, table=9, n_packets=0, n_bytes=0, priority=100,ip,reg9=0/0x2,metadata=0x1,nw_src=10.244.1.1 actions=drop
 cookie=0xceae16d0, duration=58077.814s, table=9, n_packets=0, n_bytes=0, priority=100,ip,reg9=0/0x2,metadata=0x1,nw_src=10.244.2.1 actions=drop
 cookie=0xceae16d0, duration=58077.814s, table=9, n_packets=0, n_bytes=0, priority=100,ip,reg9=0/0x2,metadata=0x1,nw_src=10.244.2.255 actions=drop
 cookie=0xf0e89e58, duration=58077.302s, table=9, n_packets=0, n_bytes=0, priority=100,ip,reg9=0/0x2,metadata=0x5,nw_src=10.0.2.19 actions=drop
 cookie=0xf0e89e58, duration=58077.302s, table=9, n_packets=0, n_bytes=0, priority=100,ip,reg9=0/0x2,metadata=0x5,nw_src=10.0.15.255 actions=drop
 cookie=0x3baf9e27, duration=58077.302s, table=9, n_packets=0, n_bytes=0, priority=100,ip,reg9=0/0x2,metadata=0x5,nw_src=100.64.1.255 actions=drop
 cookie=0x3baf9e27, duration=58077.302s, table=9, n_packets=0, n_bytes=0, priority=100,ip,reg9=0/0x2,metadata=0x5,nw_src=100.64.1.2 actions=drop
 cookie=0x1fc2d04a, duration=58078.067s, table=9, n_packets=0, n_bytes=0, priority=100,ip,metadata=0x1,nw_dst=224.0.0.0/4 actions=drop
 cookie=0x98c6081d, duration=58077.302s, table=9, n_packets=0, n_bytes=0, priority=100,ip,metadata=0x5,nw_dst=224.0.0.0/4 actions=drop
 cookie=0x1a9c47d6, duration=58078.057s, table=9, n_packets=0, n_bytes=0, priority=100,ipv6,metadata=0x1,ipv6_src=fe80::200:ff:fe77:c46 actions=drop
 cookie=0x2bbea951, duration=58078.056s, table=9, n_packets=0, n_bytes=0, priority=100,ipv6,metadata=0x1,ipv6_src=fe80::200:ff:feee:c6f9 actions=drop
 cookie=0xbf952404, duration=58077.815s, table=9, n_packets=0, n_bytes=0, priority=100,ipv6,metadata=0x1,ipv6_src=fe80::200:ff:fe37:698b actions=drop
 cookie=0xa3b9ab0c, duration=58077.302s, table=9, n_packets=0, n_bytes=0, priority=100,ipv6,metadata=0x5,ipv6_src=fe80::200:ff:fe90:9ce1 actions=drop
 cookie=0x1215d874, duration=58077.302s, table=9, n_packets=28, n_bytes=2000, priority=100,ipv6,metadata=0x5,ipv6_src=fe80::526b:4bff:fec3:9878 actions=drop
 cookie=0x1fc2d04a, duration=58077.847s, table=9, n_packets=0, n_bytes=0, priority=100,ip,metadata=0x1,nw_src=0.0.0.0/8 actions=drop
 cookie=0x1fc2d04a, duration=58077.847s, table=9, n_packets=0, n_bytes=0, priority=100,ip,metadata=0x1,nw_src=127.0.0.0/8 actions=drop
 cookie=0x98c6081d, duration=58077.302s, table=9, n_packets=0, n_bytes=0, priority=100,ip,metadata=0x5,nw_src=0.0.0.0/8 actions=drop
 cookie=0x98c6081d, duration=58077.301s, table=9, n_packets=0, n_bytes=0, priority=100,ip,metadata=0x5,nw_src=127.0.0.0/8 actions=drop
 cookie=0x6d72f186, duration=58078.069s, table=9, n_packets=0, n_bytes=0, priority=90,icmp6,metadata=0x1,nw_ttl=255,icmp_type=136,icmp_code=0 actions=push:NXM_NX_XXREG0[],push:NXM_OF_ETH_SRC[],push:NXM_NX_ND_TLL[],push:NXM_NX_ND_TARGET[],pop:NXM_NX_XXREG0[],pop:NXM_OF_ETH_SRC[],controller(userdata=00.00.00.04.00.00.00.00),pop:NXM_OF_ETH_SRC[],pop:NXM_NX_XXREG0[]
 cookie=0xe1837110, duration=58077.847s, table=9, n_packets=0, n_bytes=0, priority=80,icmp6,metadata=0x1,nw_ttl=255,icmp_type=135,icmp_code=0 actions=push:NXM_NX_XXREG0[],push:NXM_OF_ETH_SRC[],push:NXM_NX_ND_SLL[],push:NXM_NX_IPV6_SRC[],pop:NXM_NX_XXREG0[],pop:NXM_OF_ETH_SRC[],controller(userdata=00.00.00.04.00.00.00.00),pop:NXM_OF_ETH_SRC[],pop:NXM_NX_XXREG0[]
 cookie=0x9496be5f, duration=58077.301s, table=9, n_packets=0, n_bytes=0, priority=80,icmp6,metadata=0x5,nw_ttl=255,icmp_type=135,icmp_code=0 actions=push:NXM_NX_XXREG0[],push:NXM_OF_ETH_SRC[],push:NXM_NX_ND_SLL[],push:NXM_NX_IPV6_SRC[],pop:NXM_NX_XXREG0[],pop:NXM_OF_ETH_SRC[],controller(userdata=00.00.00.04.00.00.00.00),pop:NXM_OF_ETH_SRC[],pop:NXM_NX_XXREG0[]
 cookie=0x7942c29, duration=58077.301s, table=9, n_packets=0, n_bytes=0, priority=90,icmp6,metadata=0x5,nw_ttl=255,icmp_type=136,icmp_code=0 actions=push:NXM_NX_XXREG0[],push:NXM_OF_ETH_SRC[],push:NXM_NX_ND_TLL[],push:NXM_NX_ND_TARGET[],pop:NXM_NX_XXREG0[],pop:NXM_OF_ETH_SRC[],controller(userdata=00.00.00.04.00.00.00.00),pop:NXM_OF_ETH_SRC[],pop:NXM_NX_XXREG0[]
 cookie=0xa930ea0f, duration=58078.067s, table=9, n_packets=8, n_bytes=480, priority=90,arp,reg14=0x2,metadata=0x1,arp_spa=10.244.1.0/24,arp_tpa=10.244.1.1,arp_op=1 actions=push:NXM_NX_REG0[],push:NXM_OF_ETH_SRC[],push:NXM_NX_ARP_SHA[],push:NXM_OF_ARP_SPA[],pop:NXM_NX_REG0[],pop:NXM_OF_ETH_SRC[],controller(userdata=00.00.00.01.00.00.00.00),pop:NXM_OF_ETH_SRC[],pop:NXM_NX_REG0[],move:NXM_OF_ETH_SRC[]->NXM_OF_ETH_DST[],mod_dl_src:00:00:00:ee:c6:f9,load:0x2->NXM_OF_ARP_OP[],move:NXM_NX_ARP_SHA[]->NXM_NX_ARP_THA[],load:0xeec6f9->NXM_NX_ARP_SHA[],move:NXM_OF_ARP_SPA[]->NXM_OF_ARP_TPA[],load:0xaf40101->NXM_OF_ARP_SPA[],load:0x2->NXM_NX_REG15[],load:0x1->NXM_NX_REG10[0],resubmit(,32)
 cookie=0xb17c1b98, duration=58077.847s, table=9, n_packets=0, n_bytes=0, priority=90,arp,reg14=0x1,metadata=0x1,arp_spa=100.64.1.0/24,arp_tpa=100.64.1.1,arp_op=1 actions=push:NXM_NX_REG0[],push:NXM_OF_ETH_SRC[],push:NXM_NX_ARP_SHA[],push:NXM_OF_ARP_SPA[],pop:NXM_NX_REG0[],pop:NXM_OF_ETH_SRC[],controller(userdata=00.00.00.01.00.00.00.00),pop:NXM_OF_ETH_SRC[],pop:NXM_NX_REG0[],move:NXM_OF_ETH_SRC[]->NXM_OF_ETH_DST[],mod_dl_src:00:00:00:77:0c:46,load:0x2->NXM_OF_ARP_OP[],move:NXM_NX_ARP_SHA[]->NXM_NX_ARP_THA[],load:0x770c46->NXM_NX_ARP_SHA[],move:NXM_OF_ARP_SPA[]->NXM_OF_ARP_TPA[],load:0x64400101->NXM_OF_ARP_SPA[],load:0x1->NXM_NX_REG15[],load:0x1->NXM_NX_REG10[0],resubmit(,32)
 cookie=0x92187c9, duration=58077.815s, table=9, n_packets=0, n_bytes=0, priority=90,arp,reg14=0x3,metadata=0x1,arp_spa=10.244.2.0/24,arp_tpa=10.244.2.1,arp_op=1 actions=push:NXM_NX_REG0[],push:NXM_OF_ETH_SRC[],push:NXM_NX_ARP_SHA[],push:NXM_OF_ARP_SPA[],pop:NXM_NX_REG0[],pop:NXM_OF_ETH_SRC[],controller(userdata=00.00.00.01.00.00.00.00),pop:NXM_OF_ETH_SRC[],pop:NXM_NX_REG0[],move:NXM_OF_ETH_SRC[]->NXM_OF_ETH_DST[],mod_dl_src:00:00:00:37:69:8b,load:0x2->NXM_OF_ARP_OP[],move:NXM_NX_ARP_SHA[]->NXM_NX_ARP_THA[],load:0x37698b->NXM_NX_ARP_SHA[],move:NXM_OF_ARP_SPA[]->NXM_OF_ARP_TPA[],load:0xaf40201->NXM_OF_ARP_SPA[],load:0x3->NXM_NX_REG15[],load:0x1->NXM_NX_REG10[0],resubmit(,32)
 cookie=0xe2bfc682, duration=58077.303s, table=9, n_packets=0, n_bytes=0, priority=90,arp,reg14=0x1,metadata=0x5,arp_spa=100.64.1.0/24,arp_tpa=100.64.1.2,arp_op=1 actions=push:NXM_NX_REG0[],push:NXM_OF_ETH_SRC[],push:NXM_NX_ARP_SHA[],push:NXM_OF_ARP_SPA[],pop:NXM_NX_REG0[],pop:NXM_OF_ETH_SRC[],controller(userdata=00.00.00.01.00.00.00.00),pop:NXM_OF_ETH_SRC[],pop:NXM_NX_REG0[],move:NXM_OF_ETH_SRC[]->NXM_OF_ETH_DST[],mod_dl_src:00:00:00:90:9c:e1,load:0x2->NXM_OF_ARP_OP[],move:NXM_NX_ARP_SHA[]->NXM_NX_ARP_THA[],load:0x909ce1->NXM_NX_ARP_SHA[],move:NXM_OF_ARP_SPA[]->NXM_OF_ARP_TPA[],load:0x64400102->NXM_OF_ARP_SPA[],load:0x1->NXM_NX_REG15[],load:0x1->NXM_NX_REG10[0],resubmit(,32)
 cookie=0x12efeb83, duration=58078.067s, table=9, n_packets=0, n_bytes=0, priority=90,icmp,metadata=0x1,nw_dst=100.64.1.1,icmp_type=8,icmp_code=0 actions=push:NXM_OF_IP_SRC[],push:NXM_OF_IP_DST[],pop:NXM_OF_IP_SRC[],pop:NXM_OF_IP_DST[],load:0xff->NXM_NX_IP_TTL[],load:0->NXM_OF_ICMP_TYPE[],load:0x1->NXM_NX_REG10[0],resubmit(,10)
 cookie=0xd3dc5db1, duration=58078.057s, table=9, n_packets=0, n_bytes=0, priority=90,icmp,metadata=0x1,nw_dst=10.244.1.1,icmp_type=8,icmp_code=0 actions=push:NXM_OF_IP_SRC[],push:NXM_OF_IP_DST[],pop:NXM_OF_IP_SRC[],pop:NXM_OF_IP_DST[],load:0xff->NXM_NX_IP_TTL[],load:0->NXM_OF_ICMP_TYPE[],load:0x1->NXM_NX_REG10[0],resubmit(,10)
 cookie=0x4db03d6d, duration=58077.814s, table=9, n_packets=0, n_bytes=0, priority=90,icmp,metadata=0x1,nw_dst=10.244.2.1,icmp_type=8,icmp_code=0 actions=push:NXM_OF_IP_SRC[],push:NXM_OF_IP_DST[],pop:NXM_OF_IP_SRC[],pop:NXM_OF_IP_DST[],load:0xff->NXM_NX_IP_TTL[],load:0->NXM_OF_ICMP_TYPE[],load:0x1->NXM_NX_REG10[0],resubmit(,10)
 cookie=0x4045c36e, duration=58077.302s, table=9, n_packets=0, n_bytes=0, priority=90,icmp,metadata=0x5,nw_dst=100.64.1.2,icmp_type=8,icmp_code=0 actions=push:NXM_OF_IP_SRC[],push:NXM_OF_IP_DST[],pop:NXM_OF_IP_SRC[],pop:NXM_OF_IP_DST[],load:0xff->NXM_NX_IP_TTL[],load:0->NXM_OF_ICMP_TYPE[],load:0x1->NXM_NX_REG10[0],resubmit(,10)
 cookie=0xde387b80, duration=58077.301s, table=9, n_packets=0, n_bytes=0, priority=90,icmp,metadata=0x5,nw_dst=10.0.2.19,icmp_type=8,icmp_code=0 actions=push:NXM_OF_IP_SRC[],push:NXM_OF_IP_DST[],pop:NXM_OF_IP_SRC[],pop:NXM_OF_IP_DST[],load:0xff->NXM_NX_IP_TTL[],load:0->NXM_OF_ICMP_TYPE[],load:0x1->NXM_NX_REG10[0],resubmit(,10)
 cookie=0xf104919c, duration=58078.067s, table=9, n_packets=0, n_bytes=0, priority=90,icmp6,reg14=0x1,metadata=0x1,ipv6_dst=ff02::1:ff77:c46,nw_ttl=255,icmp_type=135,icmp_code=0,nd_target=fe80::200:ff:fe77:c46 actions=push:NXM_NX_XXREG0[],push:NXM_OF_ETH_SRC[],push:NXM_NX_ND_SLL[],push:NXM_NX_IPV6_SRC[],pop:NXM_NX_XXREG0[],pop:NXM_OF_ETH_SRC[],controller(userdata=00.00.00.04.00.00.00.00),pop:NXM_OF_ETH_SRC[],pop:NXM_NX_XXREG0[],controller(userdata=00.00.00.0c.00.00.00.00.00.19.00.10.80.00.08.06.00.00.00.77.0c.46.00.00.00.19.00.18.80.00.34.10.fe.80.00.00.00.00.00.00.02.00.00.ff.fe.77.0c.46.00.19.00.18.80.00.3e.10.fe.80.00.00.00.00.00.00.02.00.00.ff.fe.77.0c.46.00.19.00.10.80.00.42.06.00.00.00.77.0c.46.00.00.ff.ff.00.18.00.00.23.20.00.06.00.20.00.00.00.00.00.01.1c.04.00.01.1e.04.ff.ff.00.18.00.00.23.20.00.07.00.00.00.01.14.04.00.00.00.00.00.00.00.01.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.20.00.00.00)
 cookie=0xf104919c, duration=58078.057s, table=9, n_packets=0, n_bytes=0, priority=90,icmp6,reg14=0x1,metadata=0x1,ipv6_dst=fe80::200:ff:fe77:c46,nw_ttl=255,icmp_type=135,icmp_code=0,nd_target=fe80::200:ff:fe77:c46 actions=push:NXM_NX_XXREG0[],push:NXM_OF_ETH_SRC[],push:NXM_NX_ND_SLL[],push:NXM_NX_IPV6_SRC[],pop:NXM_NX_XXREG0[],pop:NXM_OF_ETH_SRC[],controller(userdata=00.00.00.04.00.00.00.00),pop:NXM_OF_ETH_SRC[],pop:NXM_NX_XXREG0[],controller(userdata=00.00.00.0c.00.00.00.00.00.19.00.10.80.00.08.06.00.00.00.77.0c.46.00.00.00.19.00.18.80.00.34.10.fe.80.00.00.00.00.00.00.02.00.00.ff.fe.77.0c.46.00.19.00.18.80.00.3e.10.fe.80.00.00.00.00.00.00.02.00.00.ff.fe.77.0c.46.00.19.00.10.80.00.42.06.00.00.00.77.0c.46.00.00.ff.ff.00.18.00.00.23.20.00.06.00.20.00.00.00.00.00.01.1c.04.00.01.1e.04.ff.ff.00.18.00.00.23.20.00.07.00.00.00.01.14.04.00.00.00.00.00.00.00.01.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.20.00.00.00)
 cookie=0x36275646, duration=58078.056s, table=9, n_packets=0, n_bytes=0, priority=90,icmp6,reg14=0x2,metadata=0x1,ipv6_dst=ff02::1:ffee:c6f9,nw_ttl=255,icmp_type=135,icmp_code=0,nd_target=fe80::200:ff:feee:c6f9 actions=push:NXM_NX_XXREG0[],push:NXM_OF_ETH_SRC[],push:NXM_NX_ND_SLL[],push:NXM_NX_IPV6_SRC[],pop:NXM_NX_XXREG0[],pop:NXM_OF_ETH_SRC[],controller(userdata=00.00.00.04.00.00.00.00),pop:NXM_OF_ETH_SRC[],pop:NXM_NX_XXREG0[],controller(userdata=00.00.00.0c.00.00.00.00.00.19.00.10.80.00.08.06.00.00.00.ee.c6.f9.00.00.00.19.00.18.80.00.34.10.fe.80.00.00.00.00.00.00.02.00.00.ff.fe.ee.c6.f9.00.19.00.18.80.00.3e.10.fe.80.00.00.00.00.00.00.02.00.00.ff.fe.ee.c6.f9.00.19.00.10.80.00.42.06.00.00.00.ee.c6.f9.00.00.ff.ff.00.18.00.00.23.20.00.06.00.20.00.00.00.00.00.01.1c.04.00.01.1e.04.ff.ff.00.18.00.00.23.20.00.07.00.00.00.01.14.04.00.00.00.00.00.00.00.01.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.20.00.00.00)
 cookie=0x36275646, duration=58077.815s, table=9, n_packets=0, n_bytes=0, priority=90,icmp6,reg14=0x2,metadata=0x1,ipv6_dst=fe80::200:ff:feee:c6f9,nw_ttl=255,icmp_type=135,icmp_code=0,nd_target=fe80::200:ff:feee:c6f9 actions=push:NXM_NX_XXREG0[],push:NXM_OF_ETH_SRC[],push:NXM_NX_ND_SLL[],push:NXM_NX_IPV6_SRC[],pop:NXM_NX_XXREG0[],pop:NXM_OF_ETH_SRC[],controller(userdata=00.00.00.04.00.00.00.00),pop:NXM_OF_ETH_SRC[],pop:NXM_NX_XXREG0[],controller(userdata=00.00.00.0c.00.00.00.00.00.19.00.10.80.00.08.06.00.00.00.ee.c6.f9.00.00.00.19.00.18.80.00.34.10.fe.80.00.00.00.00.00.00.02.00.00.ff.fe.ee.c6.f9.00.19.00.18.80.00.3e.10.fe.80.00.00.00.00.00.00.02.00.00.ff.fe.ee.c6.f9.00.19.00.10.80.00.42.06.00.00.00.ee.c6.f9.00.00.ff.ff.00.18.00.00.23.20.00.06.00.20.00.00.00.00.00.01.1c.04.00.01.1e.04.ff.ff.00.18.00.00.23.20.00.07.00.00.00.01.14.04.00.00.00.00.00.00.00.01.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.20.00.00.00)
 cookie=0x9be7d62e, duration=58077.815s, table=9, n_packets=0, n_bytes=0, priority=90,icmp6,reg14=0x3,metadata=0x1,ipv6_dst=ff02::1:ff37:698b,nw_ttl=255,icmp_type=135,icmp_code=0,nd_target=fe80::200:ff:fe37:698b actions=push:NXM_NX_XXREG0[],push:NXM_OF_ETH_SRC[],push:NXM_NX_ND_SLL[],push:NXM_NX_IPV6_SRC[],pop:NXM_NX_XXREG0[],pop:NXM_OF_ETH_SRC[],controller(userdata=00.00.00.04.00.00.00.00),pop:NXM_OF_ETH_SRC[],pop:NXM_NX_XXREG0[],controller(userdata=00.00.00.0c.00.00.00.00.00.19.00.10.80.00.08.06.00.00.00.37.69.8b.00.00.00.19.00.18.80.00.34.10.fe.80.00.00.00.00.00.00.02.00.00.ff.fe.37.69.8b.00.19.00.18.80.00.3e.10.fe.80.00.00.00.00.00.00.02.00.00.ff.fe.37.69.8b.00.19.00.10.80.00.42.06.00.00.00.37.69.8b.00.00.ff.ff.00.18.00.00.23.20.00.06.00.20.00.00.00.00.00.01.1c.04.00.01.1e.04.ff.ff.00.18.00.00.23.20.00.07.00.00.00.01.14.04.00.00.00.00.00.00.00.01.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.20.00.00.00)
 cookie=0x9be7d62e, duration=58077.815s, table=9, n_packets=0, n_bytes=0, priority=90,icmp6,reg14=0x3,metadata=0x1,ipv6_dst=fe80::200:ff:fe37:698b,nw_ttl=255,icmp_type=135,icmp_code=0,nd_target=fe80::200:ff:fe37:698b actions=push:NXM_NX_XXREG0[],push:NXM_OF_ETH_SRC[],push:NXM_NX_ND_SLL[],push:NXM_NX_IPV6_SRC[],pop:NXM_NX_XXREG0[],pop:NXM_OF_ETH_SRC[],controller(userdata=00.00.00.04.00.00.00.00),pop:NXM_OF_ETH_SRC[],pop:NXM_NX_XXREG0[],controller(userdata=00.00.00.0c.00.00.00.00.00.19.00.10.80.00.08.06.00.00.00.37.69.8b.00.00.00.19.00.18.80.00.34.10.fe.80.00.00.00.00.00.00.02.00.00.ff.fe.37.69.8b.00.19.00.18.80.00.3e.10.fe.80.00.00.00.00.00.00.02.00.00.ff.fe.37.69.8b.00.19.00.10.80.00.42.06.00.00.00.37.69.8b.00.00.ff.ff.00.18.00.00.23.20.00.06.00.20.00.00.00.00.00.01.1c.04.00.01.1e.04.ff.ff.00.18.00.00.23.20.00.07.00.00.00.01.14.04.00.00.00.00.00.00.00.01.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.20.00.00.00)
 cookie=0x790755, duration=58077.302s, table=9, n_packets=0, n_bytes=0, priority=90,icmp6,reg14=0x1,metadata=0x5,ipv6_dst=ff02::1:ff90:9ce1,nw_ttl=255,icmp_type=135,icmp_code=0,nd_target=fe80::200:ff:fe90:9ce1 actions=push:NXM_NX_XXREG0[],push:NXM_OF_ETH_SRC[],push:NXM_NX_ND_SLL[],push:NXM_NX_IPV6_SRC[],pop:NXM_NX_XXREG0[],pop:NXM_OF_ETH_SRC[],controller(userdata=00.00.00.04.00.00.00.00),pop:NXM_OF_ETH_SRC[],pop:NXM_NX_XXREG0[],controller(userdata=00.00.00.0c.00.00.00.00.00.19.00.10.80.00.08.06.00.00.00.90.9c.e1.00.00.00.19.00.18.80.00.34.10.fe.80.00.00.00.00.00.00.02.00.00.ff.fe.90.9c.e1.00.19.00.18.80.00.3e.10.fe.80.00.00.00.00.00.00.02.00.00.ff.fe.90.9c.e1.00.19.00.10.80.00.42.06.00.00.00.90.9c.e1.00.00.ff.ff.00.18.00.00.23.20.00.06.00.20.00.00.00.00.00.01.1c.04.00.01.1e.04.ff.ff.00.18.00.00.23.20.00.07.00.00.00.01.14.04.00.00.00.00.00.00.00.01.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.20.00.00.00)
 cookie=0x790755, duration=58077.302s, table=9, n_packets=0, n_bytes=0, priority=90,icmp6,reg14=0x1,metadata=0x5,ipv6_dst=fe80::200:ff:fe90:9ce1,nw_ttl=255,icmp_type=135,icmp_code=0,nd_target=fe80::200:ff:fe90:9ce1 actions=push:NXM_NX_XXREG0[],push:NXM_OF_ETH_SRC[],push:NXM_NX_ND_SLL[],push:NXM_NX_IPV6_SRC[],pop:NXM_NX_XXREG0[],pop:NXM_OF_ETH_SRC[],controller(userdata=00.00.00.04.00.00.00.00),pop:NXM_OF_ETH_SRC[],pop:NXM_NX_XXREG0[],controller(userdata=00.00.00.0c.00.00.00.00.00.19.00.10.80.00.08.06.00.00.00.90.9c.e1.00.00.00.19.00.18.80.00.34.10.fe.80.00.00.00.00.00.00.02.00.00.ff.fe.90.9c.e1.00.19.00.18.80.00.3e.10.fe.80.00.00.00.00.00.00.02.00.00.ff.fe.90.9c.e1.00.19.00.10.80.00.42.06.00.00.00.90.9c.e1.00.00.ff.ff.00.18.00.00.23.20.00.06.00.20.00.00.00.00.00.01.1c.04.00.01.1e.04.ff.ff.00.18.00.00.23.20.00.07.00.00.00.01.14.04.00.00.00.00.00.00.00.01.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.20.00.00.00)
 cookie=0x97137868, duration=58077.301s, table=9, n_packets=0, n_bytes=0, priority=90,icmp6,reg14=0x2,metadata=0x5,ipv6_dst=ff02::1:ffc3:9878,nw_ttl=255,icmp_type=135,icmp_code=0,nd_target=fe80::526b:4bff:fec3:9878 actions=push:NXM_NX_XXREG0[],push:NXM_OF_ETH_SRC[],push:NXM_NX_ND_SLL[],push:NXM_NX_IPV6_SRC[],pop:NXM_NX_XXREG0[],pop:NXM_OF_ETH_SRC[],controller(userdata=00.00.00.04.00.00.00.00),pop:NXM_OF_ETH_SRC[],pop:NXM_NX_XXREG0[],controller(userdata=00.00.00.0c.00.00.00.00.00.19.00.10.80.00.08.06.50.6b.4b.c3.98.78.00.00.00.19.00.18.80.00.34.10.fe.80.00.00.00.00.00.00.52.6b.4b.ff.fe.c3.98.78.00.19.00.18.80.00.3e.10.fe.80.00.00.00.00.00.00.52.6b.4b.ff.fe.c3.98.78.00.19.00.10.80.00.42.06.50.6b.4b.c3.98.78.00.00.ff.ff.00.18.00.00.23.20.00.06.00.20.00.00.00.00.00.01.1c.04.00.01.1e.04.ff.ff.00.18.00.00.23.20.00.07.00.00.00.01.14.04.00.00.00.00.00.00.00.01.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.20.00.00.00)
 cookie=0x97137868, duration=58077.301s, table=9, n_packets=0, n_bytes=0, priority=90,icmp6,reg14=0x2,metadata=0x5,ipv6_dst=fe80::526b:4bff:fec3:9878,nw_ttl=255,icmp_type=135,icmp_code=0,nd_target=fe80::526b:4bff:fec3:9878 actions=push:NXM_NX_XXREG0[],push:NXM_OF_ETH_SRC[],push:NXM_NX_ND_SLL[],push:NXM_NX_IPV6_SRC[],pop:NXM_NX_XXREG0[],pop:NXM_OF_ETH_SRC[],controller(userdata=00.00.00.04.00.00.00.00),pop:NXM_OF_ETH_SRC[],pop:NXM_NX_XXREG0[],controller(userdata=00.00.00.0c.00.00.00.00.00.19.00.10.80.00.08.06.50.6b.4b.c3.98.78.00.00.00.19.00.18.80.00.34.10.fe.80.00.00.00.00.00.00.52.6b.4b.ff.fe.c3.98.78.00.19.00.18.80.00.3e.10.fe.80.00.00.00.00.00.00.52.6b.4b.ff.fe.c3.98.78.00.19.00.10.80.00.42.06.50.6b.4b.c3.98.78.00.00.ff.ff.00.18.00.00.23.20.00.06.00.20.00.00.00.00.00.01.1c.04.00.01.1e.04.ff.ff.00.18.00.00.23.20.00.07.00.00.00.01.14.04.00.00.00.00.00.00.00.01.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.20.00.00.00)
 cookie=0xd17084b7, duration=58078.057s, table=9, n_packets=0, n_bytes=0, priority=90,icmp6,metadata=0x1,ipv6_dst=fe80::200:ff:fe77:c46,icmp_type=128,icmp_code=0 actions=push:NXM_NX_IPV6_SRC[],push:NXM_NX_IPV6_DST[],pop:NXM_NX_IPV6_SRC[],pop:NXM_NX_IPV6_DST[],load:0xff->NXM_NX_IP_TTL[],load:0x81->NXM_NX_ICMPV6_TYPE[],load:0x1->NXM_NX_REG10[0],resubmit(,10)
 cookie=0xa3502ca2, duration=58078.056s, table=9, n_packets=0, n_bytes=0, priority=90,icmp6,metadata=0x1,ipv6_dst=fe80::200:ff:feee:c6f9,icmp_type=128,icmp_code=0 actions=push:NXM_NX_IPV6_SRC[],push:NXM_NX_IPV6_DST[],pop:NXM_NX_IPV6_SRC[],pop:NXM_NX_IPV6_DST[],load:0xff->NXM_NX_IP_TTL[],load:0x81->NXM_NX_ICMPV6_TYPE[],load:0x1->NXM_NX_REG10[0],resubmit(,10)
 cookie=0xae74b7be, duration=58077.814s, table=9, n_packets=0, n_bytes=0, priority=90,icmp6,metadata=0x1,ipv6_dst=fe80::200:ff:fe37:698b,icmp_type=128,icmp_code=0 actions=push:NXM_NX_IPV6_SRC[],push:NXM_NX_IPV6_DST[],pop:NXM_NX_IPV6_SRC[],pop:NXM_NX_IPV6_DST[],load:0xff->NXM_NX_IP_TTL[],load:0x81->NXM_NX_ICMPV6_TYPE[],load:0x1->NXM_NX_REG10[0],resubmit(,10)
 cookie=0x4eba7432, duration=58077.302s, table=9, n_packets=0, n_bytes=0, priority=90,icmp6,metadata=0x5,ipv6_dst=fe80::200:ff:fe90:9ce1,icmp_type=128,icmp_code=0 actions=push:NXM_NX_IPV6_SRC[],push:NXM_NX_IPV6_DST[],pop:NXM_NX_IPV6_SRC[],pop:NXM_NX_IPV6_DST[],load:0xff->NXM_NX_IP_TTL[],load:0x81->NXM_NX_ICMPV6_TYPE[],load:0x1->NXM_NX_REG10[0],resubmit(,10)
 cookie=0x83caa85b, duration=58077.301s, table=9, n_packets=0, n_bytes=0, priority=90,icmp6,metadata=0x5,ipv6_dst=fe80::526b:4bff:fec3:9878,icmp_type=128,icmp_code=0 actions=push:NXM_NX_IPV6_SRC[],push:NXM_NX_IPV6_DST[],pop:NXM_NX_IPV6_SRC[],pop:NXM_NX_IPV6_DST[],load:0xff->NXM_NX_IP_TTL[],load:0x81->NXM_NX_ICMPV6_TYPE[],load:0x1->NXM_NX_REG10[0],resubmit(,10)
 cookie=0x8d841f4c, duration=58077.815s, table=9, n_packets=0, n_bytes=0, priority=90,arp,metadata=0x1,arp_op=2 actions=push:NXM_NX_REG0[],push:NXM_OF_ETH_SRC[],push:NXM_NX_ARP_SHA[],push:NXM_OF_ARP_SPA[],pop:NXM_NX_REG0[],pop:NXM_OF_ETH_SRC[],controller(userdata=00.00.00.01.00.00.00.00),pop:NXM_OF_ETH_SRC[],pop:NXM_NX_REG0[]
 cookie=0x2b308d28, duration=58077.301s, table=9, n_packets=3878, n_bytes=232680, priority=90,arp,metadata=0x5,arp_op=2 actions=push:NXM_NX_REG0[],push:NXM_OF_ETH_SRC[],push:NXM_NX_ARP_SHA[],push:NXM_OF_ARP_SPA[],pop:NXM_NX_REG0[],pop:NXM_OF_ETH_SRC[],controller(userdata=00.00.00.01.00.00.00.00),pop:NXM_OF_ETH_SRC[],pop:NXM_NX_REG0[]
 cookie=0x3fcf92cf, duration=58077.302s, table=9, n_packets=0, n_bytes=0, priority=90,arp,reg14=0x2,metadata=0x5,arp_spa=10.0.0.0/20,arp_tpa=10.0.2.19,arp_op=1 actions=push:NXM_NX_REG0[],push:NXM_OF_ETH_SRC[],push:NXM_NX_ARP_SHA[],push:NXM_OF_ARP_SPA[],pop:NXM_NX_REG0[],pop:NXM_OF_ETH_SRC[],controller(userdata=00.00.00.01.00.00.00.00),pop:NXM_OF_ETH_SRC[],pop:NXM_NX_REG0[],move:NXM_OF_ETH_SRC[]->NXM_OF_ETH_DST[],mod_dl_src:50:6b:4b:c3:98:78,load:0x2->NXM_OF_ARP_OP[],move:NXM_NX_ARP_SHA[]->NXM_NX_ARP_THA[],load:0x506b4bc39878->NXM_NX_ARP_SHA[],move:NXM_OF_ARP_SPA[]->NXM_OF_ARP_TPA[],load:0xa000213->NXM_OF_ARP_SPA[],load:0x2->NXM_NX_REG15[],load:0x1->NXM_NX_REG10[0],resubmit(,32)
 cookie=0x46286001, duration=58078.069s, table=9, n_packets=0, n_bytes=0, priority=80,tcp6,metadata=0x1,ipv6_dst=fe80::200:ff:fe77:c46,nw_frag=not_later actions=controller(userdata=00.00.00.0b.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.01.26.10.00.80.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.01.28.10.00.80.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.01.26.10.00.80.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.01.28.10.00.80.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0xd8d459ae, duration=58078.069s, table=9, n_packets=0, n_bytes=0, priority=80,udp6,metadata=0x1,ipv6_dst=fe80::200:ff:fe77:c46,nw_frag=not_later actions=controller(userdata=00.00.00.0a.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.01.26.10.00.80.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.01.28.10.00.80.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.01.26.10.00.80.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.01.28.10.00.80.00.00.00.00.00.00.00.19.00.10.00.01.3a.01.ff.00.00.00.00.00.00.00.00.19.00.10.80.00.3a.01.01.00.00.00.00.00.00.00.00.19.00.10.80.00.3c.01.04.00.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0xa2e9f4f6, duration=58078.067s, table=9, n_packets=0, n_bytes=0, priority=80,tcp6,metadata=0x1,ipv6_dst=fe80::200:ff:feee:c6f9,nw_frag=not_later actions=controller(userdata=00.00.00.0b.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.01.26.10.00.80.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.01.28.10.00.80.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.01.26.10.00.80.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.01.28.10.00.80.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0xe19c6e49, duration=58077.815s, table=9, n_packets=0, n_bytes=0, priority=80,udp6,metadata=0x1,ipv6_dst=fe80::200:ff:feee:c6f9,nw_frag=not_later actions=controller(userdata=00.00.00.0a.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.01.26.10.00.80.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.01.28.10.00.80.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.01.26.10.00.80.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.01.28.10.00.80.00.00.00.00.00.00.00.19.00.10.00.01.3a.01.ff.00.00.00.00.00.00.00.00.19.00.10.80.00.3a.01.01.00.00.00.00.00.00.00.00.19.00.10.80.00.3c.01.04.00.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0xf10354fc, duration=58077.815s, table=9, n_packets=0, n_bytes=0, priority=80,udp6,metadata=0x1,ipv6_dst=fe80::200:ff:fe37:698b,nw_frag=not_later actions=controller(userdata=00.00.00.0a.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.01.26.10.00.80.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.01.28.10.00.80.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.01.26.10.00.80.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.01.28.10.00.80.00.00.00.00.00.00.00.19.00.10.00.01.3a.01.ff.00.00.00.00.00.00.00.00.19.00.10.80.00.3a.01.01.00.00.00.00.00.00.00.00.19.00.10.80.00.3c.01.04.00.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0x8da95250, duration=58077.814s, table=9, n_packets=0, n_bytes=0, priority=80,tcp6,metadata=0x1,ipv6_dst=fe80::200:ff:fe37:698b,nw_frag=not_later actions=controller(userdata=00.00.00.0b.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.01.26.10.00.80.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.01.28.10.00.80.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.01.26.10.00.80.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.01.28.10.00.80.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0x292d0748, duration=58078.067s, table=9, n_packets=0, n_bytes=0, priority=80,udp,metadata=0x1,nw_dst=100.64.1.1,nw_frag=not_later actions=controller(userdata=00.00.00.0a.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.0e.04.00.20.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.10.04.00.20.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.0e.04.00.20.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.10.04.00.20.00.00.00.00.00.00.00.19.00.10.00.01.3a.01.ff.00.00.00.00.00.00.00.00.19.00.10.80.00.26.01.03.00.00.00.00.00.00.00.00.19.00.10.80.00.28.01.03.00.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0xcceae0c5, duration=58078.056s, table=9, n_packets=0, n_bytes=0, priority=80,udp,metadata=0x1,nw_dst=10.244.1.1,nw_frag=not_later actions=controller(userdata=00.00.00.0a.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.0e.04.00.20.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.10.04.00.20.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.0e.04.00.20.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.10.04.00.20.00.00.00.00.00.00.00.19.00.10.00.01.3a.01.ff.00.00.00.00.00.00.00.00.19.00.10.80.00.26.01.03.00.00.00.00.00.00.00.00.19.00.10.80.00.28.01.03.00.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0xef88c2da, duration=58078.056s, table=9, n_packets=0, n_bytes=0, priority=80,tcp,metadata=0x1,nw_dst=10.244.1.1,nw_frag=not_later actions=controller(userdata=00.00.00.0b.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.0e.04.00.20.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.10.04.00.20.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.0e.04.00.20.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.10.04.00.20.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0x1b5ebceb, duration=58077.815s, table=9, n_packets=0, n_bytes=0, priority=80,tcp,metadata=0x1,nw_dst=100.64.1.1,nw_frag=not_later actions=controller(userdata=00.00.00.0b.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.0e.04.00.20.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.10.04.00.20.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.0e.04.00.20.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.10.04.00.20.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0x39377fce, duration=58077.814s, table=9, n_packets=0, n_bytes=0, priority=80,udp,metadata=0x1,nw_dst=10.244.2.1,nw_frag=not_later actions=controller(userdata=00.00.00.0a.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.0e.04.00.20.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.10.04.00.20.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.0e.04.00.20.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.10.04.00.20.00.00.00.00.00.00.00.19.00.10.00.01.3a.01.ff.00.00.00.00.00.00.00.00.19.00.10.80.00.26.01.03.00.00.00.00.00.00.00.00.19.00.10.80.00.28.01.03.00.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0xd3be3b39, duration=58077.814s, table=9, n_packets=0, n_bytes=0, priority=80,tcp,metadata=0x1,nw_dst=10.244.2.1,nw_frag=not_later actions=controller(userdata=00.00.00.0b.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.0e.04.00.20.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.10.04.00.20.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.0e.04.00.20.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.10.04.00.20.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0x231402bc, duration=58077.847s, table=9, n_packets=0, n_bytes=0, priority=80,arp,reg14=0x2,metadata=0x1,arp_spa=10.244.1.0/24,arp_op=1 actions=push:NXM_NX_REG0[],push:NXM_OF_ETH_SRC[],push:NXM_NX_ARP_SHA[],push:NXM_OF_ARP_SPA[],pop:NXM_NX_REG0[],pop:NXM_OF_ETH_SRC[],controller(userdata=00.00.00.01.00.00.00.00),pop:NXM_OF_ETH_SRC[],pop:NXM_NX_REG0[]
 cookie=0x54bac12a, duration=58077.846s, table=9, n_packets=0, n_bytes=0, priority=80,arp,reg14=0x1,metadata=0x1,arp_spa=100.64.1.0/24,arp_op=1 actions=push:NXM_NX_REG0[],push:NXM_OF_ETH_SRC[],push:NXM_NX_ARP_SHA[],push:NXM_OF_ARP_SPA[],pop:NXM_NX_REG0[],pop:NXM_OF_ETH_SRC[],controller(userdata=00.00.00.01.00.00.00.00),pop:NXM_OF_ETH_SRC[],pop:NXM_NX_REG0[]
 cookie=0xd978561b, duration=58077.814s, table=9, n_packets=0, n_bytes=0, priority=80,arp,reg14=0x3,metadata=0x1,arp_spa=10.244.2.0/24,arp_op=1 actions=push:NXM_NX_REG0[],push:NXM_OF_ETH_SRC[],push:NXM_NX_ARP_SHA[],push:NXM_OF_ARP_SPA[],pop:NXM_NX_REG0[],pop:NXM_OF_ETH_SRC[],controller(userdata=00.00.00.01.00.00.00.00),pop:NXM_OF_ETH_SRC[],pop:NXM_NX_REG0[]
 cookie=0xdd5ede02, duration=58077.302s, table=9, n_packets=0, n_bytes=0, priority=80,arp,reg14=0x1,metadata=0x5,arp_spa=100.64.1.0/24,arp_op=1 actions=push:NXM_NX_REG0[],push:NXM_OF_ETH_SRC[],push:NXM_NX_ARP_SHA[],push:NXM_OF_ARP_SPA[],pop:NXM_NX_REG0[],pop:NXM_OF_ETH_SRC[],controller(userdata=00.00.00.01.00.00.00.00),pop:NXM_OF_ETH_SRC[],pop:NXM_NX_REG0[]
 cookie=0xeb423826, duration=58077.301s, table=9, n_packets=486357, n_bytes=29181330, priority=80,arp,reg14=0x2,metadata=0x5,arp_spa=10.0.0.0/20,arp_op=1 actions=push:NXM_NX_REG0[],push:NXM_OF_ETH_SRC[],push:NXM_NX_ARP_SHA[],push:NXM_OF_ARP_SPA[],pop:NXM_NX_REG0[],pop:NXM_OF_ETH_SRC[],controller(userdata=00.00.00.01.00.00.00.00),pop:NXM_OF_ETH_SRC[],pop:NXM_NX_REG0[]
 cookie=0x9d8ab43e, duration=58078.069s, table=9, n_packets=0, n_bytes=0, priority=70,ip,metadata=0x1,nw_dst=10.244.1.1,nw_frag=not_later actions=controller(userdata=00.00.00.0a.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.0e.04.00.20.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.10.04.00.20.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.0e.04.00.20.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.10.04.00.20.00.00.00.00.00.00.00.19.00.10.00.01.3a.01.ff.00.00.00.00.00.00.00.00.19.00.10.80.00.26.01.03.00.00.00.00.00.00.00.00.19.00.10.80.00.28.01.02.00.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0xabd97dcc, duration=58078.067s, table=9, n_packets=0, n_bytes=0, priority=70,ip,metadata=0x1,nw_dst=100.64.1.1,nw_frag=not_later actions=controller(userdata=00.00.00.0a.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.0e.04.00.20.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.10.04.00.20.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.0e.04.00.20.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.10.04.00.20.00.00.00.00.00.00.00.19.00.10.00.01.3a.01.ff.00.00.00.00.00.00.00.00.19.00.10.80.00.26.01.03.00.00.00.00.00.00.00.00.19.00.10.80.00.28.01.02.00.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0xd56b153, duration=58077.814s, table=9, n_packets=0, n_bytes=0, priority=70,ip,metadata=0x1,nw_dst=10.244.2.1,nw_frag=not_later actions=controller(userdata=00.00.00.0a.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.0e.04.00.20.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.10.04.00.20.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.0e.04.00.20.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.10.04.00.20.00.00.00.00.00.00.00.19.00.10.00.01.3a.01.ff.00.00.00.00.00.00.00.00.19.00.10.80.00.26.01.03.00.00.00.00.00.00.00.00.19.00.10.80.00.28.01.02.00.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0x5c88bf7e, duration=58078.067s, table=9, n_packets=0, n_bytes=0, priority=70,ipv6,metadata=0x1,ipv6_dst=fe80::200:ff:feee:c6f9,nw_frag=not_later actions=controller(userdata=00.00.00.0a.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.01.26.10.00.80.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.01.28.10.00.80.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.01.26.10.00.80.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.01.28.10.00.80.00.00.00.00.00.00.00.19.00.10.00.01.3a.01.ff.00.00.00.00.00.00.00.00.19.00.10.80.00.3a.01.01.00.00.00.00.00.00.00.00.19.00.10.80.00.3c.01.03.00.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0x9786673d, duration=58078.057s, table=9, n_packets=0, n_bytes=0, priority=70,ipv6,metadata=0x1,ipv6_dst=fe80::200:ff:fe77:c46,nw_frag=not_later actions=controller(userdata=00.00.00.0a.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.01.26.10.00.80.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.01.28.10.00.80.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.01.26.10.00.80.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.01.28.10.00.80.00.00.00.00.00.00.00.19.00.10.00.01.3a.01.ff.00.00.00.00.00.00.00.00.19.00.10.80.00.3a.01.01.00.00.00.00.00.00.00.00.19.00.10.80.00.3c.01.03.00.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0x1f6d213, duration=58077.814s, table=9, n_packets=0, n_bytes=0, priority=70,ipv6,metadata=0x1,ipv6_dst=fe80::200:ff:fe37:698b,nw_frag=not_later actions=controller(userdata=00.00.00.0a.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.01.26.10.00.80.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.01.28.10.00.80.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.01.26.10.00.80.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.01.28.10.00.80.00.00.00.00.00.00.00.19.00.10.00.01.3a.01.ff.00.00.00.00.00.00.00.00.19.00.10.80.00.3a.01.01.00.00.00.00.00.00.00.00.19.00.10.80.00.3c.01.03.00.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0xe89a6b0b, duration=58078.057s, table=9, n_packets=0, n_bytes=0, priority=60,ip,metadata=0x1,nw_dst=100.64.1.1 actions=drop
 cookie=0xb4f9c87c, duration=58077.815s, table=9, n_packets=0, n_bytes=0, priority=60,ip,metadata=0x1,nw_dst=10.244.1.1 actions=drop
 cookie=0xb2292734, duration=58077.814s, table=9, n_packets=0, n_bytes=0, priority=60,ip,metadata=0x1,nw_dst=10.244.2.1 actions=drop
 cookie=0x4a222093, duration=58077.847s, table=9, n_packets=0, n_bytes=0, priority=60,ipv6,metadata=0x1,ipv6_dst=fe80::200:ff:feee:c6f9 actions=drop
 cookie=0x19875956, duration=58077.815s, table=9, n_packets=0, n_bytes=0, priority=60,ipv6,metadata=0x1,ipv6_dst=fe80::200:ff:fe77:c46 actions=drop
 cookie=0x239dd4e4, duration=58077.815s, table=9, n_packets=0, n_bytes=0, priority=60,ipv6,metadata=0x1,ipv6_dst=fe80::200:ff:fe37:698b actions=drop
 cookie=0xf568815b, duration=58077.302s, table=9, n_packets=0, n_bytes=0, priority=60,ipv6,metadata=0x5,ipv6_dst=fe80::200:ff:fe90:9ce1 actions=drop
 cookie=0x44607a64, duration=58077.302s, table=9, n_packets=0, n_bytes=0, priority=60,ipv6,metadata=0x5,ipv6_dst=fe80::526b:4bff:fec3:9878 actions=drop
 cookie=0x933ce955, duration=58077.847s, table=9, n_packets=0, n_bytes=0, priority=50,metadata=0x1,dl_dst=ff:ff:ff:ff:ff:ff actions=drop
 cookie=0x60be133a, duration=58077.302s, table=9, n_packets=0, n_bytes=0, priority=50,metadata=0x5,dl_dst=ff:ff:ff:ff:ff:ff actions=drop
 cookie=0x39de4d68, duration=58078.057s, table=9, n_packets=0, n_bytes=0, priority=40,ip,reg14=0x1,metadata=0x1,nw_ttl=0,nw_frag=not_later actions=controller(userdata=00.00.00.0a.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.00.19.00.10.80.00.26.01.0b.00.00.00.00.00.00.00.00.19.00.10.80.00.28.01.00.00.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.06.00.20.00.00.00.00.00.00.0e.04.00.00.10.04.00.19.00.10.80.00.16.04.64.40.01.01.00.00.00.00.00.19.00.10.00.01.3a.01.ff.00.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0x3c308ead, duration=58078.057s, table=9, n_packets=0, n_bytes=0, priority=40,ip,reg14=0x2,metadata=0x1,nw_ttl=0,nw_frag=not_later actions=controller(userdata=00.00.00.0a.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.00.19.00.10.80.00.26.01.0b.00.00.00.00.00.00.00.00.19.00.10.80.00.28.01.00.00.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.06.00.20.00.00.00.00.00.00.0e.04.00.00.10.04.00.19.00.10.80.00.16.04.0a.f4.01.01.00.00.00.00.00.19.00.10.00.01.3a.01.ff.00.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0x39de4d68, duration=58077.846s, table=9, n_packets=0, n_bytes=0, priority=40,ip,reg14=0x1,metadata=0x1,nw_ttl=1,nw_frag=not_later actions=controller(userdata=00.00.00.0a.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.00.19.00.10.80.00.26.01.0b.00.00.00.00.00.00.00.00.19.00.10.80.00.28.01.00.00.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.06.00.20.00.00.00.00.00.00.0e.04.00.00.10.04.00.19.00.10.80.00.16.04.64.40.01.01.00.00.00.00.00.19.00.10.00.01.3a.01.ff.00.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0x3c308ead, duration=58077.815s, table=9, n_packets=0, n_bytes=0, priority=40,ip,reg14=0x2,metadata=0x1,nw_ttl=1,nw_frag=not_later actions=controller(userdata=00.00.00.0a.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.00.19.00.10.80.00.26.01.0b.00.00.00.00.00.00.00.00.19.00.10.80.00.28.01.00.00.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.06.00.20.00.00.00.00.00.00.0e.04.00.00.10.04.00.19.00.10.80.00.16.04.0a.f4.01.01.00.00.00.00.00.19.00.10.00.01.3a.01.ff.00.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0xff5e09bc, duration=58077.814s, table=9, n_packets=0, n_bytes=0, priority=40,ip,reg14=0x3,metadata=0x1,nw_ttl=1,nw_frag=not_later actions=controller(userdata=00.00.00.0a.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.00.19.00.10.80.00.26.01.0b.00.00.00.00.00.00.00.00.19.00.10.80.00.28.01.00.00.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.06.00.20.00.00.00.00.00.00.0e.04.00.00.10.04.00.19.00.10.80.00.16.04.0a.f4.02.01.00.00.00.00.00.19.00.10.00.01.3a.01.ff.00.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0xff5e09bc, duration=58077.814s, table=9, n_packets=0, n_bytes=0, priority=40,ip,reg14=0x3,metadata=0x1,nw_ttl=0,nw_frag=not_later actions=controller(userdata=00.00.00.0a.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.00.19.00.10.80.00.26.01.0b.00.00.00.00.00.00.00.00.19.00.10.80.00.28.01.00.00.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.06.00.20.00.00.00.00.00.00.0e.04.00.00.10.04.00.19.00.10.80.00.16.04.0a.f4.02.01.00.00.00.00.00.19.00.10.00.01.3a.01.ff.00.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0x1007dd53, duration=58077.302s, table=9, n_packets=0, n_bytes=0, priority=40,ip,reg14=0x2,metadata=0x5,nw_ttl=1,nw_frag=not_later actions=controller(userdata=00.00.00.0a.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.00.19.00.10.80.00.26.01.0b.00.00.00.00.00.00.00.00.19.00.10.80.00.28.01.00.00.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.06.00.20.00.00.00.00.00.00.0e.04.00.00.10.04.00.19.00.10.80.00.16.04.0a.00.02.13.00.00.00.00.00.19.00.10.00.01.3a.01.ff.00.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0xa0405310, duration=58077.302s, table=9, n_packets=0, n_bytes=0, priority=40,ip,reg14=0x1,metadata=0x5,nw_ttl=1,nw_frag=not_later actions=controller(userdata=00.00.00.0a.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.00.19.00.10.80.00.26.01.0b.00.00.00.00.00.00.00.00.19.00.10.80.00.28.01.00.00.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.06.00.20.00.00.00.00.00.00.0e.04.00.00.10.04.00.19.00.10.80.00.16.04.64.40.01.02.00.00.00.00.00.19.00.10.00.01.3a.01.ff.00.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0x1007dd53, duration=58077.302s, table=9, n_packets=0, n_bytes=0, priority=40,ip,reg14=0x2,metadata=0x5,nw_ttl=0,nw_frag=not_later actions=controller(userdata=00.00.00.0a.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.00.19.00.10.80.00.26.01.0b.00.00.00.00.00.00.00.00.19.00.10.80.00.28.01.00.00.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.06.00.20.00.00.00.00.00.00.0e.04.00.00.10.04.00.19.00.10.80.00.16.04.0a.00.02.13.00.00.00.00.00.19.00.10.00.01.3a.01.ff.00.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0xa0405310, duration=58077.302s, table=9, n_packets=0, n_bytes=0, priority=40,ip,reg14=0x1,metadata=0x5,nw_ttl=0,nw_frag=not_later actions=controller(userdata=00.00.00.0a.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1b.00.00.00.00.02.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.04.06.00.30.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.1c.00.00.00.00.02.06.00.30.00.00.00.00.00.00.00.19.00.10.80.00.26.01.0b.00.00.00.00.00.00.00.00.19.00.10.80.00.28.01.00.00.00.00.00.00.00.00.ff.ff.00.18.00.00.23.20.00.06.00.20.00.00.00.00.00.00.0e.04.00.00.10.04.00.19.00.10.80.00.16.04.64.40.01.02.00.00.00.00.00.19.00.10.00.01.3a.01.ff.00.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.0a.00.00.00)
 cookie=0x290e1c4, duration=58078.056s, table=9, n_packets=0, n_bytes=0, priority=30,ip,metadata=0x1,nw_ttl=0 actions=drop
 cookie=0x290e1c4, duration=58077.846s, table=9, n_packets=0, n_bytes=0, priority=30,ip,metadata=0x1,nw_ttl=1 actions=drop
 cookie=0x7c6169ac, duration=58077.301s, table=9, n_packets=0, n_bytes=0, priority=30,ip,metadata=0x5,nw_ttl=1 actions=drop
 cookie=0x7c6169ac, duration=58077.301s, table=9, n_packets=0, n_bytes=0, priority=30,ip,metadata=0x5,nw_ttl=0 actions=drop
 cookie=0xd349bc2c, duration=58078.067s, table=9, n_packets=1673224, n_bytes=61558280345, priority=0,metadata=0x1 actions=resubmit(,10)
 cookie=0x1b4e4eb9, duration=58078.067s, table=9, n_packets=1789319, n_bytes=61569108361, priority=0,metadata=0x3 actions=resubmit(,10)
 cookie=0xb4c7bd32, duration=58078.056s, table=9, n_packets=203090, n_bytes=70216033, priority=0,metadata=0x2 actions=resubmit(,10)
 cookie=0x6cd0a7dc, duration=58077.787s, table=9, n_packets=1470133, n_bytes=61488064222, priority=0,metadata=0x4 actions=resubmit(,10)
 cookie=0xb3f57b3, duration=58077.303s, table=9, n_packets=710300, n_bytes=101131271, priority=0,metadata=0x6 actions=resubmit(,10)
 cookie=0x44037734, duration=58077.302s, table=9, n_packets=204957, n_bytes=70436259, priority=0,metadata=0x5 actions=resubmit(,10)
 cookie=0x6d4a3402, duration=58078.067s, table=10, n_packets=203090, n_bytes=70216033, priority=0,metadata=0x2 actions=resubmit(,11)
 cookie=0xcb16a27a, duration=58078.056s, table=10, n_packets=1789319, n_bytes=61569108361, priority=0,metadata=0x3 actions=resubmit(,11)
 cookie=0xc56d48f8, duration=58077.847s, table=10, n_packets=1673224, n_bytes=61558280345, priority=0,metadata=0x1 actions=resubmit(,11)
 cookie=0xd151d357, duration=58077.787s, table=10, n_packets=1470133, n_bytes=61488064222, priority=0,metadata=0x4 actions=resubmit(,11)
 cookie=0xc81f24dd, duration=58077.302s, table=10, n_packets=204957, n_bytes=70436259, priority=0,metadata=0x5 actions=resubmit(,11)
 cookie=0x9954b606, duration=58077.302s, table=10, n_packets=710300, n_bytes=101131271, priority=0,metadata=0x6 actions=resubmit(,11)
 cookie=0x8b07dc7b, duration=58076.214s, table=11, n_packets=0, n_bytes=0, priority=110,tcp6,metadata=0x3,tcp_flags=rst actions=resubmit(,12)
 cookie=0x8b07dc7b, duration=58076.214s, table=11, n_packets=6, n_bytes=360, priority=110,tcp,metadata=0x3,tcp_flags=rst actions=resubmit(,12)
 cookie=0x15d9f34f, duration=58059.635s, table=11, n_packets=0, n_bytes=0, priority=110,tcp6,metadata=0x4,tcp_flags=rst actions=resubmit(,12)
 cookie=0x15d9f34f, duration=58059.635s, table=11, n_packets=0, n_bytes=0, priority=110,tcp,metadata=0x4,tcp_flags=rst actions=resubmit(,12)
 cookie=0x8b07dc7b, duration=58076.214s, table=11, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x3,nw_ttl=255,icmp_type=133,icmp_code=0 actions=resubmit(,12)
 cookie=0x8b07dc7b, duration=58076.214s, table=11, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x3,nw_ttl=255,icmp_type=134,icmp_code=0 actions=resubmit(,12)
 cookie=0x8b07dc7b, duration=58076.214s, table=11, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x3,nw_ttl=255,icmp_type=136,icmp_code=0 actions=resubmit(,12)
 cookie=0x8b07dc7b, duration=58076.214s, table=11, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x3,nw_ttl=255,icmp_type=135,icmp_code=0 actions=resubmit(,12)
 cookie=0x15d9f34f, duration=58059.635s, table=11, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x4,nw_ttl=255,icmp_type=134,icmp_code=0 actions=resubmit(,12)
 cookie=0x15d9f34f, duration=58059.635s, table=11, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x4,nw_ttl=255,icmp_type=135,icmp_code=0 actions=resubmit(,12)
 cookie=0x15d9f34f, duration=58059.635s, table=11, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x4,nw_ttl=255,icmp_type=133,icmp_code=0 actions=resubmit(,12)
 cookie=0x15d9f34f, duration=58059.635s, table=11, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x4,nw_ttl=255,icmp_type=136,icmp_code=0 actions=resubmit(,12)
 cookie=0x8b07dc7b, duration=58076.214s, table=11, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x3,icmp_type=1 actions=resubmit(,12)
 cookie=0x8b07dc7b, duration=58076.214s, table=11, n_packets=0, n_bytes=0, priority=110,icmp,metadata=0x3,icmp_type=3 actions=resubmit(,12)
 cookie=0x15d9f34f, duration=58059.635s, table=11, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x4,icmp_type=1 actions=resubmit(,12)
 cookie=0x15d9f34f, duration=58059.635s, table=11, n_packets=0, n_bytes=0, priority=110,icmp,metadata=0x4,icmp_type=3 actions=resubmit(,12)
 cookie=0xe208bb79, duration=58076.214s, table=11, n_packets=114816, n_bytes=63806298, priority=110,ip,reg14=0x1,metadata=0x3 actions=resubmit(,12)
 cookie=0xe208bb79, duration=58076.214s, table=11, n_packets=0, n_bytes=0, priority=110,ipv6,reg14=0x1,metadata=0x3 actions=resubmit(,12)
 cookie=0x6814d3db, duration=58059.635s, table=11, n_packets=1470133, n_bytes=61488064222, priority=110,ip,reg14=0x1,metadata=0x4 actions=resubmit(,12)
 cookie=0x6814d3db, duration=58059.635s, table=11, n_packets=0, n_bytes=0, priority=110,ipv6,reg14=0x1,metadata=0x4 actions=resubmit(,12)
 cookie=0x3ceb24b1, duration=58077.302s, table=11, n_packets=114816, n_bytes=63806298, priority=90,ip,metadata=0x5,nw_dst=10.0.2.19 actions=ct(table=12,zone=NXM_NX_REG12[0..15],nat)
 cookie=0x42992442, duration=58077.302s, table=11, n_packets=0, n_bytes=0, priority=100,ip,metadata=0x5,nw_dst=100.64.1.2 actions=ct(table=12,zone=NXM_NX_REG12[0..15],nat)
 cookie=0x28cf4159, duration=58076.214s, table=11, n_packets=0, n_bytes=0, priority=100,ipv6,metadata=0x3 actions=load:0x1->NXM_NX_XXREG0[96],resubmit(,12)
 cookie=0x28cf4159, duration=58076.214s, table=11, n_packets=1674476, n_bytes=61505300449, priority=100,ip,metadata=0x3 actions=load:0x1->NXM_NX_XXREG0[96],resubmit(,12)
 cookie=0x9f9ada06, duration=58059.635s, table=11, n_packets=0, n_bytes=0, priority=100,ip,metadata=0x4 actions=load:0x1->NXM_NX_XXREG0[96],resubmit(,12)
 cookie=0x9f9ada06, duration=58059.635s, table=11, n_packets=0, n_bytes=0, priority=100,ipv6,metadata=0x4 actions=load:0x1->NXM_NX_XXREG0[96],resubmit(,12)
 cookie=0xfc29f62b, duration=58078.067s, table=11, n_packets=1673224, n_bytes=61558280345, priority=0,metadata=0x1 actions=resubmit(,12)
 cookie=0x780b9f5d, duration=58078.067s, table=11, n_packets=21, n_bytes=1254, priority=0,metadata=0x3 actions=resubmit(,12)
 cookie=0x5609af0e, duration=58077.846s, table=11, n_packets=203090, n_bytes=70216033, priority=0,metadata=0x2 actions=resubmit(,12)
 cookie=0xe8a09767, duration=58077.787s, table=11, n_packets=0, n_bytes=0, priority=0,metadata=0x4 actions=resubmit(,12)
 cookie=0xb9357a4f, duration=58077.302s, table=11, n_packets=710300, n_bytes=101131271, priority=0,metadata=0x6 actions=resubmit(,12)
 cookie=0x2c2f0e47, duration=58077.301s, table=11, n_packets=90141, n_bytes=6629961, priority=0,metadata=0x5 actions=resubmit(,12)
 cookie=0xaac2b6f5, duration=58078.069s, table=12, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x2,nw_ttl=255,icmp_type=134,icmp_code=0 actions=resubmit(,13)
 cookie=0x9dbb70c3, duration=58078.067s, table=12, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x3,nw_ttl=255,icmp_type=135,icmp_code=0 actions=resubmit(,13)
 cookie=0xaac2b6f5, duration=58078.067s, table=12, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x2,nw_ttl=255,icmp_type=135,icmp_code=0 actions=resubmit(,13)
 cookie=0x9dbb70c3, duration=58078.057s, table=12, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x3,nw_ttl=255,icmp_type=136,icmp_code=0 actions=resubmit(,13)
 cookie=0xaac2b6f5, duration=58078.056s, table=12, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x2,nw_ttl=255,icmp_type=136,icmp_code=0 actions=resubmit(,13)
 cookie=0xaac2b6f5, duration=58077.847s, table=12, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x2,nw_ttl=255,icmp_type=133,icmp_code=0 actions=resubmit(,13)
 cookie=0x9dbb70c3, duration=58077.847s, table=12, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x3,nw_ttl=255,icmp_type=133,icmp_code=0 actions=resubmit(,13)
 cookie=0x9dbb70c3, duration=58077.846s, table=12, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x3,nw_ttl=255,icmp_type=134,icmp_code=0 actions=resubmit(,13)
 cookie=0x5dbc66af, duration=58077.787s, table=12, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x4,nw_ttl=255,icmp_type=136,icmp_code=0 actions=resubmit(,13)
 cookie=0x5dbc66af, duration=58077.787s, table=12, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x4,nw_ttl=255,icmp_type=135,icmp_code=0 actions=resubmit(,13)
 cookie=0x5dbc66af, duration=58077.786s, table=12, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x4,nw_ttl=255,icmp_type=133,icmp_code=0 actions=resubmit(,13)
 cookie=0x5dbc66af, duration=58077.786s, table=12, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x4,nw_ttl=255,icmp_type=134,icmp_code=0 actions=resubmit(,13)
 cookie=0x2bf010ec, duration=58077.302s, table=12, n_packets=611, n_bytes=42770, priority=110,icmp6,metadata=0x6,nw_ttl=255,icmp_type=133,icmp_code=0 actions=resubmit(,13)
 cookie=0x2bf010ec, duration=58077.302s, table=12, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x6,nw_ttl=255,icmp_type=135,icmp_code=0 actions=resubmit(,13)
 cookie=0x2bf010ec, duration=58077.301s, table=12, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x6,nw_ttl=255,icmp_type=134,icmp_code=0 actions=resubmit(,13)
 cookie=0x2bf010ec, duration=58077.301s, table=12, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x6,nw_ttl=255,icmp_type=136,icmp_code=0 actions=resubmit(,13)
 cookie=0x604c2a9c, duration=58077.814s, table=12, n_packets=88154, n_bytes=6398035, priority=100,ip,metadata=0x3,nw_dst=10.96.0.1 actions=load:0x1->NXM_NX_XXREG0[96],resubmit(,13)
 cookie=0x51a12d81, duration=58077.303s, table=12, n_packets=0, n_bytes=0, priority=100,ip,metadata=0x4,nw_dst=10.96.0.1 actions=load:0x1->NXM_NX_XXREG0[96],resubmit(,13)
 cookie=0x4119309c, duration=58071.594s, table=12, n_packets=0, n_bytes=0, priority=100,ip,metadata=0x4,nw_dst=10.96.0.10 actions=load:0x1->NXM_NX_XXREG0[96],resubmit(,13)
 cookie=0x3a4a2b81, duration=58071.594s, table=12, n_packets=0, n_bytes=0, priority=100,ip,metadata=0x3,nw_dst=10.96.0.10 actions=load:0x1->NXM_NX_XXREG0[96],resubmit(,13)
 cookie=0x86bcbc19, duration=58054.582s, table=12, n_packets=1470116, n_bytes=61488084668, priority=100,ip,metadata=0x3,nw_dst=10.104.114.18 actions=load:0x1->NXM_NX_XXREG0[96],resubmit(,13)
 cookie=0x835c4de2, duration=58054.582s, table=12, n_packets=0, n_bytes=0, priority=100,ip,metadata=0x4,nw_dst=10.104.114.18 actions=load:0x1->NXM_NX_XXREG0[96],resubmit(,13)
 cookie=0xaf971d46, duration=58077.302s, table=12, n_packets=1867, n_bytes=220226, priority=50,ipv6,metadata=0x5 actions=load:0x1->NXM_NX_REG10[0],ct(table=13,zone=NXM_NX_REG11[0..15],nat)
 cookie=0xaf971d46, duration=58077.302s, table=12, n_packets=203090, n_bytes=70216033, priority=50,ip,metadata=0x5 actions=load:0x1->NXM_NX_REG10[0],ct(table=13,zone=NXM_NX_REG11[0..15],nat)
 cookie=0x8b2dc891, duration=58078.057s, table=12, n_packets=231049, n_bytes=74625658, priority=0,metadata=0x3 actions=resubmit(,13)
 cookie=0x4799f868, duration=58078.056s, table=12, n_packets=203090, n_bytes=70216033, priority=0,metadata=0x2 actions=resubmit(,13)
 cookie=0xc731532f, duration=58077.846s, table=12, n_packets=1673224, n_bytes=61558280345, priority=0,metadata=0x1 actions=resubmit(,13)
 cookie=0x6c12203, duration=58077.787s, table=12, n_packets=1470133, n_bytes=61488064222, priority=0,metadata=0x4 actions=resubmit(,13)
 cookie=0x302b4a5b, duration=58077.302s, table=12, n_packets=709689, n_bytes=101088501, priority=0,metadata=0x6 actions=resubmit(,13)
 cookie=0x3deeddb7, duration=58077.302s, table=12, n_packets=0, n_bytes=0, priority=0,metadata=0x5 actions=resubmit(,13)
 cookie=0x3d09ee98, duration=58078.067s, table=13, n_packets=1674482, n_bytes=61505300809, priority=100,ip,reg0=0x1/0x1,metadata=0x3 actions=ct(table=14,zone=NXM_NX_REG13[0..15])
 cookie=0x7ecf813c, duration=58078.057s, table=13, n_packets=0, n_bytes=0, priority=100,ip,reg0=0x1/0x1,metadata=0x2 actions=ct(table=14,zone=NXM_NX_REG13[0..15])
 cookie=0x3d09ee98, duration=58078.056s, table=13, n_packets=0, n_bytes=0, priority=100,ipv6,reg0=0x1/0x1,metadata=0x3 actions=ct(table=14,zone=NXM_NX_REG13[0..15])
 cookie=0x7ecf813c, duration=58077.847s, table=13, n_packets=0, n_bytes=0, priority=100,ipv6,reg0=0x1/0x1,metadata=0x2 actions=ct(table=14,zone=NXM_NX_REG13[0..15])
 cookie=0xfb50b5f, duration=58077.787s, table=13, n_packets=0, n_bytes=0, priority=100,ip,reg0=0x1/0x1,metadata=0x4 actions=ct(table=14,zone=NXM_NX_REG13[0..15])
 cookie=0xfb50b5f, duration=58077.786s, table=13, n_packets=0, n_bytes=0, priority=100,ipv6,reg0=0x1/0x1,metadata=0x4 actions=ct(table=14,zone=NXM_NX_REG13[0..15])
 cookie=0xba493cef, duration=58077.303s, table=13, n_packets=0, n_bytes=0, priority=100,ipv6,reg0=0x1/0x1,metadata=0x6 actions=ct(table=14,zone=NXM_NX_REG13[0..15])
 cookie=0xba493cef, duration=58077.302s, table=13, n_packets=0, n_bytes=0, priority=100,ip,reg0=0x1/0x1,metadata=0x6 actions=ct(table=14,zone=NXM_NX_REG13[0..15])
 cookie=0x95c287c2, duration=58078.067s, table=13, n_packets=1673224, n_bytes=61558280345, priority=0,metadata=0x1 actions=resubmit(,14)
 cookie=0xe24b31b3, duration=58078.057s, table=13, n_packets=114837, n_bytes=63807552, priority=0,metadata=0x3 actions=resubmit(,14)
 cookie=0x95321f0d, duration=58077.846s, table=13, n_packets=203090, n_bytes=70216033, priority=0,metadata=0x2 actions=resubmit(,14)
 cookie=0xe52902b7, duration=58077.786s, table=13, n_packets=1470133, n_bytes=61488064222, priority=0,metadata=0x4 actions=resubmit(,14)
 cookie=0xf9ccfebd, duration=58077.302s, table=13, n_packets=204957, n_bytes=70436259, priority=0,metadata=0x5 actions=resubmit(,14)
 cookie=0xb4ea63f7, duration=58077.302s, table=13, n_packets=710300, n_bytes=101131271, priority=0,metadata=0x6 actions=resubmit(,14)
 cookie=0xf3221d5c, duration=58076.214s, table=14, n_packets=0, n_bytes=0, priority=65535,ct_state=+est+rpl+trk,ct_label=0x1/0x1,metadata=0x3 actions=drop
 cookie=0xbcfa523d, duration=58059.635s, table=14, n_packets=0, n_bytes=0, priority=65535,ct_state=+est+rpl+trk,ct_label=0x1/0x1,metadata=0x4 actions=drop
 cookie=0xa192c1f9, duration=58076.214s, table=14, n_packets=0, n_bytes=0, priority=65535,ct_state=-new-est+rel-inv+trk,ct_label=0/0x1,metadata=0x3 actions=resubmit(,15)
 cookie=0x7d6e8323, duration=58059.635s, table=14, n_packets=0, n_bytes=0, priority=65535,ct_state=-new-est+rel-inv+trk,ct_label=0/0x1,metadata=0x4 actions=resubmit(,15)
 cookie=0xc46ec6dd, duration=58076.214s, table=14, n_packets=0, n_bytes=0, priority=65535,icmp6,metadata=0x3,nw_ttl=255,icmp_type=136,icmp_code=0 actions=resubmit(,15)
 cookie=0xc46ec6dd, duration=58076.214s, table=14, n_packets=0, n_bytes=0, priority=65535,icmp6,metadata=0x3,nw_ttl=255,icmp_type=135,icmp_code=0 actions=resubmit(,15)
 cookie=0xb32716b2, duration=58059.635s, table=14, n_packets=0, n_bytes=0, priority=65535,icmp6,metadata=0x4,nw_ttl=255,icmp_type=135,icmp_code=0 actions=resubmit(,15)
 cookie=0xb32716b2, duration=58059.635s, table=14, n_packets=0, n_bytes=0, priority=65535,icmp6,metadata=0x4,nw_ttl=255,icmp_type=136,icmp_code=0 actions=resubmit(,15)
 cookie=0xf3221d5c, duration=58076.214s, table=14, n_packets=0, n_bytes=0, priority=65535,ct_state=+inv+trk,metadata=0x3 actions=drop
 cookie=0xbcfa523d, duration=58059.635s, table=14, n_packets=0, n_bytes=0, priority=65535,ct_state=+inv+trk,metadata=0x4 actions=drop
 cookie=0x29c7e7e7, duration=58076.214s, table=14, n_packets=58007, n_bytes=5510752, priority=65535,ct_state=-new+est-rel+rpl-inv+trk,ct_label=0/0x1,metadata=0x3 actions=resubmit(,15)
 cookie=0x83e75fc5, duration=58059.635s, table=14, n_packets=0, n_bytes=0, priority=65535,ct_state=-new+est-rel+rpl-inv+trk,ct_label=0/0x1,metadata=0x4 actions=resubmit(,15)
 cookie=0xc682137d, duration=58076.214s, table=14, n_packets=11620, n_bytes=859952, priority=1,ct_state=-est+trk,ip,metadata=0x3 actions=load:0x1->NXM_NX_XXREG0[97],resubmit(,15)
 cookie=0xc682137d, duration=58076.214s, table=14, n_packets=0, n_bytes=0, priority=1,ct_state=-est+trk,ipv6,metadata=0x3 actions=load:0x1->NXM_NX_XXREG0[97],resubmit(,15)
 cookie=0x87d06f06, duration=58059.635s, table=14, n_packets=0, n_bytes=0, priority=1,ct_state=-est+trk,ipv6,metadata=0x4 actions=load:0x1->NXM_NX_XXREG0[97],resubmit(,15)
 cookie=0x87d06f06, duration=58059.635s, table=14, n_packets=0, n_bytes=0, priority=1,ct_state=-est+trk,ip,metadata=0x4 actions=load:0x1->NXM_NX_XXREG0[97],resubmit(,15)
 cookie=0xc682137d, duration=58076.214s, table=14, n_packets=0, n_bytes=0, priority=1,ct_state=+est+trk,ct_label=0x1/0x1,ipv6,metadata=0x3 actions=load:0x1->NXM_NX_XXREG0[97],resubmit(,15)
 cookie=0xc682137d, duration=58076.214s, table=14, n_packets=0, n_bytes=0, priority=1,ct_state=+est+trk,ct_label=0x1/0x1,ip,metadata=0x3 actions=load:0x1->NXM_NX_XXREG0[97],resubmit(,15)
 cookie=0x87d06f06, duration=58059.635s, table=14, n_packets=0, n_bytes=0, priority=1,ct_state=+est+trk,ct_label=0x1/0x1,ip,metadata=0x4 actions=load:0x1->NXM_NX_XXREG0[97],resubmit(,15)
 cookie=0x87d06f06, duration=58059.635s, table=14, n_packets=0, n_bytes=0, priority=1,ct_state=+est+trk,ct_label=0x1/0x1,ipv6,metadata=0x4 actions=load:0x1->NXM_NX_XXREG0[97],resubmit(,15)
 cookie=0x92217a09, duration=58078.069s, table=14, n_packets=1719693, n_bytes=61562780823, priority=0,metadata=0x3 actions=resubmit(,15)
 cookie=0xc17569d2, duration=58078.057s, table=14, n_packets=1673224, n_bytes=61558280345, priority=0,metadata=0x1 actions=resubmit(,15)
 cookie=0x638aeeb6, duration=58077.847s, table=14, n_packets=203090, n_bytes=70216033, priority=0,metadata=0x2 actions=resubmit(,15)
 cookie=0x912a924a, duration=58077.787s, table=14, n_packets=1470133, n_bytes=61488064222, priority=0,metadata=0x4 actions=resubmit(,15)
 cookie=0x15290b00, duration=58077.302s, table=14, n_packets=710300, n_bytes=101131271, priority=0,metadata=0x6 actions=resubmit(,15)
 cookie=0x70444df5, duration=58077.302s, table=14, n_packets=204957, n_bytes=70436259, priority=0,metadata=0x5 actions=resubmit(,15)
 cookie=0xc106b1af, duration=58078.057s, table=15, n_packets=0, n_bytes=0, priority=129,ipv6,reg14=0x2,metadata=0x1,ipv6_dst=fe80::/64 actions=dec_ttl(),move:NXM_NX_IPV6_DST[]->NXM_NX_XXREG0[],load:0x20000fffeeec6f9->NXM_NX_XXREG1[0..63],load:0xfe80000000000000->NXM_NX_XXREG1[64..127],mod_dl_src:00:00:00:ee:c6:f9,load:0x2->NXM_NX_REG15[],load:0x1->NXM_NX_REG10[0],resubmit(,16)
 cookie=0x8580c440, duration=58077.815s, table=15, n_packets=0, n_bytes=0, priority=129,ipv6,reg14=0x1,metadata=0x1,ipv6_dst=fe80::/64 actions=dec_ttl(),move:NXM_NX_IPV6_DST[]->NXM_NX_XXREG0[],load:0x20000fffe770c46->NXM_NX_XXREG1[0..63],load:0xfe80000000000000->NXM_NX_XXREG1[64..127],mod_dl_src:00:00:00:77:0c:46,load:0x1->NXM_NX_REG15[],load:0x1->NXM_NX_REG10[0],resubmit(,16)
 cookie=0x91b8ca5a, duration=58077.815s, table=15, n_packets=0, n_bytes=0, priority=129,ipv6,reg14=0x3,metadata=0x1,ipv6_dst=fe80::/64 actions=dec_ttl(),move:NXM_NX_IPV6_DST[]->NXM_NX_XXREG0[],load:0x20000fffe37698b->NXM_NX_XXREG1[0..63],load:0xfe80000000000000->NXM_NX_XXREG1[64..127],mod_dl_src:00:00:00:37:69:8b,load:0x3->NXM_NX_REG15[],load:0x1->NXM_NX_REG10[0],resubmit(,16)
 cookie=0x8dfe8e8b, duration=58077.301s, table=15, n_packets=0, n_bytes=0, priority=129,ipv6,reg14=0x2,metadata=0x5,ipv6_dst=fe80::/64 actions=dec_ttl(),move:NXM_NX_IPV6_DST[]->NXM_NX_XXREG0[],load:0x526b4bfffec39878->NXM_NX_XXREG1[0..63],load:0xfe80000000000000->NXM_NX_XXREG1[64..127],mod_dl_src:50:6b:4b:c3:98:78,load:0x2->NXM_NX_REG15[],load:0x1->NXM_NX_REG10[0],resubmit(,16)
 cookie=0x99e21991, duration=58077.301s, table=15, n_packets=0, n_bytes=0, priority=129,ipv6,reg14=0x1,metadata=0x5,ipv6_dst=fe80::/64 actions=dec_ttl(),move:NXM_NX_IPV6_DST[]->NXM_NX_XXREG0[],load:0x20000fffe909ce1->NXM_NX_XXREG1[0..63],load:0xfe80000000000000->NXM_NX_XXREG1[64..127],mod_dl_src:00:00:00:90:9c:e1,load:0x1->NXM_NX_REG15[],load:0x1->NXM_NX_REG10[0],resubmit(,16)
 cookie=0xa11b4f47, duration=58078.056s, table=15, n_packets=114816, n_bytes=63806298, priority=49,ip,metadata=0x1,nw_dst=10.244.1.0/24 actions=dec_ttl(),move:NXM_OF_IP_DST[]->NXM_NX_XXREG0[96..127],load:0xaf40101->NXM_NX_XXREG0[64..95],mod_dl_src:00:00:00:ee:c6:f9,load:0x2->NXM_NX_REG15[],load:0x1->NXM_NX_REG10[0],resubmit(,16)
 cookie=0xa4d55315, duration=58077.815s, table=15, n_packets=0, n_bytes=0, priority=49,ip,metadata=0x1,nw_dst=100.64.1.0/24 actions=dec_ttl(),move:NXM_OF_IP_DST[]->NXM_NX_XXREG0[96..127],load:0x64400101->NXM_NX_XXREG0[64..95],mod_dl_src:00:00:00:77:0c:46,load:0x1->NXM_NX_REG15[],load:0x1->NXM_NX_REG10[0],resubmit(,16)
 cookie=0xf27c7dcf, duration=58077.814s, table=15, n_packets=1470133, n_bytes=61488064222, priority=49,ip,metadata=0x1,nw_dst=10.244.2.0/24 actions=dec_ttl(),move:NXM_OF_IP_DST[]->NXM_NX_XXREG0[96..127],load:0xaf40201->NXM_NX_XXREG0[64..95],mod_dl_src:00:00:00:37:69:8b,load:0x3->NXM_NX_REG15[],load:0x1->NXM_NX_REG10[0],resubmit(,16)
 cookie=0xdd5a9e3, duration=58077.302s, table=15, n_packets=0, n_bytes=0, priority=49,ip,metadata=0x5,nw_dst=100.64.1.0/24 actions=dec_ttl(),move:NXM_OF_IP_DST[]->NXM_NX_XXREG0[96..127],load:0x64400102->NXM_NX_XXREG0[64..95],mod_dl_src:00:00:00:90:9c:e1,load:0x1->NXM_NX_REG15[],load:0x1->NXM_NX_REG10[0],resubmit(,16)
 cookie=0xc7d37986, duration=58077.303s, table=15, n_packets=0, n_bytes=0, priority=48,ip,metadata=0x1,nw_src=10.244.2.0/24 actions=dec_ttl(),load:0x64400103->NXM_NX_XXREG0[96..127],load:0x64400101->NXM_NX_XXREG0[64..95],mod_dl_src:00:00:00:77:0c:46,load:0x1->NXM_NX_REG15[],load:0x1->NXM_NX_REG10[0],resubmit(,16)
 cookie=0x86609539, duration=58077.303s, table=15, n_packets=88274, n_bytes=6409735, priority=48,ip,metadata=0x1,nw_src=10.244.1.0/24 actions=dec_ttl(),load:0x64400102->NXM_NX_XXREG0[96..127],load:0x64400101->NXM_NX_XXREG0[64..95],mod_dl_src:00:00:00:77:0c:46,load:0x1->NXM_NX_REG15[],load:0x1->NXM_NX_REG10[0],resubmit(,16)
 cookie=0xdea85015, duration=58077.302s, table=15, n_packets=88154, n_bytes=6397999, priority=41,ip,metadata=0x5,nw_dst=10.0.0.0/20 actions=dec_ttl(),move:NXM_OF_IP_DST[]->NXM_NX_XXREG0[96..127],load:0xa000213->NXM_NX_XXREG0[64..95],mod_dl_src:50:6b:4b:c3:98:78,load:0x2->NXM_NX_REG15[],load:0x1->NXM_NX_REG10[0],resubmit(,16)
 cookie=0x61eb534f, duration=58077.302s, table=15, n_packets=114816, n_bytes=63806298, priority=33,ip,metadata=0x5,nw_dst=10.244.0.0/16 actions=dec_ttl(),load:0x64400101->NXM_NX_XXREG0[96..127],load:0x64400102->NXM_NX_XXREG0[64..95],mod_dl_src:00:00:00:90:9c:e1,load:0x1->NXM_NX_REG15[],load:0x1->NXM_NX_REG10[0],resubmit(,16)
 cookie=0x4c5b30b4, duration=58077.303s, table=15, n_packets=0, n_bytes=0, priority=1,ip,metadata=0x1 actions=dec_ttl(),load:0x64400102->NXM_NX_XXREG0[96..127],load:0x64400101->NXM_NX_XXREG0[64..95],mod_dl_src:00:00:00:77:0c:46,load:0x1->NXM_NX_REG15[],load:0x1->NXM_NX_REG10[0],resubmit(,16)
 cookie=0x800186e8, duration=58077.301s, table=15, n_packets=120, n_bytes=11736, priority=1,ip,metadata=0x5 actions=dec_ttl(),load:0xa000001->NXM_NX_XXREG0[96..127],load:0xa000213->NXM_NX_XXREG0[64..95],mod_dl_src:50:6b:4b:c3:98:78,load:0x2->NXM_NX_REG15[],load:0x1->NXM_NX_REG10[0],resubmit(,16)
 cookie=0x17c5c97, duration=58078.056s, table=15, n_packets=1789320, n_bytes=61569151527, priority=0,metadata=0x3 actions=resubmit(,16)
 cookie=0x34bf2951, duration=58077.846s, table=15, n_packets=203090, n_bytes=70216033, priority=0,metadata=0x2 actions=resubmit(,16)
 cookie=0x2e2c004a, duration=58077.787s, table=15, n_packets=1470133, n_bytes=61488064222, priority=0,metadata=0x4 actions=resubmit(,16)
 cookie=0xbf9e1169, duration=58077.302s, table=15, n_packets=710300, n_bytes=101131271, priority=0,metadata=0x6 actions=resubmit(,16)
 cookie=0x58269046, duration=58078.067s, table=16, n_packets=0, n_bytes=0, priority=100,reg0=0xaf40102,reg15=0x2,metadata=0x1 actions=mod_dl_dst:0a:6c:a2:17:be:39,resubmit(,17)
 cookie=0x798cea2d, duration=58077.303s, table=16, n_packets=88274, n_bytes=6409735, priority=100,reg0=0x64400102,reg15=0x1,metadata=0x1 actions=mod_dl_dst:00:00:00:90:9c:e1,resubmit(,17)
 cookie=0x5c06adb4, duration=58077.303s, table=16, n_packets=0, n_bytes=0, priority=100,reg0=0xaf40202,reg15=0x3,metadata=0x1 actions=mod_dl_dst:fa:32:1b:ec:f8:de,resubmit(,17)
 cookie=0x261f5123, duration=58077.303s, table=16, n_packets=0, n_bytes=0, priority=100,reg0=0x64400103,reg15=0x1,metadata=0x1 actions=mod_dl_dst:00:00:00:d3:e3:de,resubmit(,17)
 cookie=0x1b8e26a7, duration=58077.301s, table=16, n_packets=114816, n_bytes=63806298, priority=100,reg0=0x64400101,reg15=0x1,metadata=0x5 actions=mod_dl_dst:00:00:00:77:0c:46,resubmit(,17)
 cookie=0x92a8a3e2, duration=58077.301s, table=16, n_packets=0, n_bytes=0, priority=100,reg0=0x64400103,reg15=0x1,metadata=0x5 actions=mod_dl_dst:00:00:00:d3:e3:de,resubmit(,17)
 cookie=0xa145e688, duration=58076.198s, table=16, n_packets=57290, n_bytes=31891102, priority=100,reg0=0xaf40103,reg15=0x2,metadata=0x1 actions=mod_dl_dst:0a:00:00:00:00:01,resubmit(,17)
 cookie=0xb29e19ca, duration=58076.153s, table=16, n_packets=57406, n_bytes=31903432, priority=100,reg0=0xaf40104,reg15=0x2,metadata=0x1 actions=mod_dl_dst:0a:00:00:00:00:02,resubmit(,17)
 cookie=0x3882ffc, duration=58059.621s, table=16, n_packets=1470133, n_bytes=61488064222, priority=100,reg0=0xaf40203,reg15=0x3,metadata=0x1 actions=mod_dl_dst:0a:00:00:00:00:03,resubmit(,17)
 cookie=0xf303224c, duration=58036.886s, table=16, n_packets=120, n_bytes=11764, priority=100,reg0=0xaf40105,reg15=0x2,metadata=0x1 actions=mod_dl_dst:0a:00:00:00:00:04,resubmit(,17)
 cookie=0xaef4fb40, duration=58077.303s, table=16, n_packets=0, n_bytes=0, priority=100,reg0=0xfe800000,reg1=0,reg2=0x20000ff,reg3=0xfe909ce1,reg15=0x1,metadata=0x1 actions=mod_dl_dst:00:00:00:90:9c:e1,resubmit(,17)
 cookie=0xc9ecda79, duration=58077.303s, table=16, n_packets=0, n_bytes=0, priority=100,reg0=0xfe800000,reg1=0,reg2=0x20000ff,reg3=0xfed3e3de,reg15=0x1,metadata=0x1 actions=mod_dl_dst:00:00:00:d3:e3:de,resubmit(,17)
 cookie=0x164a0233, duration=58077.302s, table=16, n_packets=0, n_bytes=0, priority=100,reg0=0xfe800000,reg1=0,reg2=0x20000ff,reg3=0xfe770c46,reg15=0x1,metadata=0x5 actions=mod_dl_dst:00:00:00:77:0c:46,resubmit(,17)
 cookie=0xe7328dfd, duration=58077.301s, table=16, n_packets=0, n_bytes=0, priority=100,reg0=0xfe800000,reg1=0,reg2=0x20000ff,reg3=0xfed3e3de,reg15=0x1,metadata=0x5 actions=mod_dl_dst:00:00:00:d3:e3:de,resubmit(,17)
 cookie=0xcddf36a9, duration=58078.069s, table=16, n_packets=1789320, n_bytes=61569151527, priority=0,metadata=0x3 actions=resubmit(,17)
 cookie=0x549767e8, duration=58077.815s, table=16, n_packets=203090, n_bytes=70216033, priority=0,metadata=0x2 actions=resubmit(,17)
 cookie=0xf64a7e4e, duration=58077.787s, table=16, n_packets=1470133, n_bytes=61488064222, priority=0,metadata=0x4 actions=resubmit(,17)
 cookie=0x45997dba, duration=58077.301s, table=16, n_packets=710300, n_bytes=101131271, priority=0,metadata=0x6 actions=resubmit(,17)
 cookie=0xf1467c61, duration=58078.067s, table=16, n_packets=0, n_bytes=0, priority=0,ipv6,metadata=0x1 actions=mod_dl_dst:00:00:00:00:00:00,resubmit(,66),resubmit(,17)
 cookie=0x76ca49cd, duration=58077.846s, table=16, n_packets=0, n_bytes=0, priority=0,ip,metadata=0x1 actions=push:NXM_NX_REG0[],push:NXM_NX_XXREG0[96..127],pop:NXM_NX_REG0[],mod_dl_dst:00:00:00:00:00:00,resubmit(,66),pop:NXM_NX_REG0[],resubmit(,17)
 cookie=0xfb466dab, duration=58077.302s, table=16, n_packets=0, n_bytes=0, priority=0,ipv6,metadata=0x5 actions=mod_dl_dst:00:00:00:00:00:00,resubmit(,66),resubmit(,17)
 cookie=0x312a6c76, duration=58077.301s, table=16, n_packets=88274, n_bytes=6409735, priority=0,ip,metadata=0x5 actions=push:NXM_NX_REG0[],push:NXM_NX_XXREG0[96..127],pop:NXM_NX_REG0[],mod_dl_dst:00:00:00:00:00:00,resubmit(,66),pop:NXM_NX_REG0[],resubmit(,17)
 cookie=0xed0fa60b, duration=58077.814s, table=17, n_packets=1662863, n_bytes=61504484023, priority=65535,ct_state=-new+est-rel-inv+trk,metadata=0x3 actions=load:0x1->NXM_NX_XXREG0[98],resubmit(,18)
 cookie=0x35128855, duration=58077.303s, table=17, n_packets=0, n_bytes=0, priority=65535,ct_state=-new+est-rel-inv+trk,metadata=0x4 actions=load:0x1->NXM_NX_XXREG0[98],resubmit(,18)
 cookie=0x54083c63, duration=58078.057s, table=17, n_packets=203090, n_bytes=70216033, priority=0,metadata=0x2 actions=resubmit(,18)
 cookie=0x136e0ae7, duration=58077.847s, table=17, n_packets=126457, n_bytes=64667504, priority=0,metadata=0x3 actions=resubmit(,18)
 cookie=0x5a27df24, duration=58077.847s, table=17, n_packets=1673223, n_bytes=61558280255, priority=0,metadata=0x1 actions=resubmit(,18)
 cookie=0x573fba00, duration=58077.781s, table=17, n_packets=1470133, n_bytes=61488064222, priority=0,metadata=0x4 actions=resubmit(,18)
 cookie=0xac941c47, duration=58077.302s, table=17, n_packets=203090, n_bytes=70216033, priority=0,metadata=0x5 actions=resubmit(,18)
 cookie=0x608ee87e, duration=58077.301s, table=17, n_packets=710300, n_bytes=101131271, priority=0,metadata=0x6 actions=resubmit(,18)
 cookie=0xfb68bda0, duration=58077.814s, table=18, n_packets=12, n_bytes=888, priority=120,ct_state=+new+trk,tcp,metadata=0x3,nw_dst=10.96.0.1,tp_dst=443 actions=group:1
 cookie=0x3822c0b7, duration=58077.303s, table=18, n_packets=0, n_bytes=0, priority=120,ct_state=+new+trk,tcp,metadata=0x4,nw_dst=10.96.0.1,tp_dst=443 actions=group:1
 cookie=0x84c89280, duration=58071.594s, table=18, n_packets=0, n_bytes=0, priority=120,ct_state=+new+trk,tcp,metadata=0x3,nw_dst=10.96.0.10,tp_dst=53 actions=group:3
 cookie=0x3fab981f, duration=58071.594s, table=18, n_packets=0, n_bytes=0, priority=120,ct_state=+new+trk,tcp,metadata=0x4,nw_dst=10.96.0.10,tp_dst=53 actions=group:3
 cookie=0x4d20cb1e, duration=58071.549s, table=18, n_packets=0, n_bytes=0, priority=120,ct_state=+new+trk,udp,metadata=0x4,nw_dst=10.96.0.10,tp_dst=53 actions=group:3
 cookie=0xe787f020, duration=58071.549s, table=18, n_packets=0, n_bytes=0, priority=120,ct_state=+new+trk,udp,metadata=0x3,nw_dst=10.96.0.10,tp_dst=53 actions=group:3
 cookie=0x3fbd6f6f, duration=58054.582s, table=18, n_packets=2, n_bytes=148, priority=120,ct_state=+new+trk,tcp,metadata=0x3,nw_dst=10.104.114.18,tp_dst=10005 actions=group:2
 cookie=0xb00eb3e3, duration=58054.582s, table=18, n_packets=0, n_bytes=0, priority=120,ct_state=+new+trk,tcp,metadata=0x4,nw_dst=10.104.114.18,tp_dst=10005 actions=group:2
 cookie=0x96a0342e, duration=58078.069s, table=18, n_packets=0, n_bytes=0, priority=100,ipv6,reg0=0x4/0x4,metadata=0x3 actions=ct(table=19,zone=NXM_NX_REG13[0..15],nat)
 cookie=0xf67f5201, duration=58078.069s, table=18, n_packets=0, n_bytes=0, priority=100,ip,reg0=0x4/0x4,metadata=0x2 actions=ct(table=19,zone=NXM_NX_REG13[0..15],nat)
 cookie=0x96a0342e, duration=58078.069s, table=18, n_packets=1662863, n_bytes=61504484023, priority=100,ip,reg0=0x4/0x4,metadata=0x3 actions=ct(table=19,zone=NXM_NX_REG13[0..15],nat)
 cookie=0xf67f5201, duration=58078.067s, table=18, n_packets=0, n_bytes=0, priority=100,ipv6,reg0=0x4/0x4,metadata=0x2 actions=ct(table=19,zone=NXM_NX_REG13[0..15],nat)
 cookie=0xc8ef8a5e, duration=58077.787s, table=18, n_packets=0, n_bytes=0, priority=100,ipv6,reg0=0x4/0x4,metadata=0x4 actions=ct(table=19,zone=NXM_NX_REG13[0..15],nat)
 cookie=0xc8ef8a5e, duration=58077.787s, table=18, n_packets=0, n_bytes=0, priority=100,ip,reg0=0x4/0x4,metadata=0x4 actions=ct(table=19,zone=NXM_NX_REG13[0..15],nat)
 cookie=0x2949fa9e, duration=58077.302s, table=18, n_packets=0, n_bytes=0, priority=100,ipv6,reg0=0x4/0x4,metadata=0x6 actions=ct(table=19,zone=NXM_NX_REG13[0..15],nat)
 cookie=0x2949fa9e, duration=58077.302s, table=18, n_packets=0, n_bytes=0, priority=100,ip,reg0=0x4/0x4,metadata=0x6 actions=ct(table=19,zone=NXM_NX_REG13[0..15],nat)
 cookie=0x15bc0aa1, duration=58078.067s, table=18, n_packets=0, n_bytes=0, priority=100,ipv6,reg0=0x2/0x2,metadata=0x3 actions=ct(commit,zone=NXM_NX_REG13[0..15],exec(load:0->NXM_NX_CT_LABEL[0])),resubmit(,19)
 cookie=0xb2d08e34, duration=58078.057s, table=18, n_packets=0, n_bytes=0, priority=100,ip,reg0=0x2/0x2,metadata=0x2 actions=ct(commit,zone=NXM_NX_REG13[0..15],exec(load:0->NXM_NX_CT_LABEL[0])),resubmit(,19)
 cookie=0x15bc0aa1, duration=58078.057s, table=18, n_packets=11606, n_bytes=858916, priority=100,ip,reg0=0x2/0x2,metadata=0x3 actions=ct(commit,zone=NXM_NX_REG13[0..15],exec(load:0->NXM_NX_CT_LABEL[0])),resubmit(,19)
 cookie=0xb2d08e34, duration=58077.846s, table=18, n_packets=0, n_bytes=0, priority=100,ipv6,reg0=0x2/0x2,metadata=0x2 actions=ct(commit,zone=NXM_NX_REG13[0..15],exec(load:0->NXM_NX_CT_LABEL[0])),resubmit(,19)
 cookie=0x45ef63ee, duration=58077.787s, table=18, n_packets=0, n_bytes=0, priority=100,ipv6,reg0=0x2/0x2,metadata=0x4 actions=ct(commit,zone=NXM_NX_REG13[0..15],exec(load:0->NXM_NX_CT_LABEL[0])),resubmit(,19)
 cookie=0x45ef63ee, duration=58077.781s, table=18, n_packets=0, n_bytes=0, priority=100,ip,reg0=0x2/0x2,metadata=0x4 actions=ct(commit,zone=NXM_NX_REG13[0..15],exec(load:0->NXM_NX_CT_LABEL[0])),resubmit(,19)
 cookie=0xad6922b7, duration=58077.302s, table=18, n_packets=0, n_bytes=0, priority=100,ipv6,reg0=0x2/0x2,metadata=0x6 actions=ct(commit,zone=NXM_NX_REG13[0..15],exec(load:0->NXM_NX_CT_LABEL[0])),resubmit(,19)
 cookie=0xad6922b7, duration=58077.302s, table=18, n_packets=0, n_bytes=0, priority=100,ip,reg0=0x2/0x2,metadata=0x6 actions=ct(commit,zone=NXM_NX_REG13[0..15],exec(load:0->NXM_NX_CT_LABEL[0])),resubmit(,19)
 cookie=0xaacf3ddb, duration=58078.067s, table=18, n_packets=0, n_bytes=0, priority=100,ipv6,metadata=0x1,dl_dst=00:00:00:00:00:00 actions=controller(userdata=00.00.00.09.00.00.00.00.ff.ff.00.18.00.00.23.20.00.06.00.80.00.00.00.00.00.01.de.10.00.01.2e.10.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.20.00.00.00)
 cookie=0x1535f82e, duration=58077.815s, table=18, n_packets=0, n_bytes=0, priority=100,ip,metadata=0x1,dl_dst=00:00:00:00:00:00 actions=controller(userdata=00.00.00.00.00.00.00.00.00.19.00.10.80.00.06.06.ff.ff.ff.ff.ff.ff.00.00.ff.ff.00.18.00.00.23.20.00.06.00.20.00.40.00.00.00.01.de.10.00.00.20.04.ff.ff.00.18.00.00.23.20.00.06.00.20.00.60.00.00.00.01.de.10.00.00.22.04.00.19.00.10.80.00.2a.02.00.01.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.20.00.00.00)
 cookie=0xfa2cbdd8, duration=58077.302s, table=18, n_packets=6, n_bytes=444, priority=100,ip,metadata=0x5,dl_dst=00:00:00:00:00:00 actions=controller(userdata=00.00.00.00.00.00.00.00.00.19.00.10.80.00.06.06.ff.ff.ff.ff.ff.ff.00.00.ff.ff.00.18.00.00.23.20.00.06.00.20.00.40.00.00.00.01.de.10.00.00.20.04.ff.ff.00.18.00.00.23.20.00.06.00.20.00.60.00.00.00.01.de.10.00.00.22.04.00.19.00.10.80.00.2a.02.00.01.00.00.00.00.00.00.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.20.00.00.00)
 cookie=0x7ee220f2, duration=58077.302s, table=18, n_packets=0, n_bytes=0, priority=100,ipv6,metadata=0x5,dl_dst=00:00:00:00:00:00 actions=controller(userdata=00.00.00.09.00.00.00.00.ff.ff.00.18.00.00.23.20.00.06.00.80.00.00.00.00.00.01.de.10.00.01.2e.10.ff.ff.00.10.00.00.23.20.00.0e.ff.f8.20.00.00.00)
 cookie=0x649d87e5, duration=58078.068s, table=18, n_packets=203090, n_bytes=70216033, priority=0,metadata=0x2 actions=resubmit(,19)
 cookie=0x375a4da9, duration=58078.058s, table=18, n_packets=1673223, n_bytes=61558280255, priority=0,metadata=0x1 actions=resubmit(,32)
 cookie=0xfbee841a, duration=58077.847s, table=18, n_packets=114837, n_bytes=63807552, priority=0,metadata=0x3 actions=resubmit(,19)
 cookie=0x81c49769, duration=58077.787s, table=18, n_packets=1470133, n_bytes=61488064222, priority=0,metadata=0x4 actions=resubmit(,19)
 cookie=0xe49575f9, duration=58077.303s, table=18, n_packets=710300, n_bytes=101131271, priority=0,metadata=0x6 actions=resubmit(,19)
 cookie=0xe2262fa9, duration=58077.303s, table=18, n_packets=203084, n_bytes=70215589, priority=0,metadata=0x5 actions=resubmit(,32)
 cookie=0x41ba5414, duration=58077.816s, table=19, n_packets=0, n_bytes=0, priority=100,arp,reg14=0x2,metadata=0x3,arp_tpa=10.244.1.2,arp_op=1 actions=resubmit(,20)
 cookie=0xe281e1f6, duration=58072.094s, table=19, n_packets=0, n_bytes=0, priority=100,arp,reg14=0x4,metadata=0x3,arp_tpa=10.244.1.4,arp_op=1 actions=resubmit(,20)
 cookie=0x5dd7ec96, duration=58071.177s, table=19, n_packets=0, n_bytes=0, priority=100,arp,reg14=0x3,metadata=0x3,arp_tpa=10.244.1.3,arp_op=1 actions=resubmit(,20)
 cookie=0xeff94c2e, duration=58033.558s, table=19, n_packets=0, n_bytes=0, priority=100,arp,reg14=0x5,metadata=0x3,arp_tpa=10.244.1.5,arp_op=1 actions=resubmit(,20)
 cookie=0x7712cb25, duration=58077.816s, table=19, n_packets=2, n_bytes=120, priority=50,arp,metadata=0x3,arp_tpa=10.244.1.2,arp_op=1 actions=move:NXM_OF_ETH_SRC[]->NXM_OF_ETH_DST[],mod_dl_src:0a:6c:a2:17:be:39,load:0x2->NXM_OF_ARP_OP[],move:NXM_NX_ARP_SHA[]->NXM_NX_ARP_THA[],load:0xa6ca217be39->NXM_NX_ARP_SHA[],move:NXM_OF_ARP_SPA[]->NXM_OF_ARP_TPA[],load:0xaf40102->NXM_OF_ARP_SPA[],move:NXM_NX_REG14[]->NXM_NX_REG15[],load:0x1->NXM_NX_REG10[0],resubmit(,32)
 cookie=0x4dc461f, duration=58077.304s, table=19, n_packets=0, n_bytes=0, priority=50,arp,metadata=0x4,arp_tpa=10.244.2.2,arp_op=1 actions=move:NXM_OF_ETH_SRC[]->NXM_OF_ETH_DST[],mod_dl_src:fa:32:1b:ec:f8:de,load:0x2->NXM_OF_ARP_OP[],move:NXM_NX_ARP_SHA[]->NXM_NX_ARP_THA[],load:0xfa321becf8de->NXM_NX_ARP_SHA[],move:NXM_OF_ARP_SPA[]->NXM_OF_ARP_TPA[],load:0xaf40202->NXM_OF_ARP_SPA[],move:NXM_NX_REG14[]->NXM_NX_REG15[],load:0x1->NXM_NX_REG10[0],resubmit(,32)
 cookie=0xc668fa9d, duration=58072.094s, table=19, n_packets=1, n_bytes=42, priority=50,arp,metadata=0x3,arp_tpa=10.244.1.4,arp_op=1 actions=move:NXM_OF_ETH_SRC[]->NXM_OF_ETH_DST[],mod_dl_src:0a:00:00:00:00:02,load:0x2->NXM_OF_ARP_OP[],move:NXM_NX_ARP_SHA[]->NXM_NX_ARP_THA[],load:0xa0000000002->NXM_NX_ARP_SHA[],move:NXM_OF_ARP_SPA[]->NXM_OF_ARP_TPA[],load:0xaf40104->NXM_OF_ARP_SPA[],move:NXM_NX_REG14[]->NXM_NX_REG15[],load:0x1->NXM_NX_REG10[0],resubmit(,32)
 cookie=0x3cbb6c66, duration=58071.177s, table=19, n_packets=1, n_bytes=42, priority=50,arp,metadata=0x3,arp_tpa=10.244.1.3,arp_op=1 actions=move:NXM_OF_ETH_SRC[]->NXM_OF_ETH_DST[],mod_dl_src:0a:00:00:00:00:01,load:0x2->NXM_OF_ARP_OP[],move:NXM_NX_ARP_SHA[]->NXM_NX_ARP_THA[],load:0xa0000000001->NXM_NX_ARP_SHA[],move:NXM_OF_ARP_SPA[]->NXM_OF_ARP_TPA[],load:0xaf40103->NXM_OF_ARP_SPA[],move:NXM_NX_REG14[]->NXM_NX_REG15[],load:0x1->NXM_NX_REG10[0],resubmit(,32)
 cookie=0x644e0e63, duration=58056.444s, table=19, n_packets=0, n_bytes=0, priority=50,arp,metadata=0x4,arp_tpa=10.244.2.3,arp_op=1 actions=move:NXM_OF_ETH_SRC[]->NXM_OF_ETH_DST[],mod_dl_src:0a:00:00:00:00:03,load:0x2->NXM_OF_ARP_OP[],move:NXM_NX_ARP_SHA[]->NXM_NX_ARP_THA[],load:0xa0000000003->NXM_NX_ARP_SHA[],move:NXM_OF_ARP_SPA[]->NXM_OF_ARP_TPA[],load:0xaf40203->NXM_OF_ARP_SPA[],move:NXM_NX_REG14[]->NXM_NX_REG15[],load:0x1->NXM_NX_REG10[0],resubmit(,32)
 cookie=0xfe3b7b2f, duration=58033.558s, table=19, n_packets=0, n_bytes=0, priority=50,arp,metadata=0x3,arp_tpa=10.244.1.5,arp_op=1 actions=move:NXM_OF_ETH_SRC[]->NXM_OF_ETH_DST[],mod_dl_src:0a:00:00:00:00:04,load:0x2->NXM_OF_ARP_OP[],move:NXM_NX_ARP_SHA[]->NXM_NX_ARP_THA[],load:0xa0000000004->NXM_NX_ARP_SHA[],move:NXM_OF_ARP_SPA[]->NXM_OF_ARP_TPA[],load:0xaf40105->NXM_OF_ARP_SPA[],move:NXM_NX_REG14[]->NXM_NX_REG15[],load:0x1->NXM_NX_REG10[0],resubmit(,32)
 cookie=0xb7374f19, duration=58078.068s, table=19, n_packets=203090, n_bytes=70216033, priority=0,metadata=0x2 actions=resubmit(,20)
 cookie=0xbc447e28, duration=58078.058s, table=19, n_packets=1789333, n_bytes=61569109309, priority=0,metadata=0x3 actions=resubmit(,20)
 cookie=0x1e496a04, duration=58077.788s, table=19, n_packets=1470133, n_bytes=61488064222, priority=0,metadata=0x4 actions=resubmit(,20)
 cookie=0x4a270180, duration=58077.302s, table=19, n_packets=710300, n_bytes=101131271, priority=0,metadata=0x6 actions=resubmit(,20)
 cookie=0x13c187be, duration=58078.068s, table=20, n_packets=203090, n_bytes=70216033, priority=0,metadata=0x2 actions=resubmit(,21)
 cookie=0x963bb5d5, duration=58078.068s, table=20, n_packets=1789333, n_bytes=61569109309, priority=0,metadata=0x3 actions=resubmit(,21)
 cookie=0x6ac7d7c7, duration=58077.788s, table=20, n_packets=1470133, n_bytes=61488064222, priority=0,metadata=0x4 actions=resubmit(,21)
 cookie=0x9cea02a9, duration=58077.302s, table=20, n_packets=710300, n_bytes=101131271, priority=0,metadata=0x6 actions=resubmit(,21)
 cookie=0x9381f416, duration=58078.068s, table=21, n_packets=203090, n_bytes=70216033, priority=0,metadata=0x2 actions=resubmit(,22)
 cookie=0x73e15948, duration=58077.848s, table=21, n_packets=1789333, n_bytes=61569109309, priority=0,metadata=0x3 actions=resubmit(,22)
 cookie=0x2d03dfd6, duration=58077.782s, table=21, n_packets=1470133, n_bytes=61488064222, priority=0,metadata=0x4 actions=resubmit(,22)
 cookie=0xf6383fee, duration=58077.303s, table=21, n_packets=710300, n_bytes=101131271, priority=0,metadata=0x6 actions=resubmit(,22)
 cookie=0xc7507ed0, duration=58078.068s, table=22, n_packets=203090, n_bytes=70216033, priority=0,metadata=0x2 actions=resubmit(,23)
 cookie=0xcf7444bd, duration=58078.057s, table=22, n_packets=1789333, n_bytes=61569109309, priority=0,metadata=0x3 actions=resubmit(,23)
 cookie=0x11dd739c, duration=58077.782s, table=22, n_packets=1470133, n_bytes=61488064222, priority=0,metadata=0x4 actions=resubmit(,23)
 cookie=0xf7d05d2e, duration=58077.303s, table=22, n_packets=710300, n_bytes=101131271, priority=0,metadata=0x6 actions=resubmit(,23)
 cookie=0x121aa49, duration=58078.058s, table=23, n_packets=203090, n_bytes=70216033, priority=0,metadata=0x2 actions=resubmit(,24)
 cookie=0x2c9d3509, duration=58077.848s, table=23, n_packets=1789333, n_bytes=61569109309, priority=0,metadata=0x3 actions=resubmit(,24)
 cookie=0xea306ca1, duration=58077.787s, table=23, n_packets=1470133, n_bytes=61488064222, priority=0,metadata=0x4 actions=resubmit(,24)
 cookie=0xed6483fa, duration=58077.303s, table=23, n_packets=710300, n_bytes=101131271, priority=0,metadata=0x6 actions=resubmit(,24)
 cookie=0xa058b6ca, duration=58078.058s, table=24, n_packets=0, n_bytes=0, priority=100,metadata=0x2,dl_dst=01:00:00:00:00:00/01:00:00:00:00:00 actions=load:0xffff->NXM_NX_REG15[],resubmit(,32)
 cookie=0x9975c41c, duration=58078.057s, table=24, n_packets=4, n_bytes=270, priority=100,metadata=0x3,dl_dst=01:00:00:00:00:00/01:00:00:00:00:00 actions=load:0xffff->NXM_NX_REG15[],resubmit(,32)
 cookie=0x5dc50cf4, duration=58077.787s, table=24, n_packets=0, n_bytes=0, priority=100,metadata=0x4,dl_dst=01:00:00:00:00:00/01:00:00:00:00:00 actions=load:0xffff->NXM_NX_REG15[],resubmit(,32)
 cookie=0x98d114b5, duration=58077.303s, table=24, n_packets=492130, n_bytes=29636128, priority=100,metadata=0x6,dl_dst=01:00:00:00:00:00/01:00:00:00:00:00 actions=load:0xffff->NXM_NX_REG15[],resubmit(,32)
 cookie=0x5d2744bb, duration=58078.068s, table=24, n_packets=114816, n_bytes=63806298, priority=50,metadata=0x2,dl_dst=00:00:00:77:0c:46 actions=load:0x1->NXM_NX_REG15[],resubmit(,32)
 cookie=0xea3eae5c, duration=58078.057s, table=24, n_packets=58007, n_bytes=5510752, priority=50,metadata=0x3,dl_dst=0a:6c:a2:17:be:39 actions=load:0x2->NXM_NX_REG15[],resubmit(,32)
 cookie=0x3e19bf59, duration=58078.057s, table=24, n_packets=1558413, n_bytes=61494495891, priority=50,metadata=0x3,dl_dst=00:00:00:ee:c6:f9 actions=load:0x1->NXM_NX_REG15[],resubmit(,32)
 cookie=0xf38e46f5, duration=58077.788s, table=24, n_packets=0, n_bytes=0, priority=50,metadata=0x4,dl_dst=00:00:00:37:69:8b actions=load:0x1->NXM_NX_REG15[],resubmit(,32)
 cookie=0x6e4d1db2, duration=58077.304s, table=24, n_packets=0, n_bytes=0, priority=50,metadata=0x2,dl_dst=00:00:00:d3:e3:de actions=load:0x3->NXM_NX_REG15[],resubmit(,32)
 cookie=0x4a993643, duration=58077.304s, table=24, n_packets=88274, n_bytes=6409735, priority=50,metadata=0x2,dl_dst=00:00:00:90:9c:e1 actions=load:0x2->NXM_NX_REG15[],resubmit(,32)
 cookie=0x3570066d, duration=58077.304s, table=24, n_packets=0, n_bytes=0, priority=50,metadata=0x4,dl_dst=fa:32:1b:ec:f8:de actions=load:0x2->NXM_NX_REG15[],resubmit(,32)
 cookie=0x9e61b3d8, duration=58077.303s, table=24, n_packets=114822, n_bytes=63806658, priority=50,metadata=0x6,dl_dst=50:6b:4b:c3:98:78 actions=load:0x2->NXM_NX_REG15[],resubmit(,32)
 cookie=0x6bc176b9, duration=58076.199s, table=24, n_packets=86328, n_bytes=34538566, priority=50,metadata=0x3,dl_dst=0a:00:00:00:00:01 actions=load:0x3->NXM_NX_REG15[],resubmit(,32)
 cookie=0x7b181f48, duration=58076.154s, table=24, n_packets=86455, n_bytes=34551706, priority=50,metadata=0x3,dl_dst=0a:00:00:00:00:02 actions=load:0x4->NXM_NX_REG15[],resubmit(,32)
 cookie=0xa68766, duration=58059.622s, table=24, n_packets=1470133, n_bytes=61488064222, priority=50,metadata=0x4,dl_dst=0a:00:00:00:00:03 actions=load:0x3->NXM_NX_REG15[],resubmit(,32)
 cookie=0x31e3d07c, duration=58036.887s, table=24, n_packets=126, n_bytes=12124, priority=50,metadata=0x3,dl_dst=0a:00:00:00:00:04 actions=load:0x5->NXM_NX_REG15[],resubmit(,32)
 cookie=0x9e70aac1, duration=58077.303s, table=24, n_packets=103348, n_bytes=7688485, priority=0,metadata=0x6 actions=load:0xfffe->NXM_NX_REG15[],resubmit(,32)
 cookie=0x0, duration=58078.885s, table=32, n_packets=0, n_bytes=0, priority=150,reg10=0x10/0x10 actions=resubmit(,33)
 cookie=0x0, duration=58078.885s, table=32, n_packets=0, n_bytes=0, priority=150,reg10=0x2/0x2 actions=resubmit(,33)
 cookie=0x0, duration=58077.848s, table=32, n_packets=4, n_bytes=270, priority=100,reg15=0xffff,metadata=0x3 actions=load:0x1->NXM_NX_REG15[],resubmit(,34),load:0xffff->NXM_NX_REG15[],resubmit(,33)
 cookie=0x0, duration=58077.816s, table=32, n_packets=0, n_bytes=0, priority=100,reg15=0xffff,metadata=0x2 actions=load:0x1->NXM_NX_REG15[],resubmit(,34),load:0xffff->NXM_NX_REG15[],load:0x2->NXM_NX_TUN_ID[0..23],set_field:0xffff->tun_metadata0,move:NXM_NX_REG14[0..14]->NXM_NX_TUN_METADATA0[16..30],output:"ovn-9e8fa0-0",resubmit(,33)
 cookie=0x0, duration=58077.787s, table=32, n_packets=0, n_bytes=0, priority=100,reg15=0xffff,metadata=0x4 actions=load:0x1->NXM_NX_REG15[],resubmit(,34),load:0xffff->NXM_NX_REG15[],load:0x4->NXM_NX_TUN_ID[0..23],set_field:0xffff->tun_metadata0,move:NXM_NX_REG14[0..14]->NXM_NX_TUN_METADATA0[16..30],output:"ovn-9e8fa0-0"
 cookie=0x0, duration=58077.304s, table=32, n_packets=0, n_bytes=0, priority=100,reg15=0x2,metadata=0x4 actions=load:0x4->NXM_NX_TUN_ID[0..23],set_field:0x2->tun_metadata0,move:NXM_NX_REG14[0..14]->NXM_NX_TUN_METADATA0[16..30],output:"ovn-9e8fa0-0"
 cookie=0x0, duration=58077.304s, table=32, n_packets=0, n_bytes=0, priority=100,reg15=0x3,metadata=0x2 actions=load:0x2->NXM_NX_TUN_ID[0..23],set_field:0x3->tun_metadata0,move:NXM_NX_REG14[0..14]->NXM_NX_TUN_METADATA0[16..30],output:"ovn-9e8fa0-0"
 cookie=0x0, duration=58056.452s, table=32, n_packets=1470133, n_bytes=61488064222, priority=100,reg15=0x3,metadata=0x4 actions=load:0x4->NXM_NX_TUN_ID[0..23],set_field:0x3->tun_metadata0,move:NXM_NX_REG14[0..14]->NXM_NX_TUN_METADATA0[16..30],output:"ovn-9e8fa0-0"
 cookie=0x0, duration=58078.885s, table=32, n_packets=4579044, n_bytes=123368953123, priority=0 actions=resubmit(,33)
 cookie=0x0, duration=58078.070s, table=33, n_packets=58009, n_bytes=5510836, priority=100,reg15=0x2,metadata=0x3 actions=load:0x1->NXM_NX_REG13[],load:0x5->NXM_NX_REG11[],load:0x6->NXM_NX_REG12[],resubmit(,34)
 cookie=0x0, duration=58078.070s, table=33, n_packets=114816, n_bytes=63806298, priority=100,reg15=0x1,metadata=0x2 actions=load:0x2->NXM_NX_REG11[],load:0x3->NXM_NX_REG12[],resubmit(,34)
 cookie=0x0, duration=58078.070s, table=33, n_packets=1558413, n_bytes=61494495891, priority=100,reg15=0x1,metadata=0x3 actions=load:0x5->NXM_NX_REG11[],load:0x6->NXM_NX_REG12[],resubmit(,34)
 cookie=0x0, duration=58078.070s, table=33, n_packets=114824, n_bytes=63806778, priority=100,reg15=0x2,metadata=0x1 actions=load:0x7->NXM_NX_REG11[],load:0x4->NXM_NX_REG12[],resubmit(,34)
 cookie=0x0, duration=58078.070s, table=33, n_packets=88274, n_bytes=6409735, priority=100,reg15=0x1,metadata=0x1 actions=load:0x7->NXM_NX_REG11[],load:0x4->NXM_NX_REG12[],resubmit(,34)
 cookie=0x0, duration=58077.788s, table=33, n_packets=1470133, n_bytes=61488064222, priority=100,reg15=0x3,metadata=0x1 actions=load:0x7->NXM_NX_REG11[],load:0x4->NXM_NX_REG12[],resubmit(,34)
 cookie=0x0, duration=58077.787s, table=33, n_packets=0, n_bytes=0, priority=100,reg15=0x1,metadata=0x4 actions=load:0x9->NXM_NX_REG11[],load:0x8->NXM_NX_REG12[],resubmit(,34)
 cookie=0x0, duration=58077.304s, table=33, n_packets=0, n_bytes=0, priority=100,reg15=0x1,metadata=0x6 actions=load:0xa->NXM_NX_REG13[],load:0xc->NXM_NX_REG11[],load:0xb->NXM_NX_REG12[],resubmit(,34)
 cookie=0x0, duration=58077.304s, table=33, n_packets=88274, n_bytes=6409735, priority=100,reg15=0x2,metadata=0x2 actions=load:0x2->NXM_NX_REG11[],load:0x3->NXM_NX_REG12[],resubmit(,34)
 cookie=0x0, duration=58077.304s, table=33, n_packets=114816, n_bytes=63806298, priority=100,reg15=0x1,metadata=0x5 actions=load:0xe->NXM_NX_REG11[],load:0xd->NXM_NX_REG12[],resubmit(,34)
 cookie=0x0, duration=58077.304s, table=33, n_packets=0, n_bytes=0, priority=100,reg15=0xffff,metadata=0x2 actions=load:0x2->NXM_NX_REG15[],resubmit(,34),load:0xffff->NXM_NX_REG15[]
 cookie=0x0, duration=58077.304s, table=33, n_packets=88274, n_bytes=6409543, priority=100,reg15=0x2,metadata=0x5 actions=load:0xe->NXM_NX_REG11[],load:0xd->NXM_NX_REG12[],resubmit(,34)
 cookie=0x0, duration=58077.304s, table=33, n_packets=114822, n_bytes=63806658, priority=100,reg15=0x2,metadata=0x6 actions=load:0xc->NXM_NX_REG11[],load:0xb->NXM_NX_REG12[],resubmit(,34)
 cookie=0x0, duration=58077.303s, table=33, n_packets=103348, n_bytes=7688485, priority=100,reg15=0xfffe,metadata=0x6 actions=load:0xa->NXM_NX_REG13[],load:0x1->NXM_NX_REG15[],resubmit(,34),load:0xfffe->NXM_NX_REG15[]
 cookie=0x0, duration=58077.302s, table=33, n_packets=492130, n_bytes=29636128, priority=100,reg15=0xffff,metadata=0x6 actions=load:0xa->NXM_NX_REG13[],load:0x1->NXM_NX_REG15[],resubmit(,34),load:0x2->NXM_NX_REG15[],resubmit(,34),load:0xffff->NXM_NX_REG15[]
 cookie=0x0, duration=58072.094s, table=33, n_packets=86456, n_bytes=34551766, priority=100,reg15=0x4,metadata=0x3 actions=load:0xf->NXM_NX_REG13[],load:0x5->NXM_NX_REG11[],load:0x6->NXM_NX_REG12[],resubmit(,34)
 cookie=0x0, duration=58071.187s, table=33, n_packets=86329, n_bytes=34538626, priority=100,reg15=0x3,metadata=0x3 actions=load:0x10->NXM_NX_REG13[],load:0x5->NXM_NX_REG11[],load:0x6->NXM_NX_REG12[],resubmit(,34)
 cookie=0x0, duration=58078.068s, table=33, n_packets=4, n_bytes=270, priority=100,reg15=0xffff,metadata=0x3 actions=load:0xf->NXM_NX_REG13[],load:0x4->NXM_NX_REG15[],resubmit(,34),load:0x1->NXM_NX_REG13[],load:0x2->NXM_NX_REG15[],resubmit(,34),load:0x11->NXM_NX_REG13[],load:0x5->NXM_NX_REG15[],resubmit(,34),load:0x10->NXM_NX_REG13[],load:0x3->NXM_NX_REG15[],resubmit(,34),load:0xffff->NXM_NX_REG15[]
 cookie=0x0, duration=58033.569s, table=33, n_packets=1230436, n_bytes=81267444, priority=100,reg15=0x5,metadata=0x3 actions=load:0x11->NXM_NX_REG13[],load:0x5->NXM_NX_REG11[],load:0x6->NXM_NX_REG12[],resubmit(,34)
 cookie=0x0, duration=58078.070s, table=34, n_packets=0, n_bytes=0, priority=100,reg10=0/0x1,reg14=0x2,reg15=0x2,metadata=0x1 actions=drop
 cookie=0x0, duration=58078.070s, table=34, n_packets=1, n_bytes=90, priority=100,reg10=0/0x1,reg14=0x2,reg15=0x2,metadata=0x3 actions=drop
 cookie=0x0, duration=58078.070s, table=34, n_packets=0, n_bytes=0, priority=100,reg10=0/0x1,reg14=0x1,reg15=0x1,metadata=0x3 actions=drop
 cookie=0x0, duration=58078.070s, table=34, n_packets=0, n_bytes=0, priority=100,reg10=0/0x1,reg14=0x1,reg15=0x1,metadata=0x1 actions=drop
 cookie=0x0, duration=58078.070s, table=34, n_packets=0, n_bytes=0, priority=100,reg10=0/0x1,reg14=0x1,reg15=0x1,metadata=0x2 actions=drop
 cookie=0x0, duration=58077.788s, table=34, n_packets=0, n_bytes=0, priority=100,reg10=0/0x1,reg14=0x3,reg15=0x3,metadata=0x1 actions=drop
 cookie=0x0, duration=58077.787s, table=34, n_packets=0, n_bytes=0, priority=100,reg10=0/0x1,reg14=0x1,reg15=0x1,metadata=0x4 actions=drop
 cookie=0x0, duration=58077.304s, table=34, n_packets=6, n_bytes=252, priority=100,reg10=0/0x1,reg14=0x2,reg15=0x2,metadata=0x6 actions=drop
 cookie=0x0, duration=58077.304s, table=34, n_packets=0, n_bytes=0, priority=100,reg10=0/0x1,reg14=0x2,reg15=0x2,metadata=0x2 actions=drop
 cookie=0x0, duration=58077.304s, table=34, n_packets=0, n_bytes=0, priority=100,reg10=0/0x1,reg14=0x1,reg15=0x1,metadata=0x5 actions=drop
 cookie=0x0, duration=58077.304s, table=34, n_packets=507204, n_bytes=30915070, priority=100,reg10=0/0x1,reg14=0x1,reg15=0x1,metadata=0x6 actions=drop
 cookie=0x0, duration=58077.304s, table=34, n_packets=0, n_bytes=0, priority=100,reg10=0/0x1,reg14=0x2,reg15=0x2,metadata=0x5 actions=drop
 cookie=0x0, duration=58072.094s, table=34, n_packets=1, n_bytes=60, priority=100,reg10=0/0x1,reg14=0x4,reg15=0x4,metadata=0x3 actions=drop
 cookie=0x0, duration=58071.187s, table=34, n_packets=1, n_bytes=60, priority=100,reg10=0/0x1,reg14=0x3,reg15=0x3,metadata=0x3 actions=drop
 cookie=0x0, duration=58033.569s, table=34, n_packets=1, n_bytes=60, priority=100,reg10=0/0x1,reg14=0x5,reg15=0x5,metadata=0x3 actions=drop
 cookie=0x0, duration=58078.885s, table=34, n_packets=5794284, n_bytes=123448929879, priority=0 actions=load:0->NXM_NX_REG0[],load:0->NXM_NX_REG1[],load:0->NXM_NX_REG2[],load:0->NXM_NX_REG3[],load:0->NXM_NX_REG4[],load:0->NXM_NX_REG5[],load:0->NXM_NX_REG6[],load:0->NXM_NX_REG7[],load:0->NXM_NX_REG8[],load:0->NXM_NX_REG9[],resubmit(,40)
 cookie=0x294ae2ec, duration=58078.070s, table=40, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x3,nw_ttl=255,icmp_type=135,icmp_code=0 actions=resubmit(,41)
 cookie=0x2492e966, duration=58078.070s, table=40, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x2,nw_ttl=255,icmp_type=135,icmp_code=0 actions=resubmit(,41)
 cookie=0x2492e966, duration=58078.068s, table=40, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x2,nw_ttl=255,icmp_type=133,icmp_code=0 actions=resubmit(,41)
 cookie=0x294ae2ec, duration=58078.068s, table=40, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x3,nw_ttl=255,icmp_type=134,icmp_code=0 actions=resubmit(,41)
 cookie=0x2492e966, duration=58078.068s, table=40, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x2,nw_ttl=255,icmp_type=134,icmp_code=0 actions=resubmit(,41)
 cookie=0x2492e966, duration=58078.068s, table=40, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x2,nw_ttl=255,icmp_type=136,icmp_code=0 actions=resubmit(,41)
 cookie=0x294ae2ec, duration=58078.057s, table=40, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x3,nw_ttl=255,icmp_type=136,icmp_code=0 actions=resubmit(,41)
 cookie=0x294ae2ec, duration=58078.057s, table=40, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x3,nw_ttl=255,icmp_type=133,icmp_code=0 actions=resubmit(,41)
 cookie=0x6cd6283f, duration=58077.788s, table=40, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x4,nw_ttl=255,icmp_type=134,icmp_code=0 actions=resubmit(,41)
 cookie=0x6cd6283f, duration=58077.788s, table=40, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x4,nw_ttl=255,icmp_type=136,icmp_code=0 actions=resubmit(,41)
 cookie=0x6cd6283f, duration=58077.787s, table=40, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x4,nw_ttl=255,icmp_type=135,icmp_code=0 actions=resubmit(,41)
 cookie=0x6cd6283f, duration=58077.787s, table=40, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x4,nw_ttl=255,icmp_type=133,icmp_code=0 actions=resubmit(,41)
 cookie=0x571f5d27, duration=58077.303s, table=40, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x6,nw_ttl=255,icmp_type=135,icmp_code=0 actions=resubmit(,41)
 cookie=0x571f5d27, duration=58077.303s, table=40, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x6,nw_ttl=255,icmp_type=134,icmp_code=0 actions=resubmit(,41)
 cookie=0x571f5d27, duration=58077.302s, table=40, n_packets=611, n_bytes=42770, priority=110,icmp6,metadata=0x6,nw_ttl=255,icmp_type=133,icmp_code=0 actions=resubmit(,41)
 cookie=0x571f5d27, duration=58077.302s, table=40, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x6,nw_ttl=255,icmp_type=136,icmp_code=0 actions=resubmit(,41)
 cookie=0xebb55620, duration=58077.815s, table=40, n_packets=3019626, n_bytes=61650363579, priority=100,ip,metadata=0x3 actions=load:0x1->NXM_NX_XXREG0[96],resubmit(,41)
 cookie=0xebb55620, duration=58077.815s, table=40, n_packets=1, n_bytes=90, priority=100,ipv6,metadata=0x3 actions=load:0x1->NXM_NX_XXREG0[96],resubmit(,41)
 cookie=0x528061a3, duration=58077.304s, table=40, n_packets=0, n_bytes=0, priority=100,ip,metadata=0x4 actions=load:0x1->NXM_NX_XXREG0[96],resubmit(,41)
 cookie=0x528061a3, duration=58077.304s, table=40, n_packets=0, n_bytes=0, priority=100,ipv6,metadata=0x4 actions=load:0x1->NXM_NX_XXREG0[96],resubmit(,41)
 cookie=0x967e4fde, duration=58078.068s, table=40, n_packets=203090, n_bytes=70216033, priority=0,metadata=0x2 actions=resubmit(,41)
 cookie=0xef0ff294, duration=58077.848s, table=40, n_packets=1673231, n_bytes=61558280735, priority=0,metadata=0x1 actions=resubmit(,41)
 cookie=0x65f98daa, duration=58077.847s, table=40, n_packets=26, n_bytes=1524, priority=0,metadata=0x3 actions=resubmit(,41)
 cookie=0xce62f38d, duration=58077.782s, table=40, n_packets=0, n_bytes=0, priority=0,metadata=0x4 actions=resubmit(,41)
 cookie=0x13cc05ba, duration=58077.303s, table=40, n_packets=203090, n_bytes=70215841, priority=0,metadata=0x5 actions=resubmit(,41)
 cookie=0xb3bd782e, duration=58077.303s, table=40, n_packets=694609, n_bytes=99809307, priority=0,metadata=0x6 actions=resubmit(,41)
 cookie=0x69aa08f6, duration=58076.215s, table=41, n_packets=0, n_bytes=0, priority=110,tcp6,metadata=0x3,tcp_flags=rst actions=resubmit(,42)
 cookie=0x69aa08f6, duration=58076.215s, table=41, n_packets=6, n_bytes=324, priority=110,tcp,metadata=0x3,tcp_flags=rst actions=resubmit(,42)
 cookie=0x267e2635, duration=58059.636s, table=41, n_packets=0, n_bytes=0, priority=110,tcp,metadata=0x4,tcp_flags=rst actions=resubmit(,42)
 cookie=0x267e2635, duration=58059.636s, table=41, n_packets=0, n_bytes=0, priority=110,tcp6,metadata=0x4,tcp_flags=rst actions=resubmit(,42)
 cookie=0x69aa08f6, duration=58076.215s, table=41, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x3,nw_ttl=255,icmp_type=134,icmp_code=0 actions=resubmit(,42)
 cookie=0x69aa08f6, duration=58076.215s, table=41, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x3,nw_ttl=255,icmp_type=133,icmp_code=0 actions=resubmit(,42)
 cookie=0x69aa08f6, duration=58076.215s, table=41, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x3,nw_ttl=255,icmp_type=135,icmp_code=0 actions=resubmit(,42)
 cookie=0x69aa08f6, duration=58076.215s, table=41, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x3,nw_ttl=255,icmp_type=136,icmp_code=0 actions=resubmit(,42)
 cookie=0x267e2635, duration=58059.636s, table=41, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x4,nw_ttl=255,icmp_type=136,icmp_code=0 actions=resubmit(,42)
 cookie=0x267e2635, duration=58059.636s, table=41, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x4,nw_ttl=255,icmp_type=134,icmp_code=0 actions=resubmit(,42)
 cookie=0x267e2635, duration=58059.636s, table=41, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x4,nw_ttl=255,icmp_type=135,icmp_code=0 actions=resubmit(,42)
 cookie=0x267e2635, duration=58059.636s, table=41, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x4,nw_ttl=255,icmp_type=133,icmp_code=0 actions=resubmit(,42)
 cookie=0x895a6098, duration=58076.215s, table=41, n_packets=0, n_bytes=0, priority=110,ipv6,reg15=0x1,metadata=0x3 actions=resubmit(,42)
 cookie=0x895a6098, duration=58076.215s, table=41, n_packets=1558402, n_bytes=61494495267, priority=110,ip,reg15=0x1,metadata=0x3 actions=resubmit(,42)
 cookie=0x4ebd9137, duration=58059.636s, table=41, n_packets=0, n_bytes=0, priority=110,ipv6,reg15=0x1,metadata=0x4 actions=resubmit(,42)
 cookie=0x4ebd9137, duration=58059.636s, table=41, n_packets=0, n_bytes=0, priority=110,ip,reg15=0x1,metadata=0x4 actions=resubmit(,42)
 cookie=0x69aa08f6, duration=58076.215s, table=41, n_packets=0, n_bytes=0, priority=110,icmp,metadata=0x3,icmp_type=3 actions=resubmit(,42)
 cookie=0x69aa08f6, duration=58076.215s, table=41, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x3,icmp_type=1 actions=resubmit(,42)
 cookie=0x267e2635, duration=58059.636s, table=41, n_packets=0, n_bytes=0, priority=110,icmp6,metadata=0x4,icmp_type=1 actions=resubmit(,42)
 cookie=0x267e2635, duration=58059.636s, table=41, n_packets=0, n_bytes=0, priority=110,icmp,metadata=0x4,icmp_type=3 actions=resubmit(,42)
 cookie=0x13f2f6e6, duration=58077.303s, table=41, n_packets=0, n_bytes=0, priority=100,ip,reg10=0x8/0x8,metadata=0x5 actions=ct(commit,table=42,zone=NXM_NX_REG12[0..15],nat(src=100.64.1.2))
 cookie=0x13f2f6e6, duration=58077.303s, table=41, n_packets=0, n_bytes=0, priority=100,ipv6,reg10=0x8/0x8,metadata=0x5 actions=ct(commit,table=42,zone=NXM_NX_REG12[0..15],nat(src=100.64.1.2))
 cookie=0xc3398e6e, duration=58076.215s, table=41, n_packets=0, n_bytes=0, priority=100,ipv6,metadata=0x3 actions=load:0x1->NXM_NX_XXREG0[96],resubmit(,42)
 cookie=0xc3398e6e, duration=58076.215s, table=41, n_packets=1461218, n_bytes=155867988, priority=100,ip,metadata=0x3 actions=load:0x1->NXM_NX_XXREG0[96],resubmit(,42)
 cookie=0x2e228c9f, duration=58059.636s, table=41, n_packets=0, n_bytes=0, priority=100,ip,metadata=0x4 actions=load:0x1->NXM_NX_XXREG0[96],resubmit(,42)
 cookie=0x2e228c9f, duration=58059.636s, table=41, n_packets=0, n_bytes=0, priority=100,ipv6,metadata=0x4 actions=load:0x1->NXM_NX_XXREG0[96],resubmit(,42)
 cookie=0x7e48bf51, duration=58077.303s, table=41, n_packets=88268, n_bytes=6409291, priority=17,ip,metadata=0x5,nw_src=10.244.0.0/16 actions=ct(commit,table=42,zone=NXM_NX_REG12[0..15],nat(src=10.0.2.19))
 cookie=0x31faf744, duration=58078.058s, table=41, n_packets=1673231, n_bytes=61558280735, priority=0,metadata=0x1 actions=resubmit(,42)
 cookie=0x2203e72, duration=58078.057s, table=41, n_packets=27, n_bytes=1614, priority=0,metadata=0x3 actions=resubmit(,42)
 cookie=0xa46750ae, duration=58077.847s, table=41, n_packets=203090, n_bytes=70216033, priority=0,metadata=0x2 actions=resubmit(,42)
 cookie=0x652a0767, duration=58077.787s, table=41, n_packets=0, n_bytes=0, priority=0,metadata=0x4 actions=resubmit(,42)
 cookie=0x7abec1f9, duration=58077.303s, table=41, n_packets=695220, n_bytes=99852077, priority=0,metadata=0x6 actions=resubmit(,42)
 cookie=0x690e191b, duration=58077.302s, table=41, n_packets=114822, n_bytes=63806550, priority=0,metadata=0x5 actions=resubmit(,42)
 cookie=0xf99cf495, duration=58078.070s, table=42, n_packets=3019626, n_bytes=61650363579, priority=100,ip,reg0=0x1/0x1,metadata=0x3 actions=ct(table=43,zone=NXM_NX_REG13[0..15])
 cookie=0xf99cf495, duration=58078.068s, table=42, n_packets=1, n_bytes=90, priority=100,ipv6,reg0=0x1/0x1,metadata=0x3 actions=ct(table=43,zone=NXM_NX_REG13[0..15])
 cookie=0xaf16916f, duration=58078.068s, table=42, n_packets=0, n_bytes=0, priority=100,ip,reg0=0x1/0x1,metadata=0x2 actions=ct(table=43,zone=NXM_NX_REG13[0..15])
 cookie=0xaf16916f, duration=58078.057s, table=42, n_packets=0, n_bytes=0, priority=100,ipv6,reg0=0x1/0x1,metadata=0x2 actions=ct(table=43,zone=NXM_NX_REG13[0..15])
 cookie=0xdedb2e39, duration=58077.787s, table=42, n_packets=0, n_bytes=0, priority=100,ipv6,reg0=0x1/0x1,metadata=0x4 actions=ct(table=43,zone=NXM_NX_REG13[0..15])
 cookie=0xdedb2e39, duration=58077.782s, table=42, n_packets=0, n_bytes=0, priority=100,ip,reg0=0x1/0x1,metadata=0x4 actions=ct(table=43,zone=NXM_NX_REG13[0..15])
 cookie=0x256a2bf2, duration=58077.303s, table=42, n_packets=0, n_bytes=0, priority=100,ip,reg0=0x1/0x1,metadata=0x6 actions=ct(table=43,zone=NXM_NX_REG13[0..15])
 cookie=0x256a2bf2, duration=58077.302s, table=42, n_packets=0, n_bytes=0, priority=100,ipv6,reg0=0x1/0x1,metadata=0x6 actions=ct(table=43,zone=NXM_NX_REG13[0..15])
 cookie=0x3eea8f4, duration=58078.068s, table=42, n_packets=26, n_bytes=1524, priority=0,metadata=0x3 actions=resubmit(,43)
 cookie=0x537708b6, duration=58078.058s, table=42, n_packets=1673231, n_bytes=61558280735, priority=0,metadata=0x1 actions=resubmit(,43)
 cookie=0x567cb6d2, duration=58077.847s, table=42, n_packets=203090, n_bytes=70216033, priority=0,metadata=0x2 actions=resubmit(,43)
 cookie=0x425ff8f0, duration=58077.787s, table=42, n_packets=0, n_bytes=0, priority=0,metadata=0x4 actions=resubmit(,43)
 cookie=0x849f28f6, duration=58077.303s, table=42, n_packets=695220, n_bytes=99852077, priority=0,metadata=0x6 actions=resubmit(,43)
 cookie=0xb19b16d, duration=58077.303s, table=42, n_packets=203090, n_bytes=70215841, priority=0,metadata=0x5 actions=resubmit(,43)
 cookie=0x50b8c3f5, duration=58077.815s, table=43, n_packets=3008002, n_bytes=61649546463, priority=65535,ct_state=-new+est-rel-inv+trk,metadata=0x3 actions=load:0x1->NXM_NX_XXREG0[98],resubmit(,44)
 cookie=0xed815a2e, duration=58077.304s, table=43, n_packets=0, n_bytes=0, priority=65535,ct_state=-new+est-rel-inv+trk,metadata=0x4 actions=load:0x1->NXM_NX_XXREG0[98],resubmit(,44)
 cookie=0x758d0661, duration=58078.070s, table=43, n_packets=88274, n_bytes=6409735, priority=100,reg15=0x1,metadata=0x1 actions=resubmit(,64)
 cookie=0xe80151b9, duration=58078.068s, table=43, n_packets=114824, n_bytes=63806778, priority=100,reg15=0x2,metadata=0x1 actions=resubmit(,64)
 cookie=0xd18a0b90, duration=58077.816s, table=43, n_packets=1470133, n_bytes=61488064222, priority=100,reg15=0x3,metadata=0x1 actions=resubmit(,64)
 cookie=0x5d96cb9f, duration=58077.303s, table=43, n_packets=114816, n_bytes=63806298, priority=100,reg15=0x1,metadata=0x5 actions=resubmit(,64)
 cookie=0xf5322026, duration=58077.302s, table=43, n_packets=88274, n_bytes=6409543, priority=100,reg15=0x2,metadata=0x5 actions=resubmit(,64)
 cookie=0x54149cb7, duration=58078.070s, table=43, n_packets=11648, n_bytes=861668, priority=0,metadata=0x3 actions=resubmit(,44)
 cookie=0x34e05559, duration=58077.847s, table=43, n_packets=203090, n_bytes=70216033, priority=0,metadata=0x2 actions=resubmit(,44)
 cookie=0x5f30742b, duration=58077.787s, table=43, n_packets=0, n_bytes=0, priority=0,metadata=0x4 actions=resubmit(,44)
 cookie=0xde2ecfbf, duration=58077.304s, table=43, n_packets=695220, n_bytes=99852077, priority=0,metadata=0x6 actions=resubmit(,44)
 cookie=0x8ceff86a, duration=58076.215s, table=44, n_packets=0, n_bytes=0, priority=65535,icmp6,metadata=0x3,nw_ttl=255,icmp_type=135,icmp_code=0 actions=resubmit(,45)
 cookie=0x8ceff86a, duration=58076.215s, table=44, n_packets=0, n_bytes=0, priority=65535,icmp6,metadata=0x3,nw_ttl=255,icmp_type=136,icmp_code=0 actions=resubmit(,45)
 cookie=0x858bed9d, duration=58059.636s, table=44, n_packets=0, n_bytes=0, priority=65535,icmp6,metadata=0x4,nw_ttl=255,icmp_type=136,icmp_code=0 actions=resubmit(,45)
 cookie=0x858bed9d, duration=58059.636s, table=44, n_packets=0, n_bytes=0, priority=65535,icmp6,metadata=0x4,nw_ttl=255,icmp_type=135,icmp_code=0 actions=resubmit(,45)
 cookie=0xc1576414, duration=58076.215s, table=44, n_packets=0, n_bytes=0, priority=65535,ct_state=+est+rpl+trk,ct_label=0x1/0x1,metadata=0x3 actions=drop
 cookie=0x82c8ac6, duration=58059.636s, table=44, n_packets=0, n_bytes=0, priority=65535,ct_state=+est+rpl+trk,ct_label=0x1/0x1,metadata=0x4 actions=drop
 cookie=0xa78e3b19, duration=58076.215s, table=44, n_packets=1, n_bytes=102, priority=65535,ct_state=-new-est+rel-inv+trk,ct_label=0/0x1,metadata=0x3 actions=resubmit(,45)
 cookie=0xf2b90f0b, duration=58059.637s, table=44, n_packets=0, n_bytes=0, priority=65535,ct_state=-new-est+rel-inv+trk,ct_label=0/0x1,metadata=0x4 actions=resubmit(,45)
 cookie=0xbf47a6ae, duration=58076.215s, table=44, n_packets=1403128, n_bytes=150572004, priority=65535,ct_state=-new+est-rel+rpl-inv+trk,ct_label=0/0x1,metadata=0x3 actions=resubmit(,45)
 cookie=0xc6074744, duration=58059.636s, table=44, n_packets=0, n_bytes=0, priority=65535,ct_state=-new+est-rel+rpl-inv+trk,ct_label=0/0x1,metadata=0x4 actions=resubmit(,45)
 cookie=0xc1576414, duration=58076.215s, table=44, n_packets=0, n_bytes=0, priority=65535,ct_state=+inv+trk,metadata=0x3 actions=drop
 cookie=0x82c8ac6, duration=58059.636s, table=44, n_packets=0, n_bytes=0, priority=65535,ct_state=+inv+trk,metadata=0x4 actions=drop
 cookie=0xa46afc1b, duration=58076.215s, table=44, n_packets=11602, n_bytes=858548, priority=2001,ct_state=+new-est+trk,ip,metadata=0x3,nw_src=10.244.1.2 actions=load:0x1->NXM_NX_XXREG0[97],resubmit(,45)
 cookie=0xe56d720f, duration=58059.636s, table=44, n_packets=0, n_bytes=0, priority=2001,ct_state=+new-est+trk,ip,metadata=0x4,nw_src=10.244.2.2 actions=load:0x1->NXM_NX_XXREG0[97],resubmit(,45)
 cookie=0x205c683a, duration=58076.215s, table=44, n_packets=46483, n_bytes=4437070, priority=2001,ct_state=-new+est-rpl+trk,ct_label=0/0x1,ip,metadata=0x3,nw_src=10.244.1.2 actions=resubmit(,45)
 cookie=0xa46afc1b, duration=58076.215s, table=44, n_packets=0, n_bytes=0, priority=2001,ct_state=-new+est-rpl+trk,ct_label=0x1/0x1,ip,metadata=0x3,nw_src=10.244.1.2 actions=load:0x1->NXM_NX_XXREG0[97],resubmit(,45)
 cookie=0xe56d720f, duration=58059.636s, table=44, n_packets=0, n_bytes=0, priority=2001,ct_state=-new+est-rpl+trk,ct_label=0x1/0x1,ip,metadata=0x4,nw_src=10.244.2.2 actions=load:0x1->NXM_NX_XXREG0[97],resubmit(,45)
 cookie=0x6d363e9b, duration=58059.636s, table=44, n_packets=0, n_bytes=0, priority=2001,ct_state=-new+est-rpl+trk,ct_label=0/0x1,ip,metadata=0x4,nw_src=10.244.2.2 actions=resubmit(,45)
 cookie=0x38dc09c8, duration=58076.215s, table=44, n_packets=0, n_bytes=0, priority=1,ct_state=+est+trk,ct_label=0x1/0x1,ipv6,metadata=0x3 actions=load:0x1->NXM_NX_XXREG0[97],resubmit(,45)
 cookie=0x38dc09c8, duration=58076.215s, table=44, n_packets=0, n_bytes=0, priority=1,ct_state=+est+trk,ct_label=0x1/0x1,ip,metadata=0x3 actions=load:0x1->NXM_NX_XXREG0[97],resubmit(,45)
 cookie=0x4c1c690c, duration=58059.636s, table=44, n_packets=0, n_bytes=0, priority=1,ct_state=+est+trk,ct_label=0x1/0x1,ip,metadata=0x4 actions=load:0x1->NXM_NX_XXREG0[97],resubmit(,45)
 cookie=0x4c1c690c, duration=58059.636s, table=44, n_packets=0, n_bytes=0, priority=1,ct_state=+est+trk,ct_label=0x1/0x1,ipv6,metadata=0x4 actions=load:0x1->NXM_NX_XXREG0[97],resubmit(,45)
 cookie=0x38dc09c8, duration=58076.215s, table=44, n_packets=0, n_bytes=0, priority=1,ct_state=-est+trk,ipv6,metadata=0x3 actions=load:0x1->NXM_NX_XXREG0[97],resubmit(,45)
 cookie=0x38dc09c8, duration=58076.215s, table=44, n_packets=18, n_bytes=1404, priority=1,ct_state=-est+trk,ip,metadata=0x3 actions=load:0x1->NXM_NX_XXREG0[97],resubmit(,45)
 cookie=0x4c1c690c, duration=58059.636s, table=44, n_packets=0, n_bytes=0, priority=1,ct_state=-est+trk,ipv6,metadata=0x4 actions=load:0x1->NXM_NX_XXREG0[97],resubmit(,45)
 cookie=0x4c1c690c, duration=58059.636s, table=44, n_packets=0, n_bytes=0, priority=1,ct_state=-est+trk,ip,metadata=0x4 actions=load:0x1->NXM_NX_XXREG0[97],resubmit(,45)
 cookie=0xaa05e17, duration=58078.058s, table=44, n_packets=1558418, n_bytes=61494539003, priority=0,metadata=0x3 actions=resubmit(,45)
 cookie=0xf2f12e2d, duration=58077.848s, table=44, n_packets=203090, n_bytes=70216033, priority=0,metadata=0x2 actions=resubmit(,45)
 cookie=0xecc7d19b, duration=58077.788s, table=44, n_packets=0, n_bytes=0, priority=0,metadata=0x4 actions=resubmit(,45)
 cookie=0xf067ee46, duration=58077.303s, table=44, n_packets=695220, n_bytes=99852077, priority=0,metadata=0x6 actions=resubmit(,45)
 cookie=0x52cb703, duration=58077.848s, table=45, n_packets=203090, n_bytes=70216033, priority=0,metadata=0x2 actions=resubmit(,46)
 cookie=0x4cf3d586, duration=58077.847s, table=45, n_packets=3019650, n_bytes=61650408131, priority=0,metadata=0x3 actions=resubmit(,46)
 cookie=0xb1ba4cd8, duration=58077.787s, table=45, n_packets=0, n_bytes=0, priority=0,metadata=0x4 actions=resubmit(,46)
 cookie=0x3f783fdd, duration=58077.302s, table=45, n_packets=695220, n_bytes=99852077, priority=0,metadata=0x6 actions=resubmit(,46)
 cookie=0x8439229d, duration=58078.070s, table=46, n_packets=3019650, n_bytes=61650408131, priority=0,metadata=0x3 actions=resubmit(,47)
 cookie=0x133c852c, duration=58077.848s, table=46, n_packets=203090, n_bytes=70216033, priority=0,metadata=0x2 actions=resubmit(,47)
 cookie=0xa79bb3f, duration=58077.787s, table=46, n_packets=0, n_bytes=0, priority=0,metadata=0x4 actions=resubmit(,47)
 cookie=0xfe72c608, duration=58077.303s, table=46, n_packets=695220, n_bytes=99852077, priority=0,metadata=0x6 actions=resubmit(,47)
 cookie=0x22ecdd68, duration=58078.070s, table=47, n_packets=0, n_bytes=0, priority=100,ipv6,reg0=0x4/0x4,metadata=0x3 actions=ct(table=48,zone=NXM_NX_REG13[0..15],nat)
 cookie=0x22ecdd68, duration=58078.068s, table=47, n_packets=3008002, n_bytes=61649546463, priority=100,ip,reg0=0x4/0x4,metadata=0x3 actions=ct(table=48,zone=NXM_NX_REG13[0..15],nat)
 cookie=0xf6fa364f, duration=58078.058s, table=47, n_packets=0, n_bytes=0, priority=100,ip,reg0=0x4/0x4,metadata=0x2 actions=ct(table=48,zone=NXM_NX_REG13[0..15],nat)
 cookie=0xf6fa364f, duration=58077.848s, table=47, n_packets=0, n_bytes=0, priority=100,ipv6,reg0=0x4/0x4,metadata=0x2 actions=ct(table=48,zone=NXM_NX_REG13[0..15],nat)
 cookie=0x43ad6611, duration=58077.788s, table=47, n_packets=0, n_bytes=0, priority=100,ip,reg0=0x4/0x4,metadata=0x4 actions=ct(table=48,zone=NXM_NX_REG13[0..15],nat)
 cookie=0x43ad6611, duration=58077.782s, table=47, n_packets=0, n_bytes=0, priority=100,ipv6,reg0=0x4/0x4,metadata=0x4 actions=ct(table=48,zone=NXM_NX_REG13[0..15],nat)
 cookie=0x349ae249, duration=58077.303s, table=47, n_packets=0, n_bytes=0, priority=100,ipv6,reg0=0x4/0x4,metadata=0x6 actions=ct(table=48,zone=NXM_NX_REG13[0..15],nat)
 cookie=0x349ae249, duration=58077.302s, table=47, n_packets=0, n_bytes=0, priority=100,ip,reg0=0x4/0x4,metadata=0x6 actions=ct(table=48,zone=NXM_NX_REG13[0..15],nat)
 cookie=0x74aec291, duration=58078.068s, table=47, n_packets=0, n_bytes=0, priority=100,ip,reg0=0x2/0x2,metadata=0x2 actions=ct(commit,zone=NXM_NX_REG13[0..15],exec(load:0->NXM_NX_CT_LABEL[0])),resubmit(,48)
 cookie=0xc53d745d, duration=58078.058s, table=47, n_packets=11620, n_bytes=859952, priority=100,ip,reg0=0x2/0x2,metadata=0x3 actions=ct(commit,zone=NXM_NX_REG13[0..15],exec(load:0->NXM_NX_CT_LABEL[0])),resubmit(,48)
 cookie=0xc53d745d, duration=58077.847s, table=47, n_packets=0, n_bytes=0, priority=100,ipv6,reg0=0x2/0x2,metadata=0x3 actions=ct(commit,zone=NXM_NX_REG13[0..15],exec(load:0->NXM_NX_CT_LABEL[0])),resubmit(,48)
 cookie=0x74aec291, duration=58077.847s, table=47, n_packets=0, n_bytes=0, priority=100,ipv6,reg0=0x2/0x2,metadata=0x2 actions=ct(commit,zone=NXM_NX_REG13[0..15],exec(load:0->NXM_NX_CT_LABEL[0])),resubmit(,48)
 cookie=0xf80aa719, duration=58077.788s, table=47, n_packets=0, n_bytes=0, priority=100,ip,reg0=0x2/0x2,metadata=0x4 actions=ct(commit,zone=NXM_NX_REG13[0..15],exec(load:0->NXM_NX_CT_LABEL[0])),resubmit(,48)
 cookie=0xf80aa719, duration=58077.787s, table=47, n_packets=0, n_bytes=0, priority=100,ipv6,reg0=0x2/0x2,metadata=0x4 actions=ct(commit,zone=NXM_NX_REG13[0..15],exec(load:0->NXM_NX_CT_LABEL[0])),resubmit(,48)
 cookie=0x88902d6a, duration=58077.303s, table=47, n_packets=0, n_bytes=0, priority=100,ip,reg0=0x2/0x2,metadata=0x6 actions=ct(commit,zone=NXM_NX_REG13[0..15],exec(load:0->NXM_NX_CT_LABEL[0])),resubmit(,48)
 cookie=0x88902d6a, duration=58077.302s, table=47, n_packets=0, n_bytes=0, priority=100,ipv6,reg0=0x2/0x2,metadata=0x6 actions=ct(commit,zone=NXM_NX_REG13[0..15],exec(load:0->NXM_NX_CT_LABEL[0])),resubmit(,48)
 cookie=0x6a00b8d4, duration=58078.070s, table=47, n_packets=28, n_bytes=1716, priority=0,metadata=0x3 actions=resubmit(,48)
 cookie=0x8759c193, duration=58077.848s, table=47, n_packets=203090, n_bytes=70216033, priority=0,metadata=0x2 actions=resubmit(,48)
 cookie=0x4150ff82, duration=58077.787s, table=47, n_packets=0, n_bytes=0, priority=0,metadata=0x4 actions=resubmit(,48)
 cookie=0x3f9e7364, duration=58077.302s, table=47, n_packets=695220, n_bytes=99852077, priority=0,metadata=0x6 actions=resubmit(,48)
 cookie=0xd772131d, duration=58078.068s, table=48, n_packets=3019646, n_bytes=61650343163, priority=0,metadata=0x3 actions=resubmit(,49)
 cookie=0xe262d02c, duration=58077.847s, table=48, n_packets=203090, n_bytes=70216033, priority=0,metadata=0x2 actions=resubmit(,49)
 cookie=0x46658059, duration=58077.787s, table=48, n_packets=0, n_bytes=0, priority=0,metadata=0x4 actions=resubmit(,49)
 cookie=0x4f5e79ba, duration=58077.303s, table=48, n_packets=695220, n_bytes=99852077, priority=0,metadata=0x6 actions=resubmit(,49)
 cookie=0x3fd5d142, duration=58078.068s, table=49, n_packets=10, n_bytes=630, priority=100,metadata=0x3,dl_dst=01:00:00:00:00:00/01:00:00:00:00:00 actions=resubmit(,64)
 cookie=0x497b5acb, duration=58077.847s, table=49, n_packets=0, n_bytes=0, priority=100,metadata=0x2,dl_dst=01:00:00:00:00:00/01:00:00:00:00:00 actions=resubmit(,64)
 cookie=0xff0920d8, duration=58077.787s, table=49, n_packets=0, n_bytes=0, priority=100,metadata=0x4,dl_dst=01:00:00:00:00:00/01:00:00:00:00:00 actions=resubmit(,64)
 cookie=0xe383aa0b, duration=58077.303s, table=49, n_packets=492130, n_bytes=29636128, priority=100,metadata=0x6,dl_dst=01:00:00:00:00:00/01:00:00:00:00:00 actions=resubmit(,64)
 cookie=0x12cd9430, duration=58078.058s, table=49, n_packets=58009, n_bytes=5510836, priority=50,reg15=0x2,metadata=0x3 actions=resubmit(,64)
 cookie=0x5abdc498, duration=58077.847s, table=49, n_packets=1558412, n_bytes=61494474257, priority=50,reg15=0x1,metadata=0x3 actions=resubmit(,64)
 cookie=0xaac915ca, duration=58077.816s, table=49, n_packets=114816, n_bytes=63806298, priority=50,reg15=0x1,metadata=0x2 actions=resubmit(,64)
 cookie=0x33e60c30, duration=58077.782s, table=49, n_packets=0, n_bytes=0, priority=50,reg15=0x1,metadata=0x4 actions=resubmit(,64)
 cookie=0x45397c68, duration=58077.304s, table=49, n_packets=88274, n_bytes=6409735, priority=50,reg15=0x2,metadata=0x2 actions=resubmit(,64)
 cookie=0x5e4577df, duration=58077.303s, table=49, n_packets=114822, n_bytes=63806658, priority=50,reg15=0x2,metadata=0x6 actions=resubmit(,64)
 cookie=0xe9772d9c, duration=58077.302s, table=49, n_packets=88268, n_bytes=6409291, priority=50,reg15=0x1,metadata=0x6 actions=resubmit(,64)
 cookie=0x7cc6c026, duration=58072.094s, table=49, n_packets=86456, n_bytes=34551766, priority=50,reg15=0x4,metadata=0x3 actions=resubmit(,64)
 cookie=0x3ea85f75, duration=58071.187s, table=49, n_packets=86329, n_bytes=34538626, priority=50,reg15=0x3,metadata=0x3 actions=resubmit(,64)
 cookie=0xee390356, duration=58033.569s, table=49, n_packets=1230430, n_bytes=81267048, priority=50,reg15=0x5,metadata=0x3 actions=resubmit(,64)
 cookie=0x0, duration=58078.070s, table=64, n_packets=88274, n_bytes=6409735, priority=100,reg10=0x1/0x1,reg15=0x1,metadata=0x1 actions=push:NXM_OF_IN_PORT[],load:0->NXM_OF_IN_PORT[],resubmit(,65),pop:NXM_OF_IN_PORT[]
 cookie=0x0, duration=58078.070s, table=64, n_packets=0, n_bytes=0, priority=100,reg10=0x1/0x1,reg15=0x1,metadata=0x3 actions=push:NXM_OF_IN_PORT[],load:0->NXM_OF_IN_PORT[],resubmit(,65),pop:NXM_OF_IN_PORT[]
 cookie=0x0, duration=58078.070s, table=64, n_packets=0, n_bytes=0, priority=100,reg10=0x1/0x1,reg15=0x1,metadata=0x2 actions=push:NXM_OF_IN_PORT[],load:0->NXM_OF_IN_PORT[],resubmit(,65),pop:NXM_OF_IN_PORT[]
 cookie=0x0, duration=58078.070s, table=64, n_packets=2, n_bytes=84, priority=100,reg10=0x1/0x1,reg15=0x2,metadata=0x3 actions=push:NXM_OF_IN_PORT[],load:0->NXM_OF_IN_PORT[],resubmit(,65),pop:NXM_OF_IN_PORT[]
 cookie=0x0, duration=58078.070s, table=64, n_packets=114824, n_bytes=63806778, priority=100,reg10=0x1/0x1,reg15=0x2,metadata=0x1 actions=push:NXM_OF_IN_PORT[],load:0->NXM_OF_IN_PORT[],resubmit(,65),pop:NXM_OF_IN_PORT[]
 cookie=0x0, duration=58077.787s, table=64, n_packets=0, n_bytes=0, priority=100,reg10=0x1/0x1,reg15=0x1,metadata=0x4 actions=push:NXM_OF_IN_PORT[],load:0->NXM_OF_IN_PORT[],resubmit(,65),pop:NXM_OF_IN_PORT[]
 cookie=0x0, duration=58077.782s, table=64, n_packets=1470133, n_bytes=61488064222, priority=100,reg10=0x1/0x1,reg15=0x3,metadata=0x1 actions=push:NXM_OF_IN_PORT[],load:0->NXM_OF_IN_PORT[],resubmit(,65),pop:NXM_OF_IN_PORT[]
 cookie=0x0, duration=58077.304s, table=64, n_packets=0, n_bytes=0, priority=100,reg10=0x1/0x1,reg15=0x2,metadata=0x6 actions=push:NXM_OF_IN_PORT[],load:0->NXM_OF_IN_PORT[],resubmit(,65),pop:NXM_OF_IN_PORT[]
 cookie=0x0, duration=58077.304s, table=64, n_packets=0, n_bytes=0, priority=100,reg10=0x1/0x1,reg15=0x1,metadata=0x6 actions=push:NXM_OF_IN_PORT[],load:0->NXM_OF_IN_PORT[],resubmit(,65),pop:NXM_OF_IN_PORT[]
 cookie=0x0, duration=58077.304s, table=64, n_packets=88274, n_bytes=6409543, priority=100,reg10=0x1/0x1,reg15=0x2,metadata=0x5 actions=push:NXM_OF_IN_PORT[],load:0->NXM_OF_IN_PORT[],resubmit(,65),pop:NXM_OF_IN_PORT[]
 cookie=0x0, duration=58077.304s, table=64, n_packets=0, n_bytes=0, priority=100,reg10=0x1/0x1,reg15=0x2,metadata=0x2 actions=push:NXM_OF_IN_PORT[],load:0->NXM_OF_IN_PORT[],resubmit(,65),pop:NXM_OF_IN_PORT[]
 cookie=0x0, duration=58077.304s, table=64, n_packets=114816, n_bytes=63806298, priority=100,reg10=0x1/0x1,reg15=0x1,metadata=0x5 actions=push:NXM_OF_IN_PORT[],load:0->NXM_OF_IN_PORT[],resubmit(,65),pop:NXM_OF_IN_PORT[]
 cookie=0x0, duration=58072.094s, table=64, n_packets=1, n_bytes=60, priority=100,reg10=0x1/0x1,reg15=0x4,metadata=0x3 actions=push:NXM_OF_IN_PORT[],load:0->NXM_OF_IN_PORT[],resubmit(,65),pop:NXM_OF_IN_PORT[]
 cookie=0x0, duration=58071.187s, table=64, n_packets=1, n_bytes=60, priority=100,reg10=0x1/0x1,reg15=0x3,metadata=0x3 actions=push:NXM_OF_IN_PORT[],load:0->NXM_OF_IN_PORT[],resubmit(,65),pop:NXM_OF_IN_PORT[]
 cookie=0x0, duration=58033.569s, table=64, n_packets=0, n_bytes=0, priority=100,reg10=0x1/0x1,reg15=0x5,metadata=0x3 actions=push:NXM_OF_IN_PORT[],load:0->NXM_OF_IN_PORT[],resubmit(,65),pop:NXM_OF_IN_PORT[]
 cookie=0x0, duration=58078.885s, table=64, n_packets=3917952, n_bytes=61820411069, priority=0 actions=resubmit(,65)
 cookie=0x0, duration=58078.070s, table=65, n_packets=88274, n_bytes=6409735, priority=100,reg15=0x1,metadata=0x1 actions=clone(ct_clear,load:0->NXM_NX_REG11[],load:0->NXM_NX_REG12[],load:0->NXM_NX_REG13[],load:0x2->NXM_NX_REG11[],load:0x3->NXM_NX_REG12[],load:0x2->OXM_OF_METADATA[],load:0x1->NXM_NX_REG14[],load:0->NXM_NX_REG10[],load:0->NXM_NX_REG15[],load:0->NXM_NX_REG0[],load:0->NXM_NX_REG1[],load:0->NXM_NX_REG2[],load:0->NXM_NX_REG3[],load:0->NXM_NX_REG4[],load:0->NXM_NX_REG5[],load:0->NXM_NX_REG6[],load:0->NXM_NX_REG7[],load:0->NXM_NX_REG8[],load:0->NXM_NX_REG9[],load:0->NXM_OF_IN_PORT[],resubmit(,8))
 cookie=0x0, duration=58078.070s, table=65, n_packets=58012, n_bytes=5511016, priority=100,reg15=0x2,metadata=0x3 actions=output:"k8s-nd-sjc3a-c1"
 cookie=0x0, duration=58078.070s, table=65, n_packets=114824, n_bytes=63806778, priority=100,reg15=0x2,metadata=0x1 actions=clone(ct_clear,load:0->NXM_NX_REG11[],load:0->NXM_NX_REG12[],load:0->NXM_NX_REG13[],load:0x5->NXM_NX_REG11[],load:0x6->NXM_NX_REG12[],load:0x3->OXM_OF_METADATA[],load:0x1->NXM_NX_REG14[],load:0->NXM_NX_REG10[],load:0->NXM_NX_REG15[],load:0->NXM_NX_REG0[],load:0->NXM_NX_REG1[],load:0->NXM_NX_REG2[],load:0->NXM_NX_REG3[],load:0->NXM_NX_REG4[],load:0->NXM_NX_REG5[],load:0->NXM_NX_REG6[],load:0->NXM_NX_REG7[],load:0->NXM_NX_REG8[],load:0->NXM_NX_REG9[],load:0->NXM_OF_IN_PORT[],resubmit(,8))
 cookie=0x0, duration=58078.070s, table=65, n_packets=1558416, n_bytes=61494474527, priority=100,reg15=0x1,metadata=0x3 actions=clone(ct_clear,load:0->NXM_NX_REG11[],load:0->NXM_NX_REG12[],load:0->NXM_NX_REG13[],load:0x7->NXM_NX_REG11[],load:0x4->NXM_NX_REG12[],load:0x1->OXM_OF_METADATA[],load:0x2->NXM_NX_REG14[],load:0->NXM_NX_REG10[],load:0->NXM_NX_REG15[],load:0->NXM_NX_REG0[],load:0->NXM_NX_REG1[],load:0->NXM_NX_REG2[],load:0->NXM_NX_REG3[],load:0->NXM_NX_REG4[],load:0->NXM_NX_REG5[],load:0->NXM_NX_REG6[],load:0->NXM_NX_REG7[],load:0->NXM_NX_REG8[],load:0->NXM_NX_REG9[],load:0->NXM_OF_IN_PORT[],resubmit(,8))
 cookie=0x0, duration=58078.070s, table=65, n_packets=114816, n_bytes=63806298, priority=100,reg15=0x1,metadata=0x2 actions=clone(ct_clear,load:0->NXM_NX_REG11[],load:0->NXM_NX_REG12[],load:0->NXM_NX_REG13[],load:0x7->NXM_NX_REG11[],load:0x4->NXM_NX_REG12[],load:0x1->OXM_OF_METADATA[],load:0x1->NXM_NX_REG14[],load:0->NXM_NX_REG10[],load:0->NXM_NX_REG15[],load:0->NXM_NX_REG0[],load:0->NXM_NX_REG1[],load:0->NXM_NX_REG2[],load:0->NXM_NX_REG3[],load:0->NXM_NX_REG4[],load:0->NXM_NX_REG5[],load:0->NXM_NX_REG6[],load:0->NXM_NX_REG7[],load:0->NXM_NX_REG8[],load:0->NXM_NX_REG9[],load:0->NXM_OF_IN_PORT[],resubmit(,8))
 cookie=0x0, duration=58077.787s, table=65, n_packets=1470133, n_bytes=61488064222, priority=100,reg15=0x3,metadata=0x1 actions=clone(ct_clear,load:0->NXM_NX_REG11[],load:0->NXM_NX_REG12[],load:0->NXM_NX_REG13[],load:0x9->NXM_NX_REG11[],load:0x8->NXM_NX_REG12[],load:0x4->OXM_OF_METADATA[],load:0x1->NXM_NX_REG14[],load:0->NXM_NX_REG10[],load:0->NXM_NX_REG15[],load:0->NXM_NX_REG0[],load:0->NXM_NX_REG1[],load:0->NXM_NX_REG2[],load:0->NXM_NX_REG3[],load:0->NXM_NX_REG4[],load:0->NXM_NX_REG5[],load:0->NXM_NX_REG6[],load:0->NXM_NX_REG7[],load:0->NXM_NX_REG8[],load:0->NXM_NX_REG9[],load:0->NXM_OF_IN_PORT[],resubmit(,8))
 cookie=0x0, duration=58077.782s, table=65, n_packets=0, n_bytes=0, priority=100,reg15=0x1,metadata=0x4 actions=clone(ct_clear,load:0->NXM_NX_REG11[],load:0->NXM_NX_REG12[],load:0->NXM_NX_REG13[],load:0x7->NXM_NX_REG11[],load:0x4->NXM_NX_REG12[],load:0x1->OXM_OF_METADATA[],load:0x3->NXM_NX_REG14[],load:0->NXM_NX_REG10[],load:0->NXM_NX_REG15[],load:0->NXM_NX_REG0[],load:0->NXM_NX_REG1[],load:0->NXM_NX_REG2[],load:0->NXM_NX_REG3[],load:0->NXM_NX_REG4[],load:0->NXM_NX_REG5[],load:0->NXM_NX_REG6[],load:0->NXM_NX_REG7[],load:0->NXM_NX_REG8[],load:0->NXM_NX_REG9[],load:0->NXM_OF_IN_PORT[],resubmit(,8))
 cookie=0x0, duration=58077.304s, table=65, n_packets=88274, n_bytes=6409543, priority=100,reg15=0x2,metadata=0x5 actions=clone(ct_clear,load:0->NXM_NX_REG11[],load:0->NXM_NX_REG12[],load:0->NXM_NX_REG13[],load:0xc->NXM_NX_REG11[],load:0xb->NXM_NX_REG12[],load:0x6->OXM_OF_METADATA[],load:0x2->NXM_NX_REG14[],load:0->NXM_NX_REG10[],load:0->NXM_NX_REG15[],load:0->NXM_NX_REG0[],load:0->NXM_NX_REG1[],load:0->NXM_NX_REG2[],load:0->NXM_NX_REG3[],load:0->NXM_NX_REG4[],load:0->NXM_NX_REG5[],load:0->NXM_NX_REG6[],load:0->NXM_NX_REG7[],load:0->NXM_NX_REG8[],load:0->NXM_NX_REG9[],load:0->NXM_OF_IN_PORT[],resubmit(,8))
 cookie=0x0, duration=58077.304s, table=65, n_packets=114816, n_bytes=63806298, priority=100,reg15=0x1,metadata=0x5 actions=clone(ct_clear,load:0->NXM_NX_REG11[],load:0->NXM_NX_REG12[],load:0->NXM_NX_REG13[],load:0x2->NXM_NX_REG11[],load:0x3->NXM_NX_REG12[],load:0x2->OXM_OF_METADATA[],load:0x2->NXM_NX_REG14[],load:0->NXM_NX_REG10[],load:0->NXM_NX_REG15[],load:0->NXM_NX_REG0[],load:0->NXM_NX_REG1[],load:0->NXM_NX_REG2[],load:0->NXM_NX_REG3[],load:0->NXM_NX_REG4[],load:0->NXM_NX_REG5[],load:0->NXM_NX_REG6[],load:0->NXM_NX_REG7[],load:0->NXM_NX_REG8[],load:0->NXM_NX_REG9[],load:0->NXM_OF_IN_PORT[],resubmit(,8))
 cookie=0x0, duration=58077.304s, table=65, n_packets=88274, n_bytes=6409735, priority=100,reg15=0x2,metadata=0x2 actions=clone(ct_clear,load:0->NXM_NX_REG11[],load:0->NXM_NX_REG12[],load:0->NXM_NX_REG13[],load:0xe->NXM_NX_REG11[],load:0xd->NXM_NX_REG12[],load:0x5->OXM_OF_METADATA[],load:0x1->NXM_NX_REG14[],load:0->NXM_NX_REG10[],load:0->NXM_NX_REG15[],load:0->NXM_NX_REG0[],load:0->NXM_NX_REG1[],load:0->NXM_NX_REG2[],load:0->NXM_NX_REG3[],load:0->NXM_NX_REG4[],load:0->NXM_NX_REG5[],load:0->NXM_NX_REG6[],load:0->NXM_NX_REG7[],load:0->NXM_NX_REG8[],load:0->NXM_NX_REG9[],load:0->NXM_OF_IN_PORT[],resubmit(,8))
 cookie=0x0, duration=58077.304s, table=65, n_packets=88274, n_bytes=6409543, priority=100,reg15=0x1,metadata=0x6 actions=output:"k8s-patch-br-in"
 cookie=0x0, duration=58077.304s, table=65, n_packets=606946, n_bytes=93442534, priority=100,reg15=0x2,metadata=0x6 actions=clone(ct_clear,load:0->NXM_NX_REG11[],load:0->NXM_NX_REG12[],load:0->NXM_NX_REG13[],load:0xe->NXM_NX_REG11[],load:0xd->NXM_NX_REG12[],load:0x5->OXM_OF_METADATA[],load:0x2->NXM_NX_REG14[],load:0->NXM_NX_REG10[],load:0->NXM_NX_REG15[],load:0->NXM_NX_REG0[],load:0->NXM_NX_REG1[],load:0->NXM_NX_REG2[],load:0->NXM_NX_REG3[],load:0->NXM_NX_REG4[],load:0->NXM_NX_REG5[],load:0->NXM_NX_REG6[],load:0->NXM_NX_REG7[],load:0->NXM_NX_REG8[],load:0->NXM_NX_REG9[],load:0->NXM_OF_IN_PORT[],resubmit(,8))
 cookie=0x0, duration=58072.094s, table=65, n_packets=86458, n_bytes=34551886, priority=100,reg15=0x4,metadata=0x3 actions=output:"enp94s0_0"
 cookie=0x0, duration=58071.187s, table=65, n_packets=86330, n_bytes=34538686, priority=100,reg15=0x3,metadata=0x3 actions=output:"enp94s0_1"
 cookie=0x0, duration=58033.569s, table=65, n_packets=1230430, n_bytes=81267048, priority=100,reg15=0x5,metadata=0x3 actions=output:"enp94s0_2"
 cookie=0x0, duration=58077.046s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa000016,reg15=0x2,metadata=0x5 actions=mod_dl_dst:00:16:3e:80:00:16
 cookie=0x0, duration=58076.729s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa000101,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:d3:e9:54
 cookie=0x0, duration=58076.602s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa000201,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:dd:f3:1e
 cookie=0x0, duration=58076.394s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa000213,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:c3:98:78
 cookie=0x0, duration=58076.300s, table=66, n_packets=120, n_bytes=11736, priority=100,reg0=0xa000001,reg15=0x2,metadata=0x5 actions=mod_dl_dst:00:aa:aa:aa:aa:aa
 cookie=0x0, duration=58075.094s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa00010f,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:d3:e6:88
 cookie=0x0, duration=58074.843s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa000008,reg15=0x2,metadata=0x5 actions=mod_dl_dst:00:16:3e:80:00:08
 cookie=0x0, duration=58074.471s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa00010e,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:08:74:8a
 cookie=0x0, duration=58074.280s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa000209,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:c3:9b:f4
 cookie=0x0, duration=58074.042s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa000203,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:dd:ef:62
 cookie=0x0, duration=58073.879s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa000104,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:08:73:92
 cookie=0x0, duration=58073.849s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa00010a,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:1c:ab:de
 cookie=0x0, duration=58073.502s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa000102,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:c3:98:e8
 cookie=0x0, duration=58073.431s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa000207,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:dd:f3:d6
 cookie=0x0, duration=58073.403s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa000106,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:c3:a0:d8
 cookie=0x0, duration=58072.772s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa000109,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:d3:e9:88
 cookie=0x0, duration=58072.059s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa000205,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:dd:ef:46
 cookie=0x0, duration=58072.019s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa000103,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:c3:a3:bc
 cookie=0x0, duration=58072.008s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa000111,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:08:72:ae
 cookie=0x0, duration=58071.763s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa000067,reg15=0x2,metadata=0x5 actions=mod_dl_dst:ac:1f:6b:8a:6a:6f
 cookie=0x0, duration=58071.711s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa00010c,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:c3:aa:5c
 cookie=0x0, duration=58071.689s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xaf40104,reg15=0x2,metadata=0x1 actions=mod_dl_dst:0a:00:00:00:00:02
 cookie=0x0, duration=58071.215s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa000107,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:1c:ac:8a
 cookie=0x0, duration=58070.835s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xaf40103,reg15=0x2,metadata=0x1 actions=mod_dl_dst:0a:00:00:00:00:01
 cookie=0x0, duration=58070.802s, table=66, n_packets=88148, n_bytes=6397555, priority=100,reg0=0xa000212,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:c3:98:6c
 cookie=0x0, duration=58070.697s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa00010d,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:c3:a3:d4
 cookie=0x0, duration=58070.362s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa00010b,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:08:71:e2
 cookie=0x0, duration=58070.157s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa000206,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:dd:ef:52
 cookie=0x0, duration=58069.766s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa000204,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:dd:fa:e6
 cookie=0x0, duration=58069.698s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa00020c,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:c3:9c:04
 cookie=0x0, duration=58068.944s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa000074,reg15=0x2,metadata=0x5 actions=mod_dl_dst:ac:1f:6b:8b:24:84
 cookie=0x0, duration=58068.870s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa000105,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:d3:ea:60
 cookie=0x0, duration=58068.826s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa000211,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:c3:9c:78
 cookie=0x0, duration=58068.753s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa000202,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:dd:fa:da
 cookie=0x0, duration=58061.284s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa000015,reg15=0x2,metadata=0x5 actions=mod_dl_dst:00:16:3e:80:00:15
 cookie=0x0, duration=58061.265s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa000216,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:c3:95:44
 cookie=0x0, duration=58013.056s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xaf40105,reg15=0x2,metadata=0x1 actions=mod_dl_dst:0a:00:00:00:00:04
 cookie=0x0, duration=57990.968s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xaf40203,reg15=0x3,metadata=0x1 actions=mod_dl_dst:0a:00:00:00:00:03
 cookie=0x0, duration=57807.648s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa000210,reg15=0x2,metadata=0x5 actions=mod_dl_dst:50:6b:4b:c3:98:68
 cookie=0x0, duration=14415.110s, table=66, n_packets=0, n_bytes=0, priority=100,reg0=0xa00000e,reg15=0x2,metadata=0x5 actions=mod_dl_dst:00:16:3e:80:00:0e
```

# conntrack table dump
```
# conntrack -L
tcp      6 86398 ESTABLISHED src=10.244.1.3 dst=10.0.2.18 sport=59320 dport=6443 src=10.0.2.18 dst=10.0.2.19 sport=6443 dport=59320 [ASSURED] mark=0 zone=13 use=1
tcp      6 86398 ESTABLISHED src=10.0.2.19 dst=10.0.2.18 sport=54162 dport=6443 src=10.0.2.18 dst=10.0.2.19 sport=6443 dport=54162 [ASSURED] mark=0 zone=64000 use=1
tcp      6 86398 ESTABLISHED src=10.0.2.19 dst=10.0.2.18 sport=59320 dport=6443 src=10.0.2.18 dst=10.0.2.19 sport=6443 dport=59320 [ASSURED] mark=0 zone=64000 use=1
tcp      6 21 TIME_WAIT src=10.0.0.22 dst=10.0.2.19 sport=39850 dport=22 src=10.0.2.19 dst=10.0.0.22 sport=22 dport=39850 [ASSURED] mark=0 use=1
tcp      6 86398 ESTABLISHED src=10.0.2.19 dst=10.0.2.18 sport=58194 dport=6443 src=10.0.2.18 dst=10.0.2.19 sport=6443 dport=58194 [ASSURED] mark=0 use=1
tcp      6 86398 ESTABLISHED src=10.244.1.4 dst=10.96.0.1 sport=54162 dport=443 src=10.0.2.18 dst=10.244.1.4 sport=6443 dport=54162 [ASSURED] mark=0 zone=15 use=1
tcp      6 116 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34958 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34958 [ASSURED] mark=0 zone=15 use=1
tcp      6 116 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34958 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34958 [ASSURED] mark=0 zone=1 use=1
tcp      6 116 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34958 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34958 [ASSURED] mark=0 use=1
tcp      6 111 TIME_WAIT src=10.0.0.22 dst=10.0.2.19 sport=40786 dport=22 src=10.0.2.19 dst=10.0.0.22 sport=22 dport=40786 [ASSURED] mark=0 use=1
tcp      6 86379 ESTABLISHED src=10.0.2.19 dst=10.0.2.18 sport=58072 dport=6443 src=10.0.2.18 dst=10.0.2.19 sport=6443 dport=58072 [ASSURED] mark=0 use=1
tcp      6 102 TIME_WAIT src=10.0.0.21 dst=10.0.2.19 sport=37176 dport=22 src=10.0.2.19 dst=10.0.0.21 sport=22 dport=37176 [ASSURED] mark=0 use=1
tcp      6 300 ESTABLISHED src=10.244.1.5 dst=10.104.114.18 sport=60550 dport=10005 src=10.244.2.3 dst=10.244.1.5 sport=5001 dport=60550 [ASSURED] mark=0 zone=17 use=5
icmp     1 29 src=10.0.0.21 dst=10.0.2.19 type=8 code=0 id=7776 src=10.0.2.19 dst=10.0.0.21 type=0 code=0 id=7776 mark=0 use=1
tcp      6 86399 ESTABLISHED src=10.0.0.6 dst=10.0.2.19 sport=44126 dport=22 src=10.0.2.19 dst=10.0.0.6 sport=22 dport=44126 [ASSURED] mark=0 use=1
tcp      6 16 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34516 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34516 [ASSURED] mark=0 zone=15 use=1
tcp      6 16 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34516 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34516 [ASSURED] mark=0 zone=1 use=1
tcp      6 16 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34516 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34516 [ASSURED] mark=0 use=2
tcp      6 8 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38018 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38018 [ASSURED] mark=0 zone=16 use=1
tcp      6 8 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38018 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38018 [ASSURED] mark=0 zone=1 use=1
tcp      6 8 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38018 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38018 [ASSURED] mark=0 use=1
tcp      6 28 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38106 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38106 [ASSURED] mark=0 zone=16 use=1
tcp      6 28 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38106 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38106 [ASSURED] mark=0 zone=1 use=1
tcp      6 28 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38106 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38106 [ASSURED] mark=0 use=1
tcp      6 46 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34648 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34648 [ASSURED] mark=0 zone=15 use=1
tcp      6 46 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34648 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34648 [ASSURED] mark=0 zone=1 use=1
tcp      6 46 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34648 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34648 [ASSURED] mark=0 use=1
tcp      6 30 TIME_WAIT src=10.0.0.22 dst=10.0.2.19 sport=39966 dport=22 src=10.0.2.19 dst=10.0.0.22 sport=22 dport=39966 [ASSURED] mark=0 use=1
udp      17 27 src=10.0.0.22 dst=10.0.2.19 sport=37286 dport=623 [UNREPLIED] src=10.0.2.19 dst=10.0.0.22 sport=623 dport=37286 mark=0 use=1
tcp      6 48 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38194 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38194 [ASSURED] mark=0 zone=16 use=1
tcp      6 48 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38194 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38194 [ASSURED] mark=0 zone=1 use=1
tcp      6 48 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38194 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38194 [ASSURED] mark=0 use=1
udp      17 18 src=10.0.2.19 dst=208.67.222.222 sport=58369 dport=53 src=208.67.222.222 dst=10.0.2.19 sport=53 dport=58369 mark=0 use=1
tcp      6 95 TIME_WAIT src=10.0.0.22 dst=10.0.2.19 sport=40628 dport=22 src=10.0.2.19 dst=10.0.0.22 sport=22 dport=40628 [ASSURED] mark=0 use=1
tcp      6 86398 ESTABLISHED src=10.244.1.3 dst=10.96.0.1 sport=59320 dport=443 src=10.0.2.18 dst=10.244.1.3 sport=6443 dport=59320 [ASSURED] mark=0 zone=16 use=1
icmp     1 29 src=10.0.0.22 dst=10.0.2.19 type=8 code=0 id=9688 src=10.0.2.19 dst=10.0.0.22 type=0 code=0 id=9688 mark=0 use=1
tcp      6 96 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34868 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34868 [ASSURED] mark=0 zone=15 use=1
tcp      6 96 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34868 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34868 [ASSURED] mark=0 zone=1 use=1
tcp      6 96 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34868 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34868 [ASSURED] mark=0 use=1
tcp      6 106 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34912 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34912 [ASSURED] mark=0 zone=15 use=1
tcp      6 106 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34912 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34912 [ASSURED] mark=0 zone=1 use=1
tcp      6 106 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34912 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34912 [ASSURED] mark=0 use=1
icmp     1 22 src=10.0.0.21 dst=10.0.2.19 type=8 code=0 id=7614 src=10.0.2.19 dst=10.0.0.21 type=0 code=0 id=7614 mark=0 use=1
tcp      6 76 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34780 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34780 [ASSURED] mark=0 zone=15 use=1
tcp      6 76 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34780 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34780 [ASSURED] mark=0 zone=1 use=1
tcp      6 76 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34780 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34780 [ASSURED] mark=0 use=1
tcp      6 68 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38282 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38282 [ASSURED] mark=0 zone=16 use=1
tcp      6 68 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38282 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38282 [ASSURED] mark=0 zone=1 use=1
tcp      6 68 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38282 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38282 [ASSURED] mark=0 use=1
tcp      6 98 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38414 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38414 [ASSURED] mark=0 zone=16 use=1
tcp      6 98 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38414 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38414 [ASSURED] mark=0 zone=1 use=1
tcp      6 98 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38414 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38414 [ASSURED] mark=0 use=1
udp      17 168 src=127.0.0.1 dst=127.0.0.53 sport=49051 dport=53 src=127.0.0.53 dst=127.0.0.1 sport=53 dport=49051 [ASSURED] mark=0 use=1
tcp      6 6 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34472 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34472 [ASSURED] mark=0 zone=15 use=1
tcp      6 6 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34472 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34472 [ASSURED] mark=0 zone=1 use=1
tcp      6 6 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34472 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34472 [ASSURED] mark=0 use=2
tcp      6 86 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34824 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34824 [ASSURED] mark=0 zone=15 use=1
tcp      6 86 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34824 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34824 [ASSURED] mark=0 zone=1 use=1
tcp      6 86 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34824 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34824 [ASSURED] mark=0 use=1
tcp      6 86398 ESTABLISHED src=10.244.1.4 dst=10.0.2.18 sport=54162 dport=6443 src=10.0.2.18 dst=10.0.2.19 sport=6443 dport=54162 [ASSURED] mark=0 zone=13 use=1
tcp      6 56 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34692 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34692 [ASSURED] mark=0 zone=15 use=1
tcp      6 56 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34692 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34692 [ASSURED] mark=0 zone=1 use=1
tcp      6 56 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34692 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34692 [ASSURED] mark=0 use=1
tcp      6 26 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34560 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34560 [ASSURED] mark=0 zone=15 use=1
tcp      6 26 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34560 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34560 [ASSURED] mark=0 zone=1 use=1
tcp      6 26 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34560 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34560 [ASSURED] mark=0 use=1
tcp      6 118 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38504 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38504 [ASSURED] mark=0 zone=16 use=1
tcp      6 118 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38504 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38504 [ASSURED] mark=0 zone=1 use=1
tcp      6 118 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38504 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38504 [ASSURED] mark=0 use=1
udp      17 157 src=127.0.0.1 dst=127.0.0.53 sport=56441 dport=53 src=127.0.0.53 dst=127.0.0.1 sport=53 dport=56441 [ASSURED] mark=0 use=1
tcp      6 55 TIME_WAIT src=10.0.0.21 dst=10.0.2.19 sport=36634 dport=22 src=10.0.2.19 dst=10.0.0.21 sport=22 dport=36634 [ASSURED] mark=0 use=1
tcp      6 88 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38370 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38370 [ASSURED] mark=0 zone=16 use=1
tcp      6 88 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38370 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38370 [ASSURED] mark=0 zone=1 use=1
tcp      6 88 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38370 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38370 [ASSURED] mark=0 use=1
tcp      6 66 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34736 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34736 [ASSURED] mark=0 zone=15 use=1
tcp      6 66 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34736 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34736 [ASSURED] mark=0 zone=1 use=1
tcp      6 66 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34736 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34736 [ASSURED] mark=0 use=1
udp      17 18 src=127.0.0.1 dst=127.0.0.53 sport=45445 dport=53 src=127.0.0.53 dst=127.0.0.1 sport=53 dport=45445 mark=0 use=1
udp      17 30 src=10.0.2.19 dst=10.0.2.22 sport=46104 dport=6081 [UNREPLIED] src=10.0.2.22 dst=10.0.2.19 sport=6081 dport=46104 mark=0 use=22
tcp      6 18 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38062 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38062 [ASSURED] mark=0 zone=16 use=1
tcp      6 18 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38062 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38062 [ASSURED] mark=0 zone=1 use=1
tcp      6 18 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38062 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38062 [ASSURED] mark=0 use=1
udp      17 7 src=127.0.0.1 dst=127.0.0.53 sport=48274 dport=53 src=127.0.0.53 dst=127.0.0.1 sport=53 dport=48274 mark=0 use=1
tcp      6 36 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34604 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34604 [ASSURED] mark=0 zone=15 use=1
tcp      6 36 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34604 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34604 [ASSURED] mark=0 zone=1 use=1
tcp      6 36 TIME_WAIT src=10.244.1.2 dst=10.244.1.4 sport=34604 dport=8080 src=10.244.1.4 dst=10.244.1.2 sport=8080 dport=34604 [ASSURED] mark=0 use=1
tcp      6 89 TIME_WAIT src=10.0.0.21 dst=10.0.2.19 sport=36992 dport=22 src=10.0.2.19 dst=10.0.0.21 sport=22 dport=36992 [ASSURED] mark=0 use=1
tcp      6 86399 ESTABLISHED src=10.0.2.18 dst=10.0.2.19 sport=33514 dport=10250 src=10.0.2.19 dst=10.0.2.18 sport=10250 dport=33514 [ASSURED] mark=0 use=1
tcp      6 53 TIME_WAIT src=10.0.0.22 dst=10.0.2.19 sport=40164 dport=22 src=10.0.2.19 dst=10.0.0.22 sport=22 dport=40164 [ASSURED] mark=0 use=1
tcp      6 108 TIME_WAIT src=10.0.2.19 dst=91.189.88.152 sport=51596 dport=80 src=91.189.88.152 dst=10.0.2.19 sport=80 dport=51596 [ASSURED] mark=0 use=1
tcp      6 25 TIME_WAIT src=10.0.0.21 dst=10.0.2.19 sport=36332 dport=22 src=10.0.2.19 dst=10.0.0.21 sport=22 dport=36332 [ASSURED] mark=0 use=1
udp      17 18 src=10.0.2.19 dst=208.67.222.222 sport=47790 dport=53 src=208.67.222.222 dst=10.0.2.19 sport=53 dport=47790 mark=0 use=1
tcp      6 108 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38458 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38458 [ASSURED] mark=0 zone=16 use=1
tcp      6 108 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38458 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38458 [ASSURED] mark=0 zone=1 use=1
tcp      6 108 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38458 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38458 [ASSURED] mark=0 use=2
tcp      6 86392 ESTABLISHED src=10.0.2.19 dst=10.0.2.18 sport=58078 dport=6443 src=10.0.2.18 dst=10.0.2.19 sport=6443 dport=58078 [ASSURED] mark=0 use=1
tcp      6 58 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38238 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38238 [ASSURED] mark=0 zone=16 use=1
tcp      6 58 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38238 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38238 [ASSURED] mark=0 zone=1 use=1
tcp      6 58 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38238 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38238 [ASSURED] mark=0 use=1
tcp      6 86399 ESTABLISHED src=127.0.0.1 dst=127.0.0.1 sport=53236 dport=42465 src=127.0.0.1 dst=127.0.0.1 sport=42465 dport=53236 [ASSURED] mark=0 use=1
udp      17 7 src=127.0.0.1 dst=127.0.0.53 sport=58069 dport=53 src=127.0.0.53 dst=127.0.0.1 sport=53 dport=58069 mark=0 use=1
tcp      6 86399 ESTABLISHED src=10.0.2.19 dst=10.0.2.18 sport=60232 dport=6642 src=10.0.2.18 dst=10.0.2.19 sport=6642 dport=60232 [ASSURED] mark=0 use=1
tcp      6 86398 ESTABLISHED src=10.0.2.19 dst=10.0.2.18 sport=58086 dport=6443 src=10.0.2.18 dst=10.0.2.19 sport=6443 dport=58086 [ASSURED] mark=0 use=1
tcp      6 37 TIME_WAIT src=10.0.0.21 dst=10.0.2.19 sport=36452 dport=22 src=10.0.2.19 dst=10.0.0.21 sport=22 dport=36452 [ASSURED] mark=0 use=1
udp      17 10 src=0.0.0.0 dst=255.255.255.255 sport=68 dport=67 [UNREPLIED] src=255.255.255.255 dst=0.0.0.0 sport=67 dport=68 mark=0 use=1
tcp      6 38 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38150 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38150 [ASSURED] mark=0 zone=16 use=1
tcp      6 38 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38150 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38150 [ASSURED] mark=0 zone=1 use=1
tcp      6 38 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38150 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38150 [ASSURED] mark=0 use=1
icmp     1 22 src=10.0.0.22 dst=10.0.2.19 type=8 code=0 id=9588 src=10.0.2.19 dst=10.0.0.22 type=0 code=0 id=9588 mark=0 use=1
tcp      6 78 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38326 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38326 [ASSURED] mark=0 zone=16 use=1
tcp      6 78 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38326 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38326 [ASSURED] mark=0 zone=1 use=1
tcp      6 78 TIME_WAIT src=10.244.1.2 dst=10.244.1.3 sport=38326 dport=8080 src=10.244.1.3 dst=10.244.1.2 sport=8080 dport=38326 [ASSURED] mark=0 use=1
udp      17 30 src=10.0.2.22 dst=10.0.2.19 sport=35556 dport=6081 [UNREPLIED] src=10.0.2.19 dst=10.0.2.22 sport=6081 dport=35556 mark=0 use=1
tcp      6 83 TIME_WAIT src=10.0.0.22 dst=10.0.2.19 sport=40502 dport=22 src=10.0.2.19 dst=10.0.0.22 sport=22 dport=40502 [ASSURED] mark=0 use=1
udp      17 18 src=10.0.2.19 dst=208.67.222.222 sport=56043 dport=53 src=208.67.222.222 dst=10.0.2.19 sport=53 dport=56043 mark=0 use=1
conntrack v1.4.4 (conntrack-tools): 115 flow entries have been shown.
```
