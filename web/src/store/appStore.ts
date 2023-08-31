import { writable } from "svelte/store";
import type { PlayerScore, PromptProgress } from "../types";

const name = writable("");
const leaderboard = writable<PlayerScore[]>([])
const id = writable("");
const prompt = writable("")
const promptProgress = writable<PromptProgress>()

export const store = {
    id,
    name,
    leaderboard,
    prompt,
    promptProgress
}