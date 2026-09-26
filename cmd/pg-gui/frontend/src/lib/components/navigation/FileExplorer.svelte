<script lang="ts" module>
	type FS = Node | null;
	// addtional data is FSTag
	export type FSItem = TreeItem<NodeMode>;
	const NewFSGroup = NewTreeGroup<NodeMode>;
	const NewFSLeaf = NewTreeLeaf<NodeMode>;
	const emptyFSItem = { id: '' } as FSItem;
	class FileExplorerState {
		items = $state<FSItem[]>([]);
		selectedItem = $state<FSItem>(emptyFSItem);
		activedItem = $state<FSItem>();

		reload() {
			FileExploreService.ReloadRootView()
				.then((val) => {
					this.items = parseNodes(val);
					Events.Emit(DataEvent.FileExplorerLoaded);
				})
				.catch(msgAlert.error);
		}
		selectItem(item: FSItem | undefined) {
			if (!item) {
				this.selectedItem = emptyFSItem;
				return;
			}
			this.selectedItem = item;
			if (item.type === 'leaf') {
				this.activedItem = item;
			}
		}
		/**
		 * find the first item whose id equals path in items
		 * @param path
		 */
		find(path: string): FSItem | undefined {
			return this.findItem(this.items, path);
		}
		/**
		 * find the first item parent whose sub has id equals path in items
		 * @param path
		 */
		findPre(path: string): FSItem | undefined {
			return this.findPreItem(this.items, path);
		}
		/**
		 * find routes to item
		 * @param path
		 */
		findRoutes(path: string): FSItem[] {
			return this.findItemRoutes(this.items, path);
		}

		findItem(items: FSItem[], path: string): FSItem | undefined {
			for (const i of items) {
				if (path === i.id) {
					return i;
				}
				if (i.sub && path.startsWith(i.id)) {
					return this.findItem(i.sub, path);
				}
			}
		}
		findPreItem(items: FSItem[], path: string): FSItem | undefined {
			for (const i of items) {
				if (i.sub?.some((f) => f.id === path)) {
					return i;
				}
				if (i.sub && path.startsWith(i.id)) {
					return this.findPreItem(i.sub, path);
				}
			}
		}
		findItemRoutes(items: FSItem[], path: string): FSItem[] {
			const routes: FSItem[] = [];
			for (const i of items) {
				if (path === i.id) {
					return [i];
				}
				if (i.sub && path.startsWith(i.id)) {
					routes.push(i);
					const next = this.findItemRoutes(i.sub, path);
					if (next.length > 0) {
						routes.push(...next);
						return routes;
					}
					routes.pop();
				}
			}
			return routes;
		}
	}
	const fState = new FileExplorerState();
	// remember the open states of folders
	const openStates = new Map<string, boolean>();
	export { fsItemDisplay, nodeModeIcon };

	export const fileExplorer = {
		findRoutes: ((path?: string) => {
			if (!path) return [];
			return fState.findRoutes(path);
		}).bind(fState),
		find: fState.find.bind(fState),
		getActivedFS: (() => fState.activedItem).bind(fState),
		setActivedFS: ((item: FSItem | undefined) => (fState.activedItem = item)).bind(fState),
		getSelectedFS: (() => fState.selectedItem).bind(fState),
		setSelectedFS: fState.selectItem.bind(fState),
		// recursively open folder until item
		expandToItem: ((item: FSItem) => {
			for (const r of fState.findRoutes(item.id)) {
				if (r.type === 'group' && !r.isOpen) {
					r.isOpen = true;
					openStates.set(r.id, r.isOpen);
				}
			}
		}).bind(fState),
	};
	function parseNodes(nodes: FS[] | null | undefined): FSItem[] {
		return (nodes ?? []).filter((f) => f !== null).map(nodeToFSItem);
	}
	function nodeToFSItem(node: Node): FSItem {
		if (ModeContains(node.mode, NodeMode.DirMode)) {
			const dirItem = $state(
				NewFSGroup(node.path, node.name, {
					additional: node.mode,
					sub: parseNodes(node.sub),
					isOpen: openStates.get(node.path),
					snippet: fsItemDisplay,
					on: {
						// TODO: open editor of file
						onclick: () => {
							fState.selectItem(dirItem);
							FileExploreService.LoadSubsCached(node.path)
								.then((val) => {
									openStates.set(node.path, !openStates.get(node.path));
									dirItem.sub = parseNodes(val);
								})
								.catch((err) => console.error('load cached sub:', err));
						},
						// TODO dir context menu
						oncontextmenu: (e) => {
							fState.selectItem(dirItem);
							contextmenu.open(e, itemContextMenu, true);
						},
					},
				}),
			);
			return dirItem;
		} else {
			const fileItem = $state(
				NewFSLeaf(node.path, node.name, {
					additional: node.mode,
					snippet: fsItemDisplay,
					on: {
						onclick: () => fState.selectItem(fileItem),
						// TODO file context menu
						oncontextmenu: (e) => {
							fState.selectItem(fileItem);
							contextmenu.open(e, itemContextMenu, true);
						},
					},
				}),
			);
			return fileItem;
		}
	}
