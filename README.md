sango
=====

Online compiler

## REST API

### GET /api/list
Returns a list of available environments.

#### Request
GET /api/list

#### Response
```json
[
   {
      "id":"cpp",
      "name":"C++",
      "language":"C++",
      "version":"gcc version 4.8.2"
   },
   {
      "id":"go-latest",
      "name":"Go",
      "language":"Go",
      "version":"go1.3.3 linux/amd64"
   },
   {
      "id":"mruby-head",
      "name":"mruby",
      "language":"Ruby",
      "version":"mruby 1.0.0 (2014-01-10) 378aa8a9"
   }
]
```

### GET /api/log/:id
Returns an execution log.

#### Request
GET /api/log/9bfikbRaT9a

#### Response
```json
{
   "id":"9bfikbRaT9a",
   "environment":{
      "id":"go-latest",
      "name":"Go",
      "language":"Go",
      "version":"go1.3.3 linux/amd64"
   },
   "input":{
      "files":{
         "main.go":"package main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"Hello, 世界\")\n}\n"
      },
      "stdin":""
   },
   "output":{
      "build-stdout":"",
      "build-stderr":"",
      "run-stdout":"Hello, 世界\n",
      "run-stderr":"",
      "code":0,
      "signal":0,
      "status":"Success",
      "running-time":0.001559842
   },
   "date":"2014-10-31T04:57:51.708684402-04:00"
}
```

### POST /api/run
Executes the program.

#### Request
POST /api/run

```json
{
   "environment":"go-latest",
   "input":{
      "files":{
         "main.go":"package main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"Hello, 世界\")\n}\n"
      },
      "stdin":""
   }
}
```

#### Response
```json
{
   "id":"9LCmLzSz6JS",
   "environment":{
      "id":"go-latest",
      "name":"Go",
      "language":"Go",
      "version":"go1.3.3 linux/amd64"
   },
   "input":{
      "files":{
         "main.go":"package main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"Hello, 世界\")\n}\n"
      },
      "stdin":""
   },
   "output":{
      "build-stdout":"",
      "build-stderr":"",
      "run-stdout":"Hello, 世界\n",
      "run-stderr":"",
      "code":0,
      "signal":0,
      "status":"Success",
      "running-time":0.0016396940000000001
   },
   "date":"2014-10-31T05:02:25.886802983-04:00"
}
```

### Websocket /api/run/stream
Executes the program and returns real-time output.

#### Request
Websocket /api/run/stream

```json
{
   "environment":"php5",
   "input":{
      "files":{
         "main.php":"<?php\n\nfor ($i = 0; $i < 3; $i++) {\n    echo $i;\n    sleep(1);\n}"
      },
      "stdin":""
   }
}
```

#### Response
```json
{"tag":"run-stdout","data":"0"}
{"tag":"run-stdout","data":"1"}
{"tag":"run-stdout","data":"2"}
{
   "data":{
      "id":"75wPeGYfDzd",
      "environment":{
         "id":"php5",
         "name":"PHP",
         "language":"PHP",
         "version":"PHP 5.5.9-1ubuntu4.5"
      },
      "input":{
         "files":{
            "main.php":"\u003c?php\n\nfor ($i = 0; $i \u003c 3; $i++) {\n    echo $i;\n    sleep(1);\n}"
         },
         "stdin":""
      },
      "output":{
         "build-stdout":"",
         "build-stderr":"",
         "run-stdout":"012",
         "run-stderr":"",
         "code":0,
         "signal":0,
         "status":"Success",
         "running-time":3.018414556
      },
      "date":"2014-11-03T17:14:09.98496999+09:00"
   },
   "tag":"result"
}
```
