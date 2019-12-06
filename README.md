# k8sEventExporter
kubernetes event exporter
1.配置说明
Usage of k8sPodEventExporter.exe:
  -bindPort string
        0.0.0.0:9900 (default ":9900")
  -confFile string
        k8s配置文件 (default "~/.kube/config")
  -includeNormalEvent
        是否list包括正常状态的event
  -isInCluster
        是否部署在k8s cluster (default true)
  -namespace string
        namespace (default "default")
  -timeInterval int
        获取多长时间间隔内的非warning的event，单位分钟(default 5)


(1) k8s集群外部署
example: 
k8sPodEventExporter.exe -isInCluster=false -confFile "C:\\Users\\piter\\Desktop\\admin.conf"

(2) k8s集群内部署
编译打包镜像dockerfile文件

FROM golang:1.11-alpine as build
RUN apk add --no-cache git build-base
ADD . /src
WORKDIR /src
ENV GO111MODULE on
ENV CGO_ENABLED 0
RUN go build -o k8s-event-exporter -ldflags '-extldflags "-static"'

FROM scratch
COPY --from=build /src/k8s-event-exporter /k8s-event-exporter
EXPOSE 9102
ENTRYPOINT  [ "/k8s-event-exporter" ]

-------------------------------
k8s v1.15.4版本测试通过，其它没测
-------------------------------
---
svc配置
{
   "apiVersion": "v1",
   "kind": "Service",
   "metadata": {
      "annotations": {
         "prometheus.io/scrape": "true",
         "service.beta.kubernetes.io/aws-load-balancer-backend-protocol": "http",
         "service.beta.kubernetes.io/aws-load-balancer-ssl-ports": "https"
      },
      "labels": {
         "name": "k8s-event-exporter"
      },
      "name": "k8s-event-exporter"
   },
   "spec": {
      "ports": [
         {
            "name": "http",
            "port": 9102,
            "targetPort": 9102
         }
      ],
      "externalIPs": [
         "10.1.0.2"
      ],
      "selector": {
         "app": "k8s-event-exporter"
      }
   }
}

deployment配置
---
{
   "apiVersion": "apps/v1",
   "kind": "Deployment",
   "metadata": {
      "labels": {
         "app": "k8s-event-exporter"
      },
      "name": "k8s-event-exporter"
   },
   "spec": {
      "replicas": 1,
      "revisionHistoryLimit": 2,
      "selector": {
         "matchLabels": {
            "app": "k8s-event-exporter"
         }
      },
      "strategy": {
         "type": "RollingUpdate"
      },
      "template": {
         "metadata": {
            "labels": {
               "app": "k8s-event-exporter"
            }
         },
         "spec": {
            "containers": [
               {
                  "command": [
                     "./event_exporter"
                  ],
                  "env": [ ],
                  "image": "xxx/k8s-event-exporter",  //镜像仓库
                  "imagePullPolicy": "Always",
                  "name": "k8s-event-exporter",
                  "ports": [
                     {
                        "containerPort": 9102,
                        "name": "http"
                     }
                  ],
                  "resources": {
                     "limits": {
                        "memory": "100Mi"
                     },
                     "requests": {
                        "memory": "40Mi"
                     }
                  }
               }
            ],
            "terminationGracePeriodSeconds": 30
         }
      }
   }
}


2.检查k8s集群中的svc和pod都正常启动，并且telnet 9102端口没问题之后，修改prometheus的配置，拉取event exporter的上报数据
