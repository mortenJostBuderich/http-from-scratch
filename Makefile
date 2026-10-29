image:
	docker build -t my-server .

deploy-local:
	docker build -t my-server .
	kind load docker-image my-server
	kubectl apply -f deployment.yaml
	kubectl apply -f service.yaml