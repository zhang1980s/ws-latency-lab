package resources

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lb"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/config"
	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/utils"
)

// CreateLoadBalancers creates the NLB and ALB in the middle VPC
func CreateLoadBalancers(ctx *pulumi.Context, cfg *config.Config, vpc *ec2.Vpc, subnets []pulumi.IDOutput, sg *ec2.SecurityGroup, serverVpcRef *pulumi.StackReference, serverVpcId, serverSubnetIds, serverSecurityGroupId pulumi.Output, accessLogBucket *s3.Bucket) (*lb.LoadBalancer, *lb.LoadBalancer, error) {
	// Get load balancer configurations
	nlbConfig := config.GetNetworkNlbConfig()
	albConfig := config.GetNetworkAlbConfig()

	// Create ALB
	alb, albTargetGroup, albListener, err := createAlb(ctx, cfg, vpc, subnets, sg, albConfig, serverVpcRef, serverVpcId, serverSubnetIds, serverSecurityGroupId, accessLogBucket)
	if err != nil {
		return nil, nil, err
	}

	// Create NLB
	nlb, err := createNlb(ctx, cfg, vpc, subnets, nlbConfig, alb, albTargetGroup, albListener, accessLogBucket)
	if err != nil {
		return nil, nil, err
	}

	return alb, nlb, nil
}

