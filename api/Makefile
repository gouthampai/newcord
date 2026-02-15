.PHONY: run build test clean install-deps

run:
	go run cmd/server/main.go

build:
	go build -o bin/newcord-api cmd/server/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/

install-deps:
	go mod download
	go mod tidy

k8s-deploy:
	kubectl apply -k k8s/

k8s-delete:
	kubectl delete -k k8s/

k8s-status:
	kubectl get all -n newcord

k8s-logs:
	kubectl logs -n newcord -l app=cassandra --tail=100 -f

k8s-connect:
	kubectl exec -it -n newcord cassandra-0 -- cqlsh

k8s-port-forward:
	kubectl port-forward -n newcord svc/cassandra 9042:9042
