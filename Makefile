build:
	DOCKER_BUILDKIT=0 docker build -t github.com/hsmtkk/jubilant-happiness .

run:
	docker run --rm github.com/hsmtkk/jubilant-happiness

login:
	aws ecr get-login-password --region ap-northeast-1 | docker login --username AWS --password-stdin 384447982274.dkr.ecr.ap-northeast-1.amazonaws.com

tag: build
	docker tag github.com/hsmtkk/jubilant-happiness:latest 384447982274.dkr.ecr.ap-northeast-1.amazonaws.com/jubilant-happiness:latest

push: tag
	docker push 384447982274.dkr.ecr.ap-northeast-1.amazonaws.com/jubilant-happiness:latest
