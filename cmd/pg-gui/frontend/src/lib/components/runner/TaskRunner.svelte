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
	let showBoard = $state(false);

	// ! data for test view
	{
		const id1 = crypto.randomUUID();
		const id2 = crypto.randomUUID();
		const id3 = crypto.randomUUID();
		runningTaskId.add(id1);
		taskInfo.set(id1, {
			collection: { name: 'c1', path: '/home/homoarea/collecitons/c1' },
			meta: { id: id1, path: 'div1/son1/httreq' },
		});
		taskInfo.set(id2, {
			collection: { name: 'cccc2', path: '/home/homoarea/collecitons/cccc2' },
			meta: { id: id2, path: 'div1/son1/httreq' },
		});
		taskInfo.set(id3, {
			collection: { name: 'cccc2', path: '/home/homoarea/collecitons/cccc2' },
			meta: { id: id3, path: 'div1/son1/httreq' },
		});
		for (let i = 0; i < 10; i++) {
			const id = crypto.randomUUID();
			taskInfo.set(id, {
				collection: { name: `tt${i}`, path: `/home/homoarea/collection/${i}` },
				meta: { id, path: `div1/son1/req=${id}` },
			});
		}
		const exHttpFailRes = JSON.parse(String.raw`{
    "success": false,
    "reqName": "post_with_body",
    "reqInfo": "POST http://127.0.0.1:8000/{path1}/{path2}",
    "proto": "HTTP/1.0",
    "status": "200 OK",
    "duration": 1842927,
    "tests": [
        {
            "name": "xxx",
            "passed": false,
            "err": {
                "from": 31,
                "to": 32,
                "line": 1,
                "column": 31,
                "message": "unexpected token Operator(\"=\")",
                "snippet": "\n | fromJSON(res.String()).method===\"POST\"\n | ...............................^",
                "prev": null
            }
        },
        {
            "name": "yyy",
            "passed": false,
            "err": {
                "from": 9,
                "to": 12,
                "line": 1,
                "column": 9,
                "message": "cannot use *resty.Response as argument (type string) to call fromJSON ",
                "snippet": "\n | fromJSON(res).method==\"POST\"\n | .........^",
                "prev": null
            }
        },
        {
            "name": "zzz",
            "passed": true,
            "err": null
        }
    ],
    "err": null,
    "meta": {
        "id": "${id2}",
        "path": "example/post_with_body.toml"
    },
    "data": {
        "URL": "http://127.0.0.1:8000/homo/114?q1=v1&q1=v2&q2=v1",
        "method": "POST",
        "spent": "1.8ms",
        "size": "671B",
        "request": {
            "header": {
                "Accept-Encoding": [
                    "gzip, deflate"
                ],
                "Content-Type": [
                    "multipart/form-data; boundary=587e09dbd7fedac055f79f418b70268137703a8c9911c2de12fb2395378c"
                ],
                "User-Agent": [
                    "pg@v0.1"
                ],
                "h1": [
                    "v1"
                ]
            },
            "body": "--587e09dbd7fedac055f79f418b70268137703a8c9911c2de12fb2395378c\r\nContent-Disposition: form-data; name=\"mq1\"\r\n\r\nm2\r\n--587e09dbd7fedac055f79f418b70268137703a8c9911c2de12fb2395378c\r\nContent-Disposition: form-data; name=\"abc\"\r\n\r\n123\r\n--587e09dbd7fedac055f79f418b70268137703a8c9911c2de12fb2395378c--\r\n",
            "raw": "POST /homo/114?q1=v1&q1=v2&q2=v1 HTTP/1.1\r\nHost: 127.0.0.1:8000\r\nUser-Agent: pg@v0.1\r\nContent-Length: 295\r\nAccept-Encoding: gzip, deflate\r\nContent-Type: multipart/form-data; boundary=587e09dbd7fedac055f79f418b70268137703a8c9911c2de12fb2395378c\r\nh1: v1\r\n\r\n--587e09dbd7fedac055f79f418b70268137703a8c9911c2de12fb2395378c\r\nContent-Disposition: form-data; name=\"mq1\"\r\n\r\nm2\r\n--587e09dbd7fedac055f79f418b70268137703a8c9911c2de12fb2395378c\r\nContent-Disposition: form-data; name=\"abc\"\r\n\r\n123\r\n--587e09dbd7fedac055f79f418b70268137703a8c9911c2de12fb2395378c--\r\n"
        },
        "response": {
            "header": {
                "Content-Length": [
                    "671"
                ],
                "Content-Type": [
                    "application/json; charset=utf-8"
                ],
                "Date": [
                    "Thu, 08 Oct 2026 13:16:11 GMT"
                ],
                "Server": [
                    "BaseHTTP/0.6 Python/3.14.7"
                ]
            },
            "body": "{\n  \"method\": \"POST\",\n  \"path\": \"/homo/114?q1=v1&q1=v2&q2=v1\",\n  \"headers\": {\n    \"Host\": \"127.0.0.1:8000\",\n    \"User-Agent\": \"pg@v0.1\",\n    \"Content-Length\": \"295\",\n    \"Accept-Encoding\": \"gzip, deflate\",\n    \"Content-Type\": \"multipart/form-data; boundary=587e09dbd7fedac055f79f418b70268137703a8c9911c2de12fb2395378c\",\n    \"h1\": \"v1\"\n  },\n  \"body\": \"--587e09dbd7fedac055f79f418b70268137703a8c9911c2de12fb2395378c\\r\\nContent-Disposition: form-data; name=\\\"mq1\\\"\\r\\n\\r\\nm2\\r\\n--587e09dbd7fedac055f79f418b70268137703a8c9911c2de12fb2395378c\\r\\nContent-Disposition: form-data; name=\\\"abc\\\"\\r\\n\\r\\n123\\r\\n--587e09dbd7fedac055f79f418b70268137703a8c9911c2de12fb2395378c--\\r\\n\"\n}",
            "raw": "HTTP/1.0 200 OK\r\nConnection: close\r\nContent-Length: 671\r\nContent-Type: application/json; charset=utf-8\r\nDate: Thu, 08 Oct 2026 13:16:11 GMT\r\nServer: BaseHTTP/0.6 Python/3.14.7\r\n\r\n{\n  \"method\": \"POST\",\n  \"path\": \"/homo/114?q1=v1&q1=v2&q2=v1\",\n  \"headers\": {\n    \"Host\": \"127.0.0.1:8000\",\n    \"User-Agent\": \"pg@v0.1\",\n    \"Content-Length\": \"295\",\n    \"Accept-Encoding\": \"gzip, deflate\",\n    \"Content-Type\": \"multipart/form-data; boundary=587e09dbd7fedac055f79f418b70268137703a8c9911c2de12fb2395378c\",\n    \"h1\": \"v1\"\n  },\n  \"body\": \"--587e09dbd7fedac055f79f418b70268137703a8c9911c2de12fb2395378c\\r\\nContent-Disposition: form-data; name=\\\"mq1\\\"\\r\\n\\r\\nm2\\r\\n--587e09dbd7fedac055f79f418b70268137703a8c9911c2de12fb2395378c\\r\\nContent-Disposition: form-data; name=\\\"abc\\\"\\r\\n\\r\\n123\\r\\n--587e09dbd7fedac055f79f418b70268137703a8c9911c2de12fb2395378c--\\r\\n\"\n}"
        }
    }
}`);
		const taskRes = (id: string, name: string, pass = true): TaskResWrapper => ({
			mode: NodeMode.FileMode | NodeMode.HttpReq | NodeMode.HttpGet,
			res: {
				meta: { id, path: `${id}/${name}` },
				success: pass,
				reqName: name,
				reqInfo: `GET ${name}`,
				proto: 'HTTP',
				status: `${pass ? 'pass' : 'fail'} 200 OK 20ms`,
				duration: 20 * Duration.Millisecond,
				tests: null,
				err: undefined,
				data: { method: 'GET', spent: '20ms', size: '0B', URL: '' },
			} as TaskResHttp,
		});
		let res1 = $state<TaskResWrapper[]>([]);
		for (let i = 0; i < 3; i++) {
			res1.push(taskRes(id1, `req1-${i}`));
		}
		taskResMap.set(id1, res1);
		let res2 = $state<TaskResWrapper[]>([]);
		res2.push({
			mode: NodeMode.FileMode | NodeMode.HttpReq | NodeMode.HttpPost,
			res: exHttpFailRes,
		});
		for (let i = 0; i < 1; i++) {
			res2.push(taskRes(id2, `req2-${i}`));
		}
		taskResMap.set(id2, res2);
		let res3 = $state<TaskResWrapper[]>([]);
		for (let i = 0; i < 1; i++) {
			res3.push(taskRes(id3, `req3-${i}`));
		}
		res3.push(taskRes(id3, `req3-${233}`, false));
		taskResMap.set(id3, res3);
		// showBoard = true;
		// ! ^ test data
	}

	function genID(): TaskID {
		return crypto.randomUUID();
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
			})
			.catch(msgAlert.error)
			.finally(() => runningTaskId.delete(id));
	}
	function pushTaskInfo(info: TaskInfo) {
		taskInfo.set(info.meta.id, info);
	}
	export const taskRunner = {
		genID,
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

	let selectedId = $state<string>();

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
		const unToggleRunBoard = Events.On(MenuEvent.ToggleRunBoard, () => (showBoard = !showBoard));
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
	<input id="run-board-drawer-1" bind:checked={showBoard} type="checkbox" class="drawer-toggle" />
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
