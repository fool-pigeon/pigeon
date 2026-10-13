<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	import { modal } from '@comp/base/Modal.svelte';
	import TreeMenu from '@comp/base/TreeMenu.svelte';
	import EnvSwitch from '@comp/context/EnvSwitch.svelte';
	import { msgAlert } from '@comp/message/AlertNotify.svelte';
	import {
		fileExplorer,
		NewFSGroup,
		NewFSLeaf,
		nodeModeIcon,
		type FS,
		type FSItem,
	} from '@comp/navigation/FileExplorer.svelte';
	import { ModeContains } from '@lib/utils/asserts';
	import { MenuEvent } from '@pg/gui';
	import {
		FesEvent,
		FileExploreService,
		NodeMode,
		TaskRunService,
		type FesEvtUpdateSubs,
		type FileMeta,
		type Node,
		type Task,
	} from '@services';
	import { Events } from '@wailsio/runtime';
	import { onMount, untrack, type Snippet } from 'svelte';
	import { SvelteSet } from 'svelte/reactivity';
	import { taskRunner } from './TaskRunner.svelte';

	let collectionFile = $state<FileMeta>();
	let envFile = $state<FileMeta>();
	let pathSet = new SvelteSet<string>();
	let root = $state<FSItem>({ id: '', sub: [] as FSItem[] } as FSItem);
	function findItem(item: FSItem, path: string): FSItem | undefined {
		if (item.id === path) return item;
		if (item.sub) {
			for (const i of item.sub) {
				if (path.startsWith(i.id)) {
					const fs = findItem(i, path);
					if (fs) return fs;
				}
			}
		}
	}
	$effect(() => {
		const c = collectionFile;
		if (c) {
			untrack(() => {
				pathSet.clear();
				root = { id: '', sub: [] as FSItem[] } as FSItem;
				FileExploreService.ReloadRootView();
			});
		}
	});
	onMount(() => {
		const unRunTask = Events.On(MenuEvent.RunTask, () => modal.open(Modal, 'Run Task'));
		const unReload = Events.On(FesEvent.FesReloadRoot, (v) => {
			const data = v.data as FesEvtUpdateSubs;
			root.sub = parseNodes(data.subs?.filter((v) => v?.path === collectionFile?.path) ?? []);
		});
		const unUpdateSubs = Events.On(FesEvent.FesUpdateSubs, (v) => {
			const data = v.data as FesEvtUpdateSubs;
			const fs = findItem(root, data.path);
			if (fs) {
				fs.sub = parseNodes(data.subs);
			}
		});
		return () => {
			unRunTask();
			unReload();
			unUpdateSubs();
		};
	});

	function parseNodes(nodes: FS[] | null | undefined): FSItem[] {
		return (nodes ?? []).filter((f) => f !== null).map((node) => nodeToFSItem(node));
	}
	function nodeToFSItem(node: Node, snippet: Snippet<[item: FSItem]> = treeItemDisplay): FSItem {
		if (ModeContains(node.mode, NodeMode.DirMode)) {
			const dirItem = $state(
				NewFSGroup(node.path, node.name, {
					snippet,
					additional: node.mode,
					sub: parseNodes(node.sub),
					on: { onclick: () => FileExploreService.LoadSubsCached(node.path).catch(msgAlert.error) },
				}),
			);
			return dirItem;
		} else {
			const fileItem = $state(
				NewFSLeaf(node.path, node.name, {
					snippet,
					additional: node.mode,
					on: { onclick: () => setCheck(fileItem, !pathSet.has(fileItem.id)) },
				}),
			);
			return fileItem;
		}
	}

	/**
	 * @returns string[] the path only full check and exclude duplicate folders and files from the selection
	 */
	function compactPathSet(): string[] {
		function mark(item: FSItem): boolean {
			if (!item.sub?.length) {
				if (pathSet.has(item.id)) {
					full.add(item);
					return true;
				}
				return false;
			}
			let all = true;
			for (const c of item.sub) if (!mark(c)) all = false;
			if (all) {
				full.add(item);
				return true;
			}
			return false;
		}
		function collect(item: FSItem) {
			if (full.has(item)) {
				out.push(item.id);
				return;
			}
			if (!item.sub?.length) {
				if (pathSet.has(item.id)) out.push(item.id);
				return;
			}
			for (const c of item.sub) collect(c);
		}

		const full = new Set<FSItem>();
		const out: string[] = [];

		for (const c of root.sub ?? []) mark(c);
		for (const c of root.sub ?? []) collect(c);
		return out;
	}

	function leaves(item: FSItem): string[] {
		return item.sub?.length ? item.sub.flatMap(leaves) : [item.id];
	}
	function setCheck(item: FSItem, on: boolean) {
		for (const id of leaves(item)) on ? pathSet.add(id) : pathSet.delete(id);
	}
	function checkState(item: FSItem): boolean | null {
		const ids = leaves(item);
		const n = ids.filter((id) => pathSet.has(id)).length;
		return n === 0 ? false : ids.length === n ? true : null;
	}
