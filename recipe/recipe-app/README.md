### Steps to Deploy the Recipe App

1. **Clone the Repository**:
   
   Clone the repository to your local machine.

2. **Navigate to the Project Directory**:
   ```bash
   cd recipe-app
3. **Build the Container Image**:
   ```bash
   podman build --platform linux/amd64 -t recipe_app:v1 .
4. **Run the Application Locally**:
   ```bash
   podman run -d --name recipe_app -p 5001:5000 localhost/recipe_app:v1
5. **Tag and Push the Image to Image Repository**:
   ```bash
   podman tag <imageid> quay.io/r<id>/recipe:v1
   podman push quay.io/<id>/recipe:v1
