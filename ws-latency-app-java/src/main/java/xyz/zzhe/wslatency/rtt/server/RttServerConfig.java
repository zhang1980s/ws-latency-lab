package xyz.zzhe.wslatency.rtt.server;

/**
 * Configuration for the WebSocket RTT server.
 */
public class RttServerConfig {
    private int port = 10443;
    private int payloadSize = 100;
    private boolean secure = false;
    private String keyStorePath = null;
    private String keyStorePassword = null;

    /**
     * Default constructor with default values.
     */
    public RttServerConfig() {
        // Default values set in field declarations
    }

    /**
     * Get the port for the server to listen on.
     *
     * @return The port number
     */
    public int getPort() {
        return port;
    }

    /**
     * Set the port for the server to listen on.
     *
     * @param port The port number
     */
    public void setPort(int port) {
        this.port = port;
    }

    /**
     * Get the size of the event payload in bytes.
     *
     * @return Payload size in bytes
     */
    public int getPayloadSize() {
        return payloadSize;
    }

    /**
     * Set the size of the event payload in bytes.
     *
     * @param payloadSize Payload size in bytes
     */
    public void setPayloadSize(int payloadSize) {
        this.payloadSize = payloadSize;
    }

    /**
     * Check if secure WebSocket (WSS) should be used.
     *
     * @return true if secure, false otherwise
     */
    public boolean isSecure() {
        return secure;
    }

    /**
     * Set whether to use secure WebSocket (WSS).
     *
     * @param secure true to use secure WebSocket, false otherwise
     */
    public void setSecure(boolean secure) {
        this.secure = secure;
    }

    /**
     * Get the path to the keystore file for SSL/TLS.
     *
     * @return Keystore file path
     */
    public String getKeyStorePath() {
        return keyStorePath;
    }

    /**
     * Set the path to the keystore file for SSL/TLS.
     *
     * @param keyStorePath Keystore file path
     */
    public void setKeyStorePath(String keyStorePath) {
        this.keyStorePath = keyStorePath;
    }

    /**
     * Get the password for the keystore.
     *
     * @return Keystore password
     */
    public String getKeyStorePassword() {
        return keyStorePassword;
    }

    /**
     * Set the password for the keystore.
     *
     * @param keyStorePassword Keystore password
     */
    public void setKeyStorePassword(String keyStorePassword) {
        this.keyStorePassword = keyStorePassword;
    }

    @Override
    public String toString() {
        return "RttServerConfig{" +
                "port=" + port +
                ", payloadSize=" + payloadSize +
                ", secure=" + secure +
                '}';
    }
}