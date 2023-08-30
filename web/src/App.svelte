<script lang="ts">
  import { onMount } from "svelte";
  import { fade } from "svelte/transition";
  import Name from "./components/Name.svelte";
  import { store } from "./store/appStore";
  import Leaderboard from "./components/Leaderboard.svelte";

  type Message = {
    command: string;
    payload?: any;
  };

  type PokemonData = {
    name: string;
    sprite: string;
  };

  var websocket: WebSocket;

  var messages: Message[] = [];
  var msg: string;
  var isReady = false;
  $: prompt = "";
  $: correctAnswers = new Array<PokemonData>();
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

  // const parseCommand = (msg: string): { command: string; payload?: string } => {
  //   if (!msg.includes(":")) {
  //     return { command: msg };
  //   }
  //   const splits = msg.split(":");
  //   const command = splits[0].trim();
  //   const payload = splits[1].trim();

  //   return { command, payload };
  // };

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
          // const parsedPayload = JSON.parse(payload);
          const parsedPayload = payload;
          const { name, sprite } = parsedPayload;
          correctAnswers = [...correctAnswers, { name, sprite }];
        }
        break;
      case "SYS_UPDATE_SCORE":
        score = Number(payload) ?? score;
        break;
      case "SYS_NEW_PROMPT":
        prompt = payload;
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
  <Name />
  <Leaderboard />
  <h2>{prompt}</h2>
  <div>
    {#each correctAnswers as correctAnswer}
      <div class="correctAnswer">
        <img
          alt={correctAnswer.name}
          src={correctAnswer.sprite}
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
  <button
    on:click={() => handleSend({ command: "SYS_READY", payload: !isReady })}
    >{isReady ? "WAITING FOR OTHERS" : "LET'S PLAY"}</button
  >
</main>
