import { writable } from "svelte/store";

const name = writable("");
const leaderboard = writable<{id: string, name: string, score: number}[]>([])
const id = writable("");

export const store = {
    id,
    name,
    leaderboard
}