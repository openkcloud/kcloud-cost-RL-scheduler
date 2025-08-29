module github.com/kcloud-opt/operator

go 1.21

require (
	k8s.io/api v0.28.3
	k8s.io/apimachinery v0.28.3
	k8s.io/client-go v0.28.3
	sigs.k8s.io/controller-runtime v0.16.3
	github.com/operator-framework/operator-sdk v1.32.0
	github.com/prometheus/client_golang v1.17.0
	go.uber.org/zap v1.26.0
)