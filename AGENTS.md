# BlindTag Project Agent Rules

## 데이터베이스 테이블 명명 규칙 (Database Naming Convention)
모든 데이터베이스(Supabase / PostgreSQL) 테이블은 반드시 다음 형식을 준수해야 합니다:
**`{접두사}_{메인인덱스}_{서브내용}`**

- **`d_` (Data)** : 사용자 계정, 프로필, 벽면 낙서 등 핵심 데이터
  - `d_user_account` : 유저 계정 및 프로필
  - `d_graffiti_wall` : 벽면 낙서 데이터
- **`s_` (System)** : 스테이지 마스터, 서버 환경, 시스템 기준 설정
  - `s_stage_config` : 스테이지 규격 및 난이도 설정
- **`l_` (Log)** : 게임 클리어 기록, 점수 영수증, 행동/감사 로그
  - `l_game_score` : 게임 클리어 점수 및 랭크 영수증 로그
