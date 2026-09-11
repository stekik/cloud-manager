# Creating VPC Peering in Alibaba Cloud

This tutorial explains how to create a Virtual Private Cloud (VPC) peering connection between a remote VPC network and SAP BTP, Kyma runtime in Alibaba Cloud.

## Prerequisites

* You have the Cloud Manager module added. See [Add and Delete a Kyma Module](https://help.sap.com/docs/btp/sap-business-technology-platform-internal/enable-and-disable-kyma-module?state=DRAFT&version=Internal#loio1b548e9ad4744b978b8b595288b0cb5c).
* You have the Alibaba Cloud CLI (`aliyun`) configured with credentials for both the local (Kyma) and remote accounts.

## Set Up a Test Environment in the Remote Account

1. Export the remote account and region.

   ```shell
   export REMOTE_ACCOUNT_ID=$(aliyun sts GetCallerIdentity --query AccountId --output text)
   export REMOTE_REGION=cn-shanghai
   export VPC_NAME=my-vpc
   export VPC_CIDR=192.168.0.0/16
   ```

2. Create a VPC network in the remote account.

   ```shell
   export VPC_ID=$(aliyun vpc CreateVpc \
     --RegionId $REMOTE_REGION \
     --CidrBlock $VPC_CIDR \
     --VpcName $VPC_NAME \
     --query VpcId --output text)
   ```

3. Create a vSwitch in the VPC.

   ```shell
   export ZONE_ID=$(aliyun ecs DescribeZones --RegionId $REMOTE_REGION \
     --query 'Zones.Zone[0].ZoneId' --output text)
   export VSWITCH_ID=$(aliyun vpc CreateVSwitch \
     --RegionId $REMOTE_REGION \
     --VpcId $VPC_ID \
     --ZoneId $ZONE_ID \
     --CidrBlock 192.168.0.0/24 \
     --VSwitchName my-vswitch \
     --query VSwitchId --output text)
   ```

4. Launch an ECS instance for connectivity testing.

   ```shell
   export IMAGE_ID=$(aliyun ecs DescribeImages \
     --RegionId $REMOTE_REGION \
     --OSType linux \
     --Architecture x86_64 \
     --ImageOwnerAlias system \
     --query 'Images.Image[0].ImageId' --output text)

   export INSTANCE_ID=$(aliyun ecs RunInstances \
     --RegionId $REMOTE_REGION \
     --ImageId $IMAGE_ID \
     --InstanceType ecs.t6-c1m1.small \
     --VSwitchId $VSWITCH_ID \
     --query 'InstanceIdSets.InstanceIdSet[0]' --output text)

   export PRIVATE_IP=$(aliyun ecs DescribeInstances \
     --RegionId $REMOTE_REGION \
     --InstanceIds "[\"$INSTANCE_ID\"]" \
     --query 'Instances.Instance[0].VpcAttributes.PrivateIpAddress.IpAddress[0]' --output text)
   ```

## Create VPC Peering

1. Create an `AlicloudVpcPeering` resource.

   ```shell
   kubectl apply -f - <<EOF
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: AlicloudVpcPeering
   metadata:
     name: peering-to-my-vpc
   spec:
     remoteAccountId: "$REMOTE_ACCOUNT_ID"
     remoteRegion: "$REMOTE_REGION"
     remoteVpcId: "$VPC_ID"
     deleteRemotePeering: true
   EOF
   ```

   > **Note:** For cross-region peerings, Alibaba Cloud charges for bandwidth. You can set the `bandwidth` field (in Mbit/s); if omitted, the default is 1024 Mbit/s.

2. Wait for the `AlicloudVpcPeering` CR to be in the `Ready` state.

   ```shell
   kubectl wait --for=condition=Ready alicloudvpcpeering/peering-to-my-vpc --timeout=300s
   ```

   Once the newly created `AlicloudVpcPeering` is provisioned, you should see the following message:

   ```txt
   alicloudvpcpeering.cloud-resources.kyma-project.io/peering-to-my-vpc condition met
   ```

3. Create a namespace and export its value as an environment variable.

   ```shell
   export NAMESPACE={NAMESPACE_NAME}
   kubectl create ns $NAMESPACE
   ```

4. Create a workload that pings the ECS instance in the remote network.

   ```shell
   kubectl apply -n $NAMESPACE -f - <<EOF
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: alicloudvpcpeering-demo
   spec:
     selector:
       matchLabels:
         app: alicloudvpcpeering-demo
     template:
       metadata:
         labels:
           app: alicloudvpcpeering-demo
       spec:
         containers:
         - name: my-container
           resources:
             limits:
               memory: 512Mi
               cpu: "1"
             requests:
               memory: 256Mi
               cpu: "0.2"
           image: ubuntu
           command:
             - "/bin/bash"
             - "-c"
             - "--"
           args:
            - "apt update; apt install iputils-ping -y; ping -c 20 $PRIVATE_IP"
   EOF
   ```

5. Print the logs to verify connectivity.

   ```shell
   kubectl logs -n $NAMESPACE \
     $(kubectl get pod -n $NAMESPACE -l app=alicloudvpcpeering-demo -o=jsonpath='{.items[0].metadata.name}')
   ```

   A successful output looks similar to:

   ```txt
   PING 192.168.0.1 (192.168.0.1) 56(84) bytes of data.
   64 bytes from 192.168.0.1: icmp_seq=1 ttl=63 time=2.10 ms
   ...
   20 packets transmitted, 20 received, 0% packet loss
   ```

## Next Steps

To clean up, follow these steps:

1. Remove the created workloads.

   ```shell
   kubectl delete -n $NAMESPACE deployment alicloudvpcpeering-demo
   ```

2. Remove the created `AlicloudVpcPeering` CR.

   ```shell
   kubectl delete alicloudvpcpeering peering-to-my-vpc
   ```

3. Remove the created namespace.

   ```shell
   kubectl delete namespace $NAMESPACE
   ```

4. Terminate the ECS instance and delete VPC resources in your Alibaba Cloud account.

   ```shell
   aliyun ecs StopInstances --RegionId $REMOTE_REGION \
     --InstanceId.1 $INSTANCE_ID --ForceStop true
   aliyun ecs DeleteInstances --RegionId $REMOTE_REGION \
     --InstanceId.1 $INSTANCE_ID --Force true
   aliyun vpc DeleteVSwitch --RegionId $REMOTE_REGION \
     --VSwitchId $VSWITCH_ID
   aliyun vpc DeleteVpc --RegionId $REMOTE_REGION \
     --VpcId $VPC_ID
   ```
