<script lang="ts">
  import Code from '../components/Code.svelte';
  import Layout from '../components/Layout.svelte';
  import ErrorLabel from '../components/ErrorLabel.svelte';
  import WorkspaceList from '../components/WorkspaceList.svelte';
  import CenterFormPanel from '../components/form/CenterFormPanel.svelte';
  import { api } from '../lib/api';

  const workspaces = api.kernel
    .describeWorkspaces()
    .then((workspaces) =>
      workspaces.sort((w1, w2) => w1.root.localeCompare(w2.root)),
    );
</script>

<Layout>
  <CenterFormPanel title="Workspaces">
    <h1>Workspaces</h1>
    <div>
      {#await workspaces}
        loading workspaces...
      {:then workspaces}
        <WorkspaceList {workspaces} />
      {:catch error}
        <ErrorLabel value={error} />
      {/await}
    </div>
    <hr />
    <div>
      Use <Code>exo gui</Code> in your terminal to launch into the current directory's
      workspace.
    </div>
  </CenterFormPanel>
</Layout>
