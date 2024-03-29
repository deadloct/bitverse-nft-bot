ifneq ("$(wildcard .env)", "")
	include .env
endif

NAME := bitverse-nft-bot
LOCAL_PATH := bin

clean:
	rm -rf $(LOCAL_PATH)

build: clean
	go build -o $(LOCAL_PATH)/$(NAME)

run: build
	$(LOCAL_PATH)

build_amd64: clean
	GOOS=linux GOARCH=amd64 go build -o $(LOCAL_PATH)/$(NAME)

deploy_amd64: build_amd64
	rsync -avz $(LOCAL_PATH)/ $(SSH_HOST):$(SSH_DIR)
	-ssh $(SSH_HOST) "killall $(NAME)"

build_arm: clean
	GOOS=linux GOARCH=arm64 GOARM=5 go build -o $(LOCAL_PATH)/$(NAME)

deploy_arm: build_arm
	rsync -avz $(LOCAL_PATH)/ $(SSH_HOST):$(SSH_DIR)
	-ssh $(SSH_HOST) "killall $(NAME)"

docker_build:
	docker build -t $(IMAGE_URL) .

docker_run: docker_build
	docker run -e "BITVERSE_NFT_BOT_AUTH_TOKEN=$(BITVERSE_NFT_BOT_AUTH_TOKEN)" $(IMAGE_URL)

docker_push: docker_build
	docker push $(IMAGE_URL)

k8s_deploy:
	kubectl apply -f k8s