// createAlb creates an Application Load Balancer
func createAlb(ctx *pulumi.Context, cfg *config.Config, vpc *ec2.Vpc, subnets []pulumi.IDOutput, sg *ec2.SecurityGroup, albConfig config.LoadBalancerConfig, serverVpcRef *pulumi.StackReference, serverVpcId, serverSubnetIds, serverSecurityGroupId pulumi.Output, accessLogBucket *s3.Bucket) (*lb.LoadBalancer, *lb.TargetGroup, *lb.Listener, error) {
	// Create ALB
	albName := albConfig.Name

	// Convert subnets from []pulumi.IDOutput to pulumi.StringArray
	subnetIDs := pulumi.StringArray{}
	for _, subnet := range subnets {
		subnetIDs = append(subnetIDs, subnet.ToStringOutput())
	}

	alb, err := lb.NewLoadBalancer(ctx, albName, &lb.LoadBalancerArgs{
		Name:             pulumi.String(albName),
		Internal:         pulumi.Bool(false), // Set to false to make ALB publicly accessible
		LoadBalancerType: pulumi.String("application"),
		SecurityGroups:   pulumi.StringArray{sg.ID()},
		Subnets:          subnetIDs,
		// Enable access logging to S3 bucket
		AccessLogs: &lb.LoadBalancerAccessLogsArgs{
			Bucket:  accessLogBucket.ID(),
			Enabled: pulumi.Bool(true),
			Prefix:  pulumi.String("alb-logs"),
		},
		Tags: utils.ApplyTags(ctx, albName, utils.GetNamedTags(albName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, nil, nil, err
	}

	// Note: We need to set idle_timeout.timeout_seconds to 600 (10 minutes) for WebSocket connections
	// This is currently done manually in the AWS console after deployment
	// TODO: Update this to use the correct Pulumi API when available

	// Create target group for server
	targetGroupName := config.GetTargetGroupName(albName, "server")
	targetGroup, err := lb.NewTargetGroup(ctx, targetGroupName, &lb.TargetGroupArgs{
		Name:                pulumi.String(targetGroupName),
		Port:                pulumi.Int(config.WebSocketPort),
		Protocol:            pulumi.String("HTTP"), // Use HTTP for backend connections (SSL termination at ALB)
		VpcId:               vpc.ID(),
		TargetType:          pulumi.String("ip"),
		DeregistrationDelay: pulumi.Int(30),
		IpAddressType:       pulumi.String("ipv4"),
		// Configure health check
		HealthCheck: &lb.TargetGroupHealthCheckArgs{
			Enabled:            pulumi.Bool(true),
			Path:               pulumi.String("/health"),
			Port:               pulumi.String(fmt.Sprintf("%d", config.WebSocketPort)),
			Protocol:           pulumi.String("HTTP"),
			HealthyThreshold:   pulumi.Int(3),
			UnhealthyThreshold: pulumi.Int(3),
			Timeout:            pulumi.Int(5),
			Interval:           pulumi.Int(30),
			Matcher:            pulumi.String("200-299"),
		},
		// Configure stickiness for WebSocket connections
		Stickiness: &lb.TargetGroupStickinessArgs{
			Enabled:        pulumi.Bool(true),
			Type:           pulumi.String("lb_cookie"),
			CookieDuration: pulumi.Int(86400), // 24 hours
		},
		Tags: utils.ApplyTags(ctx, targetGroupName, utils.GetNamedTags(targetGroupName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, nil, nil, err
	}

	// Create ALB listener
	listenerName := config.FormatResourceName(albName, "listener")
	albListener, err := lb.NewListener(ctx, listenerName, &lb.ListenerArgs{
		LoadBalancerArn: alb.Arn,
		Port:            pulumi.Int(albConfig.Port),
		Protocol:        pulumi.String(albConfig.Protocol),
		SslPolicy:       pulumi.String("ELBSecurityPolicy-2016-08"),
		CertificateArn:  pulumi.String(cfg.CertificateArn),
		DefaultActions: lb.ListenerDefaultActionArray{
			&lb.ListenerDefaultActionArgs{
				Type:           pulumi.String("forward"),
				TargetGroupArn: targetGroup.Arn,
			},
		},
		Tags: utils.ApplyTags(ctx, listenerName, utils.GetNamedTags(listenerName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, nil, nil, err
	}

	// Register server IP as target for ALB target group
	// Get the instance private IP from the server-vpc-stack reference
	instancePrivateIp := serverVpcRef.GetOutput(pulumi.String("instancePrivateIp"))

	// Convert the output to a string output
	serverPrivateIp := instancePrivateIp.ApplyT(func(ip interface{}) string {
		if ip == nil {
			ctx.Log.Error("Server instance private IP is nil", nil)
			return ""
		}
		privateIp, ok := ip.(string)
		if !ok {
			ctx.Log.Error(fmt.Sprintf("Expected instancePrivateIp to be string, got %T", ip), nil)
			return ""
		}
		return privateIp
	}).(pulumi.StringOutput)

	// Create target group attachment for the ALB target group
	albTgAttachment, err := lb.NewTargetGroupAttachment(ctx, config.FormatResourceName(targetGroupName, "server-attachment"), &lb.TargetGroupAttachmentArgs{
		TargetGroupArn:   targetGroup.Arn,
		TargetId:         serverPrivateIp,
		Port:             pulumi.Int(config.WebSocketPort),
		AvailabilityZone: pulumi.String("all"),
	}, pulumi.DependsOn([]pulumi.Resource{targetGroup}))
	if err != nil {
		return nil, nil, nil, err
	}

	ctx.Log.Info("Registered server instance private IP as target for ALB target group", nil)

	// Add explicit dependency to ensure the ALB target group attachment is deleted before the ALB listener
	// This helps with proper resource deletion order during destroy
	ctx.RegisterResourceOutputs(albListener, pulumi.Map{
		"albTargetGroupAttachment": albTgAttachment.ID(),
	})

	return alb, targetGroup, albListener, nil
}

// createNlb creates a Network Load Balancer
func createNlb(ctx *pulumi.Context, cfg *config.Config, vpc *ec2.Vpc, subnets []pulumi.IDOutput, nlbConfig config.LoadBalancerConfig, alb *lb.LoadBalancer, albTargetGroup *lb.TargetGroup, albListener *lb.Listener, accessLogBucket *s3.Bucket) (*lb.LoadBalancer, error) {
	// Create NLB
	nlbName := nlbConfig.Name

	// Convert subnets from []pulumi.IDOutput to pulumi.StringArray
	subnetIDs := pulumi.StringArray{}
	for _, subnet := range subnets {
		subnetIDs = append(subnetIDs, subnet.ToStringOutput())
	}

	nlb, err := lb.NewLoadBalancer(ctx, nlbName, &lb.LoadBalancerArgs{
		Name:             pulumi.String(nlbName),
		Internal:         pulumi.Bool(true), // Set to false to make NLB publicly accessible
		LoadBalancerType: pulumi.String("network"),
		Subnets:          subnetIDs,
		// Enable access logging to S3 bucket
		AccessLogs: &lb.LoadBalancerAccessLogsArgs{
			Bucket:  accessLogBucket.ID(),
			Enabled: pulumi.Bool(true),
			Prefix:  pulumi.String("nlb-logs"),
		},
		Tags: utils.ApplyTags(ctx, nlbName, utils.GetNamedTags(nlbName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, err
	}

	// Create target group for ALB
	targetGroupName := config.GetTargetGroupName(nlbName, "alb")
	targetGroup, err := lb.NewTargetGroup(ctx, targetGroupName, &lb.TargetGroupArgs{
		Name:                pulumi.String(targetGroupName),
		Port:                pulumi.Int(config.AlbPort),
		Protocol:            pulumi.String("TCP"),
		VpcId:               vpc.ID(),
		TargetType:          pulumi.String("alb"),
		DeregistrationDelay: pulumi.Int(30),
		HealthCheck: &lb.TargetGroupHealthCheckArgs{
			Enabled:            pulumi.Bool(true),
			Port:               pulumi.String(fmt.Sprintf("%d", config.AlbPort)), // Use the ALB port (443)
			Protocol:           pulumi.String("HTTPS"),                           // Use HTTPS protocol
			Path:               pulumi.String("/health"),                         // Use /health path
			HealthyThreshold:   pulumi.Int(3),
			UnhealthyThreshold: pulumi.Int(3),
			Interval:           pulumi.Int(30),
			Timeout:            pulumi.Int(5),
			Matcher:            pulumi.String("200-299"),
		},
		Tags: utils.ApplyTags(ctx, targetGroupName, utils.GetNamedTags(targetGroupName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, err
	}

	// Define listener name
	listenerName := config.FormatResourceName(nlbName, "listener")

	// Create NLB listener
	_, err = lb.NewListener(ctx, listenerName, &lb.ListenerArgs{
		LoadBalancerArn: nlb.Arn,
		Port:            pulumi.Int(nlbConfig.Port),
		Protocol:        pulumi.String(nlbConfig.Protocol),
		DefaultActions: lb.ListenerDefaultActionArray{
			&lb.ListenerDefaultActionArgs{
				Type:           pulumi.String("forward"),
				TargetGroupArn: targetGroup.Arn,
			},
		},
		Tags: utils.ApplyTags(ctx, listenerName, utils.GetNamedTags(listenerName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, err
	}

	// Register ALB as target for NLB
	nlbTgAttachment, err := lb.NewTargetGroupAttachment(ctx, config.FormatResourceName(targetGroupName, "attachment"), &lb.TargetGroupAttachmentArgs{
		TargetGroupArn: targetGroup.Arn,
		TargetId:       alb.Arn,
		Port:           pulumi.Int(config.AlbPort),
	}, pulumi.DependsOn([]pulumi.Resource{alb, targetGroup, albListener}))
	if err != nil {
		return nil, err
	}

	// Add explicit dependency to ensure the NLB target group attachment is deleted before the ALB listener
	// This helps with proper resource deletion order during destroy
	ctx.RegisterResourceOutputs(albListener, pulumi.Map{
		"nlbTargetGroupAttachment": nlbTgAttachment.ID(),
	})

	return nlb, nil
}
