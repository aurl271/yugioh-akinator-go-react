DROP TABLE IF EXISTS answers;
DROP TABLE IF EXISTS questions;
DROP TABLE IF EXISTS cards;

CREATE TABLE cards (
  id BIGSERIAL PRIMARY KEY,
  card_id BIGINT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  reading TEXT,
  "desc" TEXT,
  setcode BIGINT NOT NULL,
  type BIGINT NOT NULL,
  atk INTEGER NOT NULL,
  def INTEGER NOT NULL,
  level INTEGER NOT NULL,
  race BIGINT NOT NULL,
  attribute BIGINT NOT NULL
);

CREATE TABLE questions (
  id BIGSERIAL PRIMARY KEY,
  question_text TEXT NOT NULL UNIQUE,
  category INTEGER CHECK (category IN (0, 1)) NOT NULL,
  query TEXT,
  condition_json TEXT,
  unset_bit INTEGER NOT NULL DEFAULT 0,
  new_state INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE answers (
  id BIGSERIAL PRIMARY KEY,
  card_id BIGINT NOT NULL,
  question_id BIGINT NOT NULL,
  answer INTEGER CHECK (answer IN (1, -1)) NOT NULL,
  FOREIGN KEY (card_id) REFERENCES cards(card_id),
  FOREIGN KEY (question_id) REFERENCES questions(id),
  UNIQUE (card_id, question_id)
);