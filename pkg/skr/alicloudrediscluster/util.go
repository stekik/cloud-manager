package alicloudrediscluster

import (
	"errors"
	"maps"
	"strings"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"

	"github.com/kyma-project/cloud-manager/pkg/util"
)

func getAuthSecretName(alicloudRedis *cloudresourcesv1beta1.AlicloudRedisCluster) string {
	if alicloudRedis.Spec.AuthSecret != nil && len(alicloudRedis.Spec.AuthSecret.Name) > 0 {
		return alicloudRedis.Spec.AuthSecret.Name
	}

	return alicloudRedis.Name
}

func getAuthSecretLabels(alicloudRedis *cloudresourcesv1beta1.AlicloudRedisCluster) map[string]string {
	labelsBuilder := util.NewLabelBuilder()

	if alicloudRedis.Spec.AuthSecret != nil {
		for labelName, labelValue := range alicloudRedis.Spec.AuthSecret.Labels {
			labelsBuilder.WithCustomLabel(labelName, labelValue)
		}
	}

	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelRedisClusterStatusId, alicloudRedis.Status.Id)
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelRedisClusterNamespace, alicloudRedis.Namespace)
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelCloudManaged, "true")
	labelsBuilder.WithCloudManagerDefaults()
	pvLabels := labelsBuilder.Build()

	return pvLabels
}

func getAuthSecretAnnotations(alicloudRedis *cloudresourcesv1beta1.AlicloudRedisCluster) map[string]string {
	if alicloudRedis.Spec.AuthSecret == nil {
		return nil
	}
	result := map[string]string{}
	maps.Copy(result, alicloudRedis.Spec.AuthSecret.Annotations)
	return result
}

func getAuthSecretBaseData(kcpRedis *cloudcontrolv1beta1.RedisCluster) map[string][]byte {
	result := map[string][]byte{}

	if len(kcpRedis.Status.DiscoveryEndpoint) > 0 {
		result["discoveryEndpoint"] = []byte(kcpRedis.Status.DiscoveryEndpoint)

		splitEndpoint := strings.Split(kcpRedis.Status.DiscoveryEndpoint, ":")
		if len(splitEndpoint) >= 2 {
			host := splitEndpoint[0]
			port := splitEndpoint[1]
			result["host"] = []byte(host)
			result["port"] = []byte(port)
		}
	}

	if len(kcpRedis.Status.AuthString) > 0 {
		result["authString"] = []byte(kcpRedis.Status.AuthString)
	}

	return result
}

func parseAuthSecretExtraData(extraDataTemplates map[string]string, authSecretBaseData map[string][]byte) map[string][]byte {
	baseDataStringMap := map[string]string{}
	for k, v := range authSecretBaseData {
		baseDataStringMap[k] = string(v)
	}

	return util.ParseTemplatesMapToBytesMap(extraDataTemplates, baseDataStringMap)
}

var alicloudRedisClusterTierInstanceClassMap = map[cloudresourcesv1beta1.AlicloudRedisClusterTier]string{
	cloudresourcesv1beta1.AlicloudRedisClusterTierC3: "redis.shard.large.ce",
	cloudresourcesv1beta1.AlicloudRedisClusterTierC4: "redis.shard.xlarge.ce",
	cloudresourcesv1beta1.AlicloudRedisClusterTierC5: "redis.shard.2xlarge.ce",
	cloudresourcesv1beta1.AlicloudRedisClusterTierC6: "redis.shard.4xlarge.ce",
	cloudresourcesv1beta1.AlicloudRedisClusterTierC7: "redis.shard.8xlarge.ce",
}

func redisTierToInstanceClass(tier cloudresourcesv1beta1.AlicloudRedisClusterTier) (string, error) {
	instanceClass, exists := alicloudRedisClusterTierInstanceClassMap[tier]
	if !exists {
		return "", errors.New("unknown redis cluster tier")
	}
	return instanceClass, nil
}
