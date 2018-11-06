package ovn

import (
	"fmt"
	"strings"
	"time"
	"regexp"
	"encoding/json"

	util "github.com/openvswitch/ovn-kubernetes/go-controller/pkg/util"
	"github.com/sirupsen/logrus"
	kapi "k8s.io/api/core/v1"
)

// XXX TODOS only done for default network
func (oc *Controller) syncPods(pods []interface{}) {
	// get the list of logical switch ports (equivalent to pods)
	expectedLogicalPorts := make(map[string]bool)
	for _, podInterface := range pods {
		pod, ok := podInterface.(*kapi.Pod)
		if !ok {
			logrus.Errorf("Spurious object in syncPods: %v", podInterface)
			continue
		}
		logicalPort := fmt.Sprintf("%s_%s-ovn", pod.Namespace, pod.Name)
		expectedLogicalPorts[logicalPort] = true
	}

	// get the list of logical ports from OVN
	output, stderr, err := util.RunOVNNbctl("--data=bare", "--no-heading",
		"--columns=name", "find", "logical_switch_port", "external_ids:pod=true")
	if err != nil {
		logrus.Errorf("Error in obtaining list of logical ports, "+
			"stderr: %q, err: %v",
			stderr, err)
		return
	}
	existingLogicalPorts := strings.Fields(output)
	for _, existingPort := range existingLogicalPorts {
		if _, ok := expectedLogicalPorts[existingPort]; !ok {
			// not found, delete this logical port
			logrus.Infof("Stale logical port found: %s. This logical port will be deleted.", existingPort)
			out, stderr, err := util.RunOVNNbctl("--if-exists", "lsp-del",
				existingPort)
			if err != nil {
				logrus.Errorf("Error in deleting pod's logical port "+
					"stdout: %q, stderr: %q err: %v",
					out, stderr, err)
			}
			if !oc.portGroupSupport {
				oc.deletePodAcls(existingPort)
			}
		}
	}
}

func (oc *Controller) deletePodAcls(logicalPort string) {
	// delete the ACL rules on OVN that corresponding pod has been deleted
	uuids, stderr, err := util.RunOVNNbctl("--data=bare", "--no-heading",
		"--columns=_uuid", "find", "ACL",
		fmt.Sprintf("external_ids:logical_port=%s", logicalPort))
	if err != nil {
		logrus.Errorf("Error in getting list of acls "+
			"stdout: %q, stderr: %q, error: %v", uuids, stderr, err)
		return
	}

	if uuids == "" {
		logrus.Debugf("deletePodAcls: returning because find " +
			"returned no ACLs")
		return
	}

	uuidSlice := strings.Fields(uuids)
	for _, uuid := range uuidSlice {
		// Get logical switch
		out, stderr, err := util.RunOVNNbctl("--data=bare",
			"--no-heading", "--columns=_uuid", "find", "logical_switch",
			fmt.Sprintf("acls{>=}%s", uuid))
		if err != nil {
			logrus.Errorf("find failed to get the logical_switch of acl "+
				"uuid=%s, stderr: %q, (%v)", uuid, stderr, err)
			continue
		}

		if out == "" {
			continue
		}
		logicalSwitch := out

		_, stderr, err = util.RunOVNNbctl("--if-exists", "remove",
			"logical_switch", logicalSwitch, "acls", uuid)
		if err != nil {
			logrus.Errorf("failed to delete the allow-from rule %s for"+
				" logical_switch=%s, logical_port=%s, stderr: %q, (%v)",
				uuid, logicalSwitch, logicalPort, stderr, err)
			continue
		}
	}
}

func (oc *Controller) getLogicalPortUUID(logicalPort string) string {
	if oc.logicalPortUUIDCache[logicalPort] != "" {
		return oc.logicalPortUUIDCache[logicalPort]
	}

	out, stderr, err := util.RunOVNNbctl("--if-exists", "get",
		"logical_switch_port", logicalPort, "_uuid")
	if err != nil {
		logrus.Errorf("Error while getting uuid for logical_switch_port "+
			"%s, stderr: %q, err: %v", logicalPort, stderr, err)
		return ""
	}

	if out == "" {
		return out
	}

	oc.logicalPortUUIDCache[logicalPort] = out
	return oc.logicalPortUUIDCache[logicalPort]
}

