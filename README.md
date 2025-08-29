# kcloud-opt-operator

Kubernetes Operator for AI반도체 워크로드 최적화 (Go)

## 개요

kcloud-opt-operator는 Kubernetes 환경에서 AI반도체 워크로드의 비용 최적화를 자동화하는 Kubernetes Operator입니다. Custom Resource Definitions(CRD)를 통해 워크로드 최적화 정책을 선언적으로 관리하고, Controller 패턴으로 실제 최적화를 실행합니다.

## 주요 기능

### Custom Resource 관리
- **WorkloadOptimizer CRD**: 워크로드별 최적화 정책 정의
- **CostPolicy CRD**: 비용 제약사항 및 예산 관리
- **PowerPolicy CRD**: 전력 사용량 제약사항 관리

### 자동화된 최적화
- **스케줄링 최적화**: 비용/전력 효율적인 노드 배치
- **Auto-scaling**: 예측 기반 워크로드 스케일링
- **리소스 재배치**: 실시간 최적화를 통한 워크로드 마이그레이션

### Kubernetes 네이티브
- **Admission Webhook**: 워크로드 생성 시 자동 최적화 정책 적용
- **Finalizer**: 워크로드 삭제 시 리소스 정리
- **Event 기반**: Kubernetes 이벤트 기반 반응형 최적화

## 아키텍처

```
operator/
├── api/v1alpha1/            # CRD 정의
├── controllers/             # Controller 로직  
├── cmd/manager/             # Operator 메인
├── pkg/
│   ├── webhook/            # Admission Webhook
│   ├── scheduler/          # 스케줄링 로직
│   └── optimizer/          # 최적화 엔진
├── config/
│   ├── crd/               # CRD 매니페스트
│   ├── rbac/              # RBAC 설정
│   └── webhook/           # Webhook 설정
└── hack/                  # 개발 스크립트
```

## CRD 정의

### WorkloadOptimizer

```yaml
apiVersion: kcloud.io/v1alpha1
kind: WorkloadOptimizer
metadata:
  name: ml-training-optimizer
  namespace: ai-workloads
spec:
  workloadType: "training"
  priority: 80
  resources:
    cpu: "16"
    memory: "64Gi"
    gpu: 4
    npu: 0
  costConstraints:
    maxCostPerHour: 50.0
    preferSpot: true
    budgetLimit: 1200.0
  powerConstraints:
    maxPowerUsage: 2000.0  # Watts
    preferGreen: true
  placementPolicy:
    nodeSelector:
      accelerator: nvidia-gpu
    affinity:
    - type: "gpu_workload"
      key: "gpu.nvidia.com/class"
      value: "compute"
      weight: 100
  autoScaling:
    minReplicas: 1
    maxReplicas: 10
    metrics:
    - type: "cost"
      threshold: 80
    - type: "power"
      threshold: 1800
status:
  phase: "Optimizing"
  currentCost: 42.5
  currentPower: 1650.0
  assignedNode: "gpu-node-001"
  optimizationScore: 0.87
  conditions:
  - type: "CostOptimized"
    status: "True"
    reason: "WithinBudget"
```

## 설치 및 배포

### 개발 환경

```bash
# 의존성 다운로드
make deps

# CRD 생성
make generate

# 빌드
make build

# 로컬 실행 (kubeconfig 필요)
make run
```

### Kubernetes 배포

```bash
# CRD 설치
make install

# Operator 배포
make deploy

# 확인
kubectl get pods -n kcloud-system
kubectl get crd | grep kcloud.io
```

### Helm 설치

```bash
# Helm 차트 설치
helm install kcloud-operator ./charts/kcloud-operator \
  --namespace kcloud-system \
  --create-namespace
```

## 설정

### 환경변수

- `WATCH_NAMESPACE`: 감시할 네임스페이스 (빈 값 = 모든 네임스페이스)
- `POD_NAME`: Operator 파드 이름
- `OPERATOR_NAME`: Operator 식별자
- `CORE_SCHEDULER_URL`: Core Scheduler 서비스 URL

### RBAC 권한

```yaml
# 필요한 권한 예시
rules:
- apiGroups: [""]
  resources: ["pods", "nodes", "services"]
  verbs: ["get", "list", "watch", "create", "update", "patch"]
- apiGroups: ["apps"]
  resources: ["deployments", "replicasets"]
  verbs: ["get", "list", "watch", "create", "update", "patch"]
- apiGroups: ["kcloud.io"]
  resources: ["workloadoptimizers"]
  verbs: ["get", "list", "watch", "create", "update", "patch"]
```

