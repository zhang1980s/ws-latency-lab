package stats

import (
	"fmt"
	"math"
	"sort"
)

// LatencyStats holds the calculated statistics from latency samples
type LatencyStats struct {
	Count int
	Min   int64
	P10   int64
	P50   int64
	P90   int64
	P99   int64
	Max   int64
	Mean  float64
}

// StatisticsCalculator processes latency samples and calculates statistics
type StatisticsCalculator struct {
	samples        []int64
	prewarmCount   int
	processedCount int
}

// NewStatisticsCalculator creates a new statistics calculator with the specified prewarm count
func NewStatisticsCalculator(prewarmCount int) *StatisticsCalculator {
	return &StatisticsCalculator{
		samples:        make([]int64, 0),
		prewarmCount:   prewarmCount,
		processedCount: 0,
	}
}

// AddSample adds a latency sample, skipping the first prewarmCount samples
// Returns true if the sample was added, false if it was skipped for warm-up
func (sc *StatisticsCalculator) AddSample(latencyNs int64) bool {
	sc.processedCount++
	if sc.processedCount <= sc.prewarmCount {
		// Skip warm-up samples
		return false
	}
	sc.samples = append(sc.samples, latencyNs)
	return true
}

// Calculate calculates statistics from the collected samples
func (sc *StatisticsCalculator) Calculate() *LatencyStats {
	stats := &LatencyStats{}

	if len(sc.samples) == 0 {
		return stats
	}

	// Sort samples for percentile calculation
	sortedSamples := make([]int64, len(sc.samples))
	copy(sortedSamples, sc.samples)
	sort.Slice(sortedSamples, func(i, j int) bool {
		return sortedSamples[i] < sortedSamples[j]
	})

	// Set basic statistics
	stats.Count = len(sortedSamples)
	stats.Min = sortedSamples[0]
	stats.Max = sortedSamples[len(sortedSamples)-1]

	// Calculate percentiles
	stats.P10 = sc.calculatePercentile(sortedSamples, 0.1)
	stats.P50 = sc.calculatePercentile(sortedSamples, 0.5)
	stats.P90 = sc.calculatePercentile(sortedSamples, 0.9)
	stats.P99 = sc.calculatePercentile(sortedSamples, 0.99)

	// Calculate mean
	var sum int64
	for _, sample := range sortedSamples {
		sum += sample
	}
	stats.Mean = float64(sum) / float64(len(sortedSamples))

	return stats
}

// calculatePercentile calculates a percentile value from sorted samples
func (sc *StatisticsCalculator) calculatePercentile(sortedSamples []int64, percentile float64) int64 {
	index := int(math.Ceil(percentile*float64(len(sortedSamples)))) - 1
	if index < 0 {
		index = 0
	}
	return sortedSamples[index]
}

// GetProcessedCount returns the number of processed samples (including warm-up samples)
func (sc *StatisticsCalculator) GetProcessedCount() int {
	return sc.processedCount
}

// GetSampleCount returns the number of samples used for statistics (excluding warm-up samples)
func (sc *StatisticsCalculator) GetSampleCount() int {
	return len(sc.samples)
}

// GetSkippedCount returns the number of warm-up samples that were skipped
func (sc *StatisticsCalculator) GetSkippedCount() int {
	if sc.processedCount < sc.prewarmCount {
		return sc.processedCount
	}
	return sc.prewarmCount
}

// Clear clears all samples
func (sc *StatisticsCalculator) Clear() {
	sc.samples = sc.samples[:0]
	sc.processedCount = 0
}

// String returns a string representation of the latency statistics
func (stats *LatencyStats) String() string {
	return fmt.Sprintf("Latency Statistics (nanoseconds):\n"+
		"  Samples: %d\n"+
		"  Min: %d ns\n"+
		"  P10: %d ns\n"+
		"  P50: %d ns (median)\n"+
		"  P90: %d ns\n"+
		"  P99: %d ns\n"+
		"  Max: %d ns\n"+
		"  Mean: %.2f ns",
		stats.Count, stats.Min, stats.P10, stats.P50, stats.P90, stats.P99, stats.Max, stats.Mean)
}
