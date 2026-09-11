# AlicloudVpcPeering Custom Resource

The `alicloudvpcpeering.cloud-resources.kyma-project.io` custom resource (CR) specifies the virtual network peering between Kyma and a remote Alibaba Cloud Virtual Private Cloud (VPC) network. Virtual network peering is only possible within the networks of the same cloud provider.

Once an `AlicloudVpcPeering` CR is created and reconciled, the Cloud Manager controller creates a VPC peering connection in the Kyma cluster underlying Alibaba Cloud landscape. For cross-account peerings, the controller accepts the peering connection in the remote account automatically.

## Specification

This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Parameter                          | Type    | Description                                                                                                                                                                                                                   |
|------------------------------------|---------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **remoteAccountId**                | string  | Required. Specifies the Alibaba Cloud account UID (AliUid) of the owner of the accepter VPC.                                                                                                                                  |
| **remoteRegion**                   | string  | Required. Specifies the region ID of the accepter VPC (for example, `cn-shanghai`).                                                                                                                                           |
| **remoteVpcId**                    | string  | Required. Specifies the ID of the VPC with which you are creating the VPC peering connection.                                                                                                                                 |
| **deleteRemotePeering**            | boolean | Optional. When `true`, Cloud Manager deletes the remote route entries pointing at this connection when the CR is deleted. When `false`, only local route entries are removed. Defaults to `false`.                             |
| **remoteRouteTableUpdateStrategy** | string  | Optional. Specifies the remote route table update strategy. The value is one of the following: `AUTO`, `MATCHED`, `UNMATCHED`, or `NONE`. Defaults to `AUTO`. For more information, see [RemoteRouteTableUpdateStrategy](#remoteroutetableupdatestrategy). |
| **bandwidth**                      | integer | Optional. Specifies the bandwidth for cross-region peering connections in Mbit/s. Relevant only when the Kyma cluster region differs from `remoteRegion`. Cross-region bandwidth is billed by Alibaba Cloud. If omitted or set to `0`, defaults to 1024 Mbit/s. Ignored for same-region peerings. |

**Status:**

| Parameter                         | Type       | Description                                                                                |
|-----------------------------------|------------|--------------------------------------------------------------------------------------------|
| **state**                         | string     | Signifies the current state of the custom resource.                                        |
| **conditions**                    | \[\]object | Represents the current state of the CR's conditions.                                       |
| **conditions.lastTransitionTime** | string     | Defines the date of the last condition status change.                                      |
| **conditions.message**            | string     | Provides more details about the condition status change.                                   |
| **conditions.reason**             | string     | Defines the reason for the condition status change.                                        |
| **conditions.status** (required)  | string     | Represents the status of the condition. The value is either `True`, `False`, or `Unknown`. |
| **conditions.type**               | string     | Provides a short description of the condition.                                             |

## RemoteRouteTableUpdateStrategy

To enable private IPv4 traffic between instances in peered VPC networks, Cloud Manager adds a peering route to the route tables of the remote VPC network. The route destination is the CIDR block of the Kyma VPC network, and the next hop is the VPC peering connection instance ID.

The `remoteRouteTableUpdateStrategy` parameter specifies how Cloud Manager handles remote route tables:
- `AUTO` adds a peering route to all remote route tables.
- `MATCHED` adds a peering route to all remote route tables with the Kyma shoot name tag.
- `UNMATCHED` adds a peering route to all remote route tables without the Kyma shoot name tag.
- `NONE` does not interact with remote route tables.

## deleteRemotePeering Semantics

Unlike AWS (where both sides have a separate peering object), an AliCloud VPC peering connection is a single object owned by the requester. When `deleteRemotePeering` is set to `true`, Cloud Manager deletes the remote route entries pointing at this connection before deleting the connection itself. When `deleteRemotePeering` is `false`, only local route entries and the connection are removed; the caller is responsible for cleaning up remote routes.

## Sample Custom Resource

See an exemplary `AlicloudVpcPeering` custom resource:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AlicloudVpcPeering
metadata:
  name: peering-to-my-vpc
spec:
  remoteVpcId: vpc-bp1abc123def456gh
  remoteRegion: cn-shanghai
  remoteAccountId: "123456789012345678"
```

Cross-region peering with explicit bandwidth:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AlicloudVpcPeering
metadata:
  name: peering-to-remote-region
spec:
  remoteVpcId: vpc-bp1abc123def456gh
  remoteRegion: cn-beijing
  remoteAccountId: "123456789012345678"
  bandwidth: 2048
  deleteRemotePeering: true
```
