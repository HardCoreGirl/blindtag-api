# BlindTag API Server

BlindTag 게임을 위한 고성능 백엔드 REST API 서버입니다.
**Go (Golang)** 로 작성되었으며, **GCP Cloud Run** 서버리스 환경과 **Supabase (PostgreSQL)** 데이터베이스에 최적화되어 있습니다.

---

## 🏗️ 시스템 아키텍처

```
BlindTag Unity Client (WebGL / PC Standalone)
                     │
                     ▼ HTTPS (REST API)
      ┌───────────────────────────────┐
      │  GCP Cloud Run (blindtag_api) │
      │  - Gin Web Framework          │
      │  - Graceful Shutdown          │
      │  - Stateless Auto-scaling     │
      └──────────────┬────────────────┘
                     │ pgxpool (TCP / SSL)
                     ▼
       ┌──────────────────────────────────────────────┐
       │             Supabase PostgreSQL              │
       │  - s_stage_config (스테이지 규격 시스템 설정)    │
       │  - d_user_account (게스트/닉네임/아이콘 데이터) │
       │  - d_graffiti_wall (벽면 낙서 공유 데이터)   │
       │  - l_game_score (클리어 점수/리더보드 로그)  │
       └──────────────────────────────────────────────┘
```

---

## 📁 디렉토리 구조

```
blindtag_api/
├── cmd/
│   └── server/
│       └── main.go                  # 엔트리포인트 (Cloud Run PORT 바인딩, Graceful Shutdown)
├── internal/
│   ├── config/                      # 환경변수 로더 (PORT, DATABASE_URL, SUPABASE_URL 등)
│   ├── database/                    # Supabase PostgreSQL 연결 풀 (pgxpool)
│   ├── models/                      # 데이터 모델 및 DTO (User, Score, Graffiti)
│   ├── repository/                  # 데이터베이스 CRUD 쿼리 계층
│   ├── handlers/                    # REST API 컨트롤러 (Health, User, Score, Graffiti)
│   └── router/                      # Gin 라우터, CORS 설정, 엔드포인트 등록
├── scripts/
│   └── supabase_schema.sql          # Supabase SQL Editor용 초기 테이블/인덱스 DDL
├── .env.example                     # 로컬 개발용 환경변수 템플릿
├── .dockerignore                    # 도커 빌드 제외 파일
├── Dockerfile                       # Multi-stage 빌드 (Alpine 기반 초경량 배포)
├── cloudbuild.yaml                  # GCP Cloud Build 자동 배포 파이프라인
├── Makefile                         # 빌드 및 실행 편의 명령어
└── README.md
```

---

## 🚀 빠른 시작 (Local Development)

### 1. Supabase 데이터베이스 생성
1. [Supabase](https://supabase.com)에 로그인 후 새 프로젝트를 생성합니다.
2. 좌측 메뉴 **SQL Editor**로 이동합니다.
3. [scripts/supabase_schema.sql](file:///e:/Works/Games/BlindTag/blindtag_api/scripts/supabase_schema.sql)의 내용을 복사하여 실행합니다. (`s_stage_config`, `d_user_account`, `d_graffiti_wall`, `l_game_score` 테이블과 인덱스가 생성됩니다.)
4. **Project Settings > Database > Connection string > URI** (Transaction Pooler 모드 권장) 연결 문자열을 복사합니다.

### 2. 로컬 환경 변수 설정
```bash
cd blindtag_api
copy .env.example .env
```
`.env` 파일을 열고 복사한 Supabase `DATABASE_URL`을 입력합니다.

### 3. 로컬 서버 실행
```bash
go run ./cmd/server/main.go
```
서버가 정상 구동되면 `http://localhost:8080/health` 에서 헬스체크 결과를 확인할 수 있습니다.

---

## ☁️ GCP Cloud Run 배포 방법

### 방법 A. gcloud CLI로 원클릭 배포 (가장 간단한 방법)
GCP CLI(`gcloud`)가 설치되어 있고 로그인이 완료된 상태에서:

```bash
cd blindtag_api

gcloud run deploy blindtag-api \
  --source . \
  --region asia-northeast3 \
  --allow-unauthenticated \
  --set-env-vars APP_ENV=production,DATABASE_URL="postgresql://postgres.[ref]:[pw]@aws-0-ap-northeast-2.pooler.supabase.com:6543/postgres?sslmode=require"
```
> **Tip:** 데이터베이스 비밀번호 등 민감 정보는 GCP **Secret Manager**를 통해 `--set-secrets`로 주입하시는 것을 권장합니다.

### 방법 B. GCP Cloud Build 파이프라인 이용
```bash
gcloud builds submit --config=cloudbuild.yaml
```

---

## 📡 API 엔드포인트 명세

### 1. 헬스체크 (Cloud Run Liveness Probe)
- **`GET /health`**
  - 응답:
    ```json
    {
      "status": "ok",
      "service": "blindtag_api",
      "database": "connected",
      "uptime": "1h23m"
    }
    ```

### 2. 유저 / 프로필 관리
- **`POST /api/v1/users/guest`**
  - 새 게스트 발급 또는 기존 게스트 동기화
  - 요청 Body:
    ```json
    {
      "guest_id": "RUNNER_A1B2C3",
      "nickname": "사이버러너",
      "profile_icon_index": 2
    }
    ```
- **`GET /api/v1/users/:id`** : 유저 프로필 조회 (UUID 또는 guest_id 지원)
- **`PUT /api/v1/users/:id`** : 닉네임 및 프로필 아이콘 변경
  - 요청 Body:
    ```json
    {
      "nickname": "새로운닉네임",
      "profile_icon_index": 3
    }
    ```

### 3. 스테이지 클리어 점수 & 리더보드
- **`POST /api/v1/scores`** : 게임 클리어 점수 영수증 제출
  - 요청 Body:
    ```json
    {
      "user_id": "RUNNER_A1B2C3",
      "stage_id": 1,
      "elapsed_time": 83.5,
      "failed_checks": 1,
      "room_moves": 14,
      "final_score": 8200,
      "rank": "A"
    }
    ```
- **`GET /api/v1/leaderboard/:stage_id?limit=20`** : 스테이지별 랭킹 TOP N 조회
  - 응답:
    ```json
    {
      "stage_id": 1,
      "count": 1,
      "data": [
        {
          "rank_position": 1,
          "user_id": "RUNNER_A1B2C3",
          "nickname": "사이버러너",
          "profile_icon_index": 2,
          "stage_id": 1,
          "final_score": 8200,
          "rank": "A",
          "elapsed_time": 83.5,
          "created_at": "2026-10-21T04:30:00Z"
        }
      ]
    }
    ```

### 4. 벽면 낙서 (Graffiti) 공유
- **`POST /api/v1/graffiti`** : 벽면에 그린 낙서 저장
  - 요청 Body:
    ```json
    {
      "user_id": "RUNNER_A1B2C3",
      "author_nickname": "사이버러너",
      "author_icon_index": 2,
      "stage_id": 1,
      "room_coord_x": 2,
      "room_coord_y": 1,
      "normal_x": 0.0,
      "normal_y": 0.0,
      "normal_z": -1.0,
      "position_x": 20.0,
      "position_y": 1.5,
      "position_z": 15.0,
      "image_data": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA..."
    }
    ```
- **`GET /api/v1/graffiti?stage_id=1&room_x=2&room_y=1`** : 해당 방의 낙서 목록 조회
- **`DELETE /api/v1/graffiti/:id?user_id=RUNNER_A1B2C3`** : 본인 낙서 지우기
