import { writable } from "svelte/store";

const name = writable("");

export const store = {
    name
}