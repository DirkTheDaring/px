package shared

func Pow(base int64, pow int64) int64 {
	if pow == 0 {
		return 1
	}
	var n int64
	var i int64
	n = 1
	for i = 1; i <= pow; i++ {
		n = n * base
	}
	return n
}
