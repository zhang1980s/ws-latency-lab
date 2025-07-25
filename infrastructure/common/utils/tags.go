package utils

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// TagsMap represents a map of resource tags
type TagsMap map[string]string

// GetDefaultTags returns the default tags for all resources
func GetDefaultTags(environment, project, owner string, customTags map[string]string) TagsMap {
	tags := TagsMap{
		"Environment": environment,
		"Project":     project,
		"Owner":       owner,
		"ManagedBy":   "pulumi",
	}

	// Add custom tags
	for k, v := range customTags {
		tags[k] = v
	}

	return tags
}

// GetNamedTags returns tags with a Name tag added
func GetNamedTags(name, environment, project, owner string, customTags map[string]string) TagsMap {
	tags := GetDefaultTags(environment, project, owner, customTags)
	tags["Name"] = name
	return tags
}

// ApplyTags applies tags to a resource
func ApplyTags(ctx *pulumi.Context, name string, tags TagsMap) pulumi.StringMap {
	result := pulumi.StringMap{}
	for k, v := range tags {
		result[k] = pulumi.String(v)
	}

	ctx.Log.Info(fmt.Sprintf("Applied tags to %s: %v", name, tags), nil)
	return result
}
