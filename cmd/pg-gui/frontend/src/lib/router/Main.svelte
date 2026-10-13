<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	import TaskRunner from '@comp/runner/TaskRunner.svelte';
	import ContextMenu from '@lib/components/base/ContextMenu.svelte';
	import DynamicJoin, { NewPanel } from '@lib/components/base/DynamicJoin.svelte';
	import Modal from '@lib/components/base/Modal.svelte';
	import Editor from '@lib/components/editor/Editor.svelte';
	import AlertNotify from '@lib/components/message/AlertNotify.svelte';
	import FileExplorer from '@lib/components/navigation/FileExplorer.svelte';
	import ReqCtx from '@comp/context/ReqCtx.svelte';
	import { WML } from '@wailsio/runtime';
	import { onMount } from 'svelte';

	onMount(() => {
		WML.Reload();
	});
</script>

<div class="w-lvw h-lvh bg-base-100/50">
	<DynamicJoin panels={[NewPanel(fileExplorer, { unit: 'px', value: 170 }), NewPanel(editor)]} />
</div>
<ReqCtx />
<Modal />
<TaskRunner />
<ContextMenu />
<AlertNotify />
{#snippet fileExplorer()}
	<FileExplorer />
{/snippet}
{#snippet editor()}
	<Editor />
{/snippet}
