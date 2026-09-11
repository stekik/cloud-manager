Feature: AlicloudVpcPeering feature

  @skr @alicloud @peering
  Scenario: AlicloudVpcPeering scenario
    Given eventually timeout is "20m"

    Given there is shared SKR with "AliCloud" provider

    Given resource declaration:
      | Alias   | Kind               | ApiVersion                              | Name        | Namespace |
      | peering | AlicloudVpcPeering | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()} |           |

    Given tf module "tf" is applied:
      | source   | ./alicloud-peering-target   |
      | provider | aliyun/alicloud@1.224       |
      | region   | "ap-northeast-1"            |
      | name     | "${_.peering.name}"         |
      | vpc_cidr | "192.168.0.0/16"            |

    When resource "peering" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: AlicloudVpcPeering
      spec:
        remoteAccountId: "${tf.account_id}"
        remoteRegion: "ap-northeast-1"
        remoteVpcId: "${tf.vpc_id}"
        deleteRemotePeering: true
      """

    Then eventually "peering.status.state == 'Ready'" is ok, unless:
      | peering.status.state == 'Error' |

    When resource "peering" is deleted
    Then eventually resource "peering" does not exist

    Then tf module "tf" is destroyed
