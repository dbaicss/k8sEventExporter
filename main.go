package main

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (

	icNormalEvent bool
	timeInterval int64
	nameSpace string
	portString string
	configFile string
	isInCluster bool
)

func init() {
	//对应配置选项
	flag.BoolVar(&icNormalEvent, "includeNormalEvent", false, "是否list包括正常状态的event")
	flag.BoolVar(&isInCluster, "isInCluster", true, "是否部署在k8s cluster")
	flag.Int64Var(&timeInterval, "timeInterval", 5, "获取多长时间间隔内的Normal event,单位是分钟")
	flag.StringVar(&nameSpace, "namespace", "default", "namespace")
	flag.StringVar(&portString, "bindPort", ":9900", "0.0.0.0:9900")
	flag.StringVar(&configFile, "confFile", "~/.kube/config", "k8s配置文件")
}


func main() {
	flag.Parse()
	//实例化k8s客户端
	eventClient, err := NewEventClient()
	if err != nil {
		log.Fatalf("create the event client Failed,err: %s", err)
	}
	//prometheus注册
	eventClient.IncludeNormalEvents = icNormalEvent
	if err := prometheus.Register(NewExporter(eventClient)); err != nil {
		log.Fatalf("register exporter failed,err: %s", err)
	}

	//启动httpserver
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(portString, nil))
}
