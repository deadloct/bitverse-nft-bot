ifneq ("$(wildcard .env)", "")
	include .env
endif

IMAGE_URL=registry.digitalocean.com/deadloct/bitverse-nft-bot

build:
	go build

run: build
	./bitverse-nft-bot

clean:
	rm bitverse-nft-bot

docker_build:
	docker build -t $(IMAGE_URL) .

docker_run: docker_build
	docker run -e "BITVERSE_NFT_BOT_AUTH_TOKEN=$(BITVERSE_NFT_BOT_AUTH_TOKEN)" $(IMAGE_URL)

docker_push: docker_build
	docker push $(IMAGE_URL)

k8s_deploy:
	kubectl apply -f k8s

