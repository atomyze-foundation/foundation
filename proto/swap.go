package proto

import "strings"

func (x *Swap) TokenSymbol() string {
	parts := strings.Split(x.Token, "_")
	return parts[0]
}
