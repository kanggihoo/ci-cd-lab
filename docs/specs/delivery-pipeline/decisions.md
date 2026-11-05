# delivery-pipeline decisions

Glossary: [GLOSSARY.md](../../../GLOSSARY.md) · ADRs: [0001](../../adr/0001-root-owned-fixed-deploy-helper.md), [0002](../../adr/0002-dev-and-prod-on-one-vps.md)

이번 반복(iteration 1)은 단계 0–2(계약·위협 모델 → PR CI → 이미지·버전)다. D-1~D-17은 이번 반복의 결정이고, D-18~D-22는 단계 3 이후 반복에서 grill로 확정할 provisional 결정이다. 각 결정이 막는 공격의 설명은 `~/Desktop/ai-tutor/cicd/`에서 학습 세션으로 푼다(`Rejected` 줄에 한 줄 이유를 둔다).

## D-1 저장소와 범위
- Status: settled
- Decision: 로컬 `~/Desktop/vps/ci-cd-lab`, GitHub `kanggihoo/ci-cd-lab`(공개). 이번 반복은 단계 0–2. GitHub 저장소 생성과 push는 사용자가 지시한 뒤에만 한다.
- Rejected: 스펙을 `ai-tutor` 저장소에 두기 → 저장소가 다른 프로젝트라 CI 워크플로와 섞인다.

## D-2 앱 구성
- Status: settled
- Decision: Go API(`net/http`) + Vite 정적 FE 한 페이지 + PostgreSQL. 인증 없음. API는 `/healthz`, `/api/version`, 단일 리소스 `notes`(목록 조회·추가)만 둔다. 앱은 CI/CD 실습이 목적이라 얇게 유지한다.
- Rejected: Spring + Flyway → 이번 목적에 불필요. 처음부터 k8s → 학습이 파이프라인 대신 오케스트레이션에 쏠린다.

## D-3 healthcheck와 버전 계약
- Status: settled
- Decision: `GET /healthz`는 DB 연결(1초 타임아웃)까지 확인한 준비 신호다. `GET /api/version`은 커밋 sha와 빌드 시각을 준다. FE는 상대 경로 `/api`를 쓰고 환경별 재빌드를 하지 않는다.
- Rejected: 환경별 `VITE_API_URL`로 FE 빌드 → 빌드 때 값이 박혀 "같은 이미지를 승격"이 깨진다.

## D-4 브랜치와 병합
- Status: settled
- Decision: `main` 트렁크 하나, `develop` 없음. dev/prod는 환경 이름이다. squash 병합만 허용한다. 규칙: `main` 직접 push 금지, PR 필수, `gate` 필수 체크, 최신 기준 병합, force push 금지, 관리자 우회 금지. 필수 체크 `gate`의 출처는 GitHub Actions로 고정한다. 승인 인원 수 필수는 켜지 않는다(1인이라 스스로를 막는다). 1인 운영이라 `.github/**` 변경을 독립 검토할 사람이 없는 점은 감수한다.
- Rejected: `develop`/`master` 장수 브랜치 → 한 번 빌드해 승격하는 구조에 불필요하고 GitHub는 Environment 승인이 있다. 머지 큐 → 1인 실습에 불필요, 가용 조건 미확인. 체크 출처 미고정 → 이름만 `gate`인 다른 앱·토큰의 체크도 통과한다.

## D-5 gate 필수 체크
- Status: settled
- Decision: 필수 체크는 `gate` 하나. 하위 잡의 실패·건너뜀·취소는 모두 실패로 판정한다. 경로 필터는 잡의 실행 여부만 정한다. 외부 기여자의 워크플로는 승인 후 실행하고 PR 잡은 시크릿을 쓰지 않는다.
- Rejected: 경로 필터로 잡이 건너뛰어지면 필수 체크가 비는 구성 → 검사 없이 병합될 수 있다.

