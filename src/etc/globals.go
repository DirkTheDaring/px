package etc

// --------------------------------------------------------------

var GlobalPxCluster *PxCluster

// --------------------------------------------------------------

func InOrSkipIfEmpty(haystack []string, needle string) bool {
	// we have the needle if the haystack is empty ...
	if len(haystack) == 0 {
		return true
	}
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
