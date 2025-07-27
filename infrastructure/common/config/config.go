package config

import (
	"errors"
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// Config holds the configuration for a stack
type Config struct {
	// Common configuration
	Environment    string
	Project        string
	Owner          string
	Region         string
	CertificateArn string
	KeyPairName    string

	// Server VPC configuration
	InstanceType string

	// Transit VPC configuration
	TransitInstanceType string

	// Custom tags
	CustomTags map[string]string
}

// LoadConfig loads the configuration for a stack
func LoadConfig(ctx *pulumi.Context) (*Config, error) {
	conf := config.New(ctx, "")

	// Load required configuration
	environment := conf.Require("environment")
	project := conf.Require("project")
	owner := conf.Require("owner")
	region := conf.Get("region")
	if region == "" {
		region = "ap-east-1" // Default region
	}
	certificateArn := conf.Get("certificateArn")
	keyPairName := conf.Get("keyPairName")
	if keyPairName == "" {
		keyPairName = KeyPairName // Use default from constants.go if not specified
	}

	// Load server VPC configuration
	instanceType := conf.Get("instanceType")
	if instanceType == "" {
		instanceType = ServerInstanceType // Use default from constants.go if not specified
	}

	// Load transit VPC configuration
	transitInstanceType := conf.Get("instanceType")
	if transitInstanceType == "" {
		transitInstanceType = ClientInstanceType // Use default from constants.go if not specified
	}

	// Load custom tags
	customTags := make(map[string]string)
	var tagsObj interface{}
	if err := conf.TryObject("customTags", &tagsObj); err == nil && tagsObj != nil {
		if tagsMap, ok := tagsObj.(map[string]interface{}); ok {
			for k, v := range tagsMap {
				if str, ok := v.(string); ok {
					customTags[k] = str
				}
			}
		}
	}

	// Validate configuration
	if err := validateConfig(environment, project, owner); err != nil {
		return nil, err
	}

	return &Config{
		Environment:         environment,
		Project:             project,
		Owner:               owner,
		Region:              region,
		CertificateArn:      certificateArn,
		KeyPairName:         keyPairName,
		InstanceType:        instanceType,
		TransitInstanceType: transitInstanceType,
		CustomTags:          customTags,
	}, nil
}

// validateConfig validates the configuration
func validateConfig(environment, project, owner string) error {
	// Validate required fields
	if environment == "" {
		return errors.New("environment is required")
	}
	if project == "" {
		return errors.New("project is required")
	}
	if owner == "" {
		return errors.New("owner is required")
	}

	return nil
}

// GetStackReference gets a reference to another stack
func GetStackReference(ctx *pulumi.Context, stackName string) (*pulumi.StackReference, error) {
	stackRef, err := pulumi.NewStackReference(ctx, stackName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get stack reference %s: %w", stackName, err)
	}
	return stackRef, nil
}
