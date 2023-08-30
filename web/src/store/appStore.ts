import { writable } from "svelte/store";
import type { PlayerScore, Prompt } from "../types";

const name = writable("");
const leaderboard = writable<PlayerScore[]>([])
const id = writable("");
const prompt = writable<Prompt>()

export const store = {
    id,
    name,
    leaderboard,
    prompt
}