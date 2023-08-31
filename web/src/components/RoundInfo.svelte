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
</script>

<h2>
  {#if currentPrompt && currentPromptProgress}
    {currentPrompt}
    {currentPromptProgress.remaining}/{currentPromptProgress.total}
  {:else}
    Are you ready?
  {/if}
</h2>
<div>
  {#each Object.entries(currentCorrectAnswers) as [name, sprite]}
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