## D-6 PR 검사와 차단 기준
- Status: settled
- Decision: gitleaks, govulncheck, `npm audit`, 이미지 스캔, hadolint, shellcheck, actionlint + zizmor를 PR 게이트에서 돌린다. govulncheck는 호출 경로 finding을 차단하고, npm·이미지는 high·critical이면서 수정 버전이 있으면 차단한다. 수정 불가 항목은 보고하고 수동 판단한다. 예외는 식별자·영향 판단·사유·담당자·만료일을 적어야 하며 만료되면 다시 차단한다. 스캐너 실패·DB 미수신은 성공으로 치지 않는다. 의존성과 액션 SHA 갱신은 Dependabot이 맡는다.
- Rejected: 모든 도구에 severity 하나를 일괄 적용 → 도구마다 신호의 성격이 달라 외부 CVE로 계속 빨개진다.

## D-7 시험 방식과 e2e
- Status: settled
- Decision: Go 단위 시험, 실제 Postgres(Actions `services`, 이미지 핀), FE 단위 시험은 최소로 둔다. 1단계 초반에 curl 스모크를 완성하고, 1단계 종료 전에 Chromium으로 FE→BE→DB의 CRUD 한 흐름을 시험한다. 그 시험은 FE·BE·compose·시험 코드·프록시 설정·lockfile·CI가 바뀌었거나 판단이 서지 않을 때 실행한다. 전체·위험 경로별 e2e는 이후 반복이다.
- Rejected: 매 PR 브라우저 전체 시험 → PR이 느려지고 불안정해져 게이트 구조 학습을 가린다. curl 스모크만 → 브라우저 JS→API 연결 오류를 못 잡는다.

## D-8 워크플로 보안
- Status: settled
- Decision: 워크플로 `permissions` 기본은 `contents: read`이고 저장소 기본 토큰 권한도 read로 둔다. 액션은 SHA로 고정한다. `pull_request_target`·`workflow_run`·`issue_comment`로 시크릿이나 쓰기 권한이 있는 잡을 돌리지 않는다. `actions/checkout`은 `persist-credentials: false`로 쓴다. PR 문자열을 셸에 직접 넣지 않는다(actionlint·zizmor가 검사한다). 시크릿·쓰기 권한이 있는 잡은 PR 코드·캐시·아티팩트를 실행하지 않는다. 이미지 게시는 PR 잡과 분리된 `main` push 전용 워크플로에서만 하고, Environment의 배포 브랜치는 `main`만 허용한다.
- Rejected: PR 잡과 게시 잡을 한 파일에 두고 `if`로 막기 → 같은 저장소 브랜치의 PR이 워크플로 파일(권한·조건)을 직접 고쳐 게시 권한을 켤 수 있다. 이미지 안의 의존성 내용까지 SHA 고정·출처 증명이 보증한다고 보기 → 그것들은 "누가 만들었나"만 보증하고 내용은 D-6의 스캔이 본다.

## D-9 이미지와 릴리스
- Status: settled
- Decision: `linux/amd64` 이미지 `ghcr.io/kanggihoo/ci-cd-lab/{be,fe,migrate}`를 GHCR 공개 패키지에 `sha-<커밋>` 태그로 올린다. 게시 단계는 그 태그가 이미 있으면 push를 실패시킨다. 릴리스는 세 이미지의 digest와 소스 커밋을 묶은 `release.json`(워크플로 아티팩트로 보관)이며, 복구와 배포 기록은 digest만 참조한다. 사람용 버전(`r000001`)과 GitHub Release는 prod 승격이 생기는 반복에서 만든다.
- Rejected: 비공개 패키지 → 서버에 pull 자격을 둬야 한다. sha 태그만으로 복구 → 같은 태그를 다시 push할 수 있다. 레지스트리의 태그 불변성에 기대기 → GHCR이 보장하는지 확인하지 못했고, 같은 커밋의 빌드를 재실행하면 같은 태그에 다른 digest가 올라갈 수 있다.

