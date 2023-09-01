<script lang="ts">
  import { createSocketFromId } from "../helpers/socket";

  export let onGameJoined: () => void;

  let gameId = "";
  let checkGame: Promise<{ exists: boolean }> | undefined = undefined;
  let isLoading = false;

  const joinRoom = async (roomId: string) => {
    isLoading = true;
    checkGame = fetch(`https://${location.host}/games/${roomId}`).then((r) => {
      return r.json() as unknown as { exists: boolean };
    });

    const result = await checkGame;

    if (result.exists) {
      createSocketFromId(gameId);
      onGameJoined();
    }
    isLoading = false;
  };
</script>

<div>
  <label for="roomId">Room ID:</label>
  <input bind:value={gameId} disabled={isLoading} name="roomId" />
  <button
    on:click={() => {
      joinRoom(gameId);
    }}>Join</button
  >
  {#if checkGame}
    <div class="roomResponses">
      {#await checkGame}
        Finding game...
      {:then isGameValid}
        {#if !isGameValid.exists}
          This game does not exist
        {:else}
          Joining...
        {/if}
      {:catch e}
        <span>{e}</span>
      {/await}
    </div>
  {/if}
</div>

<style>
  .roomResponses {
    display: flex;
    flex-direction: column;
  }
</style>
