# Level 5: Allow the deployment of the `Recipe` application create a HorizontalPodAutoscaler resource.

The highest capability level aims to significantly reduce/eliminate any remaining manual intervention in managing the operand. The operator should configure the Operand to auto-scale as load picks up.


# Edit the recipe_controller.go

Use our provided patch for adding the code to implement application version update support:

```shell
$ patch --strip=1 < ${WORKSHOP_REPO}/workshop/level_5/patches/0001-application-version-update.patch
patching file internal/controller/recipe_controller.go
```

## Edit the `Recipe` CR
Include the Horizontal Pod AutoScaler and the Resources Limits specification.
```yaml
hpa:
  minReplicas: 1
  maxReplicas: 3
  targetMemoryUtilization: 60
resources:
  limits:
    cpu: 100m
    memory: 128Mi
  requests:
    cpu: 5m
    memory: 64Mi
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