func (oc *Controller) getGatewayFromSwitch(logicalSwitch string) (string, string, error) {
	var gatewayIPMaskStr, stderr string
	var ok bool
	var err error

	oc.lsMutex.Lock()
	defer oc.lsMutex.Unlock()
	if gatewayIPMaskStr, ok = oc.gatewayCache[logicalSwitch]; !ok {
		gatewayIPMaskStr, stderr, err = util.RunOVNNbctl("--if-exists",
			"get", "logical_switch", logicalSwitch,
			"external_ids:gateway_ip")
		if err != nil {
			logrus.Errorf("Failed to get gateway IP:  %s, stderr: %q, %v",
				gatewayIPMaskStr, stderr, err)
			return "", "", err
		}
		if gatewayIPMaskStr == "" {
			return "", "", fmt.Errorf("Empty gateway IP in logical switch %s",
				logicalSwitch)
		}
		oc.gatewayCache[logicalSwitch] = gatewayIPMaskStr
	}
	gatewayIPMask := strings.Split(gatewayIPMaskStr, "/")
	gatewayIP := gatewayIPMask[0]
	mask := gatewayIPMask[1]
	logrus.Debugf("Gateway IP: %s, Mask: %s", gatewayIP, mask)
	return gatewayIP, mask, nil
}

//Parse the string of the form
// '[
//            { "name": "network1" }
//            { "name": "network2" }
//            ...
//    ]'
// into a list of
// ["ovn", "network1", "network2", ...]
// Note "ovn" is the default K8s OVN network, the rest are additional interfaces
// we want to add to the pod.
//
func parsePodNetworkAnnotation(panotate string) []string {
	var netlist []string

	// ovn is present, by default.
	netlist = append(netlist, "ovn")

	reg, _ := regexp.Compile("[^a-z,:A-Z0-9]+")
	n1 := reg.ReplaceAllString(panotate, "") // creates a comma seperated list, i.e. name:net1,name:net2
	ns := strings.Split(n1, ",")
	for _, n := range ns {
		networkName := strings.Split(n,":")[1]
		netlist = append(netlist, networkName)
        }

	 return netlist
}

func (oc *Controller) deleteLogicalPort(pod *kapi.Pod) {
	var nelist []string

	if pod.Spec.HostNetwork {
		return
	}

	logrus.Infof("Deleting pod: %s", pod.Name)

	// Get others, if any
	pannotation, ok := pod.Annotations["k8s.v1.cni.cncf.io/networks"]
	if !ok {
		nelist = append(nelist, "ovn")
	} else {
		nelist = parsePodNetworkAnnotation(pannotation)
		if len(nelist) == 0 {
			nelist = append(nelist, "ovn")
		}
	}

	for i := range nelist {
		netname := nelist[i]
		logicalPort := fmt.Sprintf("%s_%s-%s", pod.Namespace, pod.Name, netname)
		out, stderr, err := util.RunOVNNbctl("--if-exists", "lsp-del",
			logicalPort)
		if err != nil {
			logrus.Errorf("Error in deleting pod logical port "+
				"stdout: %q, stderr: %q, (%v)",
				out, stderr, err)
		}

		ipAddress := oc.getIPFromOvnAnnotation(pod.Annotations[netname])

		delete(oc.logicalPortCache, logicalPort)

		oc.lspMutex.Lock()
		delete(oc.lspIngressDenyCache, logicalPort)
		delete(oc.lspEgressDenyCache, logicalPort)
		delete(oc.logicalPortUUIDCache, logicalPort)
		oc.lspMutex.Unlock()

		if !oc.portGroupSupport {
			oc.deleteACLDenyOld(pod.Namespace, pod.Spec.NodeName, logicalPort,
				"Ingress")
			oc.deleteACLDenyOld(pod.Namespace, pod.Spec.NodeName, logicalPort,
				"Egress")
		}
		oc.deletePodFromNamespaceAddressSet(pod.Namespace, ipAddress)
	}
	return
}

