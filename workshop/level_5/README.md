# Level 5: Allow the deployment of the `Recipe` application create a HorizontalPodAutoscaler resource.

The highest capability level aims to significantly reduce/eliminate any remaining manual intervention in managing the operand. The operator should configure the Operand to auto-scale as load picks up.


# Enable level 5 capability

Use our provided patch for adding an HPA child resource:

```shell
$ patch --strip=1 < ${WORKSHOP_REPO}/workshop/level_5/patches/0001-hpa.patch
patching file api/v1alpha1/recipe_types.go
patching file internal/controller/recipe_controller.go
patching file internal/resources/hpa.go
$ make generate manifests
```

# Test level 5

## Ensure controller process uses latest changes

Stop the Controller (Ctrl+C), install the latest version of the CRD, and restart the controller:

```shell
make install run
```

## Edit the `Recipe` CR

```shell
$ patch --strip=1 < ${WORKSHOP_REPO}/workshop/level_5/patches/0002-enable-hpa.patch
patching file config/samples/devconfcz_v1alpha1_recipe.yaml
$ oc apply -f config/samples/devconfcz_v1alpha1_recipe.yaml
```

The Operator will provision a HorizontalPodAutoScaler like that:
```shell
kubectl get horizontalpodautoscalers.autoscaling
NAME                REFERENCE              TARGETS         MINPODS   MAXPODS   REPLICAS   AGE
recipe-sample-hpa   Recipe/recipe-sample   <unknown>/60%   1         3         0          7s
```

Let's simulate a load in our application:
```bash
apiVersion: apps/v1
kind: Deployment
metadata:
  name: loadgen-deployment
  namespace: default
  labels:
    app: loadgen
spec:
  replicas: 1
  selector:
    matchLabels:
      app: loadgen
  template:
    metadata:
      labels:
        app: loadgen
    spec:
      containers:
      - name: loadgen
        image: ghcr.io/yuriolisa/loadgen:latest
        imagePullPolicy: Always
        env:
          - name: URL
            value: "http://recipe-sample:8080/liveness"
```