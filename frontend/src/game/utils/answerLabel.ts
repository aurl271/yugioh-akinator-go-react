import type { AnswerValue } from "../../type/game";

export function answerLabel(answer: AnswerValue): string {
    switch (answer) {
        case "yes":
            return "はい";
        case "probably":
            return "たぶんはい";
        case "unknown":
            return "わからない";
        case "probably_no":
            return "たぶんいいえ";
        case "no":
            return "いいえ";
    }
}
