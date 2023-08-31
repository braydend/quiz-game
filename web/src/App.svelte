<script lang="ts">
  import { onMount } from "svelte";
  import { fade } from "svelte/transition";
  import Name from "./components/Name.svelte";
  import { store } from "./store/appStore";
  import Leaderboard from "./components/Leaderboard.svelte";
  import Timer from "./components/Timer.svelte";
  import type { Prompt } from "./types";

  type Message = {
    command: string;
    payload?: any;
  };

  let websocket: WebSocket;

  let messages: Message[] = [];
  let msg: string;
  let isReady = false;
  const correctAnswers: { [key: string]: string } = {};
  let currentPrompt: Prompt;
  $: score = 0;

  store.prompt.subscribe((p) => {
    currentPrompt = p;
  });

  onMount(() => {
    websocket = new WebSocket(`wss://${location.host}/ws/join`);

    websocket.onmessage = (e) => {
      const parsedMessage = JSON.parse(e.data) as Message;
      console.log(parsedMessage);
      if (parsedMessage.command.startsWith("SYS")) {
        handleSystemMessage(parsedMessage);
        return;
      }

      messages = [...messages, parsedMessage];
    };

    websocket.addEventListener("open", () => {
      handleSend({ command: "SYS_SYNC" });
    });
  });

  const handleSend = (msg: Message) => {
    websocket?.send(JSON.stringify(msg));
  };

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
          correctAnswers[name] = sprite;
        }
        break;
      case "SYS_UPDATE_SCORE":
        score = Number(payload) ?? score;
        break;
      case "SYS_CLEAR_PROMPT":
        store.prompt.set(undefined);
        break;
      case "SYS_PROMPT":
        store.prompt.set({
          prompt: payload.prompt,
          total: payload.totalAnswers,
          remaining: payload.remainingAnswers,
        });
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
    console.log("name updated", updatedName);
    handleSend({ command: "SYS_UPDATE_NAME", payload: updatedName });
  });
</script>

<main>
  <Timer />
  <Name />
  <Leaderboard />
  <h2>
    {#if currentPrompt}
      {currentPrompt.prompt}
      {currentPrompt.remaining}/{currentPrompt.total}
    {:else}
      Are you ready?
    {/if}
  </h2>
  <div>
    {#each Object.entries(correctAnswers) as [name, sprite]}
      <div class="correctAnswer">
        <img
          alt={name}
          src={sprite}
          width="128"
          height="128"
          in:fade={{ duration: 3000 }}
        />
      </div>
    {/each}
  </div>
  <ul>
    {#each messages as message}
      <li>{JSON.stringify(message)}</li>
    {/each}
  </ul>
  <input bind:value={msg} type="text" />
  <button on:click={() => handleSend({ command: "GUESS", payload: msg })}
    >Send</button
  >
  {#if !currentPrompt}
    <button
      on:click={() => handleSend({ command: "SYS_READY", payload: !isReady })}
      >{isReady ? "WAITING FOR OTHERS" : "LET'S PLAY"}</button
    >
  {/if}
</main>
