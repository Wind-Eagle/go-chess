package chess

func isASCII(s string) bool {
	for i := range len(s) {
		if s[i] >= 128 {
			return false
		}
	}
	return true
}

func abs[T ~int](a T) T {
	if a >= 0 {
		return a
	} else {
		return -a
	}
}
