# Level 1: Allow the deployment of the `Recipe` application

Level 1 is about automating the provisioning and configuration of our `Recipe` application.

## Scaffold Recipe API and controller

Use `operator-sdk create api` command to scaffold the Recipe API and its associated controller:

```shell
$ operator-sdk create api --group devconfcz --version v1alpha1 --kind Recipe --resource --controller
$ make manifests
```

# Edit the Recipe resource

Use our provided patch for customizing the Recipe resource definition:

```shell
$ patch --strip=1 < ${WORKSHOP_REPO}/workshop/level_1/patches/0001-recipe-type.patch
patching file api/v1alpha1/recipe_types.go
$ make manifests generate
```

# Add code to provision child resources

We're going to have to create the following child resources:
* For the frontend:
  * A Deployment
  * A Service
* For the MySQL database:
  * A Deployment
  * A Service
  * A Secret
  * A PVC
  * A ConfigMap

We create a `internal/resources/` subfolder to host the child resources definitions, and edit the controller loop to ensure those objects are created:

```shell
$ patch --strip=1 < ${WORKSHOP_REPO}/workshop/level_1/patches/0002-child-resources.patch
patching file Dockerfile
patching file internal/controller/recipe_controller.go
patching file internal/resources/configmap.go
patching file internal/resources/deployment.go
patching file internal/resources/mysqldeployment.go
patching file internal/resources/pvc.go
patching file internal/resources/secret.go
patching file internal/resources/service.go
$ make manifests
```

# Test Level 1

## Run the controller locally

Deploy the CRD in your cluster with `make install`, then run the controller locally:

```shell
make install
make run
```

## Customize the `Recipe` resource

In a new session:

```shell
$ patch --strip=1 < ${WORKSHOP_REPO}/workshop/level_1/patches/0003-example-recipe-manifest.patch 
```

## Instantiate a `Recipe` app

```shell
$ oc apply -f config/samples/devconfcz_v1alpha1_recipe.yaml
$ oc get po,deploy,pvc,secrets,configmap,service
```

# To go further...

* With the current implementation, resources are not being reconciled if they are modified externally (with the exception of the frontend's deployment replica count). We only check for the existence of a child resource.
* The main reconcile function is 238 lines long, and repeats many times the same logic for creating each child resource. One way to reduce code duplication and improve readability is to adopt a subreconciler pattern (See [this blog post](https://www.redhat.com/en/blog/subreconciler-patterns-in-action) for an example)

# [Onto Level 2...](../level_2/)