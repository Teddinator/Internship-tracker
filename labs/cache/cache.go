package cache

func sequentialSum(data []int64) int64 {
	var sum int64

	for _, value := range data {
		sum += value
	}

	return sum
}

func randomSum(data []int64, indexes []int) int64 {
	var sum int64

	for _, index := range indexes {
		sum += data[index]
	}

	return sum
}
