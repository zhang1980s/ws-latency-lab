package xyz.zzhe.wslatency.stats;

import java.util.List;

/**
 * Holds statistics about latency measurements.
 */
public class LatencyStats {
    private int count;
    private long min;
    private long max;
    private long p10;
    private long p50;
    private long p90;
    private long p99;
    private double mean;
    private List<Long> samples;

    /**
     * Default constructor.
     */
    public LatencyStats() {
        // Default values
        this.count = 0;
        this.min = 0;
        this.max = 0;
        this.p10 = 0;
        this.p50 = 0;
        this.p90 = 0;
        this.p99 = 0;
        this.mean = 0.0;
    }

    /**
     * Get the number of samples.
     *
     * @return Number of samples
     */
    public int getCount() {
        return count;
    }

    /**
     * Set the number of samples.
     *
     * @param count Number of samples
     */
    public void setCount(int count) {
        this.count = count;
    }

    /**
     * Get the minimum latency.
     *
     * @return Minimum latency in nanoseconds
     */
    public long getMin() {
        return min;
    }

    /**
     * Set the minimum latency.
     *
     * @param min Minimum latency in nanoseconds
     */
    public void setMin(long min) {
        this.min = min;
    }

    /**
     * Get the maximum latency.
     *
     * @return Maximum latency in nanoseconds
     */
    public long getMax() {
        return max;
    }

    /**
     * Set the maximum latency.
     *
     * @param max Maximum latency in nanoseconds
     */
    public void setMax(long max) {
        this.max = max;
    }

    /**
     * Get the 10th percentile latency.
     *
     * @return 10th percentile latency in nanoseconds
     */
    public long getP10() {
        return p10;
    }

    /**
     * Set the 10th percentile latency.
     *
     * @param p10 10th percentile latency in nanoseconds
     */
    public void setP10(long p10) {
        this.p10 = p10;
    }

    /**
     * Get the median (50th percentile) latency.
     *
     * @return Median latency in nanoseconds
     */
    public long getP50() {
        return p50;
    }

    /**
     * Set the median (50th percentile) latency.
     *
     * @param p50 Median latency in nanoseconds
     */
    public void setP50(long p50) {
        this.p50 = p50;
    }

    /**
     * Get the 90th percentile latency.
     *
     * @return 90th percentile latency in nanoseconds
     */
    public long getP90() {
        return p90;
    }

    /**
     * Set the 90th percentile latency.
     *
     * @param p90 90th percentile latency in nanoseconds
     */
    public void setP90(long p90) {
        this.p90 = p90;
    }

    /**
     * Get the 99th percentile latency.
     *
     * @return 99th percentile latency in nanoseconds
     */
    public long getP99() {
        return p99;
    }

    /**
     * Set the 99th percentile latency.
     *
     * @param p99 99th percentile latency in nanoseconds
     */
    public void setP99(long p99) {
        this.p99 = p99;
    }

    /**
     * Get the mean latency.
     *
     * @return Mean latency in nanoseconds
     */
    public double getMean() {
        return mean;
    }

    /**
     * Set the mean latency.
     *
     * @param mean Mean latency in nanoseconds
     */
    public void setMean(double mean) {
        this.mean = mean;
    }

    /**
     * Get the raw latency samples.
     *
     * @return List of latency samples in nanoseconds
     */
    public List<Long> getSamples() {
        return samples;
    }

    /**
     * Set the raw latency samples.
     *
     * @param samples List of latency samples in nanoseconds
     */
    public void setSamples(List<Long> samples) {
        this.samples = samples;
    }

    @Override
    public String toString() {
        return String.format(
                "Latency Statistics:%n" +
                "  Samples: %d%n" +
                "  Min: %d ns (%.3f µs, %.3f ms)%n" +
                "  P10: %d ns (%.3f µs, %.3f ms)%n" +
                "  P50: %d ns (%.3f µs, %.3f ms) (median)%n" +
                "  P90: %d ns (%.3f µs, %.3f ms)%n" +
                "  P99: %d ns (%.3f µs, %.3f ms)%n" +
                "  Max: %d ns (%.3f µs, %.3f ms)%n" +
                "  Mean: %.2f ns (%.3f µs, %.3f ms)",
                count,
                min, min/1000.0, min/1000000.0,
                p10, p10/1000.0, p10/1000000.0,
                p50, p50/1000.0, p50/1000000.0,
                p90, p90/1000.0, p90/1000000.0,
                p99, p99/1000.0, p99/1000000.0,
                max, max/1000.0, max/1000000.0,
                mean, mean/1000.0, mean/1000000.0);
    }
}