</script>

{#snippet Modal()}
	{const collections = fileExplorer.getCollections()}
	{#if collections.length <= 0}
		<div class="p-2 text-warning">
			<p>No collections opened yet.</p>
			<p>Please open collection in File Menu</p>
		</div>
	{:else}
		<form
			class="p-0 w-auto max-w-[80vw] max-h-[80vh]"
			onsubmit={(e) => {
				e.preventDefault();
				if (!root.sub?.length) {
					modal.close();
					return;
				}
				const collection = collectionFile;
				if (!collection) return;

				const c = root.sub[0];
				const id = taskRunner.genID();
				const task: Task = {
					id: id,
					path: c.id,
					param: {
						root: '',
						tasks: null,
						config: '',
						env: '',
					},
				};
				// find the root
				const q: FSItem[] = [c];
				for (let i = 0; i < q.length; i++) {
					const node = q[i];
					if (checkState(node)) {
						task.path = node.id;
						task.param.root = node.id;
						break;
					}
					for (const next of node.sub ?? []) {
						q.push(next);
					}
				}
				task.param.tasks = compactPathSet();
				task.param.env = envFile?.path ?? '';
				// tip:the config's path will be inject in RunWrapper
				// task.param.config = '';

				taskRunner.addRunningId(id);
				taskRunner.setTaskInfo({
					collection,
					meta: { id: task.id, path: task.path },
				});
				taskRunner.showBoard();
				taskRunner.selectTask(id);
				TaskRunService.RunWrapper(task)
					.catch(msgAlert.error)
					.finally(() => taskRunner.deleteRunningId(id));

				modal.close();
			}}
		>
			{@render CollectionAndEnvSelector(collections)}
			{@render PathSelector()}
			<div class="p-4 flex gap-1 justify-end items-center">
				<div>
					<button
						type="button"
						class="btn btn-ghost font-normal hover:bg-base-content/30"
						onclick={() => modal.close()}
					>
						Cancel
					</button>
					<button type="submit" class="btn btn-info font-normal"> Run </button>
				</div>
			</div>
		</form>
	{/if}
{/snippet}

{#snippet CollectionAndEnvSelector(collections: FileMeta[])}
	<div class="w-full flex gap-2 items-center text-sm">
		<label class="flex items-center">
			<span class="text-sm">Collection:</span>
			<select
				bind:value={collectionFile}
				oninvalid={(e) => e.preventDefault()}
				class="pl-1 select [&::picker(select)]:my-0 select-sm validator outline-none shadow-none"
				required
			>
				<option disabled selected value={undefined}>pick collection</option>
				{#each collections as c, i}
					<option value={c} selected={i === 0}>{c.name}</option>
				{/each}
			</select>
		</label>
		<div>
			<EnvSwitch path={collectionFile?.path} bind:current={envFile} />
		</div>
	</div>
{/snippet}

{#snippet PathSelector()}
	<div>{JSON.stringify(pathSet, null, 2)}</div>
	<button type="button" onclick={() => console.log(pathSet)}>log</button>
	{#if collectionFile}
		<div class="mt-1 p-2 text-sm border border-base-content/20">
			<span>Select paths:</span>
			<div class="min-h-0 overflow-auto scrollbar-none hover:scrollbar-thin">
				<TreeMenu items={root.sub} indentGuide={'always'} indentSize={5} enableHoverStyle />
			</div>
		</div>
	{/if}
{/snippet}

{#snippet treeItemDisplay(item: FSItem)}
	{let st = $derived(checkState(item))}
	<div title={item.id} class="flex gap-1 items-center">
		<input
			type="checkbox"
			class="checkbox checkbox-sm checkbox-primary shadow-none"
			checked={st === true}
			indeterminate={st === null}
			onclick={(e) => e.stopPropagation()}
			onchange={(e) => setCheck(item, e.currentTarget.checked)}
		/>
		{@render nodeModeIcon(item.additional)}
		<span class="truncate">{item.name}</span>
	</div>
{/snippet}
