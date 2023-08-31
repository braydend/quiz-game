<script lang="ts">
  import { onMount } from "svelte";
  import Name from "./components/Name.svelte";
  import { store } from "./store/appStore";
  import Leaderboard from "./components/Leaderboard.svelte";
  import Timer from "./components/Timer.svelte";
  import RoundInfo from "./components/RoundInfo.svelte";
  import Guesser from "./components/Guesser.svelte";
  import { createSocket, handleSend, socket } from "./helpers/socket";
  import type { Message } from "./types";

  let messages: Message[] = [];
  let msg: string;
  let isReady = false;
  let hasPrompt = false;

  store.prompt.subscribe((p) => {
    hasPrompt = p !== "";
  });

  $: score = 0;

  onMount(() => {
    let socket = createSocket();

    socket.onmessage = (e) => {
      const parsedMessage = JSON.parse(e.data) as Message;
      console.log(parsedMessage);
      if (parsedMessage.command.startsWith("SYS")) {
        handleSystemMessage(parsedMessage);
        return;
      }

      messages = [...messages, parsedMessage];
    };

    socket.addEventListener("open", () => {
      handleSend({ command: "SYS_SYNC" });
    });
  });

  const handleSystemMessage = (msg: Message) => {
    const { command, payload } = msg;

    console.log("SYSTEM MESSAGE:", msg);

    switch (command) {
      case "SYS_READY":
        isReady = payload;
        break;
      case "SYS_UPDATE_NAME":
        store.name.set(payload);
        break;
      case "SYS_CORRECT_ANSWER":
        if (payload) {
          const parsedPayload = payload;
          const { name, sprite } = parsedPayload;
          store.correctAnswers.update((prev) => {
            prev[name] = sprite;
            return prev;
          });
        }
        break;
      case "SYS_UPDATE_SCORE":
        score = Number(payload) ?? score;
        break;
      case "SYS_CLEAR_PROMPT":
        store.prompt.set("");
        break;
      case "SYS_PROMPT_PROGRESS":
        store.promptProgress.set({
          total: payload.totalAnswers,
          remaining: payload.remainingAnswers,
        });
        break;
      case "SYS_PROMPT":
        store.prompt.set(payload);
        break;
      case "SYS_UPDATE_LEADERBOARD":
        store.leaderboard.set(payload.scores);
        break;
      case "SYS_UPDATE_USER_DATA":
        store.id.set(payload.id);
        store.name.set(payload.name);
        break;
      default:
        console.error(`Unknown system command: ${msg}`);
    }
  };

  store.name.subscribe((updatedName) => {
    handleSend({ command: "SYS_UPDATE_NAME", payload: updatedName });
  });
</script>

<main>
  <Timer />
  <div class="container">
    <Name />
    {#if hasPrompt}
      <RoundInfo />
      <Guesser />
    {:else}
      <button
        on:click={() => handleSend({ command: "SYS_READY", payload: !isReady })}
        >{isReady ? "WAITING FOR OTHERS" : "LET'S PLAY"}</button
      >
    {/if}
    <Leaderboard />
  </div>
</main>

<style>
  .container {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    flex-wrap: wrap;
  }
</style>
