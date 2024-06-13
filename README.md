# devconf-operator: A demo level 5 operator built for DevConf.CZ

This project provides step-by-step instructions on how to build a Level 5 Operator. It was created for the DevConf.CZ 2024.

See the Workshop's abstract: https://pretalx.com/devconf-cz-2024/talk/YE8FEJ/

# Content

Instructions are available under the [`workshop/`](./workshop/) directoy.

Starting from an existing containerized application we'll go step-by-step through each level:
* [Level 0](./workshop/level_0/): Initialize a new operator Project
* [Level 1](./workshop/level_1/): Allow the deployment of the `Recipe` application
* [Level 2](./workshop/level_2/):
* [Level 3](./workshop/level_3/):
* [Level 4](./workshop/level_4/):
* [Level 5](./workshop/level_5/):

Happy hacking !

# The `Recipe` application

The `Recipe` application provides a way to manage cooking recipes. You can Create, Read, Update and Delete (CRUD) recipes conveniently from its web interface. It is composed of the following components, each running as their own container:
* The nginx web server
* A MySQL database for storing data

# Pre-requisites

Attendees are expected to be familiar with the operator pattern.

If you wish to follow this tutorial on your own, you'll need to:
- [Install Go](https://go.dev/doc/install)
- [Install operator-sdk](https://sdk.operatorframework.io/docs/installation/)
- Have access to an OpenShift cluster ([OpenShit Local](https://developers.redhat.com/products/openshift-local/overview) formerly CRC might be a good starting point) or Kubernetes cluster, (refer to [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/) or [Minikube](https://minikube.sigs.k8s.io/docs/start/) for a local Kubernetes cluster installations).

If you attend the Workshop in presence at DevConf.CZ, you'll be provided with access to a lab that has all pre-requisites provisioned and access to a Single Node Openshift lab.

# Resources

- [Operator Capability Level](https://sdk.operatorframework.io/docs/overview/operator-capabilities/)
- [Recipe application source code](https://github.com/rocrisp/recipe)