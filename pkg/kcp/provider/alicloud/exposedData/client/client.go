package client

import (
	"context"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	vpc "github.com/alibabacloud-go/vpc-20160428/v6/client"
	alicloudclientconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/config"
	alicloudmetrics "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/metrics"
)

type NatGatewayInfo struct {
	NatGatewayId string
	IpAddresses  []string
}

type Client interface {
	DescribeNatGateways(ctx context.Context, vpcId string) ([]NatGatewayInfo, error)
}

type ClientProvider func(ctx context.Context, region, accessKeyId, accessKeySecret, assumeRoleArn string) (Client, error)

func NewClientProvider() ClientProvider {
	return func(ctx context.Context, region, accessKeyId, accessKeySecret, assumeRoleArn string) (Client, error) {
		config := &openapi.Config{
			RegionId: new(region),
		}
		if err := alicloudclientconfig.ApplyCredentials(config, accessKeyId, accessKeySecret, assumeRoleArn); err != nil {
			return nil, err
		}
		config.Endpoint = new(fmt.Sprintf("vpc.%s.aliyuncs.com", region))
		config.HttpClient = alicloudmetrics.NewMetricsHTTPClient(region, alicloudmetrics.AccountIdFromContext(ctx))

		vpcClient, err := vpc.NewClient(config)
		if err != nil {
			return nil, fmt.Errorf("error creating alicloud vpc client for exposed data: %w", err)
		}

		return &alicloudClient{vpcClient: vpcClient, region: region}, nil
	}
}

type alicloudClient struct {
	vpcClient *vpc.Client
	region    string
}

func (c *alicloudClient) DescribeNatGateways(ctx context.Context, vpcId string) ([]NatGatewayInfo, error) {
	req := &vpc.DescribeNatGatewaysRequest{
		RegionId: new(c.region),
		VpcId:    new(vpcId),
	}

	resp, err := c.vpcClient.DescribeNatGateways(req)
	if err != nil {
		return nil, fmt.Errorf("error describing alicloud nat gateways for vpc %s: %w", vpcId, err)
	}

	var result []NatGatewayInfo
	if resp.Body == nil || resp.Body.NatGateways == nil {
		return result, nil
	}

	for _, gw := range resp.Body.NatGateways.NatGateway {
		info := NatGatewayInfo{
			NatGatewayId: tea.StringValue(gw.NatGatewayId),
		}
		if gw.IpLists != nil {
			for _, ipEntry := range gw.IpLists.IpList {
				if ip := tea.StringValue(ipEntry.IpAddress); ip != "" {
					info.IpAddresses = append(info.IpAddresses, ip)
				}
			}
		}
		result = append(result, info)
	}

	return result, nil
}
