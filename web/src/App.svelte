<script lang="ts">
  import { onMount } from "svelte";

  var websocket: WebSocket;

  var messages: string[] = [];
  var msg: string;
  var isReady = false;
  $: correctAnswers = new Array<string>();
  $: name = "";
  $: score = 0;

  onMount(() => {
    websocket = new WebSocket(`ws://${location.host}/ws/join`);

    websocket.onmessage = (e) => {
      console.log(e.data);
      if ((e.data as string).startsWith("SYS")) {
        handleSystemMessage(e.data);
      }

      messages = [...messages, e.data];
    };

    websocket.addEventListener("open", () => {
      websocket.send("SYS_SYNC");
    });
  });

  const handleSend = (msg: string) => {
    websocket.send(msg);
  };

  const parseCommand = (msg: string): { command: string; payload?: string } => {
    if (!msg.includes(":")) {
      return { command: msg };
    }
    const splits = msg.split(":");
    const command = splits[0].trim();
    const payload = splits[1].trim();

    return { command, payload };
  };

  const handleSystemMessage = (msg: string) => {
    const { command, payload } = parseCommand(msg);

    switch (command) {
      case "SYS_READY":
      case "SYS_NOT_READY":
        isReady = msg === "SYS_READY";
        break;
      case "SYS_UPDATE_NAME":
        name = payload ?? name;
        break;
      case "SYS_CORRECT_ANSWER":
        if (payload) {
          correctAnswers = [...correctAnswers, payload];
        }
        break;
      case "SYS_UPDATE_SCORE":
        score = Number(payload) ?? score;
        break;
      default:
        console.error(`Unknown system command: ${msg}`);
    }
  };
</script>

<main>
  <span>name: {name}</span>
  <span>score: {score}</span>
  <div>
    {#each correctAnswers as correctAnswer}
      <div class="correctAnswer">{correctAnswer}</div>
    {/each}
  </div>
  <ul>
    {#each messages as message}
      <li>{message}</li>
    {/each}
  </ul>
  <input bind:value={msg} type="text" />
  <button on:click={() => handleSend(msg)}>Send</button>
  <button on:click={() => handleSend("SYS_READY")}
    >{isReady ? "WAITING FOR OTHERS" : "LET'S PLAY"}</button
  >
</main>

<style>
  .correctAnswer {
  }
</style>
