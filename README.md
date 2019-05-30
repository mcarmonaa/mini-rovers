# mini-rovers

***mini-rovers*** is as proof of concept based on [src-d/rovers](https://github.com/src-d/rovers)
that takes in a list of github organizations and retrieves all their repositories
github's metatada, sending this info (endpoints, and if it's a fork) to rabbitmq.

## Usage

```sh
Usage:
  main [OPTIONS] [list]

Application Options:
      --queue=  queue name (default: mini-rovers) [$QUEUE_NAME]
      --broker= broker service URI (default: amqp://localhost:5672) [$BROKER_URI]
  -t, --token=  github authentication token  [$GH_TOKEN]

Help Options:
  -h, --help    Show this help message

Arguments:
  list:         path to a file containing a list of githug organizations(one per line)
```

To test it locally you need a rabbitmq instance running. There is `docker-compose.yml`
to run it in a container, also there is a testfile with a list of organizations:

```sh
$ docker-compose up -d
$ go run cmd/mini-rovers/main.go -t 123github456token _testdata/gh-orgs.test
[2019-05-30T14:25:58.220238016+02:00]  INFO processing data org=bblfsh page=0 repository=bblfsh/csharp-driver
[2019-05-30T14:25:58.220347282+02:00]  INFO data persisted org=bblfsh page=0 repository=bblfsh/csharp-driver
[2019-05-30T14:25:58.22037269+02:00]  INFO processing data org=bblfsh page=0 repository=bblfsh/client-go
[2019-05-30T14:25:58.220504598+02:00]  INFO data persisted org=bblfsh page=0 repository=bblfsh/client-go
[2019-05-30T14:25:58.221849298+02:00]  INFO rate limit reached, waiting 58m38s to retry org=github page=0
...

$ docker-compose down
```
