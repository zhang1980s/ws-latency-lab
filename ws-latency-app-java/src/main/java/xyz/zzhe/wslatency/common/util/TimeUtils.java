package xyz.zzhe.wslatency.common.util;

import java.time.Instant;

/**
 * Utility class for time-related operations.
 */
public class TimeUtils {
    
    /**
     * Private constructor to prevent instantiation.
     */
    private TimeUtils() {
        // Utility class should not be instantiated
    }
    
    /**
     * Get the current time in microseconds since the epoch.
     * 
     * @return Current time in microseconds
     */
    public static long getCurrentTimeMicros() {
        // Use wall clock time instead of System.nanoTime() for consistent timestamps across machines
        return Instant.now().toEpochMilli() * 1000;
    }
    
    /**
     * Get the current time in nanoseconds since the epoch.
     *
     * @return Current time in nanoseconds
     */
    // Base nanosecond time reference (initialized at class load time)
    private static final long SYSTEM_NANO_TIME_ORIGIN = System.nanoTime();
    private static final long WALL_TIME_ORIGIN = System.currentTimeMillis();
    
    /**
     * Get the current time in nanoseconds since the epoch.
     * This method ensures monotonically increasing timestamps by using
     * System.nanoTime() for elapsed time measurement and wall clock for
     * absolute time reference.
     *
     * @return Current time in nanoseconds
     */
    public static long getCurrentTimeNanos() {
        // Calculate elapsed nanos since our reference point
        long elapsedNanos = System.nanoTime() - SYSTEM_NANO_TIME_ORIGIN;
        
        // Add to our wall clock origin (converted to nanos)
        // This gives us wall clock time + elapsed nanos since startup
        return (WALL_TIME_ORIGIN * 1_000_000) + elapsedNanos;
    }
    
    /**
     * Get the current wall clock time in microseconds since the epoch.
     * This is less precise than {@link #getCurrentTimeMicros()} but is based on the system clock.
     * 
     * @return Current wall clock time in microseconds
     */
    public static long getCurrentWallClockMicros() {
        return Instant.now().toEpochMilli() * 1000;
    }
    
    /**
     * Convert nanoseconds to microseconds.
     * 
     * @param nanos Time in nanoseconds
     * @return Time in microseconds
     */
    public static long nanosToMicros(long nanos) {
        return nanos / 1000;
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