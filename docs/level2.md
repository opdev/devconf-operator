#### Update Operand image version
copy the following block of code and paste to internal/controller/recipe_controller.go in line 262

```
// Level 2: Update Operand (Recipe App)
log.Info("Reconciling Recipe App version")
found = &appsv1.Deployment{}
err = r.Get(ctx, client.ObjectKey{Name: recipe.Name, Namespace: recipe.Namespace}, found)

if err != nil {
    log.Error(err, "Failed to get Recipe App Deployment")
    return ctrl.Result{}, err
}
desiredImage := fmt.Sprintf("quay.io/rocrisp/recipe:%s", recipe.Spec.Version)
currentImage := found.Spec.Template.Spec.Containers[0].Image

if currentImage != desiredImage {
    found.Spec.Template.Spec.Containers[0].Image = desiredImage
    err = r.Update(ctx, found)
    if err != nil {
        log.Error(err, "Failed to update Recipe App version")
        return ctrl.Result{}, err
    }
}
```