

#   kubernetes定制版本

- [v1.14.10官方版本](https://github.com/cuisongliu/kubernetes/tree/v1.14.10)
- [v1.14.10的localtime的admission controller](https://github.com/cuisongliu/kubernetes/tree/v1.14.10-admission-localtime)



##  kubernetes 编译

```shell script
docker run -ti  -v /root/kube_build/kubernetes:/go/src/k8s.io/kubernetes -w /go/src/k8s.io/kubernetes \
    cuisongliu/kube-build:alpine bash

KUBE_GIT_TREE_STATE="clean"
KUBE_GIT_VERSION=v1.14.10
KUBE_BUILD_PLATFORMS=linux/amd64

rm -rf /root/kube_build/kubernetes/_output
make all WHAT=cmd/kube-apiserver GOFLAGS=-v
 
```

- KUBE_GIT_VERSION 是当前的k8s版本 这个很重要