## D-10 릴리스 검증기
- Status: settled
- Decision: `release.json`의 각 digest에 대해 출처 증명을 검사하되, 증명의 워크플로 경로를 게시 워크플로로, ref를 `refs/heads/main`으로, predicate 타입을 함께 고정한다. 세 이미지(be·fe·migrate)의 증명 속 소스 커밋이 서로 같고 `release.json`의 커밋과도 같아야 통과하며 하나라도 어긋나면 비정상 종료한다. 이 CI 검증은 빠른 피드백이고 신뢰 경계는 서버 helper다. helper는 `release.json`을 믿지 않고 처음부터 다시 검증한다. 출처 증명은 허용할 이미지의 조건일 뿐 prod 승인의 증명이 아니다.
- Rejected: "출처가 검증되면 배포를 인가" → 정상 빌드된 다른 이미지도 통과하므로 승인을 대신하지 못한다. 저장소(`--repo`)만 지정한 검증 → 이 저장소의 어떤 워크플로(PR 브랜치 포함)가 만든 증명도 통과한다. CI 검증기만 믿기 → 검증기도 저장소 코드라 악성 PR이 바꿀 수 있다.

## D-11 마이그레이션
- Status: settled
- Decision: `golang-migrate`를 일회성 컨테이너로 실행한다. PR은 새 파일 추가만 허용하고 기존 파일의 수정·삭제·이름 변경과 번호 중복·역전은 실패시킨다. expand/contract만 허용하며 자동 역적용은 하지 않고 문제는 forward-fix로 고친다.
- Rejected: down 마이그레이션 자동 실행 → 데이터 손실 위험.

## D-12 env 계약
- Status: settled
- Decision: 저장소에는 env 계약 표와 `*.env.example`만 두고 가드가 둘의 키 일치를 검사한다. 값은 암호문 형태로도 저장소에 두지 않는다.
- Rejected: 공개 저장소에 SOPS 암호문 커밋 → 암호문이 영구 공개되고 키가 나중에 새면 과거 이력이 모두 풀린다.

## D-13 로그와 비밀 위생
- Status: settled
- Decision: 비밀·내부 토폴로지(사설 IP, tailnet 이름·주소, 서버 내부 경로)·개인정보는 로그에 출력하지 않는다. compose에 적힌 컨테이너 내부 포트와 환경 이름, 공개 호스트명과 IP는 비밀이 아니며 보안은 그 은닉에 의존하지 않는다. 실행 중에 알게 되는 tailnet 값은 쓰기 전에 `add-mask`를 건다. helper는 허용 목록 형식(PASS/FAIL과 digest)만 출력하고 원문 로그는 서버에만 남긴다. 공개 로그에 env 원문 해시를 남기지 않는다. 로그를 스캔하는 도구는 이번 반복에 두지 않는다.
- Rejected: 호스트·IP를 secrets로 숨기기 → 인증서 투명성 로그와 DNS가 이미 공개한다. env 해시 공개 → 낮은 엔트로피 비밀의 추측 확인 수단이 된다. 마스킹 기능에만 맡기기 → 등록된 비밀의 정확한 문자열만 가려서 Tailscale 액션·ssh 오류가 찍는 주소는 그대로 나간다. helper 출력을 그대로 전달 → compose 로그와 마이그레이션 SQL·행 데이터가 공개 로그로 흐른다.

## D-14 절대 일어나면 안 되는 일
- Status: settled
- Decision: (1) 포크 PR이나 같은 저장소의 `main`이 아닌 브랜치 등 신뢰할 수 없는 코드가 시크릿에 닿거나 쓰기 권한으로 실행됨. (2) 비밀·내부 토폴로지가 공개 로그·저장소·이미지에 남음. (3) 출처가 검증되지 않은 이미지가 릴리스로 기록됨. (4) 같은 sha 태그를 다른 내용으로 게시하거나 배포가 태그로 이미지를 고름(검증은 "태그가 이미 있으면 게시 실패, 배포는 digest만 참조"). (5) 필수 체크가 실패·건너뜀인데 병합됨. (6) 승인 없이 prod에 배포됨. (7) dev 자격으로 prod에 접근함. (8) helper가 관리자가 설치한 것 외의 compose·스크립트를 실행함. 여덟 가지 모두 인수 기준으로 옮긴다.
- Rejected: 위협을 "포크 PR"까지만 적기 → 같은 저장소 브랜치나 탈취된 토큰으로 오는 공격이 목록에서 빠진다. "태그를 덮어쓸 수 없다"를 기준으로 삼기 → 레지스트리가 강제하지 않아 검증이 불가능하다.

