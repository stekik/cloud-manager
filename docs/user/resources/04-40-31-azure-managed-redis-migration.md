# Migrating from AzureRedisInstance / AzureRedisCluster to AzureManagedRedis

> [!WARNING]
> AzureManagedRedis is a beta feature available only per request for SAP-internal teams.

This guide explains how to move existing workloads from the legacy [`AzureRedisInstance`](./04-40-30-azure-redis-instance.md) and [`AzureRedisCluster`](./04-50-30-azure-redis-cluster.md) resources (backed by `Microsoft.Cache/Redis`, the original *Azure Cache for Redis* service) to the new [`AzureManagedRedis`](./04-40-32-azure-managed-redis.md) resource (backed by `Microsoft.Cache/redisEnterprise`, the new *Azure Managed Redis* service).

## Why migrate

Microsoft is retiring legacy Azure Cache for Redis tiers in favor of [Azure Managed Redis (AMR)](https://learn.microsoft.com/en-us/azure/redis/overview). New deployments should target `AzureManagedRedis`; existing `AzureRedisInstance` / `AzureRedisCluster` workloads continue to work but should be migrated before their underlying SKUs are deprecated.

## Migration is not in-place

There is **no in-place migration path** in Cloud Manager:

- Tier, SKU, high-availability, and clustering policy are all immutable on `AzureManagedRedis`.
- The two resource kinds are backed by different Azure services with different APIs, different SKU shapes, and different connection ports.
- No data is automatically moved. Plan a cutover that includes either a brief read-only window or an application-level dual-write/replication strategy.

The recommended pattern is:

1. Provision a new `AzureManagedRedis` alongside the existing legacy resource.
2. Repoint applications at the new connection Secret (or run dual reads/writes).
3. Migrate data using your preferred Redis tooling (e.g. `redis-cli --rdb`, `RIOT`, application-level replay). [Microsoft's migration guide](https://learn.microsoft.com/en-us/azure/redis/migrate/migrate-redis-enterprise-understand) covers options on the Azure side.
4. Once all clients are cut over, delete the legacy resource.

## Tier mapping

The legacy `redisTier` letter does not map 1:1 to AMR tiers — AMR uses different SKU families and exposes a separate `C*` clustering line.

### From AzureRedisInstance (legacy `S*`/`P*`)

| Legacy tier | Capacity (GiB) | Recommended AMR tier | AMR SKU                | Notes                                                              |
|-------------|----------------|----------------------|------------------------|--------------------------------------------------------------------|
| `S1`        | 1              | `S1`                 | `Balanced_B0`          | Dev/test single-node, no HA. Memory shape matches.                 |
| `S2`        | 2.5            | `S2` (3 GiB)         | `Balanced_B3`          | Closest AMR step up.                                               |
| `S3`        | 6              | `S3`                 | `Balanced_B5`          | Memory matches.                                                    |
| `S4`        | 13             | `S4` (12 GiB)        | `Balanced_B10`         | Slightly smaller; oversize to `S5` if your working set is near 13. |
| `S5`        | 26             | `S5` (24 GiB)        | `Balanced_B20`         | Slightly smaller; oversize to a `P*` tier if HA is desired.        |
| `P1`        | 6              | `P1`                 | `ComputeOptimized_X5`  | HA single-shard. Memory matches.                                   |
| `P2`        | 13             | `P2` (12 GiB)        | `ComputeOptimized_X10` | Slightly smaller.                                                  |
| `P3`        | 26             | `P3` (24 GiB)        | `ComputeOptimized_X20` | Slightly smaller.                                                  |
| `P4`        | 53             | `P4` (60 GiB)        | `ComputeOptimized_X50` | Larger; closest AMR step.                                          |
| `P5`        | 120            | `P5` (120 GiB)       | `ComputeOptimized_X100`| Memory matches.                                                    |

### From AzureRedisCluster (legacy `C*`)

| Legacy tier | Capacity (GiB) | Recommended AMR tier | AMR SKU                | Notes                                                       |
|-------------|----------------|----------------------|------------------------|-------------------------------------------------------------|
| `C3`        | 6              | `C3`                 | `ComputeOptimized_X5`  | Sharded HA. Memory matches.                                 |
| `C4`        | 13             | `C4` (12 GiB)        | `ComputeOptimized_X10` | Slightly smaller.                                           |
| `C5`        | 26             | `C5` (24 GiB)        | `ComputeOptimized_X20` | Slightly smaller.                                           |
| `C6`        | 53             | `C6` (60 GiB)        | `ComputeOptimized_X50` | Larger; closest AMR step.                                   |
| `C7`        | 160            | `C7` (120 GiB)       | `ComputeOptimized_X100`| Smaller. If your working set requires 160 GiB, plan ahead.  |

> [!NOTE]
> The legacy `AzureRedisCluster` exposes `shardCount` and `replicasPerPrimary`. AMR does not — sharding and replication are determined entirely by the chosen tier. There is no equivalent knob.

## Spec field mapping

| Legacy field                     | AMR equivalent                                | Notes                                                                                            |
|----------------------------------|-----------------------------------------------|--------------------------------------------------------------------------------------------------|
| `spec.redisTier`                 | `spec.redisTier`                              | Same field name; values do not overlap with legacy tier letters. See tier mapping above.         |
| `spec.ipRange`                   | `spec.ipRange`                                | Same shape. Optional; defaults to the auto-managed IpRange.                                      |
| `spec.authSecret.*`              | `spec.authSecret.*`                           | Identical (`name`, `labels`, `annotations`, `extraData`).                                        |
| `spec.redisVersion`              | *(removed)*                                   | AMR uses Redis 7.x; version is fixed by the service.                                             |
| `spec.redisConfiguration.*`      | *(removed)*                                   | `maxclients`, `maxmemory-*` knobs are not exposed on AMR. Azure-side defaults apply.             |
| `spec.shardCount`                | *(removed)*                                   | Sharding is implied by the `C*` tier.                                                            |
| `spec.replicasPerPrimary`        | *(removed)*                                   | HA is implied by the tier letter (`P*`/`C*`).                                                    |

If your existing manifests rely on `redisConfiguration` or `redisVersion`, drop those fields when porting to `AzureManagedRedis` — they will be rejected by admission validation.

## Connection secret differences

The core key names overlap, so apps that already read the well-known keys mostly need only the Secret name updated. Pay attention to the differences below:

| Secret key                       | Legacy `AzureRedisInstance` / `AzureRedisCluster` | `AzureManagedRedis`                       |
|----------------------------------|---------------------------------------------------|-------------------------------------------|
| `host`                           | Hostname (`*.redis.cache.windows.net`)            | Hostname (`*.redis.azure.net`)            |
| `port`                           | Azure TLS port (`6380`)                           | `10000`                                   |
| `primaryEndpoint`                | `<host>:<port>` — includes port                   | `<host>` only — does **not** include port |
| `authString`                     | Access key                                        | Access key                                |
| `readEndpoint` / `readHost` / `readPort` | Present (read replica / discovery endpoint) | **Not present** — single endpoint only    |

> [!IMPORTANT]
> The most common breakage during cutover is the **port change** and the **`primaryEndpoint` shape change**. Clients that hardcode `6380` or that parse `host:port` out of `primaryEndpoint` need updating. Prefer reading `host` and `port` as separate keys.

> [!NOTE]
> AMR's `primaryEndpoint` is the hostname only, not `host:port`. This differs from the legacy resources, which embed the port. AMR enforces a single fixed port (`10000`), so concatenating it adds no information.

> [!NOTE]
> AMR does not expose a separate read endpoint. Clients that send read traffic to `readHost`/`readEndpoint` must be reconfigured to use the primary endpoint.

## Networking differences

- Both legacy and AMR resources are private-only (no public network access). Connections must originate from inside the SKR network.
- AMR self-manages its own Private DNS zone (`privatelink.redis.azure.net`) and Virtual Network Link to the shoot's VNet. The legacy zone (`privatelink.redis.cache.windows.net`) is unaffected — the two coexist on the same SKR.
- Both kinds reuse the same `IpRange` mechanism; you can keep an existing `IpRange` reference, or omit it and let Cloud Manager use the default.

## Cutover checklist

- [ ] Pick the AMR tier from the [tier mapping](#tier-mapping) above. If your working set sits near the boundary of a smaller AMR step, oversize.
- [ ] Strip `redisVersion`, `redisConfiguration`, `shardCount`, and `replicasPerPrimary` from your manifests.
- [ ] Create the `AzureManagedRedis` resource. Provisioning takes 10–15 minutes — wait for `status.state: Ready`.
- [ ] Update applications to read the new connection Secret. Account for the different `host`, `port: 10000`, the hostname-only `primaryEndpoint`, and the absence of `readEndpoint`/`readHost`/`readPort`.
- [ ] Migrate Redis data using your tool of choice. Cloud Manager does not move data automatically.
- [ ] Validate end-to-end traffic against the new endpoint.
- [ ] Delete the legacy `AzureRedisInstance` / `AzureRedisCluster` resource.

## See also

- [AzureManagedRedis Custom Resource reference](./04-40-32-azure-managed-redis.md)
- [AzureRedisInstance Custom Resource reference](./04-40-30-azure-redis-instance.md) (legacy)
- [AzureRedisCluster Custom Resource reference](./04-50-30-azure-redis-cluster.md) (legacy)
- [Microsoft: Understand migrating from Redis Enterprise](https://learn.microsoft.com/en-us/azure/redis/migrate/migrate-redis-enterprise-understand)
