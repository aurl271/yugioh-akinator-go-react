// AnswerValue はバックエンドの model.AnswerValue と対応する回答文字列。
export type AnswerValue =
  | "yes"
  | "probably"
  | "unknown"
  | "probably_no"
  | "no";

// QuestionContent は画面に表示する質問の最小情報。
export type QuestionContent = {
  id: number;
  text: string;
};

// AnsweredQuestion はフロントエンドが保持する回答履歴。
export type AnsweredQuestion = {
  questionContent: QuestionContent;
  answer: AnswerValue;
};

// Hyperparameters は推理の挙動を決める設定。
export type Hyperparameters = {
  // beta が大きいほど、回答と合う候補へ確率が強く寄る。
  beta: number;
  // answerThreshold を超えたら、バックエンドはカードを回答として提示する。
  answerThreshold: number;
  // topCandidatesCount は候補欄へ表示するカード数。
  topCandidatesCount: number;
};

// DEFAULT_* はUI側で使う推理設定の初期値。
export const DEFAULT_ANSWER_THRESHOLD = 0.9;
export const DEFAULT_TOP_CANDIDATES_COUNT = 10;

// SettingHyperParameterRequest はゲーム開始APIへ送るリクエスト。
export type SettingHyperParameterRequest = {
  meta: Hyperparameters;
};

// GameRequest は質問回答APIへ送るリクエスト。
// backendは状態を保存しないため、毎回questionsに回答履歴をすべて入れる。
export type GameRequest = {
  meta: Hyperparameters;
  questions: AnsweredQuestion[];
};

// CardData は回答として提示されたカード情報。
export type CardData = {
  cardId: number;
  cardName: string;
};

// ConfirmAnswerRequest は回答確認APIへ送るリクエスト。
export type ConfirmAnswerRequest = {
  meta: Hyperparameters;
  questions: AnsweredQuestion[];
  answerCard: CardData;
  isCorrect: boolean;
};

// Candidates は候補カード一覧に表示する1件分のデータ。
export type Candidates = {
  rank: number;
  cardId: number;
  cardName: string;
  probability: number;
};

// GameResponse はゲーム系APIの共通レスポンス。
// 画面状態によってquestionやanswerCardはnullになる。
export type GameResponse = {
  meta: Hyperparameters;
  question: QuestionContent | null;
  history: AnsweredQuestion[];
  candidates: Candidates[];
  isAnswer: boolean;
  isCorrect: boolean | null;
  answerCard: CardData | null;
};
