<script lang="ts">
  import Layout from '../components/Layout.svelte';
  import Panel from '../components/Panel.svelte';
  import StringLabel from '../components/StringLabel.svelte';
  import WorkspaceNav from '../components/WorkspaceNav.svelte';
  import ComponentTable from '../components/ComponentTable.svelte';
  import { api } from '../lib/api';

  export let params = { workspace: '' };

  const workspaceId = params.workspace;
  const workspace = api.workspace(workspaceId);
  const workspaceRoute = `/workspaces/${encodeURIComponent(workspaceId)}`;
</script>

<Layout>
  <WorkspaceNav {workspaceId} active="Components" slot="navbar" />
  <Panel title="Components" backRoute={workspaceRoute}>
    <ComponentTable
      load={workspace.describeComponents}
      columns={[
        {
          title: 'id',
          component: StringLabel,
          getValue: (component) => component.id,
        },
        {
          title: 'name',
          component: StringLabel,
          getValue: (component) => component.name,
        },
        {
          title: 'type',
          component: StringLabel,
          getValue: (component) => component.type,
        },
      ]}
      actions={[
        {
          tooltip: 'Delete component',
          glyph: 'Delete',
          callback: async (component) => {
            await workspace.deleteComponent(component.id);
            window.location.reload();
          },
        },
      ]}
    />
  </Panel>
</Layout>
