package utils

func Map[T, U any](ts []T, f func(T) U) []U {
	us := make([]U, len(ts))
	for i := range ts {
		us[i] = f(ts[i])
	}
	return us
}

func Filter[T any](ts []T, f func(T) bool) []T {
	us := make([]T, 0)
	for i, val := range ts {
		if f(ts[i]) {
			us = append(us, val)
		}
	}
	return us
}
