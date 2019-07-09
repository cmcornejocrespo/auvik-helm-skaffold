# What's this for?

This is a skaffold/helm/k8s/gitlab-ci playground.

The goal is to being able to automate the build and deploy process of any given app to K8s. Ideally using skaffold, helm and potentially using gitlab CI to create a CI/CD pipeline.

## The playground

The package contains a simple Go service (app folder) that by default listens on port 80 which replies to any request with the content of the environment variable `MESSAGE`. To compile the service run `go build`.

**Tasks**

- Containerize the service using Docker. 
- Create a Kubernetes deployment and a Kubernetes service for the containerized service. It should be exposed to Internet.
- Develop a simple CI/CD pipeline that builds and deploys the service taking into account the following requirements:
    - There are two clusters: `stage` and `production`. Each repository is linked to a branch in the Git repository so a commit in `stage` should deploy to the stage cluster and a tag in `production` should deploy to the production cluster (you can mimic it with different k8s namespaces).
    - For `production` the message that the service returns has to be "I'm runnning in production!" and the server needs to listen on port 80, in `stage` the message has to be "I'm running in stage!" and needs to listen on port 8000.
    - Ideally, the pipeline should be done using Skaffold and Helm, if other technologies are used we would appreciate some documentation. Use any CI/CD platform that you find convenient, we use GitLab CI so bonus points if you use it as well (tip: you can run a Gitlab runner locally!.

---

## Requirements
- [skaffold](https://skaffold.dev/)
- [helm](https://helm.sh/)
- k8s cluster
- ingress enabled/installed
- [Gitlab runner](https://docs.gitlab.com/runner/install/)
---

## Run

We´ll use minikube for testing purposes.

```sh
minikube start
```

I've used [draft.sh](https://draft.sh/) to create the helm and Dockerfile boilerplate.

```sh
.
├── Dockerfile
├── README.md
├── app
│   ├── go.mod
│   ├── go.sum
│   └── main.go
├── skaffold-helm-app
│   ├── Chart.yaml
│   ├── templates
│   │   ├── NOTES.txt
│   │   ├── _helpers.tpl
│   │   ├── deployment.yaml
│   │   ├── ingress.yaml
│   │   └── service.yaml
│   └── values.yaml
└── skaffold.yaml
```

I made some cosmetic changes to the boilerplate to simplify the solution.

We want to have the service publicly available so I made the **assumption** that we needed ingress enabled.

In our local cluster we enable ingress, but we could have installed [this](https://github.com/nginxinc/kubernetes-ingress/blob/master/docs/installation.md) instead.

```sh
minikube addons enable ingress
```

To reuse the Docker daemon from Minikube and have the images available in k8s we need to:

```sh
eval $(minikube docker-env)
```

Rollout tiller installation:
```sh
helm init
```

Add hosts entries to fake ingress hosts

```sh
echo "$(minikube ip) skaffold-helm-app.stage.com" | sudo tee -a /etc/hosts
echo "$(minikube ip) skaffold-helm-app.production.com" | sudo tee -a /etc/hosts  
```

I've created a pipeline that should run in Gitlab. I couldn´t find the way to run the whole pipeline locally (not sure if it´s possible as [doc](https://docs.gitlab.com/runner/commands/#limitations-of-gitlab-runner-exec) says that you can only run jobs),  but conceptually we want to build, deploy to stage and deploy to production as follows:

```yaml
# .gitlab-ci.yml
stages:
  - build
  - stage
  - production
  
build:
  stage: build
  script:
    - skaffold build
stage:
  stage: stage
  script:
    - skaffold deploy -p stage
production:
  stage: production
  script:
    - skaffold deploy -p production
```

Skaffold config as follows:
```yaml
#skaffold.yaml
apiVersion: skaffold/v1beta12
kind: Config  
build:
  tagPolicy:
    envTemplate:
      template: "{{ .IMAGE_NAME }}:latest"
  artifacts:
  - image: auvik/go-app
    docker:
      dockerfile: Dockerfile
profiles:  
- name: stage
  deploy:
    helm:
      releases:
      - name: skaffold-helm-app-stage
        chartPath: skaffold-helm-app
        namespace: stage
        values:
          service.internalPort: 8000
          env.message: I'm running in stage!
          image.repository: auvik/go-app
          image.tag: latest
          ingress.domain.name: stage.com
- name: production
  deploy:
    helm:
      releases:
      - name: skaffold-helm-app-pro
        chartPath: skaffold-helm-app
        namespace: production
        values:
          service.internalPort: 80
          env.message: I'm running in production!
          image.repository: auvik/go-app
          image.tag: latest
          ingress.domain.name: production.com
```

And the helm chart can be found under the skaffold-helm-app folder.