---
status: accepted
---

# dev와 prod를 한 VPS의 분리된 compose 프로젝트로 둔다

처음에는 dev를 미니PC에 두려 했지만, VPS가 8GB RAM에 약 2~3GB만 쓰고 있어 여유가 약 5GB이고 빌드는 GitHub 서버에서 하므로 VPS에서도 충분하다. dev는 `lab-dev`, prod는 `lab-prod` compose 프로젝트로 같은 VPS에 두되, 별도 OS 사용자·Docker 망·`mem_limit`으로 나누고 dev는 tailnet 안에서만 접속한다. 미니PC는 dev 이전 연습, k3s, staging, 백업 수신에 후속으로 쓴다.

## Considered Options

- **dev를 미니PC에**: 도착·세팅을 기다려야 하고, 단계 3의 병목이 장비가 아니라 tailnet·helper 설치라서 이득이 없었다.
- **VPS를 하나 더 구입**: 비용이 들고 실습 목적에 비해 과하다.

## Consequences

- dev가 침해되면 prod까지의 거리가 같은 호스트 안이라 짧다. 그래서 OS 사용자·Docker 망·메모리 분리와 환경별 SSH 키가 필수 조건이다.
- 운영 중인 `vps-infra` 호스트에 root helper와 sudoers를 설치하므로 범위를 lab 경로로 엄격히 좁혀야 한다(다른 서비스에 영향을 주지 않을 것).
- 미니PC로 dev를 옮기는 일은 대상 서버만 바꿔 같은 helper가 동작하는지 확인하는 후속 실습이 된다.
- 같은 호스트의 기존 Jenkins는 `docker.sock`을 마운트해 root와 같은 권한이고 GitHub PAT와 SOPS 키를 가졌다(vps-infra `jenkins/compose.yml`에서 확인). Jenkins가 침해되면 lab 두 환경도 함께 무너진다. 이는 lab이 아니라 기존 서버의 위험이며 Jenkins 폐기(D-22) 때까지 잔존 위험으로 인정한다.
- Docker가 게시한 포트는 ufw를 우회한다. 그래서 lab 컨테이너 포트는 호스트 `127.0.0.1`(dev는 필요하면 tailnet IP)에만 게시한다.
- VPS 사용량 실측이 여유 부족으로 나오면 이 결정을 다시 연다.
