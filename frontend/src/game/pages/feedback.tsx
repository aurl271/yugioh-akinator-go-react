import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import lampMajin from "../../assets/lamp_majin.png";
import { useEffect } from "react";
import { useNavigate, useOutletContext } from "react-router-dom";
import type { GameOutletContext } from "../GameRoot";
import { answerLabel } from "../utils/answerLabel";
import { formatProbability } from "../utils/formatProbability";

function FeedbackPage() {
    const navigate = useNavigate();
  
    // フィードバック画面では最終結果・履歴・統計値だけを参照する。
    const {
        state: { loading,candidates,answeredQuestions,answerCard,isCorrect,questionCount,elapsedTime },
        actions: { handleFeedback },
    } = useOutletContext<GameOutletContext>();

    // 結果情報が無いまま表示されないよう、直接アクセス時はsetupへ戻す。
    useEffect(() => {
        if (answerCard === null || isCorrect === null) {
            navigate("/game/setup");
        }
    }, [answerCard, isCorrect, navigate]);

    if (answerCard === null || isCorrect === null) {
        return null;
    }

    // 最終自信度は候補1位の確率を表示用の文字列にする。
    const topCandidate = candidates[0] ?? null;
    const finalConfidence = formatProbability(topCandidate?.probability);

    return (
        <Box
            sx={{
                display: "grid",
                gridTemplateColumns: { xs: "1fr", lg: "1fr 280px" },
                gap: 2,
                alignItems: "stretch",
            }}
        >
                <Paper
                    sx={{
                        p: { xs: 3, md: 4 },
                        borderRadius: 2,
                        border: "1px solid #d8d8df",
                        display: "flex",
                        flexDirection: { xs: "column", md: "row" },
                        gap: { xs: 3, md: 5 },
                        alignItems: "center",
                    }}
                >
                    <Box
                        component="img"
                        src={lampMajin}
                        alt="ランプの魔人"
                        sx={{
                            width: { xs: 180, md: 260 },
                            height: "auto",
                            flexShrink: 0,
                        }}
                    />

                    <Stack
                        spacing={2.5}
                        sx={{
                            flex: 1,
                            width: "100%",
                            alignItems: "center",
                            textAlign: "center",
                        }}
                    >
                        <Typography
                            variant="h4"
                            component="h1"
                            sx={{
                                fontWeight: 800,
                                color: "#6d3fa0",
                            }}
                        >
                            {isCorrect === null
                            ? "結果がありません"
                            : isCorrect
                                ? "正解しました！"
                                : "失敗しました..."}
                        </Typography>
                        <Typography variant="body2" sx={{ color: "#5f5a66" }}>
                            予想したカードはこれでした{isCorrect ? "！" : ""}
                        </Typography>

                        <Box
                            sx={{
                                width: "100%",
                                maxWidth: 420,
                                border: "1px solid #7b4db1",
                                borderRadius: 1,
                                py: 2,
                                px: 3,
                                bgcolor: "#f8f8f8",
                            }}
                        >
                            <Typography variant="h4" sx={{ fontWeight: 800 }}>
                                {answerCard?.cardName}
                            </Typography>
                        </Box>

                        <Box
                            sx={{
                                width: "100%",
                                maxWidth: 420,
                                display: "flex",
                                alignItems: "center",
                                justifyContent: "center",
                                gap: 1,
                                bgcolor: "#e9f7ee",
                                color: "#1f7a3e",
                                border: "1px solid #c9ebd5",
                                borderRadius: 1,
                                py: 1,
                            }}
                        >
                            <Typography sx={{ fontWeight: 700 }}>
                                ご協力ありがとうございました！
                            </Typography>
                        </Box>

                        <Box
                            sx={{
                                width: "100%",
                                maxWidth: 460,
                                display: "grid",
                                gridTemplateColumns: "repeat(3, 1fr)",
                                bgcolor: "#f4f1f7",
                                borderRadius: 1,
                                overflow: "hidden",
                            }}
                        >
                            {[
                                ["質問数", `${questionCount}問`],
                                ["経過時間", elapsedTime],
                                ["最終自信度", finalConfidence],
                            ].map(([label, value]) => (
                                <Box
                                    key={label}
                                    sx={{
                                        py: 1.5,
                                        px: 1,
                                        borderRight: label !== "最終自信度" ? "1px solid #ded8e7" : "none",
                                    }}
                                >
                                    <Typography variant="caption" sx={{ color: "#5f5a66" }}>
                                        {label}
                                    </Typography>
                                    <Typography variant="h6" sx={{ fontWeight: 800 }}>
                                        {value}
                                    </Typography>
                                </Box>
                            ))}
                        </Box>

                        <Button
                            variant="contained"
                            size="large"
                            disabled={loading}
                            onClick={() => handleFeedback()}
                            sx={{
                                width: "100%",
                                maxWidth: 320,
                                py: 1.2,
                                bgcolor: "#6d3fa0",
                                fontWeight: 800,
                                "&:hover": {
                                    bgcolor: "#5b3387",
                                },
                            }}
                        >
                            もう一度遊ぶ
                        </Button>
                    </Stack>
                </Paper>

                <Paper
                    sx={{
                        p: 2,
                        borderRadius: 2,
                        border: "1px solid #d8d8df",
                    }}
                >
                    <Typography sx={{ fontWeight: 800, mb: 1.5 }}>
                        今回の履歴
                    </Typography>

                    <Stack
                        spacing={1.2}
                        sx={{
                            maxHeight: 340,
                            overflowY: "auto",
                            pr: 0.5,
                        }}
                    >
                        {answeredQuestions.map((item, index) => (
                            <Box
                                key={`${item.questionContent.id}-${index}`}
                                sx={{
                                    borderBottom: "1px solid #e5e2ea",
                                    pb: 1,
                                }}
                            >
                                <Typography variant="caption" sx={{ color: "#6b6570" }}>
                                    Q{index + 1}
                                </Typography>
                                <Typography variant="body2" sx={{ fontWeight: 700 }}>
                                    {item.questionContent.text}
                                </Typography>
                                <Typography variant="caption" sx={{ color: "#5f5a66" }}>
                                    {answerLabel(item.answer)}
                                </Typography>
                            </Box>
                        ))}
                    </Stack>
                </Paper>

        </Box>
    );
}

export default FeedbackPage;
