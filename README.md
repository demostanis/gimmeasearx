# gimmeasearx

Configurable, JavaScript-less Neocities alternative, written in Go!
It gives you a random searx (privacy-respecting metasearch engine) instance each time you visit the page.
You can either clone, build and use it locally using techniques below or use the Tor [hidden service](http://7tcuoi57curagdk7nsvmzedcxgwlrq2d6jach4ksa3vj72uxrzadmqqd.onion/). There's also [a wiki](https://github.com/demostanis/gimmeasearx/wiki)!

![screenshot](screenshots/2.png)

## Running with Go
You will need `git` and `go`. Once setup, run the following commands:
```sh
git clone https://github.com/demostanis/gimmeasearx.git
go run cmd/main.go
```
That's it! Open up a browser and check [localhost:8080](http://localhost:8080).
If you want .onion instances to show up, you need Tor installed and running.

## Running with Docker or Podman

```sh
docker build -t gimmeasearx .
docker run --name gimmesearx -d -ti -p 8080:8080 gimmeasearx
```

The docker instance should be up and running. You can access it via [localhost:8080](http://localhost:8080).

## Running as openrc-service

```
$ go build -o gimmeasearx ./cmd/main.go
# cp gimmeasearx /usr/local/bin
# cp openrc-service /etc/init.d/gimmeasearx
```
edit the service file and `cd` to the directory where the template directory is located.
You can also change the port via the `PORT` environment variable.
The docker instance should be up and running. You can access it via [localhost:8080](http://localhost:8080).

Licensed under GPLv3.