</script>

<script lang="ts">
	// Icons
	import IconFolder from '@lucide/svelte/icons/folder';
	import IconFolderCog from '@lucide/svelte/icons/folder-cog';
	import IconFile from '@lucide/svelte/icons/file';
	import IconRotateCw from '@lucide/svelte/icons/rotate-cw';
	import IconFolderPlus from '@lucide/svelte/icons/folder-plus';
	import IconFilePlusCorner from '@lucide/svelte/icons/file-plus-corner';
	import IconCopyMinus from '@lucide/svelte/icons/copy-minus';

	import { FileExploreService } from '@services';
	import TreeMenu, { NewTreeGroup, NewTreeLeaf, type TreeItem } from '@comp/base/TreeMenu.svelte';
	import { NodeMode, type Node } from '@services/models';
	import { Events } from '@wailsio/runtime';
	import { onMount } from 'svelte';
	import { contextmenu } from '../base/ContextMenu.svelte';
	import { msgAlert } from '../message/AlertNotify.svelte';
	import { DataEvent } from '@lib/enum/events';
	import { ModeContains } from '@lib/utils/asserts';
	import OpenCollectionModal from './OpenCollectionModal.svelte';

	onMount(() => {
		fState.reload();
	});

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
			handle: () => fState.reload(),
		},
		{
			label: 'collapse-all',
			icon: IconCopyMinus,
			handle: () => {},
		},
	];
</script>

<div class="w-full">
	<div class="flex p-1 border-b border-base-content/70">
		{#each topBarBtns as btn (btn.label)}
			<button
				title={btn.label}
				onclick={btn.handle}
				class="btn btn-ghost rounded-none size-6 p-0 text-base-content/90 hover:text-primary"
			>
				<btn.icon class="size-4.5" />
			</button>
		{/each}
	</div>
	<div class="overflow-auto">
		<TreeMenu
			bind:items={fState.items}
			bind:selectedId={fState.selectedItem.id}
			indentGuide={'always'}
			indentSize={5}
			enableHoverStyle
			enableActiveStyle
			enableSelectedStyle
		/>
	</div>
</div>
<OpenCollectionModal
	afterLoadFunc={(rootPaths) => {
		FileExploreService.AddRoot(rootPaths)
			.then((v) => (fState.items = parseNodes(v)))
			.catch(msgAlert.error);
	}}
/>

{#snippet itemContextMenu()}
	<!-- TODO: impl the function in explorer context menu -->
	<ul class="menu p-0 bg-base-100 [&_li>*]:rounded-none" oncontextmenu={(e) => e.preventDefault()}>
		{#if fState.selectedItem.type === 'group'}
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
	<div title={item.id} class="flex gap-1 items-center">
		{@render nodeModeIcon(item.additional)}
		{item.name}
	</div>
{/snippet}
{#snippet nodeModeIcon(m?: NodeMode)}
	{#if !m}
		<span>Unkown</span>
		<!-- Icon for folder -->
	{:else if ModeContains(m, NodeMode.DirMode)}
		<!-- config folder -->
		{#if ModeContains(m, NodeMode.ConfigDir)}
			<IconFolderCog
				class="size-4 text-base-content/addRootAndCloseModal80"
				strokeWidth={1}
				nonScalingStroke
			/>
		{:else}
			<IconFolder class="size-4 text-base-content/80" strokeWidth={1} nonScalingStroke />
		{/if}
		<!-- Icon for file -->
	{:else if ModeContains(m, NodeMode.FileMode)}
		{#if ModeContains(m, NodeMode.HttpGet)}
			<span class="text-req-get">GET</span>
		{:else if ModeContains(m, NodeMode.HttpPost)}
			<span class="text-req-post">POST</span>
		{:else if ModeContains(m, NodeMode.HttpPut)}
			<span class="text-req-put">PUT</span>
		{:else if ModeContains(m, NodeMode.HttpDelete)}
			<span class="text-req-delete">DEL</span>
		{:else if ModeContains(m, NodeMode.HttpHead)}
			<span class="text-req-head">HEAD</span>
		{:else if ModeContains(m, NodeMode.HttpPatch)}
			<span class="text-req-patch">PATCH</span>
		{:else if ModeContains(m, NodeMode.HttpOptions)}
			<span class="text-req-options">OPT</span>
		{:else if ModeContains(m, NodeMode.HttpTrace)}
			<span class="text-req-trace">TRACE</span>
		{:else if ModeContains(m, NodeMode.HttpConnect)}
			<span class="text-req-connect">CON</span>
			<!-- TODO: add more icon for req file -->
		{:else}
			<IconFile class="size-4 text-base-content/25" strokeWidth={1} nonScalingStroke />
		{/if}
	{/if}
{/snippet}
