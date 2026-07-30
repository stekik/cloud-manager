#!/usr/bin/env bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source $SCRIPT_DIR/.env
source $SCRIPT_DIR/_common.sh
source $SCRIPT_DIR/_common-alicloud.sh

alicloudInit

createUser() {
  local user_name=$1

  local exists
  exists=$(aliyun ram GetUser --UserName "$user_name" 2>/dev/null | jq -r '.User.UserName // empty')
  if [[ -z "$exists" ]]; then
    log "RAM user $user_name does not exist, creating it now..."
    aliyun ram CreateUser --UserName "$user_name" > /dev/null
    log "RAM user $user_name is created"
  else
    log "RAM user $user_name already exists"
  fi

  return 0
}

createPolicy() {
  local policy_name=$1
  local policy_file=$2

  local exists
  exists=$(aliyun ram GetPolicy --PolicyType Custom --PolicyName "$policy_name" 2>/dev/null | jq -r '.Policy.PolicyName // empty')
  if [[ -z "$exists" ]]; then
    log "RAM policy $policy_name does not exist, creating it now from $policy_file..."
    local policy_doc
    policy_doc=$(cat "$policy_file")
    aliyun ram CreatePolicy --PolicyName "$policy_name" --PolicyDocument "$policy_doc" > /dev/null
    log "RAM policy $policy_name is created"
  else
    log "RAM policy $policy_name already exists, creating new version..."
    local policy_doc
    policy_doc=$(cat "$policy_file")
    aliyun ram CreatePolicyVersion --PolicyName "$policy_name" --PolicyDocument "$policy_doc" --SetAsDefault true > /dev/null
    # Keep at most 5 versions; delete the oldest non-default ones if needed
    local versions
    versions=$(aliyun ram ListPolicyVersions --PolicyType Custom --PolicyName "$policy_name" \
      | jq -r '.PolicyVersions.PolicyVersion[] | select(.IsDefaultVersion == false) | .VersionId' \
      | sort | head -n -4)
    for v in $versions; do
      log "Deleting old policy version $v..."
      aliyun ram DeletePolicyVersion --PolicyName "$policy_name" --VersionId "$v" > /dev/null
    done
  fi

  return 0
}

attachPolicyToUser() {
  local user_name=$1
  local policy_name=$2

  log "Attaching policy $policy_name to RAM user $user_name"
  aliyun ram AttachPolicyToUser --PolicyType Custom --PolicyName "$policy_name" --UserName "$user_name" > /dev/null

  return 0
}

createUser "$SA_NAME_DEFAULT"
createPolicy "$POLICY_NAME_DEFAULT" "$POLICY_FILE_DEFAULT"
attachPolicyToUser "$SA_NAME_DEFAULT" "$POLICY_NAME_DEFAULT"
