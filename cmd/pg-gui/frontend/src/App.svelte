<script lang="ts">
	import ContextMenu from '@lib/components/base/ContextMenu.svelte';
	import DynamicJoin, { NewPanel } from '@lib/components/base/DynamicJoin.svelte';
	import Modal from '@lib/components/base/Modal.svelte';
	import FileExplorer from '@lib/components/navigation/FileExplorer.svelte';
	import { WML } from '@wailsio/runtime';
	import { onMount } from 'svelte';

	onMount(() => {
		// Event Example:
		// Events.On('time', (v: any) => {
		//   const compact = (full.match(/\d{1,2}:\d{2}:\d{2}/) || [full])[0];
		//   time = window.matchMedia('(max-width: 640px)').matches ? compact : full;
		// });
		// Wire up data-wml-openURL links (logos + footer "Docs" link).
		WML.Reload();
	});
</script>

<div class="w-lvw h-lvh bg-base-100/50">
	<DynamicJoin panels={[NewPanel(fileExplorer, { unit: 'px', value: 100 }), NewPanel(editor)]} />
</div>
<Modal />
<ContextMenu />

{#snippet fileExplorer()}
	<FileExplorer />
{/snippet}
<!-- TODO: complete the editor's page -->
{#snippet editor()}
	<div class="w-full flex justify-center items-center">
		<div>EDITOR PAGE</div>
	</div>
{/snippet}
