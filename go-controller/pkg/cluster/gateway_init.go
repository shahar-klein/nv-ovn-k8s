package cluster

import (
	"net"
	"fmt"
	"strings"

	"github.com/openvswitch/ovn-kubernetes/go-controller/pkg/util"

)

// getIPv4Address returns the ipv4 address for the network interface 'iface'.
func getIPv4Address(iface string) (string, error) {
	var ipAddress string
	intf, err := net.InterfaceByName(iface)
	if err != nil {
		return ipAddress, err
	}

	addrs, err := intf.Addrs()
	if err != nil {
		return ipAddress, err
	}

	for _, addr := range addrs {
		switch ip := addr.(type) {
		case *net.IPNet:
			if ip.IP.To4() != nil {
				ipAddress = ip.String()
			}
		}
	}
	return ipAddress, nil
}

// XXX-Experimental: TO exit out if the logical to the physical at L2. This function
// just initializes the physical bridge on the node. We have already configured the
// logical components for the network when setting up the networks in the master.
func  initL2Gateway (GatewayIntf string) error {
	_, _, err := util.RunOVSVsctl("--", "br-exists", GatewayIntf)
	if err != nil {
		// This is not a OVS bridge. We need to create a OVS bridge
		// and add GatewayIntf as a port of that bridge.
		_, err := util.NicToBridge(GatewayIntf)
		if err != nil {
			return fmt.Errorf("failed to convert %s to OVS bridge: %v",
				GatewayIntf, err)
		}
	} else {
		_, err = getIntfName(GatewayIntf)
		if err != nil {
			return fmt.Errorf("failed to get ofport for %s, error: %v",
				GatewayIntf, err)
		}
	}

	return nil
}

func (cluster *OvnClusterController) initGateway(
	nodeName string, clusterIPSubnet []string, subnet, netname, GatewayType, GatewayItf, GatewayNet, GatewayChassis string, vlanid uint32) error {

	nodeNetName := fmt.Sprintf("%s-%s", nodeName, netname)
	nodeNetName = strings.ToLower(nodeNetName)
	if GatewayType == "l2localnet" || GatewayType == "l2gateway" {
		return initL2Gateway(GatewayItf)
	}

	if GatewayType == "l3localnet" {
		return initLocalnetGateway(nodeName, netname, clusterIPSubnet,
			subnet, cluster.ClusterNetList[netname].NodePortEnable, vlanid)
	}

	if cluster.ClusterNetList[netname].GatewayNextHop == "" || cluster.ClusterNetList[netname].GatewayIntf == "" {
		// We need to get the interface details from the default gateway.
		GatewayIntf, GatewayNextHop, err := getDefaultGatewayInterfaceDetails()
		if err != nil {
			return err
		}

		if cluster.ClusterNetList[netname].GatewayNextHop == "" {
			cluster.ClusterNetList[netname].GatewayNextHop = GatewayNextHop
		}

		if cluster.ClusterNetList[netname].GatewayIntf == "" {
			cluster.ClusterNetList[netname].GatewayIntf = GatewayIntf
		}
	}

	if cluster.ClusterNetList[netname].GatewaySpareIntf {
		return initSpareGateway(nodeName, netname, clusterIPSubnet,
			subnet, cluster.ClusterNetList[netname].GatewayNextHop,
			cluster.ClusterNetList[netname].GatewayIntf,
			cluster.ClusterNetList[netname].NodePortEnable, vlanid)
	}

	bridge, gwIntf, err := initSharedGateway(nodeName, netname,
		clusterIPSubnet, subnet,
		cluster.ClusterNetList[netname].GatewayNextHop,
		cluster.ClusterNetList[netname].GatewayIntf,
		cluster.ClusterNetList[netname].NodePortEnable,
		vlanid, cluster.watchFactory)
	if err != nil {
		return err
	}
	cluster.ClusterNetList[netname].GatewayBridge = bridge
	cluster.ClusterNetList[netname].GatewayIntf = gwIntf
	return nil
}
