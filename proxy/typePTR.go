package proxy

import (
	"fmt"
	"net"
	"strings"
)

func ConvPTRtoIP(ptr string) string {
	a := strings.Split(strings.TrimSuffix(ptr, "."), ".")
	result := ""
	if len(a) == 4 {
		result = fmt.Sprintf("%s.%s.%s.%s", a[3], a[2], a[1], a[0])
	}
	if len(a) == 32 {
		var buf strings.Builder
		for i := len(a) - 1; i >= 0; i-- {
			buf.WriteString(a[i])
			if i%4 == 0 && i != 0 {
				buf.WriteString(":")
			}
		}
		result = buf.String()
	}
	return net.ParseIP(result).String()
}
