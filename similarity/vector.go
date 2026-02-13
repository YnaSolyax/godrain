package similarity

import (
	"math"
	"strings"
)

func ToVector(log string) map[string]float64 {
	logFields := strings.Fields(log)
	vector := make(map[string]float64, 10)

	for _, l := range logFields {
		vector[l]++
	}

	return vector
}

func FoundSimilar(vector1, vector2 map[string]float64) float64 {
	vector1sum, vector2sum, sum := 0.0, 0.0, 0.0

	for word, count1 := range vector1 {
		if count2, ok := vector2[word]; ok {
			sum += count1 * count2
		}

		vector1sum += count1 * count1
	}

	for _, count2 := range vector2 {
		vector2sum += count2 * count2
	}

	if vector1sum == 0 || vector2sum == 0 {
		return 0
	}

	return sum / (math.Sqrt(vector1sum) * math.Sqrt(vector2sum))
}
