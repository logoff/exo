<script lang="ts">
  export let inline = false;
</script>

<div class:inline />

<style>
  div {
    width: 16px;
    height: 16px;
    border-radius: 100%;
    animation: spin 0.667s infinite linear;
    border: 2px solid var(--spinner-grey);
    border-bottom-color: var(--spinner-grey-light);
    transition: all 0.125s;
  }

  .inline {
    display: inline-block;
  }

  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }
</style>
