<script lang="ts">
  import { store } from "../store/appStore";

  var scores: { id: string; name: string; score: number }[];

  store.leaderboard.subscribe((updatedLeaderboard) => {
    scores = updatedLeaderboard;
    scores.sort((a, b) => b.score - a.score);
  });

  var userId: string;

  store.id.subscribe((id) => {
    userId = id;
  });
</script>

<h2>Leaderboard</h2>
<table>
  <thead>
    <th>Score</th>
    <th>Player</th>
  </thead>
  <tbody>
    {#each scores as score}
      <tr class={score.id === userId ? "currentUser" : ""}>
        <td>{score.score}</td>
        <td>{score.name}</td>
      </tr>
    {/each}
  </tbody>
</table>

<style>
  .currentUser {
    background: red;
  }
</style>