## D-15 저장소 설정의 코드화
- Status: settled
- Decision: ruleset·외부 기여자 승인·Environment 등 저장소 설정은 `gh api` 스크립트로 저장소에 커밋한다. 저장소 기본 토큰 권한 read, 필수 체크 출처 고정, Environment 배포 브랜치 `main` 제한도 이 스크립트가 설정한다. 실행은 사용자만 한다(관리자 권한이 필요하며 CLI 에이전트와 CI는 실행하지 않는다).
- Rejected: 클릭으로만 설정 → 재현과 검토가 안 된다.

## D-16 0단계 산출물과 실패 시나리오
- Status: settled
- Decision: 위협 모델과 시크릿 위치표를 합친 1쪽, env 계약, health/version 계약을 문서로 만든다. 위협 모델은 "공격 → 막는 장치 → 안 막으면 → 직접 해보기" 표로 쓴다. 단계마다 일부러 깨뜨리기를 1회 한다. 1단계는 필수 체크가 실패해 병합이 막히는 것(후보: `gate`를 `exit 0`으로 고친 PR이 통과하는지 보고 "1인 운영의 `.github/**` 미검토"라는 남은 구멍을 직접 확인). 2단계는 출처 증명이 없는 이미지가 검증기에 거부되는 것(후보: 다른 워크플로가 만든 증명도 거부되는지 확인).

## D-17 코드 재사용 방침
- Status: settled
- Decision: A302 팀 저장소의 코드는 공개 저장소에 복사하지 않고 설계 아이디어만 참고해 직접 다시 작성한다. 개인 저장소 `vps-infra`의 스크립트는 재사용해도 된다.
- Rejected: A302 스크립트 복사 → 팀 코드가 공개 저장소로 나간다.

## D-18 알림
- Status: provisional
- Decision: 알림은 환경·버전·결과·링크만 담은 공통 계약에 어댑터를 붙이는 구조로 한다. 비밀과 로그 본문은 보내지 않는다.
- Default: 이번 반복은 GitHub 기본 실패 이메일만 쓴다. 배포 반복부터 Telegram 어댑터와 GitHub 이메일을 함께 쓰고, 새벽 e2e의 누락은 별도 외부 heartbeat로 감시한다. 사용자가 다른 채널을 택하면 어댑터만 바꾼다.
- Rejected: 미니PC에 자체 호스팅 ntfy → 알림 서버가 감시 대상과 같은 운명이 된다.

## D-19 환경과 장비 배치
- Status: provisional
- Decision: dev는 VPS의 `lab-dev`, prod는 VPS의 `lab-prod`(`lab.kkh-hub.tech`)로 분리한다. 별도 OS 사용자·Docker 망·`mem_limit`을 쓰고 dev는 tailnet 전용이다. Docker가 게시한 포트는 ufw를 우회하므로 dev·prod 컨테이너 포트는 호스트의 `127.0.0.1`(dev는 필요하면 tailnet IP)에만 게시한다. 인수 기준은 "공개 인터넷에서 dev가 열리지 않는다"이다. 미니PC는 후속 실습(dev 이전, k3s, staging, 백업 수신)에 쓴다.
- Default: 위 배치가 이 결정의 전부다. VPS 사용량 실측이 여유 부족으로 나오면 다시 연다.
- Rejected: ufw 규칙만으로 dev 차단 → Docker 게시 포트는 ufw를 거치지 않아 열려 있을 수 있다.
- ADR: [0002](../../adr/0002-dev-and-prod-on-one-vps.md)

