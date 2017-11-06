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

Talking to the Kubernetes cluster requires credentials that are provided to the pod as described
[Here](https://kubernetes.io/docs/concepts/configuration/secret/#service-accounts-automatically-create-and-attach-secrets-with-api-credentials).
  So no additional configuration is requried for authenticating to the kubernetes cluster.



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

For this demo, we need to disable security on Jenkins to allow HTTP POST calls to start builds.   This will
be used for Grafana running builds to rollback the canary build when not performing well.

`Jenkins -> Manage Jenkins -> Configure Global Security`  and uncheck the `Enable Security` box.


# Create new Git Repo for application

Using Github (or other git provider) create a new git project for the application.  This documentation has the 
git project located at github.com/runyontr/canary-app, and this value should be adjusted to your particular 
remote.

In your github project, navigate to  `Settings -> Integration & services` add a new service for `Jenkins (git plugin)` and enter the Jenkins root URL
which can be found inside of Jenkins `Manage Jenkins -> Configure -> Jenkins Location -> Jenkins URL`.

Like so:

![Github Integration](imgs/githubintegration.png)


This will notify your Jenkins deployment that there was a change in the git repo and Jenkins will then
pull down the repo and look for changes.


# Create Jenkins Build

On the home page click `Create new job` and select multibranch pipeline with build name `appinfo`


In the git section, add a branch source by `Git`.  Enter the git repo in the Repository URL.


  The default is to just build on the master branch, we want to add
 another branch for `canary`.  Select `Add Branch` and populate the Branch Specifer with `*/canary`.
 
 Click Save.
 
 
 
# Initial push on master


Take the contents of the `app` folder in this git repo and push it on the master branch of the repo that was created

 


## Monitoring

Install Prometheus and Grafana via [Helm](https://helm.sh/).

```
helm install stable/prometheus --name prom
helm install stable/grafana --name graf --set server.service.type=LoadBalancer --set server.adminPassword=admin \
 --set server.image=grafana/grafana:4.5.1
```

As of writing this, [Issue 9777](https://github.com/grafana/grafana/issues/9777) prevents the latest Grafana image
 from properly firing alerts, so we hard code a previous version.


 Log into Grafana at the following address
 
 ```
 kubectl get svc graf-grafana
 ```
 with the username/password `admin/admin`.



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

Import the dashboard in the `grafana` folder [here](grafana/dashboard.json)


## Create Rollback Build

In jenkins, we need to make a new build that will rollback the canary deployment. 
 We create a new Freestyle
project with name `canary-rollback`, which has a single build step:

![Rollback Build](imgs/rollbackbuild.png)



## Grafana Notifcation

Inside our grafana deployment at `/alerting/notifications`, we need to create a new notification Channel

![Grafana Notification](imgs/grafananotification.png)

The Dashboard that was imported should fire this notification when certain alerts are fired, causing the
Jenkins build to delete the canary deployment.



## View Dashboard

This dashboard has 3 graphs:

### AppInfo Response Time

This shows the 50th, 90th and 99th percentile response times for the stable and release deployments.

### Errors Per Second

This shows the number of errors per second being returned from the service.  In this basic demo this should never
be non-zero, but if the service starts spitting out errors in the canary deployment, the automatic rollback will
remove the canary deployment.

For example, if [line 77 in this file](app/appinfoservice.go) were uncommented. 


### Canary Response Times

This graph shows just the canary deployments response times.  Because of how Grafana works, we need a dedicated
graph that only shows the canary times to alert on.  This alert is configured to fire if the average response time
over the last 1 minute is greater than 2 seconds.  Additional alerts could be created for any SLOs on the system.




# CICD in Action

## Monitor

Setup a terminal to monitor response of the AppInfo service

```
kubectl get svc appinfo
```

will show the External-IP of the service.  Let this be `SERVICE_IP`

then

```
watch -n .5 --color -d "curl -s SERVICE_IP:8080/v1/appinfo | jq ."
```

will query the service 2x per second and show any changes in the response.

## Fix Response

looking at the response, the output does not properly fill in the Namespace value.  We will submit a fix
for this bug in the canary branch. In your git repo, checkout a new branch called `canary`

```
git checkout -b canary
```

and uncomment the following two lines in `appinfoservice.go`

```go
	//time.Sleep(3*time.Second)
	//info.Namespace = os.Getenv("MY_POD_NAMESPACE") //custom defined in the deployment spec
```

Don't forget to `import "time"`!  This will fix the Namespace value in the response, but will also introduce a slower response than is acceptable.


Commit and push this branch to `origin/canary`, and Github will trigger Jenkins to build and deploy a canary version of this application.
After being deployed, the `watch` loop should show some responses with the correct Namespace value, and a corresponding
`"Release": "canary"` tag.  


Even though this produces the correct response, the deployment will rollback as the Grafana dashboard monitoring
response times will fire the alert building the `rollback canary` build.  You should see the state of the 
responses in the watch terminal revert to its original pattern.


## Really Fix the Response

Go back to `appinfoservice.go` and comment back out the `time.Sleep(3*time.Second)` line (and remove the corresponding
time import).    Commit and push these changes.

This should again show fixed canary responses to the watch terminal.  The response times in Grafana should be below
the alert threshold and should continue to remain healthy.




## Merge into master

Issue a Pull Request on Github from `canary-> master`.  When pulling the PR, make sure not to squash and merge, as this
will create a new commit on Canary (hence firing the build) and will cause some race conditions between deploying 
the canary app on the canary branch build and removing the canary build in the master build.


Once the master build is complete, the response in the watch terminal should all be correct, and all have
`"Release": "stable"`.



