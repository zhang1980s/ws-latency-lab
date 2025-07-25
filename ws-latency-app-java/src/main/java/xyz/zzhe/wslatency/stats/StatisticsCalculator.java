package xyz.zzhe.wslatency.stats;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

/**
 * Calculates statistics from latency samples.
 */
public class StatisticsCalculator {
    private static final Logger logger = LoggerFactory.getLogger(StatisticsCalculator.class);

    private final List<Long> samples;
    private final int prewarmCount;
    private int processedCount;

    /**
     * Creates a new statistics calculator with the specified prewarm count.
     *
     * @param prewarmCount Number of samples to skip for warm-up
     */
    public StatisticsCalculator(int prewarmCount) {
        this.samples = new ArrayList<>();
        this.prewarmCount = prewarmCount;
        this.processedCount = 0;
    }

    /**
     * Add a latency sample.
     *
     * @param latencyNs Latency in nanoseconds
     * @return true if the sample was added (not skipped for warm-up), false otherwise
     */
   public boolean addSample(long latencyNs) {
       processedCount++;
       if (processedCount <= prewarmCount) {
           // Skip warm-up samples
           return false;
       }
       samples.add(latencyNs);
       return true;
   }

    /**
     * Calculate statistics from the collected samples.
     *
     * @return The calculated statistics
     */
    public LatencyStats calculate() {
        LatencyStats stats = new LatencyStats();
        
        if (samples.isEmpty()) {
            logger.warn("No samples to calculate statistics");
            return stats;
        }

        // Sort samples for percentile calculation
        List<Long> sortedSamples = new ArrayList<>(samples);
        Collections.sort(sortedSamples);

        // Set basic statistics
        stats.setCount(sortedSamples.size());
        stats.setMin(sortedSamples.get(0));
        stats.setMax(sortedSamples.get(sortedSamples.size() - 1));

        // Calculate percentiles
        stats.setP10(calculatePercentile(sortedSamples, 0.1));
        stats.setP50(calculatePercentile(sortedSamples, 0.5));
        stats.setP90(calculatePercentile(sortedSamples, 0.9));
        stats.setP99(calculatePercentile(sortedSamples, 0.99));

        // Calculate mean
        double sum = 0;
        for (Long sample : sortedSamples) {
            sum += sample;
        }
        stats.setMean(sum / sortedSamples.size());

        // Store samples
        stats.setSamples(sortedSamples);

        return stats;
    }

    /**
     * Calculate a percentile value from sorted samples.
     *
     * @param sortedSamples Sorted list of samples
     * @param percentile    Percentile to calculate (0.0 to 1.0)
     * @return The percentile value
     */
    private long calculatePercentile(List<Long> sortedSamples, double percentile) {
        int index = (int) Math.ceil(percentile * sortedSamples.size()) - 1;
        if (index < 0) {
            index = 0;
        }
        return sortedSamples.get(index);
    }


    /**
     * Get the number of processed samples (including warm-up samples).
     *
     * @return Number of processed samples
     */
    public int getProcessedCount() {
        return processedCount;
    }

    /**
     * Get the number of samples used for statistics (excluding warm-up samples).
     *
     * @return Number of samples
     */
    public int getSampleCount() {
        return samples.size();
    }

    /**
     * Get the number of warm-up samples that were skipped.
     *
     * @return Number of skipped samples
     */
    public int getSkippedCount() {
        return Math.min(processedCount, prewarmCount);
    }

    /**
     * Clear all samples.
     */
    public void clear() {
        samples.clear();
        processedCount = 0;
    }
}