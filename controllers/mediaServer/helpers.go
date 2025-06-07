package mediaserver

func bytesToGB(bytes int64) float64 {
	const bytesInGB = 1024 * 1024 * 1024
	return float64(bytes) / float64(bytesInGB)
}
