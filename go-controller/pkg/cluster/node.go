package cluster

import (
	"fmt"
	"net"
	"time"
	"strings"
	"encoding/json"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/openvswitch/ovn-kubernetes/go-controller/pkg/cni"
	"github.com/openvswitch/ovn-kubernetes/go-controller/pkg/config"
	"github.com/openvswitch/ovn-kubernetes/go-controller/pkg/ovn"

	kapi "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

// StartClusterNode learns the subnet assigned to it by the master controller
// and calls the SetupNode script which establishes the logical switch
func (cluster *OvnClusterController) StartClusterNode(name string) error {
	count := 300
	var err error
	var node *kapi.Node
	var subnet *net.IPNet
	var clusterSubnets []string
	var annotation string
	var ok bool

	// cluster.ClusterNetList = make(map[string]*ClusterNets)
	for count > 0 {
		if count != 300 {
			time.Sleep(time.Second)
		}
		count--

		// setup the node, create the logical switch
		node, err = cluster.Kube.GetNode(name)
		if err != nil {
			logrus.Errorf("Error starting node %s, no node found - %v", name, err)
			continue
		}

		annotation, ok = node.Annotations["networks"]
		if !ok {
			logrus.Errorf("Error starting node %s, no annotation found on node for subnet - %v", name, err)
			continue
		}
		break
	}

	if count == 0 {
		logrus.Errorf("Failed to get node/node-annotation for %s - %v", name, err)
		return err
	}

	networks := strings.Split(annotation, " ")
	for _, network := range networks {
	       var nMap map[string]string
	       netannotate, ok := node.Annotations[network]
	       if !ok {
		       logrus.Errorf("Error getting annotation for %s on node for subnet - %v", network, name, err)
		       continue
	       }
	       err = json.Unmarshal([]byte(netannotate), &nMap)
	       // XXX-TODOS Treat this as an error.
	       if err != nil {
	              continue
	       }
               vid := uint32(0)
               vlanid, err := strconv.Atoi(nMap["vlan_id"])
	       if err == nil {
                         vid = uint32(vlanid)
               }
	       cluster.ClusterNetList[network] = &ClusterNets{NetVlanID: vid,
							       NetworkSubnet: nMap["subnet"],
							       NetSriovDev: nMap["sriov_pf"],
							       GatewayNet: nMap["gateway_net"],
							       GatewayIntf: nMap["gateway_itf"],
							       GatewayType: nMap["gateway_type"],
							       GatewayNextHop :  nMap["gateway_nexthop"],
							       L2GatewayChassis: nMap["l2gateway_chassis"],
							       GatewaySpareIntf: YNtoBool(nMap["gateway_spareintf"]),
							       NetworkAllocate: YNtoBool(nMap["network_allocate"]),
							       NodePortEnable: YNtoBool(nMap["nodeport_enable"]),
							       GatewayInit: YNtoBool(nMap["init_gateways"]),
							       NetSriovOnly: YNtoBool(nMap["sriov_only"])}
		cluster.ClusterNetList[network].NetworkIPNet, err = parseNetworkIPNetEntries(nMap["subnet"])
		// XXX-TODOS Treat this as an error.
		if err != nil {
			logrus.Debugf("Error creating subnet %s", nMap["subnet"])
		}
	}
	err = setupOVNNode(name)
	if err != nil {
		logrus.Errorf("Error: Failed setting up OVN Node %s", name)
		return err
	}
	for key, _ := range cluster.ClusterNetList {
		clusterSubnets = nil
		for _, clusterSubnet := range cluster.ClusterNetList[key].NetworkIPNet {
			clusterSubnets = append(clusterSubnets, clusterSubnet.CIDR.String())
		}
		if cluster.ClusterNetList[key].NetworkAllocate {
			subnetstr := fmt.Sprintf("%s_host_subnet", key)
			sub, ok := node.Annotations[subnetstr]
			if !ok {
				logrus.Errorf("Error starting node %s, no annotation found on node for subnet - %v", name, err)
				continue
			}
			_, subnet, err = net.ParseCIDR(sub)
			if err != nil {
				logrus.Errorf("Invalid hostsubnet found for node %s - %v", node.Name, err)
				return err
			}

			logrus.Infof("Node %s ready for ovn initialization with subnet %s", node.Name, subnet.String())

			err = ovn.CreateManagementPort(node.Name, subnet.String(),
				cluster.ClusterServicesSubnet,
				clusterSubnets, key)
			if err != nil {
				return err
			}
		}

		if cluster.ClusterNetList[key].GatewayInit {
			// XXX TODOS: Take the gateway info etc. from the
			// cluster.ClusterNetList itself instead of passing it.
			err = cluster.initGateway(node.Name, clusterSubnets,
			                          subnet.String(), key, cluster.ClusterNetList[key].GatewayType, cluster.ClusterNetList[key].GatewayIntf, cluster.ClusterNetList[key].GatewayNet, cluster.ClusterNetList[key].L2GatewayChassis, cluster.ClusterNetList[key].NetVlanID)
			if err != nil {
				return err
			}
		}

		if err = config.WriteCNIConfig(); err != nil {
			return err
		}

		if cluster.OvnHA {
			err = cluster.watchNamespaceUpdate(node, subnet.String())
			return err
		}
	}
	// start the cni server
	cniServer := cni.NewCNIServer("")
	err = cniServer.Start(cni.HandleCNIRequest)

	return err
}

// If default namespace MasterOverlayIP annotation has been chaged, update
// config.OvnNorth and config.OvnSouth auth with new ovn-nb and ovn-remote
// IP address
func (cluster *OvnClusterController) updateOvnNode(masterIP string,
	node *kapi.Node, subnet string) error {
	err := config.UpdateOvnNodeAuth(masterIP)
	if err != nil {
		return err
	}
	err = setupOVNNode(node.Name)
	if err != nil {
		logrus.Errorf("Failed to setup OVN node (%v)", err)
		return err
	}

	var clusterSubnets []string

	for key, _ := range cluster.ClusterNetList {
		clusterSubnets = nil
		for _, clusterSubnet := range cluster.ClusterNetList[key].NetworkIPNet {
			clusterSubnets = append(clusterSubnets, clusterSubnet.CIDR.String())
		}
		// Recreate logical switch and management port for this node
		// XXX TODOS get it from configuration for the network
		if cluster.ClusterNetList[key].NetworkAllocate {
			err = ovn.CreateManagementPort(node.Name, subnet,
				cluster.ClusterServicesSubnet,
				clusterSubnets,
				key)
			if err != nil {
				return err
			}
		}
		// Reinit Gateway for this node if the --init-gateways flag is set
		// XXX TODOS: Take the gateway info etc. from the
		// cluster.ClusterNetList itself instead of passing it.
		if cluster.ClusterNetList[key].GatewayInit {
			err = cluster.initGateway(node.Name, clusterSubnets,
						  subnet, key, cluster.ClusterNetList[key].GatewayType, cluster.ClusterNetList[key].GatewayIntf, cluster.ClusterNetList[key].GatewayNet, cluster.ClusterNetList[key].L2GatewayChassis, cluster.ClusterNetList[key].NetVlanID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// watchNamespaceUpdate starts watching namespace resources and calls back
// the update handler logic if there is any namspace update event
func (cluster *OvnClusterController) watchNamespaceUpdate(node *kapi.Node,
	subnet string) error {
	_, err := cluster.watchFactory.AddNamespaceHandler(
		cache.ResourceEventHandlerFuncs{
			UpdateFunc: func(old, newer interface{}) {
				oldNs := old.(*kapi.Namespace)
				oldMasterIP := oldNs.Annotations[MasterOverlayIP]
				newNs := newer.(*kapi.Namespace)
				newMasterIP := newNs.Annotations[MasterOverlayIP]
				if newMasterIP != oldMasterIP {
					err := cluster.updateOvnNode(newMasterIP, node, subnet)
					if err != nil {
						logrus.Errorf("Failed to update OVN node with new "+
							"masterIP %s: %v", newMasterIP, err)
					}
				}
			},
		}, nil)
	return err
}
