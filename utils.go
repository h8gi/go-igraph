package igraph

import "C"

func booltoint(in bool) C.int {
	if in {
		return C.int(1)
	}
	return C.int(0)
}
