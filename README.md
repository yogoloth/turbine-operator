# 1. deploy

## 1.1 build
```bash
make docker-build docker-push IMG=<some-registry>/<project-name>:tag
```
## 1.2 deploy
```bash
make install
make deploy IMG=<some-registry>/<project-name>:tag
```
## 1.3 create cr
````
example: create a cr monitor pod with label "monitor: hystrix.stream"
kubectl apply -f config/samples/monitor.wangjl.dev_v1beta1_turbine.yaml

note: one monitor can monitor many hystrixs, and one cr create one monitor.
so monitor can have many.
````
## 1.4 label pod want to monitor
````
example: label pod should be monitor with label "monitor: hystrix.stream"
kubectl patch deployment xxxxx -p '{"spec":{"template":{"metadata":{"labels":{"monitor": "hystrix.stream"}}}}}'
````
## 1.5 undeploy
````
make undeploy
````
# 2. view dashboard
get nodeport_port
````
kubectl get svc turbinedashboard-[turbinemonitor_name]
````
dashboard addr is
````
http://[node_ip]]:[nodeport_port]/hystrix/monitor?stream=127.0.0.1%3A8000
````