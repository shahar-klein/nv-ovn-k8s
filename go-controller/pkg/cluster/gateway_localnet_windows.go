// +build windows

package cluster

import (
	"fmt"
)

func initLocalnetGateway(nodeName, netname string, clusterIPSubnet []string,
	subnet string, nodePortEnable bool) error {
	// TODO: Implement this
	return fmt.Errorf("Not implemented yet on Windows")
}
