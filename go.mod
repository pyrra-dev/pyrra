module github.com/pyrra-dev/pyrra

go 1.16

require (
	github.com/alecthomas/kong v0.4.0
	github.com/cespare/xxhash/v2 v2.1.2
	github.com/dgraph-io/ristretto v0.0.3
	github.com/fsnotify/fsnotify v1.5.1
	github.com/go-chi/chi/v5 v5.0.7
	github.com/go-chi/cors v1.2.0
	github.com/go-logr/logr v1.2.0
	github.com/gorilla/mux v1.8.0
	github.com/oklog/run v1.1.0
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.47.0
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/common v0.32.1
	github.com/prometheus/prometheus v1.8.2-0.20210421143221-52df5ef7a3be
	github.com/stretchr/testify v1.7.0
	golang.org/x/oauth2 v0.0.0-20210615190721-d04028783cf1
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.23.3
	k8s.io/client-go v0.21.0
	sigs.k8s.io/controller-runtime v0.9.0-beta.1.0.20210505224715-55a329c15d6b
	sigs.k8s.io/yaml v1.2.0
)
