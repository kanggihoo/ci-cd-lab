#!/usr/bin/env bash
# 저장소 설정을 코드로 적용한다(D-15). 관리자 권한이 필요하므로 사용자만 실행한다.
# 여러 번 실행해도 같은 결과가 되도록 작성했다(ruleset은 이름으로 찾아 갱신).
set -euo pipefail

REPO="${REPO:-kanggihoo/ci-cd-lab}"
RULESET_NAME="main-protection"
GITHUB_ACTIONS_APP_ID=15368  # gh api apps/github-actions --jq .id

# D-8: 워크플로 기본 토큰은 읽기 전용, PR 승인 권한 없음.
gh api -X PUT "repos/$REPO/actions/permissions/workflow" \
  -f default_workflow_permissions=read \
  -F can_approve_pull_request_reviews=false

# D-5: 외부 기여자의 워크플로는 매번 승인 후 실행.
gh api -X PUT "repos/$REPO/actions/permissions/fork-pr-contributor-approval" \
  -f approval_policy=all_external_contributors

# D-4: squash 병합만 허용.
gh api -X PATCH "repos/$REPO" \
  -F allow_squash_merge=true \
  -F allow_merge_commit=false \
  -F allow_rebase_merge=false \
  --silent

# D-4·D-5: main ruleset.
#  - pull_request: main 직접 push 금지(PR 필수). 승인 인원 0(1인 운영).
#  - required_status_checks: gate 필수, 출처는 GitHub Actions 앱으로 고정, 최신 기준 병합(strict).
#  - non_fast_forward: force push 금지. deletion: main 삭제 금지.
#  - bypass_actors 비움: 관리자도 우회 불가.
RULESET=$(cat <<JSON
{
  "name": "$RULESET_NAME",
  "target": "branch",
  "enforcement": "active",
  "bypass_actors": [],
  "conditions": { "ref_name": { "include": ["~DEFAULT_BRANCH"], "exclude": [] } },
  "rules": [
    { "type": "deletion" },
    { "type": "non_fast_forward" },
    { "type": "pull_request", "parameters": {
        "required_approving_review_count": 0,
        "dismiss_stale_reviews_on_push": false,
        "require_code_owner_review": false,
        "require_last_push_approval": false,
        "required_review_thread_resolution": false,
        "allowed_merge_methods": ["squash"]
    } },
    { "type": "required_status_checks", "parameters": {
        "strict_required_status_checks_policy": true,
        "do_not_enforce_on_create": false,
        "required_status_checks": [ { "context": "gate", "integration_id": $GITHUB_ACTIONS_APP_ID } ]
    } }
  ]
}
JSON
)

ID=$(gh api "repos/$REPO/rulesets" --jq ".[] | select(.name == \"$RULESET_NAME\") | .id")
if [ -n "$ID" ]; then
  echo "$RULESET" | gh api -X PUT "repos/$REPO/rulesets/$ID" --input - --silent
  echo "ruleset $RULESET_NAME updated (id $ID)"
else
  echo "$RULESET" | gh api -X POST "repos/$REPO/rulesets" --input - --silent
  echo "ruleset $RULESET_NAME created"
fi
