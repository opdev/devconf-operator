# Level 0 - Initialize a new operator project

## Check tools version and OCP access

```shell
$ go version
go version go1.22.3 linux/amd64
$ operator-sdk version
operator-sdk version: "v1.26.0", commit: "cbeec475e4612e19f1047ff7014342afe93f60d2", kubernetes version: "1.25.0", go version: "go1.19.3", GOOS: "linux", GOARCH: "amd64"
$ oc version
Client Version: 4.15.17
Kustomize Version: v5.0.4-0.20230601165947-6ce0bf390ce3
$ oc status
In project default on server https://api.exe-workshop-workshopdevconf1-pool-zrn7g.coreostrain.me:6443

svc/openshift - kubernetes.default.svc.cluster.local
svc/kubernetes - 172.30.0.1:443 -> 6443

View details with 'oc describe <resource>/<name>' or list resources with 'oc get all'.
```

## Clone this repository

Clone the workshop's repository to get easy access to instructions, snippets and patches:

```shell
$ echo "export WORKSHOP_REPO=~/workshop_repo" >> ~/.bashrc
$ . ~/bashrc
$ git clone git@github.com:opdev/devconf-operator.git ${WORKSHOP_REPO}
```

## Prepare local folder

Create a new directory for your project and initialize a local git repository:

```shell
$ mkdir ~/devconf-operator && cd ~/devconf-operator
$ git init
Initialized empty Git repository in /home/ec2-user/devconf-operator/.git/
```

## Initialize a new Go Operator project

Scaffold a new go operator project

```shell
$ operator-sdk init --domain opdev.com --repo github.com/opdev/devconf-operator
```

## Patch scaffolded project

```shell
$ patch < ${WORKSHOP_REPO}/workshop/level_0/patches/0001-controller-gen-version.patch
```

## [Onto Level 1...](../level_1/)