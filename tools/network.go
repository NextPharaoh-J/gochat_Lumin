package tools

import (
	"fmt"
	"strings"
)

const (
	networkSplit = "@"
)

func ParseNetwork(bind string) (network, addr string, err error) {
	if idx := strings.Index(bind, networkSplit); idx == -1 { // 不存在@
		err = fmt.Errorf("invalid network: %s,must be network@tcp:port or network@uinxsocket", bind)
		return
	} else {
		network = bind[:idx]
		addr = bind[idx+1:]
		return
	}
}
