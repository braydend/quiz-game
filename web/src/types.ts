export type PromptProgress = {
    total: number, 
    remaining: number
} | undefined;

export type PlayerScore = {
    id: string,
    name: string,
    score: number
}

export type Message = {
    command: string;
    payload?: any;
};