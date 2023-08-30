<script lang="ts">
  import { onMount } from "svelte";
  import { store } from "../store/appStore";

  let timerAnimater: Animation;
  const timerShrink: Keyframe[] = [
    { width: "100%", background: "green", offset: 0 },
    { background: "yellow", offset: 0.5 },
    { background: "red", offset: 0.85 },
    { background: "red", width: "0%", offset: 1 },
  ];
  const timerDuration: KeyframeAnimationOptions = {
    duration: 30 * 1000,
    iterations: 1,
    fill: "forwards",
  };

  onMount(() => {
    const timer = document.getElementById("timer");
    if (timer) {
      timerAnimater = timer.animate(timerShrink, timerDuration);
      timerAnimater.finish();
    }
  });

  store.prompt.subscribe(() => {
    if (timerAnimater) {
      timerAnimater.finish();
      timerAnimater.play();
    }
  });
</script>

<div id="timer" class="timer" />

<style>
  .timer {
    height: 30px;
  }
</style>
