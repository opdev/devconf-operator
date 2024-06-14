# Level 3: Enable Backup and Restore for the Deployed MySQL.

Level 3 is about making backup of the operand data in any stateful data managed by the operand. You don’t need to backup the CR itself or the k8s resources created by the operator as the operator should return all resources to the same state if the CR is recreated.

# Prerequisites
In order to achieve a shared disk for our workshop, we have to install a NFS Server with a respective command:

```bash
helm repo add nfs-ganesha-server-and-external-provisioner https://kubernetes-sigs.github.io/nfs-ganesha-server-and-external-provisioner

helm install nfs-release -f values.yaml nfs-ganesha-server-and-external-provisioner/nfs-server-provisioner
```

# Edit the recipe_controller.go

Use our provided patch for adding the code to implement application version update support:

```shell
$ patch --strip=1 < ${WORKSHOP_REPO}/workshop/level_2/patches/0001-application-version-update.patch
patching file internal/controller/recipe_controller.go
```

## Edit the `Recipe` CR
Include the Backup Policy Specification
```yaml
backupPolicySpec:
  volumeName: "-backup"
  schedule: "*/1 * * * *"
  timezone: "Europe/Berlin"
```
You should see the following results after one minute:
```shell
kubectl logs mysql-job-28640006-sd6ht -f
  2024/06/14 21:26:12 Waiting for: tcp://recipe-sample-mysql:3306
  2024/06/14 21:26:17 Connected to tcp://recipe-sample-mysql:3306
  => Running cron task manager in foreground
  Listening on crond, and wait...
  crond: crond (busybox 1.36.1) started, log level 8
  crond: USER root pid  20 cmd /backup.sh >> /mysql_backup.log 2>&1
  => Backup started at 2024-06-14 21:28:00
  ==> Dumping database: recipes
  ==> Compressing recipes with LEVEL 6
  ==> Creating symlink to latest backup: 202406142128.recipes.sql.gz
  => Backup process finished at 2024-06-14 21:28:00
```

Let's simulate an outage in our namespace:
```bash
kubectl delete -f samples/config/devconfcz_v1alpha1_recipe.yaml
```

Wait until you see the PersistentVolume Released:
```bash
kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS     CLAIM                          STORAGECLASS   REASON   AGE
pvc-f22d6d22-72a0-4240-a908-c71a6ddcb6fd   1Gi        RWX            Retain           Released   default/recipe-sample-backup   nfs                     13m
```

Let's make the PersistVolume available again
```shell
kubectl patch pv pvc-f22d6d22-72a0-4240-a908-c71a6ddcb6fd -p '{"spec":{"claimRef": null}}'
```

The PersistVolume should present an Available status:
```shell
kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS      CLAIM   STORAGECLASS   REASON   AGE
pvc-f22d6d22-72a0-4240-a908-c71a6ddcb6fd   1Gi        RWX            Retain           Available   /       nfs
```

Now, as we have clared the PersistentVolume Status we should be able to put back our application and see if the Restore works:
```bash
kubectl apply -f samples/config/devconfcz_v1alpha1_recipe.yaml
```

The expected Restore results are should be like the following:

```bash
kubectl logs mysql-restore-job-bwbjq -f
  2024/06/14 21:50:44 Connected to tcp://  recipe-sample-mysql:3306
  => Restore latest backup
  => Searching database name in /backup/202406142150.  recipes.sql.gz
  => Restore database recipes from /backup/202406142150.  recipes.sql.gz
  => Restore succeeded
  => Running cron task manager in foreground
```