## Controller 로직

### WorkloadOptimizer Controller

```go
func (r *WorkloadOptimizerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // 1. WorkloadOptimizer 리소스 조회
    var wo kcloudv1alpha1.WorkloadOptimizer
    if err := r.Get(ctx, req.NamespacedName, &wo); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // 2. 현재 상태 분석
    currentState, err := r.analyzeCurrentState(ctx, &wo)
    if err != nil {
        return ctrl.Result{}, err
    }

    // 3. 최적화 실행
    optimized, err := r.optimizeWorkload(ctx, &wo, currentState)
    if err != nil {
        return ctrl.Result{}, err
    }

    // 4. 상태 업데이트
    if err := r.updateStatus(ctx, &wo, optimized); err != nil {
        return ctrl.Result{}, err
    }

    return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
}
```

## Admission Webhook

### Mutating Webhook

자동으로 WorkloadOptimizer 정책을 Pod에 적용:

```go
func (w *WorkloadMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
    pod := &corev1.Pod{}
    if err := w.Decoder.Decode(req, pod); err != nil {
        return admission.Errored(http.StatusBadRequest, err)
    }

    // 워크로드 타입 추론
    workloadType := w.inferWorkloadType(pod)
    
    // 최적화 정책 적용
    optimized := w.applyOptimizationPolicy(pod, workloadType)
    
    return admission.PatchResponseFromRaw(req.Object.Raw, optimized)
}
```

### Validating Webhook

리소스 생성/수정 시 정책 검증:

```go
func (w *WorkloadValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
    wo := &kcloudv1alpha1.WorkloadOptimizer{}
    if err := w.Decoder.Decode(req, wo); err != nil {
        return admission.Errored(http.StatusBadRequest, err)
    }

    // 비용 제약사항 검증
    if err := w.validateCostConstraints(wo); err != nil {
        return admission.Denied(err.Error())
    }

    // 전력 제약사항 검증
    if err := w.validatePowerConstraints(wo); err != nil {
        return admission.Denied(err.Error())
    }

    return admission.Allowed("")
}
```

## 사용 예시

### 1. ML 트레이닝 워크로드 최적화

```bash
# WorkloadOptimizer 생성
kubectl apply -f - <<EOF
apiVersion: kcloud.io/v1alpha1
kind: WorkloadOptimizer
metadata:
  name: bert-training
  namespace: ml-workloads
spec:
  workloadType: "training"
  resources:
    cpu: "32"
    memory: "128Gi"
    gpu: 8
  costConstraints:
    maxCostPerHour: 100.0
    preferSpot: true
  powerConstraints:
    maxPowerUsage: 4000.0
EOF

# 상태 확인
kubectl get wo bert-training -o yaml
kubectl describe wo bert-training
```

### 2. 추론 서빙 워크로드 최적화

```bash
kubectl apply -f - <<EOF
apiVersion: kcloud.io/v1alpha1  
kind: WorkloadOptimizer
metadata:
  name: llm-serving
spec:
  workloadType: "serving"
  resources:
    cpu: "8"
    memory: "32Gi"
    gpu: 2
  costConstraints:
    maxCostPerHour: 25.0
  autoScaling:
    minReplicas: 2
    maxReplicas: 20
    metrics:
    - type: "latency"
      threshold: 100  # ms
EOF
```

## 모니터링

### Prometheus 메트릭

- `kcloud_workload_optimizations_total`: 최적화 실행 횟수
- `kcloud_cost_savings_total`: 비용 절약 누적 금액
- `kcloud_power_savings_watts`: 전력 절약량
- `kcloud_optimization_score`: 최적화 점수

### 이벤트

```bash
# Kubernetes 이벤트 확인
kubectl get events --field-selector reason=WorkloadOptimized

# Operator 로그
kubectl logs -n kcloud-system deployment/kcloud-operator -f
```

## 개발

### 요구사항

- Go 1.21+
- Kubernetes 1.25+
- controller-runtime
- operator-sdk (선택사항)

### 코드 생성

```bash
# Controller/Client 코드 생성
make generate

# CRD 매니페스트 생성
make manifests

# 모든 코드 생성
make all
```

### 테스트

```bash
# 단위 테스트
make test

# 통합 테스트 (envtest)
make test-integration

# E2E 테스트
make test-e2e
```

## 라이선스

Apache License 2.0