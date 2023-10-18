package version

import (
	"fmt"
	"os"
)

func getNamesOfFiles() []string {
	return []string{
		"/etc/issue",
		"/etc/resolv.conf",
		"/proc/meminfo",
		"/proc/cpuinfo",
		"/etc/timezone",
		"/proc/diskstats",
		"/proc/loadavg",
		"/proc/version",
		"/proc/uptime",
		"/etc/hyperledger/fabric/client.crt",
		"/etc/hyperledger/fabric/peer.crt",
	}
}

// SystemEnv returns the system environment
func SystemEnv() map[string]string {
	res := make(map[string]string)
	for _, name := range getNamesOfFiles() {
		b, err := os.ReadFile(name)
		switch {
		case err != nil:
			res[name] = fmt.Sprintf("error:'%v'", err)
		case len(b) == 0:
			res[name] = "file is empty"
		default:
			res[name] = string(b)
		}
	}
	return res
}
