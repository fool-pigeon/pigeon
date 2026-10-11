<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->
<script lang="ts" module>
	// Icons
	import IconRefreshCw from '@lucide/svelte/icons/refresh-cw';
	import IconCircleCheck from '@lucide/svelte/icons/circle-check';
	import IconCircleX from '@lucide/svelte/icons/circle-x';
	import IconTrash from '@lucide/svelte/icons/trash';
	import IconPause from '@lucide/svelte/icons/pause';
	import IconRotateCcwClock from '@lucide/svelte/icons/rotate-ccw-clock';

	import { MenuEvent } from '@pg/gui';
	import { msgAlert } from '@comp/message/AlertNotify.svelte';
	import { fileExplorer } from '@comp/navigation/FileExplorer.svelte';

	export type TaskInfo = {
		collection: FileMeta;
		meta: TaskMeta;
	};
	export type TaskResWrapper = {
		mode: NodeMode;
		res: TaskRes<unknown>;
	};
	type TaskID = string;
	let runningTaskId = new SvelteSet<TaskID>();
	let taskInfo = new SvelteMap<TaskID, TaskInfo>();
	let taskResMap = new SvelteMap<TaskID, TaskResWrapper[]>();
	let isShow = $state(false);
	let selectedId = $state<string>();

	function genID(): TaskID {
		return crypto.randomUUID();
	}
	function showBoard() {
		isShow = true;
	}
	function selectTask(id: string) {
		if (taskInfo.has(id)) {
			selectedId = id;
		} else {
			msgAlert.warn(`not exist task`);
		}
	}
	function runFolder(path: string) {
		const id = genID();
		runningTaskId.add(id);
		msgAlert.info(`run ${path} ${id}`);
		TaskRunService.RunDir(id, path)
			.then((task) => {
				const col = fileExplorer.getCollections().find((c) => assertPath.contains(c.path, path));
				pushTaskInfo({
					collection: { name: col?.name ?? '', path: col?.path ?? '' },
					meta: { id: id, path: task.param.root },
				});
				showBoard();
				selectTask(id);
			})
			.catch(msgAlert.error)
			.finally(() => runningTaskId.delete(id));
	}
	function pushTaskInfo(info: TaskInfo) {
		taskInfo.set(info.meta.id, info);
	}
	export const taskRunner = {
		genID,
		showBoard,
		selectTask,
		runFolder,
		pushTaskInfo,
	};

	/**
	 * TaskRunnerEvent
	 */
	const TaskEvent = {
		/**
		 * send data:
		 * id:string
		 * this will remove the task in list
		 */
		Delete: 'task-runner:delete',
		/**
		 * send data:
		 * id:string
		 * restart the task
		 */
		Restart: 'task-runner:restart',
		/**
		 * send data:
		 * id:string
		 * stop the task
		 */
		Stop: 'task-runner:stop',
	} as const;
	type TaskEvent = (typeof TaskEvent)[keyof typeof TaskEvent];
</script>

<script lang="ts">
	import {
		NodeMode,
		TaskRunService,
		TrsEvent,
		type FileMeta,
		type TaskMeta,
		type TaskRes,
		type TaskResHttp,
	} from '@services';
	import { Events } from '@wailsio/runtime';
	import { onMount } from 'svelte';
	import { SvelteMap, SvelteSet } from 'svelte/reactivity';
	import TaskViewer from './TaskViewer.svelte';
	import { Duration } from '@bindings/time';
	import { assertPath } from '@lib/utils/asserts';
	import { contextmenu } from '@comp/base/ContextMenu.svelte';
	import TaskRunModal from './TaskRunModal.svelte';

	onMount(() => {
		function pushRes<R extends TaskRes<unknown>>(mode: NodeMode, ev: Events.WailsEvent<string>) {
			const res = ev.data as R;
			const id = res.meta.id;
			const w = { mode, res };
			if (!taskResMap.has(id)) {
				let arr = $state<TaskResWrapper[]>([w]);
				taskResMap.set(id, arr);
			} else {
				taskResMap.get(id)!.push(w);
			}
			// taskResMap.set(id, [...(taskResMap.get(id) ?? []), w]);
		}
		const unToggleRunBoard = Events.On(MenuEvent.ToggleRunBoard, () => (isShow = !isShow));
		const unResErr = Events.On(TrsEvent.TrsResultErr, (ev) => pushRes(NodeMode.$zero, ev));
		const unResAny = Events.On(TrsEvent.TrsResultAny, (ev) => pushRes(NodeMode.$zero, ev));
		const unResHttp = Events.On(TrsEvent.TrsResultHttp, (ev) => {
			pushRes(NodeMode.FileMode | NodeMode.HttpReq, ev);
		});
		const unCancelld = Events.On(TrsEvent.TrsCancelld, (ev) => {
			const id = ev.data as string;
			if (runningTaskId.has(id)) {
				runningTaskId.delete(id);
			}
		});

		const unStop = Events.On(TaskEvent.Stop, (ev) => {
			const id = ev.data as string;
			if (!runningTaskId.has(id)) {
				msgAlert.warn('task is not running');
				return;
			}
			Events.Emit(TrsEvent.TrsCancel, id);
		});
		const unRestart = Events.On(TaskEvent.Restart, (ev) => {
			const id = ev.data as string;
			// TODO: handle restart
		});
		const unDelete = Events.On(TaskEvent.Delete, (ev) => {
			const id = ev.data as string;
			if (runningTaskId.has(id)) {
				Events.Emit(TrsEvent.TrsCancel, id);
				const ungister = Events.On(TrsEvent.TrsCancelld, (cevt) => {
					const cancelledId = cevt.data as string;
					if (cancelledId !== id) return;
					taskInfo.delete(id);
					taskResMap.delete(id);
					ungister();
				});
			} else {
				taskInfo.delete(id);
				taskResMap.delete(id);
			}
		});

		return () => {
			unToggleRunBoard();
			unResErr();
			unResAny();
			unResHttp();
			unCancelld();
			unStop();
			unRestart();
			unDelete();
		};
	});
