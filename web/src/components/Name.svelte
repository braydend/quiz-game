<script lang="ts">
  import { store } from "../store/appStore";

  var name: string;
  var editMode = false;

  store.name.subscribe((val) => {
    name = val;
  });

  const updateName = (val: string) => {
    store.name.set(val);
  };
</script>

<div>
  {#if editMode}
    <input bind:value={name} />
    <button
      on:click={() => {
        updateName(name);
        editMode = false;
      }}>Save</button
    >
  {:else}
    <span>{name}</span>
    <button
      on:click={() => {
        editMode = true;
      }}>Edit</button
    >
  {/if}
</div>
