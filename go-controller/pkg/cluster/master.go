package cluster

import (
	"fmt"
	"net"
	"strings"
	"strconv"
	"encoding/json"

	kapi "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/openvswitch/ovn-kubernetes/go-controller/pkg/ovn"
	"github.com/openvswitch/ovn-kubernetes/go-controller/pkg/util"

	"github.com/openshift/origin/pkg/util/netutils"
	"github.com/sirupsen/logrus"
)

func BoolToYN(b bool) string {
	if b {
	   return "yes"
	}
	return "no"
}

func YNtoBool(s string) bool {
	if s == "yes" {
		return true
	}
	return false
}

// RebuildOVNDatabase rebuilds db if HA option is enabled and OVN DB
// doesn't exist. First It updates k8s nodes by creating a logical
// switch for each node. Then it reads all resources from k8s and
// creates needed resources in OVN DB. Last, it updates master node
// ip in default namespace to trigger event on node.
func (cluster *OvnClusterController) RebuildOVNDatabase(
	masterNodeName string, oc *ovn.Controller) error {
	logrus.Debugf("Rebuild OVN database for cluster nodes")
	var err error
	ipChange, err := cluster.checkMasterIPChange(masterNodeName)
	if err != nil {
		logrus.Errorf("Error when checking Master Node IP Change: %v", err)
		return err
	}

	// If OvnHA options is enabled, make sure default namespace has the
	// annotation of current cluster master node's overlay IP address
	logrus.Debugf("cluster.OvnHA: %t", cluster.OvnHA)
	if cluster.OvnHA && ipChange {
		logrus.Debugf("HA is enabled and DB doesn't exist!")
		// Rebuild OVN DB for the k8s nodes
		err = cluster.UpdateDBForKubeNodes(masterNodeName)
		if err != nil {
			return err
		}
		// Rebuild OVN DB for the k8s namespaces and all the resource
		// objects inside the namespace including pods and network
		// policies
		err = cluster.UpdateKubeNsObjects(oc)
		if err != nil {
			return err
		}
		// Update master node IP annotation on default namespace
		err = cluster.UpdateMasterNodeIP(masterNodeName)
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateDBForKubeNodes rebuilds ovn db for k8s nodes
func (cluster *OvnClusterController) UpdateDBForKubeNodes(
	masterNodeName string) error {
	nodes, err := cluster.Kube.GetNodes()
	if err != nil {
		logrus.Errorf("Failed to obtain k8s nodes: %v", err)
		return err
	}
	for _, node := range nodes.Items {
		for key, _ := range cluster.ClusterNetList {
			subnetstr := fmt.Sprintf("%s_host_subnet", key)
			subnet, ok := node.Annotations[subnetstr]
			ls_name := fmt.Sprintf("%s-%s", node.Name, key)
			if ok {
				// Create a logical switch for the node
				logrus.Debugf("%s: %s", key, subnet)
				ip, localNet, err := net.ParseCIDR(subnet)
				if err != nil {
					return fmt.Errorf("Failed to parse subnet %v: %v", subnet,
						err)
				}
				ip = util.NextIP(ip)
				n, _ := localNet.Mask.Size()
				routerIPMask := fmt.Sprintf("%s/%d", ip.String(), n)
				stdout, stderr, err := util.RunOVNNbctl("--may-exist",
					"ls-add", node.Name, "--", "set", "logical_switch",
					node.Name, fmt.Sprintf("other-config:subnet=%s", subnet),
					fmt.Sprintf("external-ids:gateway_ip=%s", routerIPMask))
				if err != nil {
					logrus.Errorf("Failed to create logical switch %s for "+
						"node %s, stdout: %q, stderr: %q, error: %v",
						ls_name, node.Name, stdout, stderr, err)
					return err
				}
			}
		}
	}
	return nil
}

// UpdateKubeNsObjects rebuilds ovn db for k8s namespaces
// and pods/networkpolicies inside the namespaces.
func (cluster *OvnClusterController) UpdateKubeNsObjects(
	oc *ovn.Controller) error {
	namespaces, err := cluster.Kube.GetNamespaces()
	if err != nil {
		logrus.Errorf("Failed to get k8s namespaces: %v", err)
		return err
	}
	for _, ns := range namespaces.Items {
		oc.AddNamespace(&ns)
		pods, err := cluster.Kube.GetPods(ns.Name)
		if err != nil {
			logrus.Errorf("Failed to get k8s pods: %v", err)
			return err
		}
		for _, pod := range pods.Items {
			oc.AddLogicalPortWithIP(&pod)
		}
		endpoints, err := cluster.Kube.GetEndpoints(ns.Name)
		if err != nil {
			logrus.Errorf("Failed to get k8s endpoints: %v", err)
			return err
		}
		for _, ep := range endpoints.Items {
			er := oc.AddEndpoints(&ep)
			if er != nil {
				logrus.Errorf("Error adding endpoints: %v", er)
				return er
			}
		}
		policies, err := cluster.Kube.GetNetworkPolicies(ns.Name)
		if err != nil {
			logrus.Errorf("Failed to get k8s network policies: %v", err)
			return err
		}
		for _, policy := range policies.Items {
			oc.AddNetworkPolicy(&policy)
		}
	}
	return nil
}

// UpdateMasterNodeIP add annotations of master node IP on
// default namespace
func (cluster *OvnClusterController) UpdateMasterNodeIP(
	masterNodeName string) error {
	masterNodeIP, err := netutils.GetNodeIP(masterNodeName)
	if err != nil {
		return fmt.Errorf("Failed to obtain local IP from master node "+
			"%q: %v", masterNodeName, err)
	}

	defaultNs, err := cluster.Kube.GetNamespace(DefaultNamespace)
	if err != nil {
		return fmt.Errorf("Failed to get default namespace: %v", err)
	}

	// Get overlay ip on OVN master node from default namespace. If it
	// doesn't have it or the IP address is different than the current one,
	// set it with current master overlay IP.
	masterIP, ok := defaultNs.Annotations[MasterOverlayIP]
	if !ok || masterIP != masterNodeIP {
		err := cluster.Kube.SetAnnotationOnNamespace(defaultNs, MasterOverlayIP,
			masterNodeIP)
		if err != nil {
			return fmt.Errorf("Failed to set %s=%s on namespace %s: %v",
				MasterOverlayIP, masterNodeIP, defaultNs.Name, err)
		}
	}

	return nil
}

func (cluster *OvnClusterController) checkMasterIPChange(
	masterNodeName string) (bool, error) {
	// Check DB existence by checking if the default namespace annotated
	// IP address is the same as the master node IP. If annotated IP
	// address is different, we assume that the ovn db is crashed on the
	// old node and needs to be rebuilt on the new master node.
	masterNodeIP, err := netutils.GetNodeIP(masterNodeName)
	if err != nil {
		return false, fmt.Errorf("Failed to obtain local IP from master "+
			"node %q: %v", masterNodeName, err)
	}

	defaultNs, err := cluster.Kube.GetNamespace(DefaultNamespace)
	if err != nil {
		return false, fmt.Errorf("Failed to get default namespace: %v", err)
	}

	// Get overlay ip on OVN master node from default namespace. If the IP
	// address is different than master node IP, return true. Else, return
	// false.
	masterIP := defaultNs.Annotations[MasterOverlayIP]
	logrus.Debugf("Master IP: %s, Annotated IP: %s", masterNodeIP, masterIP)
	if masterIP != masterNodeIP {
		logrus.Debugf("Detected Master node IP is different than default " +
			"namespae annotated IP.")
		return true, nil
	}
	return false, nil
}

// XXX TODOS: These 2 fns are duplicated from ovnkube.go, need to move this to util pkg
//cidrsOverlap returns a true if the cidr range overlaps any in the list of cidr ranges
func cidrsOverlap(cidr *net.IPNet, cidrList []CIDRNetworkEntry) bool {

	for _, clusterEntry := range cidrList {
		if cidr.Contains(clusterEntry.CIDR.IP) || clusterEntry.CIDR.Contains(cidr.IP) {
			return true
		}
	}
	return false
}

// parseNetworkIPNetEntries returns the parsed set of CIDRNetworkEntries passed by the user on the command line
// These entries define the clusters network space by specifying a set of CIDR and netmaskas the SDN can allocate
// addresses from.
func parseNetworkIPNetEntries(clusterSubnetCmd string) ([]CIDRNetworkEntry, error) {
	var parsedClusterList []CIDRNetworkEntry

	clusterEntriesList := strings.Split(clusterSubnetCmd, ",")

	for _, clusterEntry := range clusterEntriesList {
		var parsedClusterEntry CIDRNetworkEntry

		splitClusterEntry := strings.Split(clusterEntry, "/")
		if len(splitClusterEntry) == 3 {
			tmp, err := strconv.ParseUint(splitClusterEntry[2], 10, 32)
			if err != nil {
				return nil, err
			}
			parsedClusterEntry.HostSubnetLength = uint32(tmp)
		} else if len(splitClusterEntry) == 2 {
			// the old hardcoded value for backwards compatability
			parsedClusterEntry.HostSubnetLength = 24
		} else {
			return nil, fmt.Errorf("cluster-cidr not formatted properly")
		}

		var err error
		_, parsedClusterEntry.CIDR, err = net.ParseCIDR(fmt.Sprintf("%s/%s", splitClusterEntry[0], splitClusterEntry[1]))
		if err != nil {
			return nil, err
		}

		//check to make sure that no cidrs overlap
		if cidrsOverlap(parsedClusterEntry.CIDR, parsedClusterList) {
			return nil, fmt.Errorf("CIDR %s overlaps with another cluster network CIDR", parsedClusterEntry.CIDR.String())
		}

		parsedClusterList = append(parsedClusterList, parsedClusterEntry)

	}

	return parsedClusterList, nil
}

// Creating a logical switch for each non-"ovn" network.
// XXX-TODOS no need for the logical router if we are configuring only l2
// localnet.
func createLogicalSwitches(netname, localSubnet string, gateway_init bool, gateway_type string, vlanid uint32, gateway_net, gateway_chassis string) error {


	ip, localSubnetNet, err := net.ParseCIDR(localSubnet)
	if err != nil {
		return fmt.Errorf("Failed to parse local subnet %v : %v", localSubnetNet, err)
	}
	ip = util.NextIP(ip)
	n, _ := localSubnetNet.Mask.Size()
	routerIPMask := fmt.Sprintf("%s/%d", ip.String(), n)

	// Create the Logical Router port.
	routerMac, stderr, err := util.RunOVNNbctl("--if-exist", "get", "logical_router_port", "rtos-"+netname, "mac")
	if err != nil {
		logrus.Errorf("Failed to get logical router port,stderr: %q, error: %v", stderr, err)
		return err
	}

	var clusterRouter string
	if routerMac == "" {
		routerMac = util.GenerateMac()
	}

	clusterRouter, err = util.GetK8sClusterRouter()
	if err != nil {
		return err
	}

	// Connect Logical Router to switch
	_, stderr, err = util.RunOVNNbctl("--may-exist", "lrp-add", clusterRouter, "rtos-"+netname, routerMac, routerIPMask)
	if err != nil {
		logrus.Errorf("Failed to add logical port to router, stderr: %q, error: %v", stderr, err)
		return err
	}

	// Create the logical switch
	stdout, stderr, err := util.RunOVNNbctl("--", "--may-exist", "ls-add", netname, "--", "set", "logical_switch", netname, "other-config:subnet="+localSubnet, "external-ids:gateway_ip="+routerIPMask)
	if err != nil {
		logrus.Errorf("Failed to create a logical switch %v, stdout: %q, stderr: %q, error: %v", netname, stdout, stderr, err)
		return err
	}

	// Connect the switch to the router.
	stdout, stderr, err = util.RunOVNNbctl("--", "--may-exist", "lsp-add", netname, "stor-"+netname, "--", "set", "logical_switch_port", "stor-"+netname, "type=router", "options:router-port=rtos-"+netname, "addresses="+"\""+routerMac+"\"")
	if err != nil {
		logrus.Errorf("Failed to add logical port to switch, stdout: %q, stderr: %q, error: %v", stdout, stderr, err)
		return err
	}

	if !gateway_init || (gateway_type != "l2localnet" && gateway_type != "l2gateway") {
		return nil
	}

	logical_switch := netname
	localport_name := netname + "_localnet"

	// Add the localnet port to the logical switch
	if gateway_type == "l2localnet" {
		if vlanid == 0 {
			_, _, err = util.RunOVNNbctl("--", "--may-exist", "lsp-add", logical_switch,  localport_name,  "--", "set", "logical_switch_port", localport_name, "type=localnet", "addresses=unknown", "options:network_name="+gateway_net)
		} else {
			vlanstr := strconv.Itoa(int(vlanid))
			_, _, err = util.RunOVNNbctl("--", "--may-exist", "lsp-add", logical_switch,  localport_name,  "--", "set", "logical_switch_port", localport_name, "type=localnet", "addresses=unknown", "options:network_name="+gateway_net, "tag_request="+vlanstr)
		}
	} else if gateway_type == "l2gateway" {
		if vlanid == 0 {
			_, _, err = util.RunOVNNbctl("--", "--may-exist", "lsp-add", logical_switch,  localport_name,  "--", "set", "logical_switch_port", localport_name, "type=l2gateway", "addresses=unknown", "options:network_name="+gateway_net, "options:l2gateway-chassis="+gateway_chassis)
		} else {
			vlanstr := strconv.Itoa(int(vlanid))
			_, _, err = util.RunOVNNbctl("--", "--may-exist", "lsp-add", logical_switch,  localport_name,  "--", "set", "logical_switch_port", localport_name, "type=l2gateway", "addresses=unknown", "options:network_name="+gateway_net, "options:l2gateway-chassis="+gateway_chassis, "tag_request="+vlanstr)
		}
	} else {
		return fmt.Errorf("createLogicalSwitches: Unknown l2gateway type %s", gateway_type)
	}
	if err != nil {
		logrus.Errorf("createLogicalSwitches: Failed to create logical switch port called %s on %s, error: %v", localport_name, logical_switch, err)
		return err
	}
	return nil
}

// StartClusterMaster runs a subnet IPAM and a controller that watches arrival/departure
// of nodes in the cluster
// On an addition to the cluster (node create), a new subnet is created for it that will translate
// to creation of a logical switch (done by the node, but could be created here at the master process too)
// Upon deletion of a node, the switch will be deleted
//
// TODO: Verify that the cluster was not already called with a different global subnet
//  If true, then either quit or perform a complete reconfiguration of the cluster (recreate switches/routers with new subnet values)
func (cluster *OvnClusterController) StartClusterMaster(masterNodeName string) error {
	masterLabel, err := labels.Parse("name=ovnkube-master")
        masterOvnPods, err := cluster.Kube.GetPodsByLabels("ovn-kubernetes", masterLabel)
	masterOvnPod := masterOvnPods.Items[0]

	var podsann map[string]string
        podsann, err = cluster.Kube.GetAnnotationsOnPod("ovn-kubernetes", masterOvnPod.GetName())
        if err != nil {
               logrus.Warningf("Error while obtaining pod annotations - %v", err)
	}

	networksAnn := podsann[OvnNetworks]
        networks := strings.Split(networksAnn, " ")

	for _, network := range networks {
               netann := podsann[network]
               var netMap map[string]string
               err = json.Unmarshal([]byte(netann), &netMap)
               if err != nil {
                         logrus.Errorf("unmarshal network annotation failed")
               }

               vid := uint32(0)
               vlanid, err := strconv.Atoi(netMap["vlan_id"])
	       if err == nil {
                         vid = uint32(vlanid)
               }

	       cluster.ClusterNetList[network] = &ClusterNets{NetVlanID: vid,
							       NetworkSubnet: netMap["subnet"],
							       NetSriovDev: netMap["sriov_pf"],
							       GatewayNet: netMap["gateway_net"],
							       GatewayIntf: netMap["gateway_itf"],
							       GatewayType: netMap["gateway_type"],
							       GatewayNextHop :  netMap["gateway_nexthop"],
							       L2GatewayChassis: netMap["l2gateway_chassis"],
							       GatewaySpareIntf: YNtoBool(netMap["gateway_spareintf"]),
							       NetworkAllocate: YNtoBool(netMap["network_allocate"]),
							       NodePortEnable: YNtoBool(netMap["nodeport_enable"]),
							       GatewayInit: YNtoBool(netMap["init_gateways"]),
							       NetSriovOnly: YNtoBool(netMap["sriov_only"])}
	       cluster.ClusterNetList[network].NetworkIPNet, err = parseNetworkIPNetEntries(netMap["subnet"])
	       if err != nil {
	              logrus.Debugf("Error parsing subnet from %s", netMap["subnet"])
	       }
	}

	existingNodes, err := cluster.Kube.GetNodes()
	if err != nil {
		logrus.Errorf("Error in getting nodes: %v", err)
		return err
	}
	for _, node := range existingNodes.Items {
		for key, _ := range cluster.ClusterNetList {
			if !cluster.ClusterNetList[key].NetworkAllocate {
				continue
			}
			cluster.ClusterNetList[key].alreadyAllocated = make([]string, 0)
			subnetstr := fmt.Sprintf("%s_host_subnet", key)
			hostsubnet, ok := node.Annotations[subnetstr]
			if ok {
				cluster.ClusterNetList[key].alreadyAllocated = append(cluster.ClusterNetList[key].alreadyAllocated, hostsubnet)
			}
		}
	}
	// NewSubnetAllocator is a subnet IPAM, which takes a CIDR (first argument)
	// and gives out subnets of length 'hostSubnetLength' (second argument)
	// but omitting any that exist in 'subrange' (third argument)
	// We'll do this for "ovn" network, for the rest, we'll use a logical switch
	// per specified network.
	masterSubnetAllocatorList := make([]*netutils.SubnetAllocator, 0)
	for key, _ := range cluster.ClusterNetList {
		if !cluster.ClusterNetList[key].NetworkAllocate {
			continue
		}
		masterSubnetAllocatorList = nil
		for _, clusterEntry := range cluster.ClusterNetList[key].NetworkIPNet {
			subrange := make([]string, 0)
			for _, allocatedRange := range cluster.ClusterNetList[key].alreadyAllocated {
				firstAddress, _, err := net.ParseCIDR(allocatedRange)
				if err != nil {
					return err
				}

				if clusterEntry.CIDR.Contains(firstAddress) {
					subrange = append(subrange, allocatedRange)
				}
			}
			subnetAllocator, err := netutils.NewSubnetAllocator(clusterEntry.CIDR.String(), 32-clusterEntry.HostSubnetLength, subrange)
			if err != nil {
				return err
			}
			masterSubnetAllocatorList = append(masterSubnetAllocatorList, subnetAllocator)
		}
		cluster.ClusterNetList[key].masterSubnetAllocatorList = masterSubnetAllocatorList
	}

	if err := cluster.SetupMaster(masterNodeName); err != nil {
		logrus.Errorf("Failed to setup master (%v)", err)
		return err
	}

	// Create the logical switches for networks that don't need a
        // per-node subnet, i.e. NetworkAllocate is false.
	for key, _ := range cluster.ClusterNetList {
		if cluster.ClusterNetList[key].NetworkAllocate {
			continue
		}
		// XXX-TODOS may not need a list of subnets in this case.
		for _, clusterEntry := range cluster.ClusterNetList[key].NetworkIPNet {
			err = createLogicalSwitches(key,  clusterEntry.CIDR.String(), cluster.ClusterNetList[key].GatewayInit, cluster.ClusterNetList[key].GatewayType, cluster.ClusterNetList[key].NetVlanID,  cluster.ClusterNetList[key].GatewayNet, cluster.ClusterNetList[key].L2GatewayChassis)
			if err != nil {
				return err
			}
		}
	}
	// now go over the 'existing' list again and create annotations for those who do not have it
	for _, node := range existingNodes.Items {
		_, ok := node.Annotations["networks"]
		if !ok {
			err := cluster.addNode(&node)
			if err != nil {
				logrus.Errorf("error creating subnet for node %s: %v", node.Name, err)
				break
			}
		}
	}

	// Watch all node events.  On creation, addNode will be called that will
	// create a subnet for the switch belonging to that node. On a delete
	// call, the subnet will be returned to the allocator as the switch is
	// deleted from ovn
	return cluster.watchNodes()
}

// SetupMaster creates the central router and load-balancers for the network
func (cluster *OvnClusterController) SetupMaster(masterNodeName string) error {
	if err := setupOVNMaster(masterNodeName); err != nil {
		return err
	}

	// Create a single common distributed router for the cluster.
	stdout, stderr, err := util.RunOVNNbctl("--", "--may-exist", "lr-add", OvnClusterRouter,
		"--", "set", "logical_router", OvnClusterRouter, "external_ids:k8s-cluster-router=yes")
	if err != nil {
		logrus.Errorf("Failed to create a single common distributed router for the cluster, "+
			"stdout: %q, stderr: %q, error: %v", stdout, stderr, err)
		return err
	}

	// Create 2 load-balancers for east-west traffic.  One handles UDP and another handles TCP.
	k8sClusterLbTCP, stderr, err := util.RunOVNNbctl("--data=bare", "--no-heading", "--columns=_uuid", "find", "load_balancer", "external_ids:k8s-cluster-lb-tcp=yes")
	if err != nil {
		logrus.Errorf("Failed to get tcp load-balancer, stderr: %q, error: %v", stderr, err)
		return err
	}

	if k8sClusterLbTCP == "" {
		stdout, stderr, err = util.RunOVNNbctl("--", "create", "load_balancer", "external_ids:k8s-cluster-lb-tcp=yes", "protocol=tcp")
		if err != nil {
			logrus.Errorf("Failed to create tcp load-balancer, stdout: %q, stderr: %q, error: %v", stdout, stderr, err)
			return err
		}
	}

	k8sClusterLbUDP, stderr, err := util.RunOVNNbctl("--data=bare", "--no-heading", "--columns=_uuid", "find", "load_balancer", "external_ids:k8s-cluster-lb-udp=yes")
	if err != nil {
		logrus.Errorf("Failed to get udp load-balancer, stderr: %q, error: %v", stderr, err)
		return err
	}
	if k8sClusterLbUDP == "" {
		stdout, stderr, err = util.RunOVNNbctl("--", "create", "load_balancer", "external_ids:k8s-cluster-lb-udp=yes", "protocol=udp")
		if err != nil {
			logrus.Errorf("Failed to create udp load-balancer, stdout: %q, stderr: %q, error: %v", stdout, stderr, err)
			return err
		}
	}

	// Create a logical switch called "join" that will be used to connect gateway routers to the distributed router.
	// The "join" will be allocated IP addresses in the range 100.64.1.0/24.
	stdout, stderr, err = util.RunOVNNbctl("--may-exist", "ls-add", "join")
	if err != nil {
		logrus.Errorf("Failed to create logical switch called \"join\", stdout: %q, stderr: %q, error: %v", stdout, stderr, err)
		return err
	}

	// Connect the distributed router to "join".
	routerMac, stderr, err := util.RunOVNNbctl("--if-exist", "get", "logical_router_port", "rtoj-"+OvnClusterRouter, "mac")
	if err != nil {
		logrus.Errorf("Failed to get logical router port rtoj-%v, stderr: %q, error: %v", OvnClusterRouter, stderr, err)
		return err
	}
	if routerMac == "" {
		routerMac = util.GenerateMac()
		stdout, stderr, err = util.RunOVNNbctl("--", "--may-exist", "lrp-add", OvnClusterRouter,
			"rtoj-"+OvnClusterRouter, routerMac, "100.64.1.1/24", "--", "set", "logical_router_port",
			"rtoj-"+OvnClusterRouter, "external_ids:connect_to_join=yes")
		if err != nil {
			logrus.Errorf("Failed to add logical router port rtoj-%v, stdout: %q, stderr: %q, error: %v",
				OvnClusterRouter, stdout, stderr, err)
			return err
		}
	}

	// Connect the switch "join" to the router.
	stdout, stderr, err = util.RunOVNNbctl("--", "--may-exist", "lsp-add", "join", "jtor-"+OvnClusterRouter,
		"--", "set", "logical_switch_port", "jtor-"+OvnClusterRouter, "type=router",
		"options:router-port=rtoj-"+OvnClusterRouter, "addresses="+"\""+routerMac+"\"")
	if err != nil {
		logrus.Errorf("Failed to add router-type logical switch port to join, stdout: %q, stderr: %q, error: %v",
			stdout, stderr, err)
		return err
	}

	// Create a lock for gateway-init to co-ordinate.
	stdout, stderr, err = util.RunOVNNbctl("--", "set", "nb_global", ".",
		"external-ids:gateway-lock=\"\"")
	if err != nil {
		logrus.Errorf("Failed to create lock for gateways, "+
			"stdout: %q, stderr: %q, error: %v", stdout, stderr, err)
		return err
	}

	return nil
}

func (cluster *OvnClusterController) addNode(node *kapi.Node) error {
	var networks string
	for key, _ := range cluster.ClusterNetList {
		if networks == "" {
			networks = fmt.Sprintf("%s", key)
		} else {
			networks = fmt.Sprintf("%s %s", networks, key)
		}

		annotation := fmt.Sprintf(`{\"subnet\":\"%s\", \"network_allocate\":\"%s\", \"init_gateways\":\"%s\",\"gateway_itf\":\"%s\", \"gateway_type\":\"%s\", \"gateway_spare_interface\":\"%s\",  \"gateway_nexthop\":\"%s\", \"gateway_net\":\"%s\",  \"l2gateway_chassis\":\"%s\", \"vlan_id\":\"%s\", \"nodeport_enable\":\"%s\", \"sriov_pf\":\"%s\", \"sriov_only\":\"%s\"}`,cluster.ClusterNetList[key].NetworkSubnet, BoolToYN(cluster.ClusterNetList[key].NetworkAllocate),
		    BoolToYN(cluster.ClusterNetList[key].GatewayInit), cluster.ClusterNetList[key].GatewayIntf,
		    cluster.ClusterNetList[key].GatewayType, BoolToYN(cluster.ClusterNetList[key].GatewaySpareIntf),
		    cluster.ClusterNetList[key].GatewayNextHop,  cluster.ClusterNetList[key].GatewayNet,
		    cluster.ClusterNetList[key].L2GatewayChassis, strconv.Itoa(int(cluster.ClusterNetList[key].NetVlanID)),
		    BoolToYN(cluster.ClusterNetList[key].NodePortEnable), cluster.ClusterNetList[key].NetSriovDev,
		    BoolToYN(cluster.ClusterNetList[key].NetSriovOnly))
		err := cluster.Kube.SetAnnotationOnNode(node, key, annotation)
		if err != nil {
			return fmt.Errorf("Error adding annotation %s for  %s: %v", key, node.Name, err)
		}
		if !cluster.ClusterNetList[key].NetworkAllocate {
			continue
		}
		// Do not create a subnet if the node already has a subnet
		subnetstr := fmt.Sprintf("%s_host_subnet", key)
		hostsubnet, ok := node.Annotations[subnetstr]
		if ok {
			// double check if the hostsubnet looks valid
			_, _, err := net.ParseCIDR(hostsubnet)
			if err == nil {
				continue
			}
		}

		// Create new subnet
		for _, possibleSubnet := range cluster.ClusterNetList[key].masterSubnetAllocatorList {
			sn, err := possibleSubnet.GetNetwork()
			if err == netutils.ErrSubnetAllocatorFull {
				logrus.Infof("Subnet %s exhausted", key)
				// Current subnet exhausted, check next possible subnet
				continue
			} else if err != nil {
				return fmt.Errorf("Error allocating network for node %s: %v", node.Name, err)
			} else {
				err = cluster.Kube.SetAnnotationOnNode(node, subnetstr, sn.String())
				if err != nil {
					// XXX TODOS Release all subnets
					_ = possibleSubnet.ReleaseNetwork(sn)
					return fmt.Errorf("Error creating subnet %s for node %s: %v", sn.String(), node.Name, err)
				}
				logrus.Infof("Created HostSubnet %s", sn.String())
			}
		}
	}
	if networks != "" {
		err := cluster.Kube.SetAnnotationOnNode(node, "networks", networks)
		if err != nil {
			// XXX TODOS Release all subnets
			return fmt.Errorf("Error creating adding networks annotation %s for  %s: %v", networks, node.Name, err)
		}
	}
	return nil
}

func (cluster *OvnClusterController) deleteNode(node *kapi.Node) error {
	for key, _ := range cluster.ClusterNetList {
		subnetstr := fmt.Sprintf("%s_host_subnet", key)
		sub, ok := node.Annotations[subnetstr]
		if !ok {
			return fmt.Errorf("Error in obtaining host subnet for node %q for deletion", node.Name)
		}

		_, subnet, err := net.ParseCIDR(sub)
		if err != nil {
			return fmt.Errorf("Error in parsing hostsubnet - %v", err)
		}
		for _, possibleSubnet := range cluster.ClusterNetList[key].masterSubnetAllocatorList {
			err = possibleSubnet.ReleaseNetwork(subnet)
			if err == nil {
				logrus.Infof("Deleted HostSubnet %s for node %s", sub, node.Name)
				break
			}
		}
		// XXX TODOS if we didn't find the subnet, then we neeed to return an error
	}
	// SubnetAllocator.network is an unexported field so the only way to figure out if a subnet is in a network is to try and delete it
	// if deletion succeeds then stop iterating, if the list is exhausted the node subnet wasn't deleteted return err
	return nil
}

func (cluster *OvnClusterController) watchNodes() error {
	_, err := cluster.watchFactory.AddNodeHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			node := obj.(*kapi.Node)
			logrus.Debugf("Added event for Node %q", node.Name)
			err := cluster.addNode(node)
			if err != nil {
				logrus.Errorf("error creating subnet for node %s: %v", node.Name, err)
			}
		},
		UpdateFunc: func(old, new interface{}) {},
		DeleteFunc: func(obj interface{}) {
			node := obj.(*kapi.Node)
			logrus.Debugf("Delete event for Node %q", node.Name)
			err := cluster.deleteNode(node)
			if err != nil {
				logrus.Errorf("Error deleting node %s: %v", node.Name, err)
			}
			err = util.RemoveNode(node.Name)
			if err != nil {
				logrus.Errorf("Failed to remove node %s (%v)", node.Name, err)
			}
		},
	}, nil)
	return err
}