func (oc *Controller) addLogicalPort(pod *kapi.Pod) {
	var out, stderr string
	var err error
	var nelist []string

	if pod.Spec.HostNetwork {
		logrus.Debugf("addLogicalPort: pod %s hostnetwork", pod.Name)
		return
	}

	// Get network list for the pod
	pannotation, ok := pod.Annotations["k8s.v1.cni.cncf.io/networks"]

	if !ok {
		nelist = append(nelist, "ovn")
	} else {
		nelist = parsePodNetworkAnnotation(pannotation)
		if len(nelist) == 0 {
			nelist = append(nelist, "ovn")
		}
	}

	for i := range nelist {
		netname := nelist[i]
		annotation, isStaticIP := pod.Annotations[netname]

		logicalSwitch := netname
		if netname == "ovn" {
			logicalSwitch = fmt.Sprintf("%s-%s", pod.Spec.NodeName, netname)
		}
		if logicalSwitch == "" {
			logrus.Errorf("Failed to find the logical switch for pod %s/%s/%s",
				pod.Namespace, pod.Name, netname)
			return
		}

		oc.lsMutex.Lock()
		if !oc.logicalSwitchCache[logicalSwitch] {
			oc.logicalSwitchCache[logicalSwitch] = true
			oc.addAllowACLFromNode(logicalSwitch)
		}
		oc.lsMutex.Unlock()

		portName := fmt.Sprintf("%s_%s-%s", pod.Namespace, pod.Name, netname)
		logrus.Debugf("Creating logical port for %s on switch %s for network %s", portName, logicalSwitch, netname)

		// If pod already has annotations, just add the lsp with static ip/mac.
		// Else, create the lsp with dynamic addresses.
		// XXX TODOS: Static IP check for other networks (other than ovn)
		if isStaticIP {
			ipAddress := oc.getIPFromOvnAnnotation(annotation)
			macAddress := oc.getMacFromOvnAnnotation(annotation)

			out, stderr, err = util.RunOVNNbctl("--may-exist", "lsp-add",
				logicalSwitch, portName, "--", "lsp-set-addresses", portName,
				fmt.Sprintf("%s %s", macAddress, ipAddress), "--", "--if-exists",
				"clear", "logical_switch_port", portName, "dynamic_addresses")
			if err != nil {
				logrus.Errorf("Failed to add logical port to switch "+
					"stdout: %q, stderr: %q (%v)",
					out, stderr, err)
				return
			}
		} else {
			out, stderr, err = util.RunOVNNbctl("--wait=sb", "--",
				"--may-exist", "lsp-add", logicalSwitch, portName,
				"--", "lsp-set-addresses",
				portName, "dynamic", "--", "set",
				"logical_switch_port", portName,
				"external-ids:namespace="+pod.Namespace,
				"external-ids:logical_switch="+logicalSwitch,
				"external-ids:pod=true")
			if err != nil {
				logrus.Errorf("Error while creating logical port %s "+
					"stdout: %q, stderr: %q (%v)",
					portName, out, stderr, err)
				return
			}
		}

		oc.logicalPortCache[portName] = logicalSwitch

		gatewayIP, mask, err := oc.getGatewayFromSwitch(logicalSwitch)
		if err != nil {
			logrus.Errorf("Error obtaining gateway address for switch %s", logicalSwitch)
			return
		}

		count := 30
		for count > 0 {
			if isStaticIP {
				out, stderr, err = util.RunOVNNbctl("get",
					"logical_switch_port", portName, "addresses")
			} else {
				out, stderr, err = util.RunOVNNbctl("get",
					"logical_switch_port", portName, "dynamic_addresses")
			}
			if err == nil && out != "[]" {
				break
			}
			if err != nil {
				logrus.Errorf("Error while obtaining addresses for %s - %v", portName,
					err)
				return
			}
			time.Sleep(time.Second)
			count--
		}
		if count == 0 {
			logrus.Errorf("Error while obtaining addresses for %s "+
				"stdout: %q, stderr: %q, (%v)", portName, out, stderr, err)
			return
		}

		// static addresses have format ["0a:00:00:00:00:01 192.168.1.3"], while
		// dynamic addresses have format "0a:00:00:00:00:01 192.168.1.3".
		outStr := strings.TrimLeft(out, `[`)
		outStr = strings.TrimRight(outStr, `]`)
		outStr = strings.Trim(outStr, `"`)
		addresses := strings.Split(outStr, " ")
		if len(addresses) != 2 {
			logrus.Errorf("Error while obtaining addresses for %s", portName)
			return
		}

		sriov_pf := ""
		sriov_only := "no"
		network_subnet := ""
		nodeAnnotation, err := oc.kube.GetAnnotationsOnNode(pod.Spec.NodeName)
		if err != nil {
			logrus.Warningf("Error while obtaining Node annotations - %v", err)
			return
		}
		currNetAnn := nodeAnnotation[netname]
		var currNetAnnMap map[string]string
		err = json.Unmarshal([]byte(currNetAnn), &currNetAnnMap)
		if err != nil {
			logrus.Errorf("unmarshal network annotation failed")
			return
		}
		sriov_pf = currNetAnnMap["sriov_pf"]
		sriov_only = currNetAnnMap["sriov_only"]
		network_subnet = currNetAnnMap["subnet"]
		if !isStaticIP {
			annotation = fmt.Sprintf(`{\"ip_address\":\"%s/%s\", \"mac_address\":\"%s\", \"gateway_ip\": \"%s\", \"sriov_pf\": \"%s\", \"sriov_only\": \"%s\", \"network_subnet\": \"%s\"}`, addresses[1], mask, addresses[0], gatewayIP, sriov_pf, sriov_only, network_subnet)
			logrus.Debugf("Annotation values: ip=%s/%s ; mac=%s ; gw=%s; sriov_pf=%s\nAnnotation=%s", addresses[1], mask, addresses[0], gatewayIP, sriov_pf, annotation)
			err = oc.kube.SetAnnotationOnPod(pod, netname, annotation)
			if err != nil {
				logrus.Errorf("Failed to set annotation on pod %s - %v", pod.Name, err)
			}
		}
		oc.addPodToNamespaceAddressSet(pod.Namespace, addresses[1])
	}

	return
}

