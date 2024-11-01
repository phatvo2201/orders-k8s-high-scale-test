brew install k6
kubectl create configmap init-sql --from-file=init.sql

kubectl port-forward service/orders-service 8000:80

