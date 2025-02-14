<script lang="ts">
  import Icon from './Icon.svelte';
  import VersionInfo from './VersionInfo.svelte';
  import NavbarRoute from './nav/NavbarRoute.svelte';
  import NavbarButton from './nav/NavbarButton.svelte';
  import { theme, themeOptions } from '../lib/theme';

  $: {
    for (const option of themeOptions) {
      document.body.classList.toggle(option, $theme === option);
    }
  }
</script>

<main>
  <nav>
    <header>
      <NavbarRoute title="Home" href="#/">
        {#if import.meta.env.MODE === 'development'}
          <img src="/deref-rounded-icon-dev.png" alt="Deref" height="24px" />
        {:else}
          <img src="/deref-rounded-icon.png" alt="Deref" height="24px" />
        {/if}
      </NavbarRoute>
    </header>
    <div class="navbar-wrapper">
      <slot name="navbar" />
    </div>
    <footer>
      <NavbarRoute title="Preferences" href="#/preferences">
        <Icon glyph="Preferences" />
      </NavbarRoute>
      <NavbarButton
        title="Give feedback on GitHub"
        on:click={() => {
          window.location.href = 'https://github.com/deref/exo/discussions';
        }}
      >
        <Icon glyph="Feedback" />
      </NavbarButton>
      <VersionInfo />
    </footer>
  </nav>
  <div class="content-wrapper">
    <slot />
  </div>
</main>

<style>
  :global(body) {
    overscroll-behavior: none;
  }

  main {
    display: grid;
    grid-template-columns: max-content 1fr;
    gap: 1px;
    height: 100vh;
    overflow: hidden;
    background: var(--layout-bg-color);
  }

  nav {
    position: relative;
    width: 48px;
    height: 100vh;
    z-index: 4;
    display: grid;
    grid-template-rows: auto 1fr max-content;
    grid-auto-flow: column;
    gap: 1px;
    background: var(--nav-bg-color);
  }

  .navbar-wrapper {
    width: 48px;
    border-top: 1px solid var(--layout-bg-color);
    border-bottom: 1px solid var(--layout-bg-color);
  }

  nav ::-webkit-scrollbar {
    display: none;
  }

  .content-wrapper {
    position: relative;
    z-index: 2;
    overflow: hidden;
  }
</style>
