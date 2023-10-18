# API

Описание api функций и примеры их вызова

# TOC

- [API](#-api)
- [TOC](#-toc)
  - [Methods BaseContract](#-methods-basecontract)
    - [QueryBuildInfo](#-querybuildinfo)
    - [QueryCoreChaincodeIDName](#-querycorechaincodeidname)
    - [QueryNameOfFiles](#-querynameoffiles)
    - [QuerySrcFile](#-querysrcfile)
    - [QuerySrcPartFile](#-querysrcpartfile)
    - [QuerySystemEnv](#-querysystemenv)
  - [Example](#-example)
- [Links](#-links)

## Methods BaseContract

Методы структуры `BaseContract`. У любого чейнкода в котрый встроен `BaseContract` есть такие методы. 

### QueryBuildInfo

```
func (bc *BaseContract) QueryBuildInfo() (*debug.BuildInfo, error)
```

QueryBuildInfo возвращает результат вычисления `debug.ReadBuildInfo()` в чейнкоде

### QueryCoreChaincodeIDName

```
func (bc *BaseContract) QueryCoreChaincodeIDName() (string, error)
```

QueryCoreChaincodeIDName возвращает значение переменной окружения `CORE_CHAINCODE_ID_NAME` в чейнкоде

### QueryNameOfFiles

```
func (bc *BaseContract) QueryNameOfFiles() ([]string, error)
```

QueryNameOfFiles возвращает список имен файлов исходников чейнкода [встроенных](embed.md) в чейнкод

### QuerySrcFile

```
func (bc *BaseContract) QuerySrcFile(name string) (string, error)
```

QuerySrcFile возвращает файл исходника чейнкода [встроенного](embed.md) в чейнкод

### QuerySrcPartFile

```
func (bc *BaseContract) QuerySrcPartFile(name string, start int, end int) (string, error)
```

QuerySrcPartFile возвращает часть файла (если файл большого размера) исходника чейнкода [встроенного](embed.md) в чейнкод

### QuerySystemEnv

```
func (bc *BaseContract) QuerySystemEnv() (map[string]string, error)
```

QuerySystemEnv возвращает системное окружение чейнкода. Это файлы: 
- `/etc/issue`
- `/etc/resolv.conf`
- `/proc/meminfo`
- `/proc/cpuinfo`
- `/etc/timezone`
- `/proc/diskstats`
- `/proc/loadavg`
- `/proc/version`
- `/proc/uptime`
- `/etc/hyperledger/fabric/client.crt`
- `/etc/hyperledger/fabric/peer.crt`

## Example

Все примеры сделаны для отправки в hlf-proxy

- QueryBuildInfo
```shell
curl -X 'POST' \
  'http://127.0.0.1:9001/query' \
  -H 'accept: */*' \
  -H 'Content-Type: application/json' \
  -d '{
  "channel": "cc",
  "chaincodeId": "cc",
  "fcn": "buildInfo",
  "args": []
}'
```

- QueryCoreChaincodeIDName
```shell
curl -X 'POST' \
  'http://127.0.0.1:9001/query' \
  -H 'accept: */*' \
  -H 'Content-Type: application/json' \
  -d '{
  "channel": "cc",
  "chaincodeId": "cc",
  "fcn": "coreChaincodeIDName",
  "args": []
}'
```

- QueryNameOfFiles
```shell
curl -X 'POST' \
  'http://127.0.0.1:9001/query' \
  -H 'accept: */*' \
  -H 'Content-Type: application/json' \
  -d '{
  "channel": "cc",
  "chaincodeId": "cc",
  "fcn": "nameOfFiles",
  "args": []
}'
```

- QuerySrcFile
```shell
curl -X 'POST' \
  'http://127.0.0.1:9001/query' \
  -H 'accept: */*' \
  -H 'Content-Type: application/json' \
  -d '{
  "channel": "cc",
  "chaincodeId": "cc",
  "fcn": "srcFile",
  "args": ["folder/token.go"]
}'
```

- QuerySrcPartFile
```shell
curl -X 'POST' \
  'http://127.0.0.1:9001/query' \
  -H 'accept: */*' \
  -H 'Content-Type: application/json' \
  -d '{
  "channel": "cc",
  "chaincodeId": "cc",
  "fcn": "srcPartFile",
  "args": ["folder/token.go", "8", "23"]
}'
```

- QuerySystemEnv
```shell
curl -X 'POST' \
  'http://127.0.0.1:9001/query' \
  -H 'accept: */*' \
  -H 'Content-Type: application/json' \
  -d '{
  "channel": "cc",
  "chaincodeId": "cc",
  "fcn": "systemEnv",
  "args": [],
  "options": {
    "targetEndpoints": ["test-peer-001.org0"]
  }
}'
```

# Links

* No
