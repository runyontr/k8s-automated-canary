# Automated Canary Deployments


This tutorial looks to expand upon the tutorial in [runyontr/k8s-canary](http://github.com/runyontr/k8s-canary)
 and show how to automate the steps inside of a CICD pipeline. 
 
 
 
 
# Setup

## Install Minikube

This guide uses Minikube to deploy Kubernetes.  Follow [these instructions](https://kubernetes.io/docs/getting-started-guides/minikube/)
 for installation guides.
 
## Start Minkiube

We do not need any special configuration for our minikube deployment.  To start the cluster run

```
minikube start
```


# Deploy Infrastructure


```
helm install -f k8s/jenkins/values.yaml stable/jenkins
```
//TODO(@trunyon) the docuemntation allows for configuring the default plugins:
Master.InstallPlugins

Then what do we need to do to configure each plugin?


Get the login information for the admin password

```
printf $(kubectl get secret --namespace default khaki-lynx-jenkins -o jsonpath="{.data.jenkins-admin-password}" | base64 --decode);echo
```



Get the port

```
NODE_PORT=$(kubectl get svc khaki-lynx-jenkins \
  --output=jsonpath='{range .spec.ports[0]}{.nodePort}')
```

Login to Jenkins at 

```
echo 192.168.99.100:${NODE_PORT}
```

## Jenkins Docker

Install the docker plugin to allow for pushing to docker.io

 To use docker in Kubernetes, we'll use a sidecar Docker-in-Docker container.  This container will
 provide access to the Docker daemon from the jnlp slave image via the port 2375.  The environment
 variable `DOCKER_HOST` is set to let the docker executable mounted in the container to know where
 to post requests to.
 
 Go to the Jenkins Configuration's Cloud section and 
 
 1) Add Container Environment Variable DOCKER_HOST=tcp://localhost:2375 to tell the container how to to talk to the docker
 daemon
 2) The Docker-In-Docker sidecar container:
   a) name = docker-in-docker
   b) Docker Image = docker:1.12-dind
   c) Command to run slave agent: /usr/local/bin/dockerd-entrypoint.sh
   d) Args to pass the command:  --storage-driver=overlay
   e) run privileged
 3) Add Volumes
   a) Empty Dir Volume
      i) Mount Path:  /var/lib/docker to get access to the underlying node's docker storage
   b) Host Path volume
      i) /usr/bin/docker -> /usr/bin/docker to get access to the underlying nodes docker client executable


The last issue to tackle is authentication with Dockerhub for pushing the app image.  We've configured a credentials


Dockerhub Credentials:

Add a username and password in the credential section with the dockerhub login with ID 'Dockerhub'.  Then we can use it
in the app's Jenkinsfile

```groovy
 docker.withRegistry('https://registry.hub.docker.com', 'Dockerhub') {
                app = docker.build("${image}:${tag}")
                app.push()
            }
```

For running this demo, you will need to create your own creditials in Jenkins and modify the image name in Jenkinsfile 
in the application folder


## Build Dependencies

Instead of installing Go and Kubectl in the pod performing each build, we've created an extension of the default
`jenkinsci/jnlp-slave` image that has Go (1.8.5) and Kubectl (1.8.0) installed.  This speeds up the time of the build,
but doesn't allow for the flexibility of the Go plugin, where the go version is controlled by the Jenkinsfile.

Kubectl is installed in the slave image at /usr/local/kubectl, and the credentials for talking to the cluster are
mounted automatically by Kubernetes, so no additional configuration is required.  For more information see [LINK](LINKME)

//TODO find actual link



## Monitoring

```
helm install stable/prometheus --set server.service.type=NodePort --name prom
helm install stable/grafana  --set server.service.type=NodePort --set server.adminPassword=admin --name graf
```

kubernetes_sd_configs:
- api_servers:
- 'https://kubernetes'

get the grafana admin password:


Log into the service at the following port (on 192.168.99.100) with the password admin

```
NODE_PORT=$(kubectl get svc graf-grafana  \
  --output=jsonpath='{range .spec.ports[0]}{.nodePort}')
```


After logging in, we need to add Prometheus as a data source:

Here are the config values that should be used:

Name: Prometheus
Type: Promethus

URL: http://prom-prometheus-server
Access: proxy


Click save and test and we should be good to go.


## Import Dashboard



#The Setup

To start, we will need a new git repository.  For ease, we chose to make a new project in github.com.

We need to replace the import packages in the application to prevent importing the packages from my project.

```
sed -i

``` 


## Create Jenkins Job
Multibranch project

Screenshots of configuration


checkout app locally


Checkout branch `canary`

`cp -R ./app/* ../appinfo/`

Create commit and push up

(this would now be automated, but since I'm runnign jenkins locally, I'll)
Talk about manual running since Jenkins is all local right now, I'll manually run the check for new commits




