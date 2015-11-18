# Protype of cf push performance

To use the prototype, clone this repository (this may take several minutes as it includes some large test files):
```
git clone https://github.com/glyn/bloblets.git
```

In the following, we assume the repository has been cloned into the `bloblets` directory.

Then install the pre-requisites and run and invoke the server as described below.

## Pre-Requisites

Install Go 1.5.x or later.

Get the compile-time dependencies:
```
go get https://github.com/cloudfoundry/cli.git
```

## Running Locally

Run the server locally as follows:
```
cd bloblets/server
PORT=8080 go run server.go
```

Push an application to the server as follows:
```
cd bloblets/client
go install . && time client localhost:8080 <test-app>
```

where `<test-app>` is the path of the application to be pushed. For a suitable example, download [zipkin-web](http://search.maven.org/remotecontent?filepath=io/zipkin/zipkin-web/1.18.0/zipkin-web-1.18.0-all.jar).


## Run Server on PWS

Install the [cf CLI](https://github.com/cloudfoundry/cli) if you don't already have it.

Build the server:
```
cd bloblets/server
GOOS=linux GOARCH=amd64 go build -o <app-dir> .
```

where `<app-dir>` is a convenient empty directory.

Then push the server:
```
cf api api.run.pivotal.io
// log in and set org/space
cf push <app-name> -p <app-dir> -c "./server" -b https://github.com/cloudfoundry/binary-buildpack.git
```

where `<app-name>` is the name chosen for the application

The output will include the hostname of the deployed server:
```
urls: <app-name>.cfapps.io
```

Push an application to the server as follows:
```
cd bloblets/client
go install . && time client <app-name>.cfapps.io:80 <test-app>
```

where `<app-name>` is the name chosen for the application and `<test-app>` is the path of the application to be pushed. For a suitable example, download [zipkin-web](http://search.maven.org/remotecontent?filepath=io/zipkin/zipkin-web/1.18.0/zipkin-web-1.18.0-all.jar).
