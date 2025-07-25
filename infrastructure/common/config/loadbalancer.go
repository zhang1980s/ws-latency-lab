package config

// Default protocols for load balancers
const (
	AlbProtocol = "HTTP"
	NlbProtocol = "TLS"
)

// LoadBalancerConfig holds the configuration for a load balancer
type LoadBalancerConfig struct {
	Name            string
	Type            string // "application" or "network"
	Port            int
	Protocol        string
	Internal        bool
	HealthCheckPath string // Added to match network_config.go
}

// GetNlbConfig returns the configuration for the NLB
// Deprecated: Use the version in network_config.go instead
func GetNlbConfigLegacy() LoadBalancerConfig {
	return LoadBalancerConfig{
		Name:     "nlb",
		Type:     "network",
		Port:     NlbPort,
		Protocol: NlbProtocol,
		Internal: false,
	}
}

// GetAlbConfig returns the configuration for the ALB
// Deprecated: Use the version in network_config.go instead
func GetAlbConfigLegacy() LoadBalancerConfig {
	return LoadBalancerConfig{
		Name:     "alb",
		Type:     "application",
		Port:     AlbPort,
		Protocol: AlbProtocol,
		Internal: true,
	}
}

// GetLoadBalancerNameLegacy returns a formatted name for a load balancer
// Deprecated: Use the version in helpers.go instead
func GetLoadBalancerNameLegacy(config LoadBalancerConfig, suffix string) string {
	if suffix == "" {
		return FormatResourceNameLegacy(config.Name, config.Type)
	}
	return FormatResourceNameLegacy(config.Name, config.Type, suffix)
}

// GetTargetGroupNameLegacy returns a formatted name for a target group
// Deprecated: Use the version in helpers.go instead
func GetTargetGroupNameLegacy(lbName string, suffix string) string {
	return FormatResourceNameLegacy(lbName, "tg", suffix)
}

// FormatResourceNameLegacy formats a resource name with a prefix and suffix
// Deprecated: Use the version in helpers.go instead
func FormatResourceNameLegacy(parts ...string) string {
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += "-"
		}
		result += part
	}
	return result
}
