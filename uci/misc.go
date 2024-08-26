package uci

func rawClone[T any](v *T) *T {
	res := *v
	return &res
}
