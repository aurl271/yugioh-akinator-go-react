import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";

// StepProgressProps.currentStep は1始まりで現在の画面位置を表す。
type StepProgressProps = {
    currentStep?: number;
};

// steps はsetup/question/confirm/feedbackの画面順と対応する表示名。
const steps = ["設定", "質問", "確認", "結果"];

function StepProgress({ currentStep = 1 }: StepProgressProps) {
    // totalSteps は下部の「現在/全体」表示に使う。
    const totalSteps = steps.length;

    return (
        <Box
            sx={{
                width: "100%",
                maxWidth: 860,
                mx: "auto",
                mb: 1.5,
                px: { xs: 1, md: 0 },
            }}
        >
            <Box
                sx={{
                    position: "relative",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "space-between",
                    "&::before": {
                        content: '""',
                        position: "absolute",
                        top: 14,
                        left: "12%",
                        right: "12%",
                        height: 2,
                        bgcolor: "#d8d4df",
                        zIndex: 0,
                    },
                }}
            >
                {steps.map((label, index) => {
                    // stepNumberは配列indexをユーザー向けの1始まり番号へ変換したもの。
                    const stepNumber = index + 1;
                    const isActive = stepNumber <= currentStep;
                    const isCurrent = stepNumber === currentStep;

                    return (
                        <Box
                            key={label}
                            sx={{
                                position: "relative",
                                zIndex: 1,
                                display: "flex",
                                flexDirection: "column",
                                alignItems: "center",
                                minWidth: 52,
                            }}
                        >
                            <Box
                                sx={{
                                    width: 28,
                                    height: 28,
                                    borderRadius: "50%",
                                    display: "flex",
                                    alignItems: "center",
                                    justifyContent: "center",
                                    bgcolor: isActive ? "#6d3fa0" : "white",
                                    color: isActive ? "white" : "#6b6570",
                                    border: isActive ? "none" : "2px solid #d8d4df",
                                    boxShadow: isCurrent ? "0 0 0 3px rgba(109, 63, 160, 0.16)" : "none",
                                    fontWeight: 800,
                                    fontSize: 13,
                                }}
                            >
                                {stepNumber}
                            </Box>
                            <Typography
                                variant="caption"
                                sx={{
                                    mt: 0.5,
                                    color: isActive ? "#3a274d" : "#6b6570",
                                    fontWeight: isActive ? 800 : 600,
                                    fontSize: 11,
                                }}
                            >
                                {label}
                            </Typography>
                        </Box>
                    );
                })}
            </Box>
            <Typography
                variant="caption"
                sx={{
                    display: "block",
                    mt: 0.5,
                    textAlign: "center",
                    color: "#5f5a66",
                    fontWeight: 700,
                    fontSize: 11,
                }}
            >
                {currentStep} / {totalSteps}
            </Typography>
        </Box>
    );
}

export default StepProgress;
