# Level 2: Seamless upgrade of the operator and the  `Recipe` application, or the operand.

Level 2 is about making upgrade as easy as possible for the user. You should support seamless upgrades of both your operator and operand

# Edit the recipe_controller.go

Use our provided patch for adding the code to implement application version update support:

```shell
$ patch --strip=1 < ${WORKSHOP_REPO}/workshop/level_2/patches/0001-application-version-update.patch
patching file internal/controller/recipe_controller.go
```


## Edit the `Recipe` cr
Modify version to 1.2.0
```shell
$ oc apply -f config/samples/devconfcz_v1alpha1_recipe.yaml
```
