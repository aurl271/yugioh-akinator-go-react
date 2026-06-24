import type { GameRequest, GameResponse, SettingHyperParameterRequest, ConfirmAnswerRequest } from "../type/game"

// API_BASE_URL はバックエンドのURL。
// Cloudflare Pagesなどに置く場合は VITE_API_BASE_URL で差し替える。
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080";

// startGame は設定値を送って、最初の質問を取得する。
export async function startGame(
  request: SettingHyperParameterRequest
): Promise<GameResponse> {
    const response = await fetch(`${API_BASE_URL}/api/game/start`,{
        method:"POST",
        headers:{
            "Content-Type": "application/json",
        },
        body: JSON.stringify(request),
    });

    if(!response.ok){
        throw new Error("failed to start game");
    }

    return response.json();
}

// answerQuestion は回答履歴を送って、次の質問または回答候補を取得する。
export async function answerQuestion(
    request: GameRequest
): Promise<GameResponse> {
    const response = await fetch(`${API_BASE_URL}/api/game/answer`,{
        method:"POST",
        headers:{
            "Content-Type": "application/json",
        },
        body: JSON.stringify(request),
    });

    if(!response.ok){
        throw new Error("failed to start game");
    }

    return response.json();
}

// confirmAnswer は提示されたカードが正しかったかをバックエンドへ送る。
export async function confirmAnswer(
    request: ConfirmAnswerRequest
): Promise<GameResponse> {
    const response = await fetch(`${API_BASE_URL}/api/game/confirm`,{
        method:"POST",
        headers:{
            "Content-Type": "application/json",
        },
        body: JSON.stringify(request),
    });

    if(!response.ok){
        throw new Error("failed to start game");
    }

    return response.json();
}
