# BE 이미지 (D-9). base 이미지는 digest로 고정한다.
# 1단계: Go 컴파일. 결과물은 정적 바이너리 하나.
FROM golang:1.26.5-alpine@sha256:0178a641fbb4858c5f1b48e34bdaabe0350a330a1b1149aabd498d0699ff5fb2 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
# commit과 빌드 시각은 /api/version이 돌려준다(D-3). 빌드할 때 --build-arg로 넘긴다.
ARG COMMIT=dev
ARG BUILT_AT=unknown
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags "-s -w -X main.commit=${COMMIT} -X main.builtAt=${BUILT_AT}" -o /out/be .

# 2단계: 실행. 셸·패키지 관리자 없는 이미지에 바이너리만 넣고, root가 아닌 사용자로 실행한다.
FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab
COPY --from=build /out/be /be
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/be"]
