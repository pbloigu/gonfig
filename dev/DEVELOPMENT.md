# Starting the dev server
Just run `go run ./server`. The server will be listening on 8080 (ui), 8081 (backend API) and 9000 (C&C channel)

Dev server utilizes [test containers for Go](https://golang.testcontainers.org/) which obviously requires docker being installed on the host and the current user having access to docker functionalities. 

Generally docker will expose all container ports to all host interaces, but if in your setup this is not the case (you've configured docker to bind container ports to specific interface only), set the environment variable `TESTCONTAINERS_HOST_OVERRIDE` before starting the dev server, e.g.
`export TESTCONTAINERS_HOST_OVERRIDE=10.0.0.2 && go run ./server`