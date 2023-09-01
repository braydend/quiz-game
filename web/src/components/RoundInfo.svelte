<script lang="ts">
  import { store } from "../store/appStore";
  import type { PromptProgress } from "../types";
  import { fade } from "svelte/transition";

  let currentCorrectAnswers: { [key: string]: string } = {};
  let currentPrompt: string;
  let currentPromptProgress: PromptProgress;

  store.correctAnswers.subscribe((a) => {
    currentCorrectAnswers = a;
  });

  store.prompt.subscribe((p) => {
    currentPrompt = p;
    store.correctAnswers.set({});
  });
  store.promptProgress.subscribe((p) => {
    currentPromptProgress = p;
  });

  const getRandomPosition = () => {
    const max = 30;
    const min = 0;
    return Math.random() * (max - min) + min;
  };
</script>

<h2>
  {#if currentPrompt && currentPromptProgress}
    {currentPrompt}
    {currentPromptProgress.remaining}/{currentPromptProgress.total}
  {:else}
    Are you ready?
  {/if}
</h2>
<div class="correctAnswers">
  {#each Object.entries(currentCorrectAnswers) as [name, sprite], index}
    <img
      class="sprite"
      alt={name}
      src={sprite}
      width="128"
      height="128"
      in:fade={{ duration: 3000 }}
      style={`z-index: ${
        index + 1
      }; margin-top: ${getRandomPosition()}vw; margin-left: ${getRandomPosition()}vw`}
    />
  {/each}
</div>

<style>
  .sprite {
    width: 128px;
    height: 128px;
    position: absolute;
  }

  .correctAnswers {
    height: 45vw;
    width: 90vw;
    background-image: url("/pokeball.svg");
    background-repeat: no-repeat;
    background-repeat: no-repeat;
    background-position: center;
    margin: 0 auto;
    border: 4px solid black;
    border-radius: 4px;
    display: flex;
  }
</style>
