package uci

func rawClone[T any](v *T) *T {
	if v == nil {
		return nil
	}
	res := *v
	return &res
}
