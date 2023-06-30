package slice

// Map is a generic function that maps a slice of type T to a slice of type U.
func Map[T, U any](ts []T, f func(T) U) []U {
	us := make([]U, len(ts))
	for i := range ts {
		us[i] = f(ts[i])
	}
	return us
}

// ForEach is a generic function that iterates over a slice of type T.
func ForEach[T any](ts []T, f func(i int, t T)) {
	for i, t := range ts {
		f(i, t)
	}
}
