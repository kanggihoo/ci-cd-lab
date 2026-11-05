---
status: accepted
---

# 배포는 root 소유의 고정 helper 하나로만 한다

공개 저장소의 CI가 서버에 배포하는데, 배포 계정이 docker 그룹에 있거나 CI가 보낸 compose를 그대로 실행하면 계정 탈취나 악성 compose(`privileged`, `/:/host` 마운트)가 곧바로 호스트 root가 된다. 그래서 배포 계정에는 docker 권한을 주지 않고, SSH forced command로 root 소유 helper 하나만 호출하게 한다. helper는 관리자가 설치한 고정 compose와 허용된 인자(환경, digest)만 쓰고, 출처 증명·소스 커밋의 `main` 포함·다운그레이드 여부를 환경 잠금 안에서 검사한다. 접속 키는 `restrict`와 forced command로 이 helper 호출만 하게 하고, 인자는 환경과 digest 형식으로만 제한한다. 잠금과 현재 릴리스 기록은 root 소유 디렉터리에 둔다.

## Considered Options

- **self-hosted 러너**: 공개 저장소의 PR 코드가 서버에서 실행될 수 있어 버렸다.
- **docker 그룹 계정 + SSH**: 사실상 root라서 버렸다.
- **CI가 compose를 보내 실행**: `main`에 들어간 compose가 곧 root 경로가 된다. 브랜치 보호는 검토 경계이지 권한 제한이 아니다.
- **Watchtower식 최신 태그 추적**: 승인된 digest를 명시하는 배포·승격·순서 실습에 맞지 않고 감사가 어렵다.

## Consequences

- compose 변경은 관리자가 서버에 다시 설치해야 하므로, 설치본과 저장소 사본이 다르면 점검에서 드러나게 한다.
- 출처 증명은 허용할 이미지의 조건이지 prod 승인의 증명이 아니다. prod 승인은 Environment 승인 뒤에만 prod 접속 자격을 CI에 주는 것으로 지킨다.
- 해당 환경의 접속 자격과 tailnet 접근이 함께 유출되면 출처가 유효한 이미지를 고르는 것까지는 막지 못한다. 유출 시 키 폐기·교체를 런북으로 둔다.
- 승인된 앱 이미지가 자신의 DB와 비밀을 악용하는 것은 이 경계로 막지 못한다.
- 다운그레이드 가드는 그 환경에 과거 성공 배포된 릴리스로의 롤백만 예외로 둔다. 안 그러면 장애 때 직전 정상 버전으로도 돌아가지 못한다.
- Tailscale은 전송 경로일 뿐이고 한 VPS의 dev·prod 분리는 환경별 키와 Environment가 맡는다.
- 아래는 알면서 감수한 것이다. 1인 운영이라 `.github/**` 변경과 prod 승인(Environment 클릭)을 독립 검토할 사람이 없다(승인은 계정 침해가 아니라 워크플로 침해를 막는다). 출처 검증은 GitHub·Sigstore에 의존해 그쪽 장애가 곧 배포 중단이고, 서버가 증명 번들을 어떻게 받는지는 spec에서 확인한다. 악성으로 판명된 `main` 커밋을 차단·폐기하는 경로는 배포 반복 전에 정한다. 공개 로그를 스캔하는 도구는 두지 않는다.
