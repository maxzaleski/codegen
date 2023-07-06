package slice

// Map is a generic function that maps a slice of type T to a slice of type U.
func Map[T, U any](ts []T, f func(T) U) []U {
	us := make([]U, len(ts))
	for i := range ts {
		us[i] = f(ts[i])
	}
	return us
}

// MapKeys is a generic function that maps a map of type T to a slice of type T.
func MapKeys[T comparable, U any](m map[T]U) []T {
	keys := make([]T, len(m))

	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}

// Filter is a generic function that filters a slice of type T.
func Filter[T any](ts []T, f func(T) bool) []T {
	var us []T
	for _, t := range ts {
		if f(t) {
			us = append(us, t)
		}
	}
	return us
}
