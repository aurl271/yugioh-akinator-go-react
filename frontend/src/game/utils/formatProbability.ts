export function formatProbability(probability: number | null | undefined): string {
    if (probability == null) {
        return "--";
    }

    return `${(probability * 100).toFixed(1)}%`;
}

export function probabilityToPercentValue(probability: number | null | undefined): number {
    if (probability == null) {
        return 0;
    }

    return probability * 100;
}
