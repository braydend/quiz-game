<script lang="ts">
  import { onMount } from "svelte";
  import Name from "./components/Name.svelte";
  import { store } from "./store/appStore";
  import Leaderboard from "./components/Leaderboard.svelte";
  import Timer from "./components/Timer.svelte";
  import RoundInfo from "./components/RoundInfo.svelte";

  type Message = {
    command: string;
    payload?: any;
  };

  let websocket: WebSocket;

  let messages: Message[] = [];
  let msg: string;
  let isReady = false;
  let hasPrompt = false;

  store.prompt.subscribe((p) => {
    hasPrompt = p !== "";
  });

  $: score = 0;

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
  <Name />
  <Leaderboard />
  <RoundInfo />
  <input bind:value={msg} type="text" />
  <button on:click={() => handleSend({ command: "GUESS", payload: msg })}
    >Send</button
  >
  {#if !hasPrompt}
    <button
      on:click={() => handleSend({ command: "SYS_READY", payload: !isReady })}
      >{isReady ? "WAITING FOR OTHERS" : "LET'S PLAY"}</button
    >
  {/if}
</main>
