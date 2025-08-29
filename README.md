# kcloud-opt-core

AI반도체 클라우드 서비스의 핵심 비용 최적화 스케줄러 (Go)

## 개요

kcloud-opt-core는 GPU/NPU 자원을 포함한 AI반도체 워크로드의 비용 최적화 스케줄링을 담당하는 핵심 모듈입니다. Kubernetes Controller 패턴으로 구현되어 클라우드 네이티브 환경에서 최적화된 성능을 제공합니다.

## 주요 기능

- **비용 인지 스케줄링**: 워크로드별 비용 제약사항을 고려한 최적 노드 배치
- **에너지 효율 최적화**: 전력 사용량 기반 스케줄링 결정
- **동적 재배치**: 실시간 비용/성능 분석 기반 워크로드 재배치
- **OpenStack 연동**: Nova, Neutron, Cyborg를 통한 자원 관리

## 아키텍처

```
core/
├── cmd/scheduler/          # 메인 애플리케이션
├── src/
│   ├── scheduler/          # 스케줄링 알고리즘
│   ├── resource_manager/   # 자원 관리
│   └── api/               # REST API
├── config/                # 설정 파일
└── tests/                 # 테스트
```

## 빌드 및 실행

### 개발 환경

```bash
# 의존성 다운로드
make deps

# 빌드
make build

# 테스트
make test

# 로컬 실행
make run
```

### Docker

```bash
# 이미지 빌드
make docker-build

# 컨테이너 실행
docker run -p 8080:8080 kcloud-opt/core:latest
```

### Kubernetes

```bash
# CRD 적용
kubectl apply -f config/crd/

# 배포
make deploy
```

## 설정

### 환경변수

- `DB_HOST`: PostgreSQL 호스트
- `REDIS_HOST`: Redis 호스트  
- `OS_AUTH_URL`: OpenStack 인증 URL
- `LOG_LEVEL`: 로그 레벨 (DEBUG, INFO, WARN, ERROR)

### 설정 파일

`config/scheduler.yaml`에서 스케줄링 정책을 설정합니다:

```yaml
scheduler:
  algorithm: cost_aware
  policies:
    cost_optimization:
      weight: 0.6
    energy_efficiency:
      weight: 0.4
```

## API 엔드포인트

- `GET /health`: 헬스체크
- `GET /ready`: 준비상태 확인
- `GET /metrics`: Prometheus 메트릭
- `POST /api/v1/schedule`: 워크로드 스케줄링 요청
- `GET /api/v1/nodes`: 노드 상태 조회

## 모니터링

- **Prometheus 메트릭**: `:9090/metrics`
- **로그**: JSON 형태 구조화 로그
- **헬스체크**: `:8080/health`

## 개발

### 요구사항

- Go 1.21+
- Docker
- kubectl (K8s 배포 시)

### 코드 품질

```bash
# 포맷팅
make fmt

# 린팅  
make lint

# 커버리지 테스트
make test-coverage
```

## 라이선스

Apache License 2.0