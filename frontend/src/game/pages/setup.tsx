import Box from '@mui/material/Box';
import Button from "@mui/material/Button";
import Paper from "@mui/material/Paper";
import Radio from "@mui/material/Radio";
import RadioGroup from "@mui/material/RadioGroup";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import { useState } from "react";
import lampMajin from "../../assets/lamp_majin.png"
import { useOutletContext } from "react-router-dom";
import type { GameOutletContext } from "../GameRoot";
import type { Hyperparameters } from "../../type/game";
import { DEFAULT_TOP_CANDIDATES_COUNT } from "../../type/game";

// DifficultyMode は画面に表示する難易度と、backendへ渡す推理設定をまとめた型。
type DifficultyMode = {
    id: string;
    title: string;
    description: string;
    meta: Hyperparameters;
};

// difficultyModes はユーザーが選べる推理モード。
// betaとanswerThresholdの組み合わせで、回答ミスへの強さを変える。
const difficultyModes: DifficultyMode[] = [
    {
        id: "forgiving",
        title: "ミスを許容しやすい",
        description: "質問は多めになりますが、途中で少し間違えても立て直しやすい設定です。",
        meta: {
            beta: 1,
            answerThreshold: 0.9999,
            topCandidatesCount: DEFAULT_TOP_CANDIDATES_COUNT,
        },
    },
    {
        id: "strict",
        title: "ミスを許容しない",
        description: "少ない質問で絞り込みやすい一方、回答ミスの影響が大きい設定です。",
        meta: {
            beta: 100,
            answerThreshold: 0.9,
            topCandidatesCount: DEFAULT_TOP_CANDIDATES_COUNT,
        },
    },
];

function SetupPage() {
    // selectedModeId はラジオボタンで選択中の難易度ID。
    const [selectedModeId, setSelectedModeId] = useState(difficultyModes[0].id)
    const {
        state: { loading },
        actions: { handleStartGame },
    } = useOutletContext<GameOutletContext>();
    // selectedMode は開始ボタンを押したときに送信する現在の設定。
    const selectedMode = difficultyModes.find((mode) => mode.id === selectedModeId) ?? difficultyModes[0];

    return (
        <Paper
            sx={{
                maxWidth: 860,
                mx: "auto",
                p: { xs: 3, md: 5 },
                borderRadius: 2,
                display: "flex",
                flexDirection: { xs: "column", md: "row" },
                alignItems: "center",
                gap: { xs: 3, md: 6 },
            }}
        >
                <Box
                    sx={{
                        width: { xs: "100%", md: 300 },
                        flexShrink: 0,
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "center",
                    }}
                >
                    <Box
                        component="img"
                        src={lampMajin}
                        alt="ランプの魔人"
                        sx={{
                            width: { xs: 190, md: 270 },
                            height: "auto",
                        }}
                    />
                </Box>

                <Stack
                    spacing={3}
                    sx={{
                        flex: 1,
                        width: "100%",
                        minWidth: 0,
                        alignItems: "center",
                    }}
                >
                    <Stack spacing={1} sx={{ textAlign: "center" }}>
                        <Typography
                            variant="h4"
                            component="h1"
                            sx={{
                                fontWeight: 800,
                                color: "#1b1720",
                            }}
                        >
                            推理の強さを設定
                        </Typography>
                        <Typography variant="body2" sx={{ color: "#5f5a66" }}>
                            回答ミスへの強さを選びます。
                        </Typography>
                    </Stack>

                    <RadioGroup
                        value={selectedModeId}
                        onChange={(event) => setSelectedModeId(event.target.value)}
                        sx={{
                            width: "100%",
                            maxWidth: 460,
                            gap: 1.4,
                        }}
                    >
                        {difficultyModes.map((mode) => {
                            const checked = mode.id === selectedModeId;

                            return (
                                <Paper
                                    key={mode.id}
                                    variant="outlined"
                                    onClick={() => setSelectedModeId(mode.id)}
                                    sx={{
                                        borderRadius: 2,
                                        borderColor: checked ? "#6d3fa0" : "#d8d8df",
                                        bgcolor: checked ? "#f7f3fb" : "#ffffff",
                                        cursor: "pointer",
                                        width: "100%",
                                    }}
                                >
                                    <Box
                                        sx={{
                                            display: "flex",
                                            alignItems: "flex-start",
                                            gap: 1,
                                            width: "100%",
                                            minWidth: 0,
                                            px: 1.4,
                                            py: 1.2,
                                            boxSizing: "border-box",
                                        }}
                                    >
                                        <Radio
                                            value={mode.id}
                                            checked={checked}
                                            onChange={() => setSelectedModeId(mode.id)}
                                            sx={{
                                                color: "#8d8497",
                                                flexShrink: 0,
                                                p: 0.5,
                                                "&.Mui-checked": {
                                                    color: "#6d3fa0",
                                                },
                                            }}
                                        />
                                        <Stack
                                            spacing={0.4}
                                            sx={{
                                                minWidth: 0,
                                                flex: 1,
                                            }}
                                        >
                                            <Typography
                                                sx={{
                                                    fontWeight: 800,
                                                    color: "#1b1720",
                                                    overflowWrap: "break-word",
                                                }}
                                            >
                                                {mode.title}
                                            </Typography>
                                            <Typography
                                                variant="body2"
                                                sx={{
                                                    color: "#5f5a66",
                                                    lineHeight: 1.6,
                                                    whiteSpace: "normal",
                                                    overflowWrap: "break-word",
                                                    wordBreak: "break-word",
                                                }}
                                            >
                                                {mode.description}
                                            </Typography>
                                        </Stack>
                                    </Box>
                                </Paper>
                            );
                        })}
                    </RadioGroup>

                    <Box
                        sx={{
                            width: "100%",
                            maxWidth: 420,
                            bgcolor: "#f7f3fb",
                            border: "1px solid #e4d8ee",
                            borderRadius: 1,
                            px: 2,
                            py: 1.5,
                            color: "#4b4356",
                        }}
                    >
                        <Typography variant="body2">
                            ミスを許容しやすい設定では、確信するまで慎重に質問します。
                            <br />
                            ミスを許容しない設定では、回答を強く信じて早めに絞り込みます。
                        </Typography>
                    </Box>

                    <Button
                        variant="contained"
                        size="large"
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
                        disabled={loading}
                        onClick={() => handleStartGame(selectedMode.meta)}
                    >
                        ゲームを開始
                    </Button>
                </Stack>
        </Paper>
    )
}

export default SetupPage