</script>

<TaskRunModal />

<div class="drawer drawer-end">
	<input id="run-board-drawer-1" bind:checked={isShow} type="checkbox" class="drawer-toggle" />
	<div class="drawer-side">
		<label for="run-board-drawer-1" aria-label="close sidebar" class="drawer-overlay"></label>
		{@render DrawerContent()}
	</div>
</div>

{#snippet DrawerContent()}
	<div class="w-[90vw] max-w-[90vw] h-full max-h-lvh min-h-0 flex">
		<div class="shrink-0 break-all h-full min-h-full overflow-auto">
			{@render TaskInfoSider()}
		</div>
		<div class="grow min-w-0 h-full min-h-0 bg-base-100">
			{#if selectedId}
				{#key selectedId}
					{let wrappers = $derived(taskResMap.get(selectedId))}
					<TaskViewer {wrappers} />
				{/key}
			{/if}
		</div>
	</div>
{/snippet}

{#snippet TaskInfoSider()}
	<ul class="menu bg-base-200 min-h-full p-0">
		<!-- Sidebar content -->
		{#if taskInfo.size <= 0}
			<div class="text-xl">
				<p>No task here.</p>
			</div>
		{:else}
			{#each taskInfo.entries() as [id, info] (id)}
				<li>
					{let isRunning = $derived(runningTaskId.has(id))}
					{let wrapper = $derived(taskResMap.get(id))}
					{let getWrong = $derived(!wrapper || wrapper.some((w) => w.res.err || !w.res.success))}
					{#snippet ContextMenu()}
						{@render TaskContextMenu(id)}
					{/snippet}
					<button
						title={`${info.collection.name} - ${info.meta.path}\n${info.collection.path} > ${info.meta.path}`}
						class="group flex gap-1 items-center text-sm border-b border-base-content/50
										{isRunning ? 'text-base-content/70' : getWrong ? 'text-error' : 'text-success'}
										{selectedId === id ? 'text-primary-content bg-primary' : ''}"
						onclick={() => (selectedId = id)}
						oncontextmenu={(e) => {
							e.stopPropagation();
							e.preventDefault();
							contextmenu.open(e, ContextMenu);
						}}
					>
						<span class="*:size-3.5 *:shrink-0">
							{#if isRunning}
								<IconRefreshCw class="animate-spin" />
							{:else if getWrong}
								<IconCircleX />
							{:else}
								<IconCircleCheck />
							{/if}
						</span>
						<span>{info.collection.name}</span>
						<span class="hidden group-hover:inline">{info.meta.path}</span>
					</button>
				</li>
			{/each}
		{/if}
	</ul>
{/snippet}

{#snippet TaskContextMenu(id: string)}
	<ul
		class="menu p-0 bg-base-100 border border-base-content/30 [&_li>*]:rounded-none"
		oncontextmenu={(e) => e.preventDefault()}
	>
		{const Labels = {
			Stop: 'Stop',
			Restart: 'Restart',
			Delete: 'Delete',
		} as const}
		{const Fields = {
			[Labels.Stop]: { on: () => Events.Emit(TaskEvent.Stop, id), icon: IconPause },
			[Labels.Restart]: { on: () => Events.Emit(TaskEvent.Restart, id), icon: IconRotateCcwClock },
			[Labels.Delete]: { on: () => Events.Emit(TaskEvent.Delete, id), icon: IconTrash },
		} as const}
		{#snippet btn(label: (typeof Labels)[keyof typeof Labels])}
			<button
				onclick={async () => {
					Fields[label].on();
					contextmenu.close();
				}}
			>
				<div class="flex gap-1 items-center">
					{const FieldIcon = Fields[label].icon}
					<FieldIcon class="size-3.5" />
					<span class="text-sm">{label}</span>
				</div>
			</button>
		{/snippet}
		{#if runningTaskId.has(id)}
			<li>{@render btn('Stop')}</li>
		{:else}
			<li>{@render btn('Restart')}</li>
		{/if}
		<div class="divider p-0 m-0 h-1"></div>
		<li class="text-error [&>button]:hover:bg-error/10">{@render btn('Delete')}</li>
	</ul>
{/snippet}