## D-20 배포 경로
- Status: provisional
- Decision: GitHub-hosted 배포 잡이 Tailscale 임시 노드(tag ACL: CI→대상 22번만, Tailscale 자격은 Environment 시크릿)로 접속해 환경별 SSH 키로 root 소유 helper 하나를 호출한다. 키는 `restrict`와 forced command를 쓰고 포워딩·pty를 금지하며 `authorized_keys`는 root 소유로 둔다. helper의 인자는 환경(`dev`/`prod`)과 digest(`sha256:`+64자리 16진)뿐이고 그 밖의 옵션은 거부한다. 배포 계정에는 docker 그룹·소켓이 없다. `tailscale up --ssh`는 끄고 그 사실을 확인한다. 환경 분리는 환경별 키와 Environment가 맡고 Tailscale은 전송 경로다. helper의 잠금·현재 릴리스 기록은 root 소유 디렉터리에 두고 `git`은 root 소유 클론에서 실행한다. 다운그레이드는 막되 그 환경에 과거 성공 배포된 릴리스로의 롤백은 허용한다. prod 접속 자격은 Environment 승인 뒤에만 CI에 제공한다.
- Default: 위 구조를 쓴다. 사용자가 Tailscale 계정 도입을 거절하면 접속 경로를 다시 정한다. 공개 22번 폐쇄는 tailnet 접속과 Hostinger 콘솔 복구를 실제 확인한 별도 관리 세션 뒤에만 한다.
- Rejected: forced command만 믿기 → 포트·에이전트 포워딩이 남아 loopback 전용 서비스로 우회할 수 있다. 모든 다운그레이드 거부 → 장애 때 직전 정상 버전으로도 못 돌아간다. Tailscale 태그로 dev·prod 분리 → 한 VPS는 한 노드·한 IP라 분리되지 않는다. Tailscale 자격을 저장소 시크릿으로 두기 → 모든 잡이 tailnet에 들어올 수 있다.
- ADR: [0001](../../adr/0001-root-owned-fixed-deploy-helper.md)

## D-21 프록시
- Status: provisional
- Decision: 각 환경에 같은 내부 Caddy 라우터(`/`→FE, `/api`→BE)를 두고 바깥 TLS 종단만 환경별로 다르게 한다. 내부 Caddy와 앱 포트는 호스트 `127.0.0.1`에만 게시한다. prod는 VPS의 기존 nginx가 종단하고 내부 Caddy는 직전 nginx만 신뢰한다. 완료 기준에 외부 HTTPS 기준 redirect·절대 URL과 헤더 위조 시험을 넣는다.
- Default: 위 구조를 쓴다. VPS 입구를 Caddy로 바꾸는 일은 별도 작업이고, k3s 단계에서 Traefik을 검토한다.
- Rejected: 지금 Traefik → Docker 소켓을 읽는 provider가 필요하고 이번 앱의 고정 라우팅에는 동적 발견이 필요 없다.

## D-22 prod·백업·마무리
- Status: provisional
- Decision: prod 승인은 "본인 수동 확인"이라 부르고 prevent-self-review는 켜지 않는다. 릴리스 보존은 성공 릴리스 묶음 N=5이며 현재 dev/prod·직전 정상·진행 중 배포가 참조하는 것은 보호한다. Jenkins 폐기는 마지막 단계에서 기존 Jenkins 작업 전체를 목록화하고 새 경로를 검증한 뒤 한다(운영 변경 주체는 하나).
- Default: 백업은 미니PC 세팅 뒤로 미룬다. prod에 의미 있는 데이터가 들어가기 전에 암호화된 오프호스트 백업과 별도 DB로의 복원 리허설을 통과해야 한다. 후보는 append-only 수신 서버이고, 보관처와 예산은 그때 사용자가 정한다.
