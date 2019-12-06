package main

import (
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"fmt"
)

type EventClient struct {
	kube                *kubernetes.Clientset
	IncludeNormalEvents bool
	TimeWindowMinutes   time.Duration
}

func NewEventClient() (client EventClient, err error) {
	var config *rest.Config
	//判断集群外的初始化配置
	if !isInCluster {
		fmt.Printf("confFile:%s\n",configFile)
		config, err = clientcmd.BuildConfigFromFlags("", configFile)
		if err != nil {
			fmt.Printf("build  config flags err:%v\n",err)
			log.Fatalf("error connecting to the client: %v", err)
		}
	}else{
		//集群内方式初始化配置
		config, err = rest.InClusterConfig()
		if err != nil {
			return
		}
	}

	//初始化客户端
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return
	}

	client = EventClient{
		kube: kubeClient,
		TimeWindowMinutes: time.Duration(timeInterval),
	}

	return
}

func (e EventClient) Scrape(ch chan<- prometheus.Metric) error {

	//过滤event类型为Warning
	opts := metav1.ListOptions{FieldSelector: "type==Warning"}

	if e.IncludeNormalEvents {
		opts = metav1.ListOptions{}
	}

	list, err := e.kube.CoreV1().Events("").List(opts)
	if err != nil {
		return err
	}

	//设置时间间隔
	timeWindow := &metav1.Time{
		Time: time.Now().Add(-e.TimeWindowMinutes * time.Minute),
	}

	for _, event := range list.Items {
		//判断在时间间隔之内的Warning状态的，有可能是人工处理的重启和删除等操作
		if event.LastTimestamp.Before(timeWindow) {
			continue
		}
		//生成pro metric
		metric, err := prometheus.NewConstMetric(
			prometheus.NewDesc(
				"event_detail",
				"events",
				[]string{"event_namespace", "event_type", "event_name", "event_message", "object_kind", "object_name", "event_reason"},
				nil,
			),
			prometheus.GaugeValue,
			float64(event.Count),
			event.InvolvedObject.Namespace,
			event.Type,
			event.Name,
			event.Message,
			event.InvolvedObject.Kind,
			event.InvolvedObject.Name,
			event.Reason,
		)
		if err != nil {
			return err
		}
		//metric塞进channel
		ch <- metric
	}

	return nil
}
