# alertmgr-sidecar service

`alertmgr-sidecar` is a mobiledgeX sidecar service for Alertmanager. It exposes a receiver and create/delete APIs to have a manipulate Alertmanager configuration file from the single common place. In addition is proxies all the requests for other Alertmanager apis to the Alertmanager running in the same pod.

## Usage

```
$ alertmgr-sidecar -h
Usage of alertmgr-sidecar:
  -alertmgrAddr string
    	Alertmanager address (default "0.0.0.0:9093")
  -alsologtostderr
    	log to standard error as well as files
  -configFile string
    	Alertmanager config file (default "/tmp/alertmanager.yml")
  -httpAddr string
    	Http API endpoint (default "0.0.0.0:9094")
  -d string
    	comma separated list of [etcd api notify dmedb dmereq locapi infra metrics upgrade info sampled]

```