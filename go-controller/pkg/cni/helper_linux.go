// +build linux

package cni

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"strconv"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/openvswitch/ovn-kubernetes/go-controller/pkg/util"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink"
	"github.com/Mellanox/sriovnet"
)

// getIPv4 returns the ipv4 address for the network interface 'iface'.
// XXX-TODO dup from gateway_init.go, need to use it from a util
func getIPv4(iface string) (string, error) {
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

func renameLink(curName, newName string) error {
	link, err := netlink.LinkByName(curName)
	if err != nil {
		return err
	}

	if err := netlink.LinkSetDown(link); err != nil {
		return err
	}
	if err := netlink.LinkSetName(link, newName); err != nil {
		return err
	}
	if err := netlink.LinkSetUp(link); err != nil {
		return err
	}

	return nil
}

func get_encap_id(itf string) (string, error) {
	ipAddress, err := getIPv4(itf)
	if err != nil {
		return "", fmt.Errorf("failed to get encap_ip for %s", itf)
	}
	if ipAddress == "" {
		britf := fmt.Sprintf("br%s", itf)
		ipAddress, err = getIPv4(britf)
		if err != nil {
			return "", fmt.Errorf("failed to get encap_ip for %s", britf)
		}
	}
	if ipAddress != "" {
		// Check if this is in the list of encap-ips.
		ip, _, err := net.ParseCIDR(ipAddress)
		if err != nil {
			return "", fmt.Errorf("failed parsing ip %s", ipAddress)
		}
		ovn_encap_ip, _, _ := util.RunOVSVsctl("get",
			"Open_vSwitch",
			".",
			fmt.Sprintf("external_ids:ovn-encap-ip"))
		encap_ips := strings.Split(ovn_encap_ip, ",")
		for _, encap_ip := range encap_ips {
			if encap_ip == ip.String() {
				return ip.String(), nil
			}
		}
	}
	return "", fmt.Errorf("failed to get encap_ip for %s", itf)
}

// XXX-TODOS Consolidate with setupInterface
func setupSRIOVInterface(netns ns.NetNS, containerID, ifName, macAddress, ipAddress, gatewayIP string, mtu int, pfName string, vfName string, vfNum int, network_subnet string) (*current.Interface, *current.Interface, error) {
	hostIface := &current.Interface{}
	contIface := &current.Interface{}

	// set hardware address
	if macAddress != "" {
		macAddr, err := net.ParseMAC(string(macAddress))
		if err != nil {
			return nil, nil, err
		}
		if err = sriovnet.UpdateVFMAC(pfName, vfNum, macAddr); err != nil {
                        return nil, nil, fmt.Errorf("failed to set vf %d macaddress: %v", vfNum, err)
                }
	}

	vfDev, err := netlink.LinkByName(vfName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to lookup vf device %q: %v", vfName, err)
	}

	if err = netlink.LinkSetUp(vfDev); err != nil {
		return nil, nil, fmt.Errorf("failed to setup vf %d device: %v", vfNum, err)
	}

	if err = netlink.LinkSetNsFd(vfDev, int(netns.Fd())); err != nil {
		return nil, nil, fmt.Errorf("failed to move vf %d to netns: %v", vfNum, err)
	}

	err = netns.Do(func(hostNS ns.NetNS) error {

		v, err := netlink.LinkByName(vfName)
		if err != nil {
			return fmt.Errorf("failed to lookup link for %q: %v", vfName, err)
		}

		contIface.Sandbox = netns.Path()
		if ipAddress != "" {
			addr, err := netlink.ParseAddr(ipAddress)
			if err != nil {
				return fmt.Errorf("failed to parse IP %s: %v", ipAddress, err)
			}
			err = netlink.AddrAdd(v, addr)
			if err != nil {
				return fmt.Errorf("failed to add IP addr %s to %s: %v", ipAddress, v, err)
			}
		}

		err = renameLink(vfName, ifName)
		if err != nil {
			return fmt.Errorf("failed to rename vf %d device %q to %q: %v", vfNum, vfName, ifName, err)
		}

		// Set the MTU
		if mtu > 0 {
			err =  netlink.LinkSetMTU(v, mtu)
			if err != nil {
				return fmt.Errorf("failed to set MTU to %d ondevice %q: %v", mtu, ifName, err)
			}
		}

		v, err = netlink.LinkByName(ifName)
		if err != nil {
			return fmt.Errorf("failed to lookup link for %q: %v", ifName, err)
		}
		// Set default gateway
		if ifName == "eth0" && gatewayIP != "" {
			gw := net.ParseIP(gatewayIP)
			if gw == nil {
				return fmt.Errorf("failed to parse gateway IP %s", gatewayIP)
			}
			err = ip.AddDefaultRoute(gw, v)
			if err != nil {
				if os.IsExist(err) {
					err = nil
				} else {
					return fmt.Errorf("failed adding route %q: %v", gatewayIP, err)
				}
			}
		}
		if ifName != "eth0"  && network_subnet != "" {
			table_str := ifName[len(ifName)-1:]
			table_no,_ := strconv.Atoi(table_str)
			// Start from 10 in case the others are being used
			// XXX-TODOS make this more generic
			table_no += 10
			subnet_ip, subnet_mask, _ := net.ParseCIDR(network_subnet)
			netmask,_ := subnet_mask.Mask.Size()
			// Create a routing table for this network: not persistent.
			// Get the default route on this interface, we'll delete them after adding rules for this
			// interface
			routes, err := netlink.RouteList(v, syscall.AF_INET)

			if gatewayIP != "" {
				dst_ip := &net.IPNet{
					IP: net.IPv4(0, 0, 0, 0),
					Mask: net.CIDRMask(0, 32),
				}
				gw_ip := net.ParseIP(gatewayIP)
				route := netlink.Route{LinkIndex: v.Attrs().Index, Dst: dst_ip, Gw: gw_ip, Table: table_no}
				// XXX-TODOS Treat this as error
				if err = netlink.RouteAdd(&route); err != nil {
					logrus.Debugf("Error adding subnet-based default route, %s,  for %s on table %d", gatewayIP, network_subnet, table_no)
					return nil
				}
			}
			src_ip := &net.IPNet{
				IP: subnet_ip,
				Mask: net.CIDRMask(netmask, 32),
			}

			// XXX-TODOS: IF we are using physical networks, then the routes need to be adjusted
			// based on the peer network.
			// XXX-TODOS Check if the new "Invert" allows adding only one rule
			s_rule := netlink.NewRule()
			s_rule.Table = table_no
			s_rule.Src = src_ip
			if err = netlink.RuleAdd(s_rule); err != nil {
				return fmt.Errorf("failed adding source rule %q: %v", ipAddress, err)
			}
			d_rule := netlink.NewRule()
			d_rule.Table = table_no
			d_rule.Dst = src_ip
			if err = netlink.RuleAdd(d_rule); err != nil {
				return fmt.Errorf("failed adding dst rule %q: %v", ipAddress, err)
			}
			// Now delete the pre-existing rules
			if len (routes) > 0 {
				for _, route := range routes {
					// XXX-TODOS Treat this as error.
					_ = netlink.RouteDel(&route)
				}
			}

		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	// rename the host end of veth pair. 
	representor := fmt.Sprintf("%s_%d", pfName, vfNum)
	hostIface.Name = representor
	// XXX-TODOS: get the MAC from the representor.
	hostIface.Mac = "2e:7c:f9:d2:41:4c"
	contIface.Mac = macAddress

	contIface.Name = ifName
	return hostIface, contIface, nil
}

func setupInterface(netns ns.NetNS, containerID, ifName, macAddress, ipAddress, gatewayIP string, mtu int, network_subnet string) (*current.Interface, *current.Interface, error) {
	hostIface := &current.Interface{}
	contIface := &current.Interface{}

	var oldHostVethName string
	err := netns.Do(func(hostNS ns.NetNS) error {
		// create the veth pair in the container and move host end into host netns
		hostVeth, containerVeth, err := ip.SetupVeth(ifName, mtu, hostNS)
		if err != nil {
			return err
		}
		hostIface.Mac = hostVeth.HardwareAddr.String()
		contIface.Name = containerVeth.Name

		link, err := netlink.LinkByName(contIface.Name)
		if err != nil {
			return fmt.Errorf("failed to lookup %s: %v", contIface.Name, err)
		}

		hwAddr, err := net.ParseMAC(macAddress)
		if err != nil {
			return fmt.Errorf("failed to parse mac address for %s: %v", contIface.Name, err)
		}
		err = netlink.LinkSetHardwareAddr(link, hwAddr)
		if err != nil {
			return fmt.Errorf("failed to add mac address %s to %s: %v", macAddress, contIface.Name, err)
		}
		contIface.Mac = macAddress
		contIface.Sandbox = netns.Path()

		addr, err := netlink.ParseAddr(ipAddress)
		if err != nil {
			return err
		}
		err = netlink.AddrAdd(link, addr)
		if err != nil {
			return fmt.Errorf("failed to add IP addr %s to %s: %v", ipAddress, contIface.Name, err)
		}

		if ifName == "eth0" && gatewayIP !=  "" {
			gw := net.ParseIP(gatewayIP)
			if gw == nil {
				return fmt.Errorf("parse ip of gateway failed")
			}
			err = ip.AddRoute(nil, gw, link)
			if err != nil {
				return err
			}
		}
		if ifName != "eth0"  && network_subnet != "" {
			table_str := ifName[len(ifName)-1:]
			table_no,_ := strconv.Atoi(table_str)
			// Start from 10 in case the others are being used
			// XXX-TODOS make this more generic
			table_no += 10
			subnet_ip, subnet_mask, _ := net.ParseCIDR(network_subnet)
			netmask,_ := subnet_mask.Mask.Size()
			// Create a routing table for this network
			// Get the default route on this interface, we'll delete it after adding rules for this
			// interface
			routes, err := netlink.RouteList(link, syscall.AF_INET)

			if gatewayIP != "" {
				dst_ip := &net.IPNet{
					IP: net.IPv4(0, 0, 0, 0),
					Mask: net.CIDRMask(0, 32),
				}
				gw_ip := net.ParseIP(gatewayIP)
				route := netlink.Route{LinkIndex: link.Attrs().Index, Dst: dst_ip, Gw: gw_ip, Table: table_no}
				// XXX-TODOS Treat this as error
				if err = netlink.RouteAdd(&route); err != nil {
					return nil
				}
			}
			src_ip := &net.IPNet{
				IP: subnet_ip,
				Mask: net.CIDRMask(netmask, 32),
			}

			s_rule := netlink.NewRule()
			s_rule.Table = table_no
			s_rule.Src = src_ip
			if err = netlink.RuleAdd(s_rule); err != nil {
				return fmt.Errorf("failed adding source rule %q: %v", ipAddress, err)
			}
			d_rule := netlink.NewRule()
			d_rule.Table = table_no
			d_rule.Dst = src_ip
			if err = netlink.RuleAdd(d_rule); err != nil {
				return fmt.Errorf("failed adding dst rule %q: %v", ipAddress, err)
			}
			// Now delete the pre-existing rules
			if len (routes) > 0 {
				for _, route := range routes {
					// XXX-TODOS Treat this as error.
					_ = netlink.RouteDel(&route)
				}
			}
		}
		oldHostVethName = hostVeth.Name

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	// rename the host end of veth pair
	hostIface.Name = fmt.Sprintf("%s_%s", ifName, containerID[:(14-len(ifName))])
	if err := renameLink(oldHostVethName, hostIface.Name); err != nil {
		return nil, nil, fmt.Errorf("failed to rename %s to %s: %v", oldHostVethName, hostIface.Name, err)
	}

	return hostIface, contIface, nil
}

// ConfigureInterface sets up the container interface
// sriovpfs is the list of PFs that can be used to give the POD a VF. They are not currently managed ny OVN, so apart from
// adding them to the pod nothing else happens.
// sdnpf is the PF from which we will take one VF and assign to the POD and also configure it instead of using a veth pair.
// if sdnpf is not provided we will fall back to vethpair.
func (pr *PodRequest) ConfigureInterface(namespace string, podName string, macAddress string, ipAddress string, gatewayIP string, mtu int, ingress, egress int64, sdnpf string, sriovonly bool, netname, network_subnet string) ([]*current.Interface, error) {
	var hostIface *current.Interface
	var contIface *current.Interface

	SDNvf := ""
	SDNvfnum := 0
	netns, err := ns.GetNS(pr.Netns)
	if err != nil {
		return nil, fmt.Errorf("failed to open netns %q: %v", pr.Netns, err)
	}
	defer netns.Close()

	encap_ip := ""
TryPV:
	if sdnpf == "" {
		hostIface, contIface, err = setupInterface(netns, pr.SandboxID, pr.IfName, macAddress, ipAddress, gatewayIP, mtu, network_subnet)
		if err != nil {
			return nil, err
		}
	} else {
		// Let's get the IP, if any, so that even if we fallback
		// to PV, we can use the IP on the provided PF as the
		// VTEP IP.
		encap_ip,_  = get_encap_id(sdnpf)

		SDNvf, SDNvfnum, err = sriovnet.GetVF(sdnpf)
                if err != nil || SDNvf == "" {
                        if sriovonly {
                                return nil,  fmt.Errorf("failed to get vf %q", pr.Netns)
                        }
                        // Fallback to PV
                        sdnpf = ""
                        goto TryPV
                }

		hostIface, contIface, err = setupSRIOVInterface(netns, pr.SandboxID, pr.IfName, macAddress, ipAddress, gatewayIP, mtu, sdnpf, SDNvf, SDNvfnum, network_subnet)
		if err != nil {
			return nil, err
		}
	}

	ifaceID := fmt.Sprintf("%s_%s-%s", namespace, podName, netname)
	// the port might exist, esp. if the VF representor is not deleted, if so try setting the external ids
	ovsArgs := []string{
		"--may-exist", "add-port", "br-int", hostIface.Name, "--", "set",
		"interface", hostIface.Name,
		fmt.Sprintf("external_ids:attached_mac=%s", macAddress),
		fmt.Sprintf("external_ids:iface-id=%s", ifaceID),
		fmt.Sprintf("external_ids:ip_address=%s", ipAddress),
		fmt.Sprintf("external_ids:sandbox=%s", pr.SandboxID),
	}
	if out, err := ovsExec(ovsArgs...); err != nil {
		return nil, fmt.Errorf("failure in plugging pod interface: %v\n  %q", err, out)
	}

	if encap_ip != "" {
		ovsArgs := []string{
			"set", "interface", hostIface.Name,
			fmt.Sprintf("external_ids:encap-ip=%s", encap_ip),
		}
		if out, err := ovsExec(ovsArgs...); err != nil {
			return nil, fmt.Errorf("failure in plugging pod interface: %v\n  %q", err, out)
		}
	}
	if SDNvf != "" {
		ovsArgs := []string{
			"set", "interface", hostIface.Name,
			fmt.Sprintf("external_ids:sriov_pf=%s", sdnpf),
		}
		if out, err := ovsExec(ovsArgs...); err != nil {
			return nil, fmt.Errorf("failure in plugging pod interface: %v\n  %q", err, out)
		}
	}
	// XXX-TODOS: not supported for VF for now
	if SDNvf == "" {
		if err := clearPodBandwidth(pr.SandboxID); err != nil {
			return nil, err
		}
		if ingress > 0 || egress > 0 {
			l, err := netlink.LinkByName(hostIface.Name)
			if err != nil {
				return nil, fmt.Errorf("failed to find host veth interface %s: %v", hostIface.Name, err)
			}
			err = netlink.LinkSetTxQLen(l, 1000)
			if err != nil {
				return nil, fmt.Errorf("failed to set host veth txqlen: %v", err)
			}
	
			if err := setPodBandwidth(pr.SandboxID, hostIface.Name, ingress, egress); err != nil {
				return nil, err
			}
		}
	}
	return []*current.Interface{hostIface, contIface}, nil
}

// PlatformSpecificCleanup deletes the OVS port
func (pr *PodRequest) PlatformSpecificCleanup() error {
	ifaceName := pr.SandboxID[:15]

	ovsArgs := []string{
		"--data=bare", "--no-heading", "--columns=name", "find", "interface", fmt.Sprintf("external_ids:sandbox=%s", pr.SandboxID),
	}
	out, err := exec.Command("ovs-vsctl", ovsArgs...).CombinedOutput()
	if err != nil {
		return nil
	}
	ovsports := strings.Fields(string(out))
	for _, ovsport := range ovsports {
		// If it is a VF, release it.
		ovsArgs := []string{
			"get", "interface", ovsport, "external_ids:sriov_pf",
		}
		out, err := exec.Command("ovs-vsctl", ovsArgs...).CombinedOutput()
		if err == nil {
			netns, err := ns.GetNS(pr.Netns)
			if err != nil {
				return fmt.Errorf("failed to open netns %q: %v", pr.Netns, err)
			}
			err = sriovnet.ReleaseVF(pr.IfName, netns)
                        if err != nil {
                                logrus.Errorf("failed sriovnet.ReleaseVF  %v", err)
			}
		}

		ovsArgs = []string{
			"del-port", "br-int", ovsport,
		}
		out, err = exec.Command("ovs-vsctl", ovsArgs...).CombinedOutput()
		if err != nil && !strings.Contains(string(out), "no port named") {
			// DEL should be idempotent; don't return an error just log it
			logrus.Warningf("failed to delete OVS port %s: %v\n  %q", ifaceName, err, string(out))
		}
	}
	_ = clearPodBandwidth(pr.SandboxID)

	return nil
}
