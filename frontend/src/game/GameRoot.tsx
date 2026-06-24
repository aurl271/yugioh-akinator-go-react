import { Outlet,useNavigate,useLocation } from "react-router-dom";
import { useState } from "react"
import AppBar from '@mui/material/AppBar';
import Box from '@mui/material/Box';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import StepProgress from "./components/StepProgress";
import type { AnsweredQuestion, QuestionContent, Candidates, CardData,SettingHyperParameterRequest, Hyperparameters,GameRequest,AnswerValue,ConfirmAnswerRequest} from "../type/game";
import {startGame,answerQuestion,confirmAnswer} from "../api/gameApi"
import CircularProgress from "@mui/material/CircularProgress";
import Alert from "@mui/material/Alert";

// calcProgressPercent は候補1位の確率を、ユーザー向けの進行度表示に変換する。
function calcProgressPercent(candidates : Candidates[],answerThreshold : number|null): number{
    if(answerThreshold === null){
        return 0;
    }
    const topProbability = candidates[0]?.probability ?? 0;
    if(answerThreshold <= 0){
        return 0;
    }

    return Math.min(100,Math.round((topProbability / answerThreshold) * 100))
}

// calcElapsedTime は開始時刻と終了時刻からフィードバック画面用の経過時間を作る。
function calcElapsedTime(startedAt : number|null,finishedAt : number|null): string{
    if(startedAt === null || finishedAt === null){
        return "0:00";
    }

    const elapsedMs = finishedAt - startedAt;
    const totalSeconds = Math.floor(elapsedMs / 1000);
    const minutes = Math.floor(totalSeconds / 60);
    const seconds = totalSeconds % 60;
    return `${minutes}:${seconds.toString().padStart(2, "0")}`;
}

// currentStepFromPath はURLから現在のゲームステップを決める。
function currentStepFromPath(pathname: string): number {
    if (pathname === "/game/question") {
        return 2;
    }
    if (pathname === "/game/confirm") {
        return 3;
    }
    if (pathname === "/game/feedback") {
        return 4;
    }

    return 1;
}

// GameOutletContext はGameRootから各ページへ渡す共有状態と操作関数。
export type GameOutletContext = {
    state: {
        meta: Hyperparameters | null;
        question: QuestionContent | null;
        candidates: Candidates[];
        answeredQuestions: AnsweredQuestion[];
        answerCard: CardData | null;
        isCorrect: boolean | null;
        progressPercent: number;
        questionCount: number;
        elapsedTime: string;
        loading: boolean;
    };
    actions: {
        handleStartGame: (selectedMeta: Hyperparameters) => Promise<void>;
        handleQuestion: (answer: AnswerValue) => Promise<void>;
        handleBackQuestion: () => Promise<void>;
        handleResetGame: () => Promise<void>;
        handleConfirm: (isCorrect: boolean) => Promise<void>;
        handleFeedback: () => void;
    };
};

