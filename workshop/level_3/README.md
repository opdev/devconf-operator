# Level 3: Enable Backup and Restore for the Deployed MySQL.

Level 3 is about making backup of the operand data in any stateful data managed by the operand. You don’t need to backup the CR itself or the k8s resources created by the operator as the operator should return all resources to the same state if the CR is recreated.

# Prerequisites
In order to achieve a shared disk for our workshop, we have to install a NFS Server with a respective command:

```shell
$ curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

$ helm repo add nfs-ganesha-server-and-external-provisioner https://kubernetes-sigs.github.io/nfs-ganesha-server-and-external-provisioner

$ helm install nfs-release -f  ${WORKSHOP_REPO}/workshop/level_3/values.yaml nfs-ganesha-server-and-external-provisioner/nfs-server-provisioner
```

# Enable backup and restore feature

Use our provided patch for adding the code to implement the backup and restore functionality:

```shell
$ patch --strip=1 < ${WORKSHOP_REPO}/workshop/level_3/patches/0001-backup-restore.patch
patching file api/v1alpha1/recipe_types.go
patching file internal/controller/recipe_controller.go
patching file internal/resources/cronjob.go
patching file internal/resources/job.go
patching file internal/resources/pvc.go
$ make manifests
```

# Test level 3

## Ensure controller process uses latest changes

Stop the Controller (Ctrl+C), install the latest version of the CRD, and restart the controller:

```shell
make install run
```

## Edit the `Recipe` CR

Include the Backup Policy Specification:

```shell
$ patch --strip=1 < ${WORKSHOP_REPO}/workshop/level_3/patches/0002-enable-cron-backup.patch
patching file config/samples/devconfcz_v1alpha1_recipe.yaml
$ oc apply -f config/samples/devconfcz_v1alpha1_recipe.yaml
```

You should see the following results after one minute:
```shell
$ oc logs mysql-job-28640006-sd6ht -f
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

## Connect to the Recipe UI and add some data

## Vizualize the backed up data

```shell
$ oc debug mysql-job-28640006-sd6ht
Starting pod/mysql-job-28640250-25q4h-debug-v4gbn ...
Pod IP: 10.128.1.86
If you don't see a command prompt, try pressing enter.
/ # ls -al /backup
total 4
drwxrwsrwx    2 root     root            70 Jun 15 01:31 .
dr-xr-xr-x    1 root     root            40 Jun 15 01:31 ..
-rw-r--r--    1 root     root           866 Jun 15 01:31 202406150131.recipes.sql.gz
lrwxrwxrwx    1 root     root            27 Jun 15 01:31 latest.recipes.sql.gz -> 202406150131.recipes.sql.gz
```

## Simulate outage and test restore

Let's simulate an outage in our namespace:
```shell
$ oc delete -f config/samples/devconfcz_v1alpha1_recipe.yaml
```

Wait until you see the PersistentVolume Released:
```shell
$ oc get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS     CLAIM                          STORAGECLASS   REASON   AGE
pvc-f22d6d22-72a0-4240-a908-c71a6ddcb6fd   1Gi        RWX            Retain           Released   default/recipe-sample-backup   nfs                     13m
```

Let's make the PersistentVolume available again
```shell
$ oc patch pv pvc-f22d6d22-72a0-4240-a908-c71a6ddcb6fd -p '{"spec":{"claimRef": null}}'
```

The PersistVolPersistentVolumeume should present an Available status:
```shell
$ oc get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS      CLAIM   STORAGECLASS   REASON   AGE
pvc-f22d6d22-72a0-4240-a908-c71a6ddcb6fd   1Gi        RWX            Retain           Available   /       nfs
```

Now, as we have cleared the PersistentVolume Status, we should be able to put back our application and see if the Restore works:
```shell
$ patch --strip=1 < ${WORKSHOP_REPO}/workshop/level_3/patches/0003-enable-restore.patch
patching file config/samples/devconfcz_v1alpha1_recipe.yaml
$ oc apply -f config/samples/devconfcz_v1alpha1_recipe.yaml
```

The Restore results should be like the following:

```shell
$ oc logs mysql-restore-job-bwbjq -f
  2024/06/14 21:50:44 Connected to tcp://  recipe-sample-mysql:3306
  => Restore latest backup
  => Searching database name in /backup/202406142150.  recipes.sql.gz
  => Restore database recipes from /backup/202406142150.  recipes.sql.gz
  => Restore succeeded
  => Running cron task manager in foreground
```

# [Onto Level 4...](../level_4/)