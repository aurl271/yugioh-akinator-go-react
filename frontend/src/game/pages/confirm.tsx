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
import { formatProbability, probabilityToPercentValue } from "../utils/formatProbability";

function ConfirmPage() {
    const navigate = useNavigate();
    // answerCardはGameRootが回答候補に到達したときだけ設定する。
    const {
        state: { loading,meta,candidates,answerCard },
        actions: { handleConfirm },
    } = useOutletContext<GameOutletContext>();

    // 確認画面に必要な状態が無ければ、途中遷移と判断してsetupへ戻す。
    useEffect(() => {
        if (meta === null || answerCard === null) {
            navigate("/game/setup");
        }
    }, [meta, answerCard, navigate]);

    if (meta === null || answerCard === null) {
        return null;
    }

    // candidatesの先頭を最終候補として扱い、自信度表示へ変換する。
    const topProbability = candidates[0]?.probability ?? null;
    const confidenceText = formatProbability(topProbability);
    const confidenceValue = probabilityToPercentValue(topProbability);
    

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
                            width: { xs: 180, md: 270 },
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
                            variant="h5"
                            component="h1"
                            sx={{ fontWeight: 800 }}
                        >
                            あなたの考えたカードはこれですか？
                        </Typography>

                        <Box
                            sx={{
                                width: "100%",
                                maxWidth: 360,
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

                        <Box sx={{ width: "100%", maxWidth: 340 }}>
                            <Typography
                                sx={{
                                    mb: 1,
                                    fontWeight: 700,
                                    textAlign: "left",
                                }}
                            >
                                自信度: {confidenceText}
                            </Typography>
                            <LinearProgress
                                variant="determinate"
                                value={confidenceValue}
                                sx={{
                                    height: 10,
                                    borderRadius: 999,
                                    bgcolor: "#ebe8ef",
                                    "& .MuiLinearProgress-bar": {
                                        bgcolor: "#6d3fa0",
                                    },
                                }}
                            />
                        </Box>

                        <Stack
                            direction={{ xs: "column", sm: "row" }}
                            spacing={1.5}
                            sx={{
                                width: "100%",
                                maxWidth: 340,
                            }}
                        >
                            <Button
                                variant="contained"
                                size="large"
                                disabled={loading}
                                onClick={() => handleConfirm(true)}
                                sx={{
                                    flex: 1,
                                    py: 1.1,
                                    bgcolor: "#6d3fa0",
                                    fontWeight: 800,
                                    "&:hover": {
                                        bgcolor: "#5b3387",
                                    },
                                }}
                            >
                                はい
                            </Button>
                            <Button
                                variant="outlined"
                                size="large"
                                disabled={loading}
                                onClick={() => handleConfirm(false)}
                                sx={{
                                    flex: 1,
                                    py: 1.1,
                                    color: "#6d3fa0",
                                    borderColor: "#b99ad6",
                                    fontWeight: 800,
                                    "&:hover": {
                                        borderColor: "#6d3fa0",
                                        bgcolor: "#f7f3fb",
                                    },
                                }}
                            >
                                いいえ
                            </Button>
                        </Stack>
                    </Stack>
                </Paper>

                <Paper
                    sx={{
                        p: 2,
                        borderRadius: 2,
                        border: "1px solid #d8d8df",
                        order: { xs: 3, lg: 2 },
                    }}
                >
                    <Typography sx={{ fontWeight: 800, mb: 1.5 }}>
                        候補
                    </Typography>

                    <Stack spacing={1.3}>
                        {candidates.map((candidate) => (
                            <Box
                                key={candidate.rank}
                                sx={{
                                    display: "grid",
                                    gridTemplateColumns: "24px 1fr auto",
                                    gap: 1,
                                    alignItems: "center",
                                    bgcolor: candidate.rank === 1 ? "#eadff3" : "transparent",
                                    borderRadius: 1,
                                    px: candidate.rank === 1 ? 1 : 0,
                                    py: candidate.rank === 1 ? 0.8 : 0,
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

        </Box>
    );
}

export default ConfirmPage;