// AddLogicalPortWithIP add logical port with static ip address
// and mac adddress for the pod
func (oc *Controller) AddLogicalPortWithIP(pod *kapi.Pod) {
	var nelist []string
	if pod.Spec.HostNetwork {
		return
	}

	// Get network list for the pod
	pannotation, ok := pod.Annotations["k8s.v1.cni.cncf.io/networks"]
	if !ok {
		nelist = append(nelist, "ovn")
	} else {
		nelist = parsePodNetworkAnnotation(pannotation)
		if len(nelist) == 0 {
			nelist = append(nelist, "ovn")
		}
	}

	for i := range nelist {
		netname := nelist[i]

		// portName := fmt.Sprintf("%s_%s", pod.Namespace, pod.Name)
		// logicalSwitch := pod.Spec.NodeName
		portName := fmt.Sprintf("%s_%s-%s", pod.Namespace, pod.Name, netname)
		logicalSwitch := netname
		if netname == "ovn" {
			logicalSwitch = fmt.Sprintf("%s-%s", pod.Spec.NodeName, netname)
		}
		logrus.Debugf("Creating logical port for %s on switch %s", portName,
			logicalSwitch)

		annotation, ok := pod.Annotations[netname]
		if !ok {
			logrus.Errorf("Failed to get ovn annotation from pod!")
			return
		}
		ipAddress := oc.getIPFromOvnAnnotation(annotation)
		macAddress := oc.getMacFromOvnAnnotation(annotation)

		stdout, stderr, err := util.RunOVNNbctl("--", "--may-exist", "lsp-add",
			logicalSwitch, portName, "--", "lsp-set-addresses", portName,
			fmt.Sprintf("%s %s", macAddress, ipAddress))
		if err != nil {
			logrus.Errorf("Failed to add logical port to switch, stdout: %q, "+
				"stderr: %q, error: %v", stdout, stderr, err)
			return
		}
	}
}
