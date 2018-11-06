package util

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

func GetNodeLogicalSwitches(nodename string) (string, error) {
	nodeLogicalSwitches, stderr, err := RunOVNNbctl("--data=bare",
		"--no-heading", "--columns=name", "find", "logical_switch",
		"external_ids:nodename="+nodename)

	if  err != nil {
		logrus.Errorf("Failed to get logical switches for node %s, stderr: %q, "+
			"error: %v", nodename, stderr, err)
			return "", err
	}
	if nodeLogicalSwitches == "" {
		return "", fmt.Errorf("Failed to get logical switches for %s", nodename)
	}

	return nodeLogicalSwitches, nil
}

func GetNodeGateways(nodename string) (string, error) {
	nodeLogicalGateways, stderr, err := RunOVNNbctl("--data=bare",
		"--no-heading", "--columns=_uuid", "find", "logical_router",
		"external_ids:nodename="+nodename)

	if  err != nil {
		logrus.Errorf("Failed to get logical gateways for node %s, stderr: %q, "+
			"error: %v", nodename, stderr, err)
			return "", err
	}
	if nodeLogicalGateways == "" {
		return "", fmt.Errorf("Failed to get logical gateways for %s", nodename)
	}

	return nodeLogicalGateways, nil
}


// RemoveNode removes all the NB DB objects created for that node.
// Since we don't have the cluster info to walk all the networks,
// we use the external_ids field to get the logical switches and
// gateways for this node.
func RemoveNode(nodeName string) error {
	// Get the cluster router
	clusterRouter, err := GetK8sClusterRouter()
	if err != nil {
		return fmt.Errorf("failed to get cluster router")
	}

	logicalSwitches, err := GetNodeLogicalSwitches(nodeName)
	if err != nil {
		return fmt.Errorf("failed to get logical swithes for %s", nodeName)
	}
	logicalSwitchesList := strings.Fields(logicalSwitches)
	for _, logicalSwitch := range logicalSwitchesList {
		// XXX TODOS: Remove the router port associated with this switch
		// Remove the logical switch associated with nodeName
		_, stderr, err := RunOVNNbctl("--if-exist", "ls-del", logicalSwitch)
		if err != nil {
			// XXX TODOS might leave things in half baked state
			return fmt.Errorf("Failed to delete logical switch %s, "+
				"stderr: %q, error: %v", nodeName, stderr, err)
		}
	}
	logicalGateways, err := GetNodeGateways(nodeName)
	if err != nil {
		return fmt.Errorf("failed to get logical gateways for %s", nodeName)
	}
	logicalGWList := strings.Fields(logicalGateways)
	for _, gatewayRouter := range logicalGWList {
		// XXX TODOS: remove the port to the join switch

		// gatewayRouter := fmt.Sprintf("GR_%s", nodeNetName)

		// Get the gateway router port's IP address (connected to join switch)
		var routerIP string
		routerIPNetwork, stderr, err := RunOVNNbctl("--if-exist", "get",
			"logical_router_port", "rtoj-"+gatewayRouter, "networks")
		if err != nil {
			return fmt.Errorf("Failed to get logical router port, stderr: %q, "+
				"error: %v", stderr, err)
		}

		if routerIPNetwork != "" {
			routerIPNetwork = strings.Trim(routerIPNetwork, "[]\"")
			if routerIPNetwork != "" {
				routerIP = strings.Split(routerIPNetwork, "/")[0]
			}
		}

		if routerIP != "" {
			// Get a list of all the routes in cluster router with this gateway
			// Router as the next hop.
			var uuids string
			uuids, stderr, err = RunOVNNbctl("--data=bare", "--no-heading",
				"--columns=_uuid", "find", "logical_router_static_route",
				"nexthop="+routerIP)
			if err != nil {
				return fmt.Errorf("Failed to fetch all routes with gateway "+
					"router %s as nexthop, stderr: %q, "+
					"error: %v", gatewayRouter, stderr, err)
			}

			// Remove all the routes in cluster router with this gateway Router
			// as the nexthop.
			routes := strings.Fields(uuids)
			for _, route := range routes {
				_, stderr, err = RunOVNNbctl("--if-exists", "remove",
					"logical_router", clusterRouter, "static_routes", route)
				if err != nil {
					logrus.Errorf("Failed to delete static route %s"+
						", stderr: %q, err = %v", route, stderr, err)
					continue
				}
			}
		}

		// Remove the patch port that connects join switch to gateway router
		_, stderr, err = RunOVNNbctl("--if-exist", "lsp-del", "jtor-"+gatewayRouter)
		if err != nil {
			return fmt.Errorf("Failed to delete logical switch port jtor-%s, "+
				"stderr: %q, error: %v", gatewayRouter, stderr, err)
		}

		// Remove the patch port that connects distributed router to node's logical switch
		_, stderr, err = RunOVNNbctl("--if-exist", "lrp-del", "rtos-"+nodeName)
		if err != nil {
			return fmt.Errorf("Failed to delete logical router port rtos-%s, "+
				"stderr: %q, error: %v", nodeName, stderr, err)
		}

		// Remove any gateway routers associated with nodeName
		_, stderr, err = RunOVNNbctl("--if-exist", "lr-del",
			gatewayRouter)
		if err != nil {
			return fmt.Errorf("Failed to delete gateway router %s, stderr: %q, "+
				"error: %v", gatewayRouter, stderr, err)
		}


		// Remove external switch
		// XXX Would have been deleted above when we removed all the logical switches for
		// this node

		// externalSwitch := "ext_" + nodeNetName
		// _, stderr, err = RunOVNNbctl("--if-exist", "ls-del",
		//	externalSwitch)
		// if err != nil {
		//	return fmt.Errorf("Failed to delete external switch %s, stderr: %q, "+
		//		"error: %v", externalSwitch, stderr, err)
		//}
	}
	return nil
}
