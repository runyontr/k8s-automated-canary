# Automated Canary Deployments


This tutorial looks to expand upon the tutorial in [runyontr/k8s-canary](http://github.com/runyontr/k8s-canary)
 and show how to automate the steps inside of a CICD pipeline. 
 
 
 
 
# Setup



# Deploy Infrastructure


```
helm install -f k8s/jenkins/values.yaml --name pretty-bird stable/jenkins
```


Get the login information for the admin password

```
printf $(kubectl get secret --namespace default pretty-bird-jenkins -o jsonpath="{.data.jenkins-admin-password}" | base64 --decode);echo
```

and the login address:
```
export SERVICE_IP=$(kubectl get svc --namespace default pretty-bird-jenkins --template "{{ range (index .status.loadBalancer.ingress 0) }}{{ . }}{{ end }}")
  echo http://$SERVICE_IP:8080/login
```


# Configure Jenkins

Navigate to  `Manage Jenkins -> Configure Systems` and ensure the following settings are present:

![Jenkins Slave](imgs/jenkins-slave.png)

![Docker In Docker](imgs/dind.png)

![Mounted Volumes](imgs/Volumes.png)


## Custom Jenkins-Slave image

This custom image is built from [this Dockerfile](jenkins/slave-image/Dockerfile) and has Golang 1.8 and
kubectl 1.8.0 pre-installed so it doesn't have to be installed fresh each run.

## Explain docker in docker sidecar
 
 Thanks to @prydonius for this approach.  By adding the `docker:1.12-dind` container as a sidecar to the JNLP
  container, and mounting the docker executable from the kubernetes node, we are able to run docker commands
  inside of JNLP (namely `docker build` and `docker push`).
 

 
 
## Dockerhub Credentials

In order to push the image we build, we'll need to have our credentials available to the docker runtime
in the slave.  Create a credentials with your dockerhub login with ID 'Dockerhub' like this:

![Dockerhub](imgs/DockerhubCredentials.png)

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


## Github Webhook

Disable security on Jenkins to allow simple POSTs to start builds.  We'll use this for both a github webhook
and grafana alerts


# Create new Git Repo for application






# Create Jenkins Build

On the home page click `Create new job` and select multibranch pipeline with build name `appinfo`


In the git section, add a branch source by `Single repo and Branch`.  The Repository URL should be the
 git repo setup in the previous section.  The default is to just build on the master branch, we want to add
 another branch for `canary`.  Select `Add Branch` and populate the Branch Specifer with `*/canary`.
 
 Click Save.
 
 
 
# Initial push on master


Take the contents of the `app` folder in this git repo and push it on the master branch of the repo that was created

 



## Build Dependencies

Instead of installing Go and Kubectl in the pod performing each build, we've created an extension of the default
`jenkinsci/jnlp-slave` image that has Go (1.8.5) and Kubectl (1.8.0) installed.  This speeds up the time of the build,
but doesn't allow for the flexibility of the Go plugin, where the go version is controlled by the Jenkinsfile.

Kubectl is installed in the slave image at /usr/local/kubectl, and the credentials for talking to the cluster are
mounted automatically by Kubernetes, so no additional configuration is required.  For more information see [LINK](LINKME)

//TODO find actual link



## Monitoring

//TODO all params to prometheus to only deploy prometheus. Don't need
node exporter, push gateway, alert.

```
helm install stable/prometheus --name prom
helm install stable/grafana --name graf --set server.service.type=LoadBalancer --set server.adminPassword=admin \
 --set server.image=grafana/grafana:4.5.1
```

As of writing this, [Issue 9777](https://github.com/grafana/grafana/issues/9777) prevents the latest from properly
firing alerts

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

```
Name: Prometheus
Type: Promethus

URL: http://prom-prometheus-server
Access: proxy
```



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



# Before demo:

 Reset git repo
```
git checkout 7ec948e
git checkout -b temp
git branch -D master
git checkout -b master
git push origin master -f
git branch -D temp
```

Delete appinfo
```
kubectl delete deployment appinfo
kubectl delete deployment appinfo-canary
```


Watch terminals

`watch --color -d kubectl get pods -l app=appinfo`
`watch --color -d "http://104.154.230.2:8080/v1/appinfo | jq ."
`


# DEMO COMMANDS

```
cp -R ../k8s-automated-canary/app/* .
git add .
git commit -m "Initial app"
git push
```



```
git checkout -b canary
```

DO FIX

```
git add .
git commit -m "Added namespace value to response object"
git push origin canary
```

Show Grafana

Show slow canary


Show Unhealthy

Show build in Jenkins that was triggered

Show no more canary pod



FIX

git add .
git commit -m "Removed sleep"
git push origin canary


Show build

Show grafana/while has good values and response times look good




ISSUE PR

Merge PR (no sqaush, since that causes new build of canary branch)

Show
