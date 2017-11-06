node {
  //This specifies what image tag for the application.  Adjust as needed
  def image = 'runyonsolutions/appinfo'
  def tag = "${env.BRANCH_NAME}-${env.BUILD_NUMBER}"

  //This specifies where to put the repo in the GOPATH to build
  def srcdir = 'github.com/runyontr/canary-app'

// Install the desired Go version

  checkout scm

  def workspace = pwd()
  stage('Run Go tests') {
    sh """
         mkdir -p /go/src/${srcdir}
         ln -sf ${workspace}/* /go/src/${srcdir}/
         cd  /go/src/${srcdir}
         go test ./...
          """
    }


     stage('Build and Push Image') {
           sh """
             cd  /go/src/${srcdir}
             CGO_ENABLED=0 GOOS=linux go build -o app *.go
             cp app ${workspace}/
           """
            docker.withRegistry('https://registry.hub.docker.com', 'Dockerhub') {
                app = docker.build("${image}:${tag}")
                app.push()
            }
     }


     stage("Deploy Application"){
     withEnv(['PATH+JENKINSHOME=/home/jenkins/bin']) {
        git_commit = sh(returnStdout: true, script: 'git rev-parse HEAD').trim()
        //Update the image in the deployment spec
        sh("sed -i 's/IMAGE_TAG/${tag}/g' ./k8s/deployment.yaml")
        sh("sed -i 's/GITCOMMIT/${git_commit}/g' ./k8s/deployment.yaml")

        switch (env.BRANCH_NAME) {
            // Roll out to canary environment
            case "canary":
                // Change deployed image in canary to the one we just built
                sh("echo Hi this is the canary branch")
                //Change the name of things to be appinfo-canary

                sh("sed -i 's/name: appinfo/name: appinfo-canary/g' ./k8s/deployment.yaml")

                //Canary only needs 1 replica
                sh("sed -i 's/replicas: 3/replicas: 1/g' ./k8s/deployment.yaml")

                //change the release value to be canary
                sh("sed -i 's#release: stable#release: canary#' ./k8s/deployment.yaml")

                sh("cat ./k8s/deployment.yaml")

                sh("kubectl apply -f k8s/deployment.yaml")

                break

            // Roll out to production
            case "master":
                // Change deployed image in canary to the one we just built
                sh("echo hi master")
                sh("cat ./k8s/deployment.yaml")
                sh("kubectl apply -f k8s/deployment.yaml")
                sh("kubectl apply -f k8s/service.yaml")

                //cleanup the canary build
                sh("kubectl delete deployment appinfo-canary")

                break

            // All other branches shouldn't be deployed
            default:

                sh("echo Not deploying application")
          } //switch
       }//env
     } //stage
} //nodes