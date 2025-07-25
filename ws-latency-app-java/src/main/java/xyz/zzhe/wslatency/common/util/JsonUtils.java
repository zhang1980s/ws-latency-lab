package xyz.zzhe.wslatency.common.util;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Utility class for JSON operations.
 */
public class JsonUtils {
    private static final Logger logger = LoggerFactory.getLogger(JsonUtils.class);
    private static final ObjectMapper mapper = new ObjectMapper();

    /**
     * Private constructor to prevent instantiation.
     */
    private JsonUtils() {
        // Utility class should not be instantiated
    }

    /**
     * Convert an object to a JSON string.
     *
     * @param object The object to convert
     * @return JSON string representation, or null if conversion fails
     */
    public static String toJson(Object object) {
        try {
            return mapper.writeValueAsString(object);
        } catch (JsonProcessingException e) {
            logger.error("Error converting object to JSON", e);
            return null;
        }
    }

    /**
     * Convert a JSON string to an object.
     *
     * @param json  The JSON string
     * @param clazz The class of the object
     * @param <T>   The type of the object
     * @return The object, or null if conversion fails
     */
    public static <T> T fromJson(String json, Class<T> clazz) {
        try {
            return mapper.readValue(json, clazz);
        } catch (JsonProcessingException e) {
            logger.error("Error converting JSON to object", e);
            return null;
        }
    }

    /**
     * Get the ObjectMapper instance.
     *
     * @return The ObjectMapper instance
     */
    public static ObjectMapper getMapper() {
        return mapper;
    }
}