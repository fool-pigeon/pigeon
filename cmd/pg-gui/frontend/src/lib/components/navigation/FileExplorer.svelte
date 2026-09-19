<script lang="ts" module>
	type FS = Node | null;
	// addtional data is FSTag
	type FSItem = TreeItem<FSTag>;
	const NewFSGroup = NewTreeGroup<FSTag>;
	const NewFSLeaf = NewTreeLeaf<FSTag>;

	let items = $state<FSItem[]>([]);
	let selectedItem = $state<FSItem>({ id: '' } as FSItem);
	let selectedId = $state('');
	let rootPathsToAdd = $state<string[]>([]);
	let selectedPath = $state<string>('');

	// remember the open states of folders
	const openStates = new Map<string, boolean>();
</script>

<script lang="ts">
	import { MenuEvent } from '@bindings/gui';
	import TreeMenu, {
		NewTreeGroup,
		NewTreeLeaf,
		type TreeItem,
	} from '@components/base/TreeMenu.svelte';
	import {
		AddRoot,
		DeleteNode,
		AddNode,
		LoadSubsCached,
		ReloadSubsCached,
		ReloadRootView,
		ParseRootPaths,
	} from '@services/fileexploreservice';
	import { FSTag, FSType, type Node } from '@services/models';
	import { Events } from '@wailsio/runtime';
	import { onDestroy, onMount } from 'svelte';
	import { modal } from '../base/Modal.svelte';
	import { SelectFolder } from '@bindings/service/pathselectservice';
	import { contextmenu } from '../base/ContextMenu.svelte';
	// Import Icons
	import IconLoaderCircle from '@lucide/svelte/icons/loader-circle';
	import IconFolder from '@lucide/svelte/icons/folder';
	import IconFolderCog from '@lucide/svelte/icons/folder-cog';
	import IconFile from '@lucide/svelte/icons/file';
	import IconRotateCw from '@lucide/svelte/icons/rotate-cw';
	import IconFolderPlus from '@lucide/svelte/icons/folder-plus';
	import IconFilePlusCorner from '@lucide/svelte/icons/file-plus-corner';
	import IconCopyMinus from '@lucide/svelte/icons/copy-minus';

	$effect(() => {
		selectedId = selectedItem.id;
	});
	onMount(() => {
		Events.On(MenuEvent.OpenCollection, () => {
			SelectFolder()
				.then((val) => {
					rootPathsToAdd = [];
					selectedPath = val;
					modal.open(selectRootPathsModal, 'Open Collection');
				})
				.catch((err) => {
					// TODO global message
					console.error(err);
				});
		});
		ReloadRootView()
			.then((val) => (items = parseNodes(val)))
			.catch((err) => console.error(err));
	});
	onDestroy(async () => {
		Events.Off(MenuEvent.OpenCollection);
	});

	/**
	 * find the first item whose id equals path in items
	 * @param items
	 * @param path
	 */
	function findItem(items: FSItem[], path: string): FSItem | undefined {
		for (const i of items) {
			if (path === i.id) {
				return i;
			}
			if (i.sub && path.startsWith(i.id)) {
				const target = findItem(i.sub, path);
				if (target) return target;
			}
		}
	}
	/**
	 * find the first item parent whose sub has id equals path in items
	 * @param items
	 * @param path
	 */
	function findPreItem(items: FSItem[], path: string): FSItem | undefined {
		for (const i of items) {
			if (i.sub?.some((f) => f.id === path)) {
				return i;
			}
			if (i.sub && path.startsWith(i.id)) {
				const pre = findPreItem(i.sub, path);
				if (pre) return pre;
			}
		}
	}
	function deleteFSItem(path: string) {
		DeleteNode(path)
			.then(() => {
				const pre = findPreItem(items, path);
				if (pre) pre.sub = pre.sub?.filter((s) => s.id !== path);
			})
			.catch((err) => console.error(err));
	}
	function selectItem(item: FSItem) {
		selectedItem = item;
	}

	function parseNodes(nodes: FS[] | null | undefined): FSItem[] {
		return (nodes ?? []).filter((f) => f !== null).map(nodeToFSItem);
	}
	function nodeToFSItem(node: Node): FSItem {
		if (FSType.DirType === node.type) {
			const dirItem = NewFSGroup(node.path, node.name, {
				sub: parseNodes(node.sub),
				isOpen: openStates.get(node.path),
				snippet: fsItemDisplay,
				on: {
					// TODO: open editor of file
					onclick: () => {
						selectItem(dirItem);
						LoadSubsCached(node.path)
							.then((val) => {
								if (openStates.get(node.path)) {
									openStates.set(node.path, false);
								} else {
									openStates.set(node.path, true);
								}
								const target = findItem(items, node.path);
								if (target) target.sub = parseNodes(val);
							})
							.then((err) => console.error('load cached sub:', err));
					},
					// TODO dir context menu
					oncontextmenu: (e) => {
						selectItem(dirItem);
						contextmenu.open(e, itemContextMenu, true);
					},
				},
			});
			return dirItem;
		} else {
			const fileItem = NewFSLeaf(node.path, node.name, {
				additional: node.tag,
				snippet: fsItemDisplay,
				on: {
					onclick: () => selectItem(fileItem),
					// TODO file context menu
					oncontextmenu: (e) => {
						selectItem(fileItem);
						contextmenu.open(e, itemContextMenu, true);
					},
				},
			});
			return fileItem;
		}
	}
	function addRootInModal() {
		AddRoot(rootPathsToAdd)
			.then((val) => {
				items = parseNodes(val);
				modal.close();
			})
			.catch((err) => {
				console.error(err);
				modal.close();
			});
	}

	// TODO:finish top bar btns of explorer
	const topBarBtns = [
		{
			label: 'add-file',
			icon: IconFilePlusCorner,
			handle: () => {},
		},
		{
			label: 'add-folder',
			icon: IconFolderPlus,
			handle: () => {},
		},
		{
			label: 'reload',
			icon: IconRotateCw,
			handle: () =>
				ReloadRootView()
					.then((val) => (items = parseNodes(val)))
					.catch((err) => console.error(err)),
		},
		{
			label: 'collapse-all',
			icon: IconCopyMinus,
			handle: () => {},
		},
	];
