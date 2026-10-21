.PHONY: run build test tidy docker-build deploy

# 로컬 개발 서버 실행
run:
	go run ./cmd/server/main.go

# 프로덕션 바이너리 빌드
build:
	go build -ldflags="-w -s" -o ./bin/blindtag-server ./cmd/server/main.go

# 의존성 정리
tidy:
	go mod tidy

# 테스트 실행
test:
	go test -v ./...

# 로컬 도커 빌드 & 실행 테스트
docker-build:
	docker build -t blindtag-api:local .

docker-run:
	docker run -p 8080:8080 --env-file .env blindtag-api:local

# gcloud CLI를 통한 단일 커맨드 Cloud Run 배포
# 사용법: make deploy PROJECT_ID=my-gcp-project
deploy:
	gcloud run deploy blindtag-api \
		--source . \
		--region asia-northeast3 \
		--allow-unauthenticated \
		--set-env-vars APP_ENV=production
