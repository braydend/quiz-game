<script lang="ts">
  import { onMount } from "svelte";
  import { fade } from "svelte/transition";

  type Message = {
    command: string;
    payload: any;
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
  $: name = "";
  $: score = 0;

  onMount(() => {
    websocket = new WebSocket(`wss://${location.host}/ws/join`);

    websocket.onmessage = (e) => {
      const parsedMessage = JSON.parse(e.data) as Message;
      console.log(parsedMessage);
      if (parsedMessage.command.startsWith("SYS")) {
        handleSystemMessage(parsedMessage);
      }

      messages = [...messages, parsedMessage];
    };

    websocket.addEventListener("open", () => {
      websocket.send("SYS_SYNC");
    });
  });

  const handleSend = (msg: string) => {
    websocket.send(msg);
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

    switch (command) {
      case "SYS_READY":
      case "SYS_NOT_READY":
        isReady = command === "SYS_READY";
        break;
      case "SYS_UPDATE_NAME":
        console.log(payload);
        name = payload ?? name;
        break;
      case "SYS_CORRECT_ANSWER":
        if (payload) {
          // const parsedPayload = JSON.parse(payload);
          const parsedPayload = payload;
          const { name, sprite } = parsedPayload;
          console.log("correct answer:", parsedPayload);
          correctAnswers = [...correctAnswers, { name, sprite }];
        }
        break;
      case "SYS_UPDATE_SCORE":
        score = Number(payload) ?? score;
        break;
      case "SYS_NEW_PROMPT":
        prompt = payload;
        break;
      default:
        console.error(`Unknown system command: ${msg}`);
    }
  };
</script>

<main>
  <span>name: {name}</span>
  <span>score: {score}</span>
  <h2>{prompt}</h2>
  <div>
    {#each correctAnswers as correctAnswer}
      <div class="correctAnswer">
        <!-- <img
          alt="pokeball"
          src="/pokeball.png"
          width="64"
          height="64"
          out:fade={{ duration: 0, delay: 1000 }}
        /> -->
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
  <button on:click={() => handleSend(msg)}>Send</button>
  <button on:click={() => handleSend("SYS_READY")}
    >{isReady ? "WAITING FOR OTHERS" : "LET'S PLAY"}</button
  >
</main>
