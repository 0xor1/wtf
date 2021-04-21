tlbx
===

## fmt & test

to gen, fmt and test everything.
```
./bin/pre
```

to clear out all docker containers and rebuild before gen, fmt and test.
(useful if you have a sql schema change and want to recreate your sql db containers).
```
./bin/pre nuke
```


## structure

* `/pkg` - reusable packages
* `/bin` - util scripts to run/build/test
* `/cmd` - executable go programs
* `/sql` - sql schemas
* `/docker` - docker-compose files for dev environment setup

## web apps

tlbx is predominantly for making web apps that follow a similar pattern, to run these apps, simply run
`./bin/run <app_name>` e.g. `./bin/run todo`. This script makes use of tmux to run all the aspects of an
app in one terminal screen so you can see everything going on in one place, if you prefer you can simply run the
commands in the `./bin/run` script manually. If you do install tmux, to kill the development services tmux
cmd `Ctrl+b &` then `y` to confirm will kill everything.