
Level 4 : **Deep Insights for the Recipe application**

It is about monitoring and alerting the recipe application.

## 
{namespace="devconf-operator-system"}

## Publish additional metrics with custom collectors by using the global registry from controller-runtime/pkg/metrics.

## One way to achieve this is to declare your collectors as global variables and then register them using init() in the controller’s package.

## First, you will need enable the Metrics by uncommenting the following line in the file config/default/kustomization.yaml

# [PROMETHEUS] To enable prometheus monitor, uncomment all sections with 'PROMETHEUS'.
- ../prometheus

## Under the “controller” directory, we implement a new function as the following
## We then record to these collectors in the controller under the **reconcile** function

```shell
$ patch --strip=1 < ${WORKSHOP_REPO}/workshop/level_4/patches/0001-add_metrics.patch
```


## Now we can build the controller
```shell
make docker-build docker-push IMG=quay.io/rocrisp/recipe_operator:3.0.0
```
## Install the controller inside the cluster
```shell
make deploy IMG=quay.io/rocrisp/recipe_operator:3.0.0
```
## Allow the service monitor of the Operator to be scraped by the Prometheus instance of the OpenShift Container Platform cluster.
```shell
oc apply -f config/prometheus/role.yaml
oc apply -f config/prometheus/rolebinding.yaml
```
## Labels the devconf-operator-system namespace to scrape for metrics, which enables OpenShift cluster monitoring for that namespace

```shell
oc label namespace devconf-operator-system openshift.io/cluster-monitoring="true"
oc get namespace devconf-operator-system --show-labels
```

For reference, https://docs.openshift.com/container-platform/4.15/operators/operator_sdk/osdk-monitoring-prometheus.html