<script lang="ts">
  import { handleSend } from "../helpers/socket";

  let pokemon: { name: string }[] = [];
  let query: string;
  const endpoint = "https://pokeapi.co/api/v2/pokemon?limit=1000";

  fetch(endpoint)
    .then((r) => r.json())
    .then((d) => {
      pokemon = d.results;
    });

  $: filteredResults =
    query === ""
      ? []
      : pokemon.filter(({ name }) => name.includes(query)).slice(0, 10);

  const handleGuess = (guess: string) => {
    handleSend({ command: "GUESS", payload: guess });
    query = "";
  };
</script>

<div>
  {#if !pokemon}
    <span>Loading...</span>
  {:else}
    <input bind:value={query} />
    {#each filteredResults as p}
      <div class="option" on:click={() => handleGuess(p.name)}>
        {p.name}
      </div>
    {/each}
  {/if}
</div>

<style>
  .option {
    margin: 1rem;
    padding: 0.5rem;
    border: 1px solid green;
  }
</style>