function GameRoot() {
    const navigate = useNavigate();
    const location = useLocation();

    // loading/error はAPI通信中や失敗時の画面表示を制御する。
    const [loading,setLoading] = useState<boolean>(false);
    const [error,setError] = useState<string|null>(null);
    // meta は現在の推理設定。履歴リセットや回答送信でも使い回す。
    const [meta,setMeta] = useState<Hyperparameters|null>(null);
    // answeredQuestions はfrontendが保持する回答履歴。backendへ毎回送る。
    const [answeredQuestions,setAnsweredQuestions] = useState<AnsweredQuestion[]>([]);
    // question は現在表示中の質問。confirm/feedbackではnullになる。
    const [question,setQuestion] = useState<QuestionContent|null>(null);
    // candidates は候補欄に表示するカード一覧。
    const [candidates,setCandidates] = useState<Candidates[]>([]);
    // answerCard/isCorrect は回答確認からフィードバック画面で使う。
    const [answerCard,setAnswerCard] = useState<CardData|null>(null);
    const [isCorrect,setIsCorrect] = useState<boolean|null>(null);
    const progressPercent = calcProgressPercent(candidates,meta?.answerThreshold ?? null);
    const questionCount = answeredQuestions.length;
    // startedAt/finishedAt はゲーム終了時に経過時間を出すための時刻。
    const [startedAt, setStartedAt] = useState<number | null>(null);
    const [finishedAt, setFinishedAt] = useState<number | null>(null);
    const elapsedTime = calcElapsedTime(startedAt,finishedAt);
    const currentStep = currentStepFromPath(location.pathname);

    // resetGameState はsetupへ戻ったとき、前回ゲームの状態をすべて消す。
    function resetGameState(): void {
        setLoading(false);
        setMeta(null);
        setAnsweredQuestions([]);
        setQuestion(null);
        setCandidates([]);
        setAnswerCard(null);
        setIsCorrect(null);
        setStartedAt(null);
        setFinishedAt(null);
    }

    const state = {
        meta,
        question,
        candidates,
        answeredQuestions,
        answerCard,
        isCorrect,
        progressPercent,
        questionCount,
        elapsedTime,
        loading,
    };

    const actions = {
        handleStartGame,
        handleQuestion,
        handleBackQuestion,
        handleResetGame,
        handleConfirm,
        handleFeedback,
    };

    // handleStartGame は選択された推理設定で最初の質問を取得する。
    async function handleStartGame(selectedMeta : Hyperparameters): Promise<void> {
        console.log("[GameRoot] start game", selectedMeta);

        const request:SettingHyperParameterRequest = {
            meta : selectedMeta,
        };

        resetGameState();
        setError(null);
        setLoading(true);

        try {
            console.log("[GameRoot] request startGame", request);
            const response = await startGame(request);
            console.log("[GameRoot] received startGame response", response);

            setMeta(response.meta);
            setQuestion(response.question);
            setCandidates(response.candidates);
            setStartedAt(Date.now());
            setError(null);

            navigate("/game/question");
        } catch (error) {
            console.error("[GameRoot] failed to start game", error);
            setError("ゲームを開始できませんでした。もう一度試してください。");
        }finally {
            setLoading(false);
        }
    }

    // handleQuestion は現在の質問への回答を履歴に追加し、次状態をbackendへ問い合わせる。
    async function handleQuestion(answer : AnswerValue): Promise<void> {
        console.log("[GameRoot] answer question", { answer });

        if(meta === null){
            console.warn("[GameRoot] missing meta while answering question");
            setError("ゲーム設定が見つかりません。設定画面からやり直してください。");
            navigate("/game/setup");
            return;
        }

        if(question === null){
            console.warn("[GameRoot] missing question while answering question");
            setError("質問が見つかりません。設定画面からやり直してください。");
            navigate("/game/setup");
            return;
        }

        const newAnsweredQuestion: AnsweredQuestion = {
            questionContent: question,
            answer,
        };

        const nextAnsweredQuestions = [...answeredQuestions, newAnsweredQuestion];

        const request: GameRequest = {
            meta,
            questions: nextAnsweredQuestions,
        };

        setError(null);
        setLoading(true);

        try {
            console.log("[GameRoot] request answerQuestion", request);
            const response = await answerQuestion(request);
            console.log("[GameRoot] received answerQuestion response", response);

            if (response.isAnswer) {
                if (response.answerCard === null) {
                    setError("予想カードが見つかりません。もう一度やり直してください。");
                    navigate("/game/setup");
                    return;
                }
                console.log("[GameRoot] answer candidate is ready; navigate to confirm");
                setMeta(response.meta);
                setAnsweredQuestions(nextAnsweredQuestions);
                setAnswerCard(response.answerCard);
                setCandidates(response.candidates);
                setError(null);
                navigate("/game/confirm");
            }else{
                console.log("[GameRoot] next question received");
                setMeta(response.meta);
                setAnsweredQuestions(nextAnsweredQuestions);
                setQuestion(response.question);
                setCandidates(response.candidates);
                setError(null);
            }
        }catch (error){
            console.error("[GameRoot] failed to answer question", error);
            setError("質問が取得できませんでした。");
        }finally{
            setLoading(false);
        }
    }

    // handleBackQuestion は最後の回答を1件取り消し、その履歴でbackendへ再問い合わせする。
    async function handleBackQuestion(): Promise<void> {
        console.log("[GameRoot] back question");

        if (meta === null) {
            console.warn("[GameRoot] missing meta while going back");
            setError("ゲーム設定が見つかりません。設定画面からやり直してください。");
            navigate("/game/setup");
            return;
        }

        if (answeredQuestions.length === 0) {
            console.log("[GameRoot] no answered question to go back");
            return;
        }

        const nextAnsweredQuestions = answeredQuestions.slice(0, -1);

        const request: GameRequest = {
            meta,
            questions: nextAnsweredQuestions,
        };

        setError(null);
        setLoading(true);

        try {
            console.log("[GameRoot] request back question", request);
            const response = await answerQuestion(request);
            console.log("[GameRoot] received back question response", response);

            setMeta(response.meta);
            setAnsweredQuestions(nextAnsweredQuestions);
            setCandidates(response.candidates);

            if (response.isAnswer) {
                if (response.answerCard === null) {
                    setError("予想カードが見つかりません。もう一度やり直してください。");
                    navigate("/game/setup");
                    return;
                }
                setAnswerCard(response.answerCard);
                setQuestion(null);
                setError(null);
                navigate("/game/confirm");
                return;
            }

            if (response.question === null) {
                setError("質問が見つかりません。設定画面からやり直してください。");
                navigate("/game/setup");
                return;
            }

            setAnswerCard(null);
            setQuestion(response.question);
            setError(null);
            navigate("/game/question");
        } catch (error) {
            console.error("[GameRoot] failed to go back question", error);
            setError("前の質問に戻れませんでした。");
        } finally {
            setLoading(false);
        }
    }

    // handleConfirm は提示カードが正しかったかを保存し、フィードバック画面へ進める。
    async function handleConfirm(isCorrect : boolean): Promise<void> {
        console.log("[GameRoot] confirm answer", { isCorrect });

        if(meta === null){
            console.warn("[GameRoot] missing meta while confirming answer");
            setError("ゲーム設定が見つかりません。設定画面からやり直してください。");
            navigate("/game/setup");
            return;
        }

        if (answerCard == null) {
            console.warn("[GameRoot] missing answerCard while confirming answer");
            setError("予想カードが見つかりません。もう一度やり直してください。");
            navigate("/game/setup");
            return;
        }

        const request : ConfirmAnswerRequest = {
            meta:meta,
            questions:answeredQuestions,
            answerCard:answerCard,
            isCorrect:isCorrect,
        };

        setError(null);
        setLoading(true);

        try {
            console.log("[GameRoot] request confirmAnswer", request);
            const response = await confirmAnswer(request);
            console.log("[GameRoot] received confirmAnswer response", response);
            setIsCorrect(response.isCorrect);
            setFinishedAt(Date.now());
            setError(null);
            navigate("/game/feedback");
        }catch (error){
            console.error("[GameRoot] failed to confirm answer", error);
            setError("回答を送信できませんでした。設定画面からやり直してください。")
            navigate("/game/setup");
        }finally{
            setLoading(false);
        }
    }

    // handleFeedback はフィードバック画面からsetupへ戻る。
    function handleFeedback(): void {
        console.log("[GameRoot] restart from feedback");
        setError(null);
        navigate("/game/setup");
    }

    // handleResetGame は同じ推理設定のまま回答履歴だけを空にする。
    async function handleResetGame(): Promise<void> {
        console.log("[GameRoot] reset question history");

        if (meta === null) {
            console.warn("[GameRoot] missing meta while resetting question history");
            setError("ゲーム設定が見つかりません。設定画面からやり直してください。");
            navigate("/game/setup");
            return;
        }

        const request: SettingHyperParameterRequest = {
            meta,
        };

        setError(null);
        setLoading(true);

        try {
            console.log("[GameRoot] request reset question history", request);
            const response = await startGame(request);
            console.log("[GameRoot] received reset question history response", response);

            if (response.question === null) {
                setError("質問が見つかりません。設定画面からやり直してください。");
                navigate("/game/setup");
                return;
            }

            setMeta(response.meta);
            setAnsweredQuestions([]);
            setQuestion(response.question);
            setCandidates(response.candidates);
            setAnswerCard(null);
            setIsCorrect(null);
            setStartedAt(Date.now());
            setFinishedAt(null);
            setError(null);
            navigate("/game/question");
        } catch (error) {
            console.error("[GameRoot] failed to reset question history", error);
            setError("質問履歴をリセットできませんでした。");
        } finally {
            setLoading(false);
        }
    }

    return (
        <Box sx={{
            minHeight: "100vh",
            bgcolor:"#dbdbdb",
        }}>
            <AppBar
            position="static"
            sx={{
                bgcolor: "#ffffff",
                boxShadow: "none",
            }}
            >
                <Toolbar
                sx={{
                display: "flex",
                justifyContent: "space-between",
                }}
                >
                    <Typography variant="h6" component="div" sx={{color:"black"}}>
                        遊戯王アキネーター
                    </Typography>

                    <Box>
                        <Button
                            component="a"
                            href=""
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            使い方
                        </Button>
                        <Button
                            component="a"
                            href=""
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            github
                        </Button>
                        <Button
                            component="a"
                            href="https://x.com/Nobu27182"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            製作者
                        </Button>
                    </Box>
                </Toolbar>
            </AppBar>

            {error !== null && (
                <Alert severity="error" sx={{ mb: 2 }}>
                    {error}
                </Alert>
            )}
            

            <Box
                sx={{
                    marginX:"50px",
                    marginY:"30px"
                }}>
                <StepProgress currentStep={currentStep} />
                <Outlet context={{ state, actions }} />
            </Box>

            {loading && (
                <Box
                    sx={{
                    position: "fixed",
                    inset: 0,
                    bgcolor: "rgba(240, 240, 240, 0.65)",
                    zIndex: 2000,
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                    }}
                >
                    <CircularProgress
                    size={56}
                    thickness={4}
                    sx={{
                        color: "#6d3fa0",
                    }}
                    />
                </Box>
            )}
        </Box>
    )
}

export default GameRoot
