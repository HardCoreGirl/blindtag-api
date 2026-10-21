-- =========================================================================
-- BlindTag Game Supabase Schema DDL (보안 강화 버전)
-- 테이블 명명 규칙: {접두사}_{메인인덱스}_{서브내용}
--   d_ : 데이터 (Data)
--   s_ : 시스템 (System)
--   l_ : 로그 (Log)
-- =========================================================================

-- UUID 확장
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =========================================================================
-- [s_] 시스템 테이블 (System)
-- =========================================================================

-- 1. 스테이지 마스터 설정 (1단계: 4x4, 2단계: 6x6, 3단계: 10x10)
-- ※ goal_coord_x, goal_coord_y: 모든 러너가 공유하는 고정 미로 출구 정답 (클라이언트에 미노출)
CREATE TABLE IF NOT EXISTS public.s_stage_config (
    stage_id INT PRIMARY KEY,
    stage_name TEXT NOT NULL,
    grid_width INT NOT NULL,
    grid_height INT NOT NULL,
    max_sprays INT NOT NULL DEFAULT 3,
    goal_coord_x INT NOT NULL DEFAULT 3,               -- 모든 러너가 공유하는 비밀 출구 방 좌표 X
    goal_coord_y INT NOT NULL DEFAULT 2,               -- 모든 러너가 공유하는 비밀 출구 방 좌표 Y
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc'::text, now())
);

-- 스테이지별 마스터 출구 좌표 등록
INSERT INTO public.s_stage_config (stage_id, stage_name, grid_width, grid_height, max_sprays, goal_coord_x, goal_coord_y, description)
VALUES 
    (1, '입문 훈련소 (Stage 1)', 4, 4, 3, 3, 2, '4x4 크기의 콤팩트한 미로. 결승 제단을 빠르게 찾는 기초 훈련 코스.'),
    (2, '뒤틀린 복도 (Stage 2)', 6, 6, 3, 4, 5, '6x6 크기의 중형 미로. 낙서 길잡이가 본격적으로 요구되는 코스.'),
    (3, '심연의 루프 (Stage 3)', 10, 10, 3, 6, 7, '10x10 (100개 방) 무한 순환 토러스 루프 미로. 최고 난이도.')
ON CONFLICT (stage_id) DO UPDATE SET 
    goal_coord_x = EXCLUDED.goal_coord_x,
    goal_coord_y = EXCLUDED.goal_coord_y;


-- =========================================================================
-- [d_] 데이터 테이블 (Data)
-- =========================================================================

-- 2. 유저 계정 및 프로필 데이터 (기기별 비밀키 해시 검증 탑재)
CREATE TABLE IF NOT EXISTS public.d_user_account (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),      -- 내부 고유 식별자 (UID / PK)
    device_uuid TEXT UNIQUE NOT NULL,                   -- 클라이언트 기기 고유 UUID (비공개)
    secret_hash TEXT NOT NULL,                          -- 클라이언트 비밀 토큰(client_secret)의 SHA-256 해시값 (비밀번호 역할)
    display_id TEXT UNIQUE NOT NULL,                    -- 공개 고유 러너 배번 (예: BT-7K9M2P)
    display_name TEXT NOT NULL,                         -- 유저 닉네임 (표시 이름)
    profile_icon_index INT NOT NULL DEFAULT 0,          -- 프로필 아이콘 번호 (0~4)
    created_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc'::text, now())
);

CREATE INDEX IF NOT EXISTS idx_d_user_account_device_uuid ON public.d_user_account(device_uuid);
CREATE INDEX IF NOT EXISTS idx_d_user_account_display_id ON public.d_user_account(display_id);


-- 3. 벽면 낙서(Graffiti) 데이터
CREATE TABLE IF NOT EXISTS public.d_graffiti_wall (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES public.d_user_account(id) ON DELETE CASCADE,
    author_display_id TEXT NOT NULL,                    -- 서명 러너 배번 (예: BT-7K9M2P)
    author_name TEXT NOT NULL,                          -- 서명 닉네임
    author_icon_index INT NOT NULL DEFAULT 0,           -- 서명 아이콘 인덱스
    stage_id INT NOT NULL REFERENCES public.s_stage_config(stage_id),
    room_coord_x INT NOT NULL,                          -- 미로 방 좌표 X
    room_coord_y INT NOT NULL,                          -- 미로 방 좌표 Y
    normal_x REAL NOT NULL DEFAULT 0,                   -- 벽 노멀 벡터 X
    normal_y REAL NOT NULL DEFAULT 0,                   -- 벽 노멀 벡터 Y
    normal_z REAL NOT NULL DEFAULT -1,                  -- 벽 노멀 벡터 Z
    position_x REAL NOT NULL DEFAULT 0,                 -- 벽면 데칼 월드 X
    position_y REAL NOT NULL DEFAULT 0,                 -- 벽면 데칼 월드 Y
    position_z REAL NOT NULL DEFAULT 0,                 -- 벽면 데칼 월드 Z
    image_data TEXT NOT NULL,                           -- 2D PNG 이미지 Base64
    created_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc'::text, now())
);

CREATE INDEX IF NOT EXISTS idx_d_graffiti_wall_room ON public.d_graffiti_wall(stage_id, room_coord_x, room_coord_y, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_d_graffiti_wall_user ON public.d_graffiti_wall(user_id);


-- =========================================================================
-- [l_] 로그 테이블 (Log)
-- =========================================================================

-- 4. 스테이지 클리어 점수 영수증 로그 (리더보드용)
CREATE TABLE IF NOT EXISTS public.l_game_score (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES public.d_user_account(id) ON DELETE CASCADE,
    stage_id INT NOT NULL REFERENCES public.s_stage_config(stage_id),
    elapsed_time REAL NOT NULL,                         -- 소요 시간(초)
    failed_checks INT NOT NULL DEFAULT 0,               -- 오답 제단 확인 횟수
    room_moves INT NOT NULL DEFAULT 0,                  -- 방 이동 횟수
    final_score INT NOT NULL,                           -- 최종 10,000점 만점 차감 점수
    rank VARCHAR(4) NOT NULL,                           -- S, A, B, C, D
    created_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc'::text, now())
);

CREATE INDEX IF NOT EXISTS idx_l_game_score_leaderboard ON public.l_game_score(stage_id, final_score DESC, elapsed_time ASC);


-- =========================================================================
-- Row Level Security (RLS) 정책
-- =========================================================================
ALTER TABLE public.s_stage_config ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.d_user_account ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.d_graffiti_wall ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.l_game_score ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Public Read Stage Config" ON public.s_stage_config FOR SELECT USING (true);
CREATE POLICY "Public Read Users" ON public.d_user_account FOR SELECT USING (true);
CREATE POLICY "Public Read Graffitis" ON public.d_graffiti_wall FOR SELECT USING (true);
CREATE POLICY "Public Read Scores" ON public.l_game_score FOR SELECT USING (true);

CREATE POLICY "Service Role Stage Config" ON public.s_stage_config FOR ALL USING (true);
CREATE POLICY "Service Role Users" ON public.d_user_account FOR ALL USING (true);
CREATE POLICY "Service Role Graffitis" ON public.d_graffiti_wall FOR ALL USING (true);
CREATE POLICY "Service Role Scores" ON public.l_game_score FOR ALL USING (true);
