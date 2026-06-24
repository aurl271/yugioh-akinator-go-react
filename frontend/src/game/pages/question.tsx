import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import LinearProgress from "@mui/material/LinearProgress";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import lampMajin from "../../assets/lamp_majin.png";
import { useEffect } from "react";
import { useNavigate, useOutletContext } from "react-router-dom";
import type { GameOutletContext } from "../GameRoot";
import type { AnswerValue } from "../../type/game";
import { answerLabel } from "../utils/answerLabel";
import { formatProbability } from "../utils/formatProbability";

// answerButtons は5段階回答の表示名・記号・APIへ送る値をまとめた定義。
const answerButtons: { label: string; mark: string; color : string; value: AnswerValue }[] = [
    { label: "はい", mark: "〇", color: "#2f9b73", value:"yes" },
    { label: "たぶんはい", mark: "△", color: "#69a84f", value:"probably"  },
    { label: "わからない", mark: "?", color: "#7f8790", value:"unknown"  },
    { label: "たぶんいいえ", mark: "△", color: "#f39a23", value:"probably_no"  },
    { label: "いいえ", mark: "×", color: "#d23b3b", value:"no"  },
];

function QuestionPage() {
    const navigate = useNavigate();

    // GameRootが持つ状態と操作関数をOutlet経由で受け取る。
    const {
        state: { loading,meta,question,candidates,answeredQuestions,progressPercent},
        actions: { handleQuestion, handleBackQuestion, handleResetGame },
    } = useOutletContext<GameOutletContext>();

    // URL直打ちなどで必要な状態が無い場合はsetupへ戻す。
    useEffect(() => {
        if (meta === null || question === null) {
            navigate("/game/setup");
        }
    }, [meta, question, navigate]);

    // setupへ戻す途中に、未初期化状態の画面を描画しない。
    if (meta === null || question === null) {
        return null;
    }

    return (
        <Box
            sx={{
                display: "grid",
                gridTemplateColumns: { xs: "1fr", lg: "240px 1fr 280px" },
                gap: 2,
                alignItems: "start",
            }}
        >
                <Paper
                    sx={{
                        p: 2,
                        borderRadius: 2,
                        order: { xs: 4, lg: 1 },
                        gridColumn: { lg: 1 },
                        gridRow: { lg: "1 / span 2" },
                    }}
                >
                    <Typography sx={{ fontWeight: 800, mb: 1.5 }}>
                        直近の回答履歴
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

                <Paper
                    sx={{
                        p: { xs: 3, md: 4 },
                        borderRadius: 2,
                        display: "flex",
                        flexDirection: { xs: "column", md: "row" },
                        gap: { xs: 3, md: 4 },
                        alignItems: "center",
                        order: { xs: 2, lg: 2 },
                        gridColumn: { lg: 2 },
                        gridRow: { lg: 2 },
                    }}
                >
                    <Box
                        component="img"
                        src={lampMajin}
                        alt="ランプの魔人"
                        sx={{
                            width: { xs: 120, md: 220 },
                            height: "auto",
                            flexShrink: 0,
                        }}
                    />

                    <Stack
                        spacing={2}
                        sx={{
                            flex: 1,
                            width: "100%",
                            alignItems: "center",
                        }}
                    >
                        <Typography
                            variant="h5"
                            component="h1"
                            sx={{
                                fontWeight: 800,
                                textAlign: "center",
                            }}
                        >
                            {question?.text}
                        </Typography>

                        <Stack spacing={1.2} sx={{ width: "100%", maxWidth: 360 }}>
                            {answerButtons.map((answer) => (
                                <Button
                                    key={answer.label}
                                    variant="outlined"
                                    disabled={loading}
                                    onClick = {() => handleQuestion(answer.value)}
                                    sx={{
                                        position: "relative",
                                        py: 1,
                                        color: "#1b1720",
                                        borderColor: "#d8d8df",
                                        bgcolor: "white",
                                        fontWeight: 700,
                                        "&:hover": {
                                            borderColor: answer.color,
                                            bgcolor: "#fbf9fd",
                                        },
                                    }}
                                >
                                    <Box
                                        component="span"
                                        sx={{
                                            position: "absolute",
                                            left: 16,
                                            width: 22,
                                            height: 22,
                                            borderRadius: "50%",
                                            display: "flex",
                                            alignItems: "center",
                                            justifyContent: "center",
                                            border: `2px solid ${answer.color}`,
                                            color: answer.color,
                                            fontSize: 14,
                                            fontWeight: 800,
                                            lineHeight: 1,
                                        }}
                                    >
                                        {answer.mark}
                                    </Box>
                                    {answer.label}
                                </Button>
                            ))}
                        </Stack>

                        <Stack
                            direction={{ xs: "column", sm: "row" }}
                            spacing={1.2}
                            sx={{
                                width: "100%",
                                maxWidth: 360,
                            }}
                        >
                            <Button
                                variant="outlined"
                                disabled={loading || answeredQuestions.length === 0}
                                onClick={() => handleBackQuestion()}
                                sx={{
                                    flex: 1,
                                    color: "#4b5563",
                                    borderColor: "#9ca3af",
                                    bgcolor: "#f9fafb",
                                    fontWeight: 800,
                                    "&:hover": {
                                        borderColor: "#6b7280",
                                        bgcolor: "#f3f4f6",
                                    },
                                }}
                            >
                                1つ戻る
                            </Button>
                            <Button
                                variant="outlined"
                                disabled={loading || answeredQuestions.length === 0}
                                onClick={() => handleResetGame()}
                                sx={{
                                    flex: 1,
                                    color: "#8a3b2f",
                                    borderColor: "#d69a8d",
                                    bgcolor: "#fff7f5",
                                    fontWeight: 800,
                                    "&:hover": {
                                        borderColor: "#b85f4f",
                                        bgcolor: "#ffefeb",
                                    },
                                }}
                            >
                                履歴リセット
                            </Button>
                        </Stack>
                    </Stack>
                </Paper>

                <Paper
                    sx={{
                        p: 2,
                        borderRadius: 2,
                        border: "1px solid #d8d8df",
                        order: { xs: 3, lg: 3 },
                        gridColumn: { lg: 3 },
                        gridRow: { lg: "1 / span 2" },
                    }}
                >
                    <Typography sx={{ fontWeight: 800, mb: 1.5 }}>
                        候補
                    </Typography>

                    <Stack spacing={1.1}>
                        {candidates.map((candidate) => (
                            <Box
                                key={candidate.rank}
                                sx={{
                                    display: "grid",
                                    gridTemplateColumns: "24px 1fr auto",
                                    gap: 1,
                                    alignItems: "center",
                                }}
                            >
                                <Typography variant="body2" sx={{ fontWeight: 800 }}>
                                    {candidate.rank}
                                </Typography>
                                <Typography variant="body2">{candidate.cardName}</Typography>
                                <Typography variant="body2" sx={{ fontVariantNumeric: "tabular-nums" }}>
                                    {formatProbability(candidate.probability)}
                                </Typography>
                            </Box>
                        ))}
                    </Stack>
                </Paper>

                <Paper
                    sx={{
                        gridColumn: { xs: "1", lg: "2" },
                        gridRow: { lg: 1 },
                        p: 2,
                        borderRadius: 2,
                        border: "1px solid #d8d8df",
                        order: { xs: 1, lg: 1 },
                    }}
                >
                    <Box
                        sx={{
                            display: "flex",
                            alignItems: "center",
                            gap: 2,
                        }}
                    >
                        <Typography variant="body2" sx={{ fontWeight: 700 }}>
                            推理進行度 {progressPercent}%
                        </Typography>
                        <LinearProgress
                            variant="determinate"
                            value={progressPercent}
                            sx={{
                                flex: 1,
                                height: 8,
                                borderRadius: 999,
                                bgcolor: "#ebe8ef",
                                "& .MuiLinearProgress-bar": {
                                    bgcolor: "#6d3fa0",
                                },
                            }}
                        />
                    </Box>
                </Paper>
        </Box>
    );
}

export default QuestionPage;
