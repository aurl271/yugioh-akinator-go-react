CREATE TABLE IF NOT EXISTS game_results (
  id BIGSERIAL PRIMARY KEY,
  answer_card_id BIGINT NOT NULL,
  answer_card_name TEXT NOT NULL,
  is_correct BOOLEAN NOT NULL,
  answered_questions JSONB NOT NULL,
  beta DOUBLE PRECISION NOT NULL,
  answer_threshold DOUBLE PRECISION NOT NULL,
  top_candidates_count INTEGER NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE game_results ENABLE ROW LEVEL SECURITY;
