module ics-csi-driver

replace (
	github.com/go-resty/resty => gopkg.in/resty.v1 v1.12.0
	k8s.io/api => k8s.io/kubernetes/staging/src/k8s.io/api v0.0.0-20200112230737-36e40fb85029
	k8s.io/apimachinery => k8s.io/kubernetes/staging/src/k8s.io/apimachinery v0.0.0-20200112230737-36e40fb85029
	k8s.io/apiserver => k8s.io/kubernetes/staging/src/k8s.io/apiserver v0.0.0-20200112230737-36e40fb85029
	k8s.io/client-go => k8s.io/kubernetes/staging/src/k8s.io/client-go v0.0.0-20200112230737-36e40fb85029
	k8s.io/component-base => k8s.io/kubernetes/staging/src/k8s.io/component-base v0.0.0-20200112230737-36e40fb85029
)

require (
	github.com/akutz/gofsutil v0.1.2
	github.com/container-storage-interface/spec v1.2.0
	github.com/go-resty/resty v1.12.0 // indirect
	github.com/inspur-ics/ics-go-sdk v0.0.0-20200119075651-075d90827b0a
	github.com/rexray/gocsi v1.1.0
	golang.org/x/lint v0.0.0-20190313153728-d0100b6bd8b3 // indirect
	golang.org/x/tools v0.0.0-20190524140312-2c0ae7006135 // indirect
	google.golang.org/grpc v1.26.0
	gopkg.in/gcfg.v1 v1.2.3
	gopkg.in/warnings.v0 v0.1.2 // indirect
	honnef.co/go/tools v0.0.0-20190523083050-ea95bdfd59fc // indirect
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.0.0
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/klog v1.0.0
	k8s.io/sample-controller v0.0.0-20180822125000-be98dc6210ab
)
