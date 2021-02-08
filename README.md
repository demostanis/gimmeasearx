# gimmeasearx

Configurable, JavaScript-less Neocities alternative, written in Go!
It gives you a random searx (privacy-respecting metasearch engine) instance each time you visit the page.
You can either clone, build and use it locally using techniques below or use the Tor [hidden service](http://7tcuoi57curagdk7nsvmzedcxgwlrq2d6jach4ksa3vj72uxrzadmqqd.onion/). There's also [a wiki](https://github.com/demostanis/gimmeasearx/wiki)!

![screenshot](screenshots/2.png)

## Running with Go
You will need `git` and `go`. Once setup, run the following commands:
```sh
git clone https://github.com/demostanis/gimmeasearx.git
go run gimmeasearx.go 
```
That's it! Open up a browser and check [localhost:8080](http://localhost:8080).

If you want .onion instances to show up, you need [Tor](https://www.torproject.org/) installed and running.

## Running with Docker or Podman

```sh
docker build -t gimmeasearx .
docker run --name gimmesearx -d -ti -p 8080:8080 gimmeasearx
```

The docker instance should be up and running. You can access it via [localhost:8080](http://localhost:8080).

## Running as openrc-service

```
$ go build gimmeasearx.go
$ sed -i "s|TEMPLATE_DIR|$PWD|" services/openrc-service
# cp gimmeasearx /usr/local/bin
# cp services/openrc-service /etc/init.d/gimmeasearx
```
Edit the service file and `cd` to the directory where the template directory is located.
You can also change the port via the `PORT` environment variable.
The docker instance should be up and running. You can access it via [localhost:8080](http://localhost:8080).

Licensed under GPLv3.

If my time spent coding this was helpful to you,
I'd be gladful to receive donations:

- Ethereum: **0xF239e7C7b1C75EFF467EE4b74CEB4002E3d00BEE**

- Bitcoin: **5cc720fb7ca0bf0807e0223946fae738**

