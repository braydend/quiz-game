<script lang="ts">
  import Name from "./components/Name.svelte";
  import { store } from "./store/appStore";
  import Leaderboard from "./components/Leaderboard.svelte";
  import Timer from "./components/Timer.svelte";
  import RoundInfo from "./components/RoundInfo.svelte";
  import Guesser from "./components/Guesser.svelte";
  import { handleSend } from "./helpers/socket";
  import RoomControl from "./components/RoomControl.svelte";

  let isInRoom = false;
  let isCurrentlyReady = false;
  let hasPrompt = false;

  store.prompt.subscribe((p) => {
    hasPrompt = p !== "";
  });

  store.isReady.subscribe((v) => (isCurrentlyReady = v));

  store.name.subscribe((updatedName) => {
    handleSend({ command: "SYS_UPDATE_NAME", payload: updatedName });
  });
</script>

<main>
  {#if isInRoom}
    <Timer />
    <div class="container">
      <Name />
      {#if hasPrompt}
        <RoundInfo />
        <Guesser />
      {:else}
        <button
          on:click={() =>
            handleSend({ command: "SYS_READY", payload: !isCurrentlyReady })}
          >{isCurrentlyReady ? "WAITING FOR OTHERS" : "LET'S PLAY"}</button
        >
      {/if}
      <Leaderboard />
    </div>
  {:else}
    <RoomControl
      onGameJoined={() => {
        isInRoom = true;
      }}
    />
  {/if}
</main>

<style>
  .container {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    flex-wrap: wrap;
  }
</style>
