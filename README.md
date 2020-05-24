# Kubernetes

- 根据首页方法编译kube-apiserver组件
- 编译之后替换master的apiserver


## 新增admission-controller

- 名称: Localtime
- 已设为默认
- pod的注解: kubernetes.io/localtime="true"
- 挂载点名称: kubernetes-localtime
- 说明
  - 如果pod的注解是kubernetes.io/localtime="true"这个值则会在当前pod中自动挂载/etc/localtime目录。


