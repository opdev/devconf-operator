# Level 2: Seamless upgrade of the operator and the  `Recipe` application, or the operand.

Level 2 is about making upgrade as easy as possible for the user. You should support seamless upgrades of both your operator and operand


# Update recipe_controller.go

Use our provided patch for adding the code to implement application version update support:

```shell
$ git add internal/controller/recipe_controller.go
$ patch --strip=1 < ${WORKSHOP_REPO}/workshop/level_2/patches/0001-application-version-update.patch
$ git diff internal/controller/recipe_controller.go
```
# Restart the controller
```shell
$ make run
```

### Modify the CR spec.version to v1.2.0

### Apply the CR
```shell
$ oc apply -f config/samples/devconfcz_v1alpha1_recipe.yaml
```
