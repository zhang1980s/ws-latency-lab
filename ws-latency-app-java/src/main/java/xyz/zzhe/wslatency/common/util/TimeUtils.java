package xyz.zzhe.wslatency.common.util;

import java.time.Instant;

/**
 * Utility class for time-related operations.
 * Optimized for microsecond precision to match Go implementation.
 */
public class TimeUtils {
    
    /**
     * Private constructor to prevent instantiation.
     */
    private TimeUtils() {
        // Utility class should not be instantiated
    }
    
    // Base microsecond time reference (initialized at class load time)
    private static final long SYSTEM_NANO_TIME_ORIGIN = System.nanoTime();
    private static final long WALL_TIME_ORIGIN = System.currentTimeMillis();
    
    /**
     * Get the current time in microseconds since the epoch.
     * This method ensures monotonically increasing timestamps by using
     * System.nanoTime() for elapsed time measurement and wall clock for
     * absolute time reference.
     *
     * @return Current time in microseconds
     */
    public static long getCurrentTimeMicros() {
        // Calculate elapsed nanos since our reference point
        long elapsedNanos = System.nanoTime() - SYSTEM_NANO_TIME_ORIGIN;
        
        // Add to our wall clock origin (converted to micros) and convert nanos to micros
        return (WALL_TIME_ORIGIN * 1_000) + (elapsedNanos / 1_000);
    }
    
    /**
     * Convert microseconds to milliseconds.
     *
     * @param micros Time in microseconds
     * @return Time in milliseconds
     */
    public static long microsToMillis(long micros) {
        return micros / 1000;
    }
    
    /**
     * Format a duration in microseconds to a human-readable string.
     * 
     * @param micros Duration in microseconds
     * @return Human-readable duration string
     */
    public static String formatMicros(long micros) {
        if (micros < 1000) {
            return micros + " µs";
        } else if (micros < 1000000) {
            return String.format("%.2f ms", micros / 1000.0);
        } else {
            return String.format("%.2f s", micros / 1000000.0);
        }
    }
}