</script>

<div class="w-full">
	<div class="flex justify-end">
		{#each topBarBtns as btn (btn.label)}
			<div class="group relative">
				<button
					onclick={btn.handle}
					class="btn btn-ghost rounded-none size-6 p-1 hover:bg-base-content/10"
				>
					<btn.icon class="size-4" />
				</button>
				<span class="absolute top-full left-0 hidden group-hover:block whitespace-nowrap">
					{btn.label}
				</span>
			</div>
		{/each}
	</div>
	<div class="overflow-auto">
		<TreeMenu
			bind:items
			bind:selectedId
			indentGuide={'always'}
			indentSize={5}
			enableHoverStyle
			enableActiveStyle
			enableSelectedStyle
		/>
	</div>
</div>

{#snippet selectRootPathsModal()}
	{#await ParseRootPaths(selectedPath)}
		<div class="h-30 w-full flex justify-center items-center">
			<IconLoaderCircle class="absolute size-10 animate-spin" strokeWidth={2} nonScalingStroke />
		</div>
	{:then paths}
		{#if paths !== null && paths.length > 0}
			<div class="p-4 min-w-sm">
				<div class="flex items-end justify-between">
					<div class="flex items-baseline gap-1">
						<span>Collections</span>
						<div class="badge badge-soft badge-info">{paths.length}</div>
					</div>
					<div>
						<button
							class="btn btn-soft btn-sm font-normal {rootPathsToAdd.length === paths.length
								? 'btn-error'
								: 'btn-primary'}"
							onclick={() => {
								if (rootPathsToAdd.length === paths.length) {
									rootPathsToAdd = [];
								} else {
									rootPathsToAdd = paths;
								}
							}}>{rootPathsToAdd.length === paths.length ? 'DropAll' : 'SelectAll'}</button
						>
					</div>
				</div>
				<div class="p-2 mt-2 flex items-center border border-base-content/30">
					{#each paths as path}
						<label class="label">
							<input
								type="checkbox"
								class="checkbox checkbox-sm checkbox-primary"
								bind:group={rootPathsToAdd}
								value={path}
							/>
							{path}
						</label>
					{/each}
				</div>
				<div class="mt-10 flex gap-1 justify-between items-center">
					<div>
						<span class="text-primary">{rootPathsToAdd.length} </span>
						<span class="text-base-content/50">of {paths.length} selected</span>
					</div>
					<div>
						<button class="btn btn-ghost font-normal" onclick={() => modal.close()}>Cancel</button>
						<button class="btn btn-primary font-normal" onclick={addRootInModal}>Load</button>
					</div>
				</div>
			</div>
		{:else}
			<div class="text-warning h-10 text-center">Not Found Collection in {selectedPath}</div>
		{/if}
	{:catch err}
		<div class="text-error h-10">cause-error: {err}</div>
	{/await}
{/snippet}
{#snippet itemContextMenu()}
	<!-- TODO: impl the function in explorer context menu -->
	<ul class="menu p-0 bg-base-100 [&_li>*]:rounded-none" oncontextmenu={(e) => e.preventDefault()}>
		{#if selectedItem?.type === 'group'}
			<li><button>new request</button></li>
			<li><button>new folder</button></li>
		{:else}
			<li><button>copy</button></li>
		{/if}
		<li><button>rename</button></li>
		<div class="divider p-0 m-0 h-1"></div>
		<li><button>info</button></li>
		<li><button>Reveal In File Manager</button></li>
		<div class="divider p-0 m-0 h-1"></div>
		<li><button>Delete</button></li>
	</ul>
{/snippet}

{#snippet fsItemDisplay(item: FSItem)}
	<div class="flex gap-1 items-center">
		{#if item.type === 'group'}
			<!-- Icon for folder -->
			{#if item.id.endsWith('.pg')}
				<IconFolderCog class="size-4 text-base-content/80" strokeWidth={1} nonScalingStroke />
			{:else}
				<IconFolder class="size-4 text-base-content/80" strokeWidth={1} nonScalingStroke />
			{/if}
			<!-- Icon for file -->
		{:else if item.additional === FSTag.HttpGet}
			<span class="text-req-get">GET</span>
		{:else if item.additional === FSTag.HttpPost}
			<span class="text-req-post">POST</span>
		{:else if item.additional === FSTag.HttpPut}
			<span class="text-req-put">PUT</span>
		{:else if item.additional === FSTag.HttpDelete}
			<span class="text-req-delete">DEL</span>
		{:else if item.additional === FSTag.HttpHead}
			<span class="text-req-head">HEAD</span>
		{:else if item.additional === FSTag.HttpPatch}
			<span class="text-req-patch">PATCH</span>
		{:else if item.additional === FSTag.HttpOptions}
			<span class="text-req-options">OPT</span>
		{:else if item.additional === FSTag.HttpTrace}
			<span class="text-req-trace">TRACE</span>
		{:else if item.additional === FSTag.HttpConnect}
			<span class="text-req-connect">CON</span>
			<!-- TODO: add more icon for req file -->
		{:else}
			<IconFile class="size-4 text-base-content/25" strokeWidth={1} nonScalingStroke />
		{/if}
		{item.name}
	</div>
{/snippet}
