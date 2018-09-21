# esalert
A go example to provide Elasticsearch alerting service.
Basically, it shows how to load alert configs in db and trigger related watchers to reload/start/stop by REST-api

Forget about x-pack!

## Requirement
- github.com/gin-gonic/gin
- github.com/jmoiron/sqlx
- github.com/koding/multiconfig
- go.uber.org/zap
- github.com/domodwyer/mailyak

## Usage
Firstly, create `alert_job` table in your db

Glide up your vendors

```shell
glide update
```

Go run this service 

Each time you create or change a alert row in db (or some backend CRUD webpage you may have), send request to trigger that alert watcher to reload.
```shell
curl -XPOST http://localhost:9000/watcher/{id}
```

Of course stopping the watcher if not more need
```shell
curl -XDELETE http://localhost:9000/watcher/{id}
```

Check running jobs
```shell
curl -XGET http://localhost:9000/watcher/
```

## More Actioner
Currently, the actions to carry out while condition in lua script meet including
- log
- http request
- wechat broadcast service
- mail

You can develop your own Actioner freely