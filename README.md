# deploy

## build
```bash
make docker-build docker-push IMG=<some-registry>/<project-name>:tag
```
## deploy
```bash
make deploy IMG=<some-registry>/<project-name>:tag
```
## create cr
````
example: create a cr monitor pod with label "monitor: hystrix.stream"
kubectl apply -f config/samples/monitor.wangjl.dev_v1beta1_turbine.yaml

note: one monitor can monitor many hystrixs, and one cr create one monitor.
so monitor can have many.
````
## label pod want to monitor
````
example: label pod should be monitor with label "monitor: hystrix.stream"
kubectl patch deployment xxxxx -p '{"spec":{"template":{"metadata":{"labels":{"monitor": "hystrix.stream"}}}}}'
````

