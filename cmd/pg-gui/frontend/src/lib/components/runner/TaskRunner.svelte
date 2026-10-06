<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->
<script lang="ts" module>
	import { MenuEvent } from '@bindings/gui';
	import { modal } from '@comp/base/Modal.svelte';
	import { msgAlert } from '@comp/message/AlertNotify.svelte';
	import { fileExplorer } from '@comp/navigation/FileExplorer.svelte';

	export type TaskResWrapper = {
		mode: NodeMode;
		res: TaskRes<unknown>;
	};
	type TaskID = string;
	let runningTaskId = $state(new SvelteSet<TaskID>());
	let taskResMap = $state(new SvelteMap<TaskID, TaskResWrapper[]>());
	let showBoard = $state(false);

	function genID(): TaskID {
		return crypto.randomUUID();
	}
	function runFolder(path: string) {
		const id = genID();
		runningTaskId.add(id);
		TaskRunService.RunDir(id, path)
			.catch(msgAlert.error)
			.finally(() => runningTaskId.delete(id));
	}
	export const taskRunner = {
		runFolder,
	};
</script>

<script lang="ts">
	import { NodeMode, TaskRunService, TrsEvent, type Task, type TaskRes } from '@services';
	import { Events } from '@wailsio/runtime';
	import { onMount } from 'svelte';
	import { SvelteMap, SvelteSet } from 'svelte/reactivity';

	onMount(() => {
		function pushRes<R extends TaskRes<unknown>>(mode: NodeMode, ev: Events.WailsEvent<string>) {
			const res = ev.data as R;
			const id = res.meta.id;
			if (!taskResMap.has(id)) {
				let arr = $state<TaskResWrapper[]>([]);
				taskResMap.set(id, arr);
			}
			taskResMap.get(id)?.push({ mode, res });
		}
		const unRunTask = Events.On(MenuEvent.RunTask, () => modal.open(RunTaskModal, 'Run Task'));
		const unToggleRunBoard = Events.On(MenuEvent.ToggleRunBoard, () => (showBoard = !showBoard));
		const unResErr = Events.On(TrsEvent.TrsResultErr, (ev) => pushRes(NodeMode.$zero, ev));
		const unResAny = Events.On(TrsEvent.TrsResultAny, (ev) => pushRes(NodeMode.$zero, ev));
		const unResHttp = Events.On(TrsEvent.TrsResultHttp, (ev) => pushRes(NodeMode.HttpReq, ev));

		return () => {
			unRunTask();
			unToggleRunBoard();
			unResErr();
			unResAny();
			unResHttp();
		};
	});
</script>

<div class="drawer">
	<input id="run-board-drawer-1" bind:checked={showBoard} type="checkbox" class="drawer-toggle" />
	<div class="drawer-side">
		<label for="run-board-drawer-1" aria-label="close sidebar" class="drawer-overlay"></label>
		<ul class="menu bg-base-200 min-h-full w-auto p-4">
			<!-- Sidebar content -->
			{#if taskResMap.size <= 0}
				<div class="text-xl">
					<p>No task here.</p>
				</div>
			{:else}
				{#each taskResMap.entries() as [id, taskRes] (id)}

				{/each}
			{/if}
		</ul>
	</div>
</div>
{#snippet RunTaskModal()}
	{const collections = fileExplorer.getCollections()}
	{#if collections.length <= 0}
		<div class="p-2 text-warning">
			<p>No collections opened yet.</p>
			<p>Please open collection in File Menu</p>
		</div>
	{:else}
		<div class="min-w-[33vw] p-1">
			{let collection = $state('')}
			<select bind:value={collection} class="select select-sm">
				{#each collections as c, i}
					<option value={c.path} selected={i === 0}>{c.name}</option>
				{/each}
			</select>
		</div>
	{/if}
{/snippet}
