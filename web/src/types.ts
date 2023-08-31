export type Prompt = {
    prompt: string, 
    total: number, 
    remaining: number
} | undefined;

export type PlayerScore = {
    id: string,
    name: string,
    score: number
}