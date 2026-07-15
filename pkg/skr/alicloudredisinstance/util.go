package alicloudredisinstance

import (
	"errors"
	"maps"
	"strings"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"

	"github.com/kyma-project/cloud-manager/pkg/util"
)

func getAuthSecretName(alicloudRedis *cloudresourcesv1beta1.AlicloudRedisInstance) string {
	if alicloudRedis.Spec.AuthSecret != nil && len(alicloudRedis.Spec.AuthSecret.Name) > 0 {
		return alicloudRedis.Spec.AuthSecret.Name
	}

	return alicloudRedis.Name
}

func getAuthSecretLabels(alicloudRedis *cloudresourcesv1beta1.AlicloudRedisInstance) map[string]string {
	labelsBuilder := util.NewLabelBuilder()

	if alicloudRedis.Spec.AuthSecret != nil {
		for labelName, labelValue := range alicloudRedis.Spec.AuthSecret.Labels {
			labelsBuilder.WithCustomLabel(labelName, labelValue)
		}
	}

	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelRedisInstanceStatusId, alicloudRedis.Status.Id)
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelRedisInstanceNamespace, alicloudRedis.Namespace)
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelCloudManaged, "true")
	labelsBuilder.WithCloudManagerDefaults()
	pvLabels := labelsBuilder.Build()

	return pvLabels
}

func getAuthSecretAnnotations(alicloudRedis *cloudresourcesv1beta1.AlicloudRedisInstance) map[string]string {
	if alicloudRedis.Spec.AuthSecret == nil {
		return nil
	}
	result := map[string]string{}
	maps.Copy(result, alicloudRedis.Spec.AuthSecret.Annotations)
	return result
}

func getAuthSecretBaseData(kcpRedis *cloudcontrolv1beta1.RedisInstance) map[string][]byte {
	result := map[string][]byte{}

	if len(kcpRedis.Status.PrimaryEndpoint) > 0 {
		result["primaryEndpoint"] = []byte(kcpRedis.Status.PrimaryEndpoint)

		splitEndpoint := strings.Split(kcpRedis.Status.PrimaryEndpoint, ":")
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

type redisTierInfo struct {
	instanceClass string
	readOnlyCount int32
}

// tair.rdb.* is the Tair DRAM-based HA instance available in ap-southeast-1 (Singapore).
// It is always a master+replica pair (NodeType=double); readOnlyCount is not applicable.
// S-tiers and P-tiers map to the same instance class — P-tiers use readOnlyCount=0 here
// because tair.rdb does not accept readOnlyCount via API (returns COMMODITY.INVALID_COMPONENT).
var alicloudRedisTierMap = map[cloudresourcesv1beta1.AlicloudRedisTier]redisTierInfo{
	cloudresourcesv1beta1.AlicloudRedisTierS1: {instanceClass: "tair.rdb.1g", readOnlyCount: 0},
	cloudresourcesv1beta1.AlicloudRedisTierS2: {instanceClass: "tair.rdb.2g", readOnlyCount: 0},
	cloudresourcesv1beta1.AlicloudRedisTierS3: {instanceClass: "tair.rdb.4g", readOnlyCount: 0},
	cloudresourcesv1beta1.AlicloudRedisTierS4: {instanceClass: "tair.rdb.8g", readOnlyCount: 0},
	cloudresourcesv1beta1.AlicloudRedisTierS5: {instanceClass: "tair.rdb.16g", readOnlyCount: 0},

	cloudresourcesv1beta1.AlicloudRedisTierP1: {instanceClass: "tair.rdb.1g", readOnlyCount: 0},
	cloudresourcesv1beta1.AlicloudRedisTierP2: {instanceClass: "tair.rdb.2g", readOnlyCount: 0},
	cloudresourcesv1beta1.AlicloudRedisTierP3: {instanceClass: "tair.rdb.4g", readOnlyCount: 0},
	cloudresourcesv1beta1.AlicloudRedisTierP4: {instanceClass: "tair.rdb.8g", readOnlyCount: 0},
	cloudresourcesv1beta1.AlicloudRedisTierP5: {instanceClass: "tair.rdb.16g", readOnlyCount: 0},
}

func redisTierToInstanceClassAndReadOnlyCount(tier cloudresourcesv1beta1.AlicloudRedisTier) (string, int32, error) {
	info, exists := alicloudRedisTierMap[tier]
	if !exists {
		return "", 0, errors.New("unknown redis tier")
	}
	return info.instanceClass, info.readOnlyCount, nil
}
