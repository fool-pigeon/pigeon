<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts" module>
	// Icons
	import IconFolder from '@lucide/svelte/icons/folder';
	import IconFolderCog from '@lucide/svelte/icons/folder-cog';
	import IconFolderOpenDot from '@lucide/svelte/icons/folder-open-dot';
	import IconFile from '@lucide/svelte/icons/file';
	import IconRotateCw from '@lucide/svelte/icons/rotate-cw';
	import IconFolderPlus from '@lucide/svelte/icons/folder-plus';
	import IconFilePlusCorner from '@lucide/svelte/icons/file-plus-corner';
	import IconCopyMinus from '@lucide/svelte/icons/copy-minus';
	import IconTrash from '@lucide/svelte/icons/trash';
	import IconSquarePen from '@lucide/svelte/icons/square-pen';
	import IconPlay from '@lucide/svelte/icons/play';

	import { FileAccessService, FileExploreService, PathSelectService } from '@services';
	import TreeMenu, { NewTreeGroup, NewTreeLeaf, type TreeItem } from '@comp/base/TreeMenu.svelte';
	import { FesEvent, NodeMode, type FesEvtUpdateSubs, type Node } from '@services/models';
	import { Events } from '@wailsio/runtime';
	import { onMount, type Snippet } from 'svelte';
	import { contextmenu } from '../base/ContextMenu.svelte';
	import { msgAlert } from '../message/AlertNotify.svelte';
	import { DataEvent } from '@lib/enum/events';
	import { ModeContains } from '@lib/utils/asserts';
	import OpenCollectionModal from './OpenCollectionModal.svelte';
	import { modal } from '@comp/base/Modal.svelte';
	import { MenuEvent } from '@bindings/gui';
	import { EditorEvent } from '@comp/editor/Editor.svelte';

	type FS = Node | null;
	// addtional data is FSTag
	export type FSItem = TreeItem<NodeMode>;
	const NewFSGroup = NewTreeGroup<NodeMode>;
	const NewFSLeaf = NewTreeLeaf<NodeMode>;
	const emptyFSItem = { id: '' } as FSItem;
	class FileExplorerState {
		root = $state<FSItem>({ id: '', sub: [] as FSItem[] } as FSItem);
		selectedItem = $state<FSItem>(emptyFSItem);
		activedItem = $state<FSItem>();

		reload() {
			FileExploreService.ReloadRootView().catch(msgAlert.error);
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
			// update req context
			Events.Emit(DataEvent.UpdateReqCtx, item.id);
		}
		/**
		 * find the first item whose id equals path in items
		 * @param path
		 */
		find(path: string): FSItem | undefined {
			return this.findItem(this.root, path);
		}
		/**
		 * find the first item parent whose sub has id equals path in items
		 * @param path
		 */
		findPre(path: string): FSItem | undefined {
			return this.findPreItem(this.root, path);
		}
		/**
		 * find routes to item
		 * @param path
		 */
		findRoutes(path: string): FSItem[] {
			return this.findItemRoutes(this.root.sub ?? [], path);
		}

		findItem(item: FSItem, path: string): FSItem | undefined {
			if (item.id === path) return item;
			if (item.sub) {
				for (const i of item.sub) {
					if (path.startsWith(i.id)) {
						const fs = this.findItem(i, path);
						if (fs) return fs;
					}
				}
			}
		}
		findPreItem(item: FSItem, path: string): FSItem | undefined {
			if (!item.sub) return undefined;
			if (item.sub.some((f) => f.id === path)) return item;
			for (const i of item.sub) {
				if (path.startsWith(i.id)) {
					const fs = this.findPreItem(i, path);
					if (fs) return fs;
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
	/**
	 * FileExplorer Event
	 */
	export const FEEvent = {
		NewFolder: 'file-explorer:new-folder',
		Reload: 'file-explorer:reload',
		CollapseAll: 'file-explorer:collapse-all',
		Reveal: 'file-explorer:reveal',
		Rename: 'file-explorer:rename',
		Delete: 'file-explorer:delete',
	} as const;
	export type FEEvent = (typeof FEEvent)[keyof typeof FEEvent];

	export const fileExplorer = {
		findRoutes: ((path?: string) => {
			if (!path) return [];
			return fState.findRoutes(path);
		}).bind(fState),
		find: fState.find.bind(fState),
		getCollections: (() => {
			return (fState.root.sub ?? []).map((v) => ({ name: v.name, path: v.id }));
		}).bind(fState),
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
						onclick: () => {
							fState.selectItem(dirItem);
							FileExploreService.LoadSubsCached(node.path)
								.then(() => openStates.set(node.path, !openStates.get(node.path)))
								.catch((err) => console.error('load cached sub:', err));
						},
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
	function dir(path: string): string {
		return path.replace(/[\\/][^\\/]*$/, '');
	}
	function basename(path: string): string | undefined {
		return path
			.replace(/[\\/]+$/, '')
			.split(/[\\/]/)
			.pop();
	}
</script>

<script lang="ts">
	onMount(() => {
		function updateSubs(v: Events.WailsEvent<string>) {
			const data = v.data as unknown as FesEvtUpdateSubs;
			const fs = fState.find(data.path);
			if (fs) {
				fs.sub = parseNodes(data.subs);
			}
		}
		fState.reload();
		const unUpdateSubs = Events.On(FesEvent.FesUpdateSubs, (v) => {
			updateSubs(v);
		});
		const unAfterRename = Events.On(FesEvent.FesRename, (v) => {
			updateSubs(v);
			Events.Emit(EditorEvent.Reload);
		});
		const unAfterReload = Events.On(FesEvent.FesRename, (v) => {
			updateSubs(v);
			Events.Emit(DataEvent.FileExplorerLoaded);
		});
		const unAfterDeletedNode = Events.On(FesEvent.FesDeletedNode, (v) => {
			const path = v.data as unknown as string;
			const fs = fState.findPre(path);
			if (fs && fs.sub) {
				fs.sub = fs.sub.filter((v) => v.id !== path);
			}
		});
		const unNewCollection = Events.On(MenuEvent.NewCollection, () => {
			FileAccessService.CreateCollection().catch(msgAlert.error);
		});
		const unNewFile = Events.On(MenuEvent.NewFile, () => {
			modal.open(NewFileModal, 'New File');
		});
		const unNewFolder = Events.On(FEEvent.NewFolder, () => {
			modal.open(NewFolderModal, 'New Folder');
		});
		const unReload = Events.On(FEEvent.Reload, () => fState.reload());
		const unCollapseAll = Events.On(FEEvent.CollapseAll, () => {
			openStates.clear();
			fState.reload();
		});
		const unReveal = Events.On(FEEvent.Reveal, () => {
			const path = fState.selectedItem.id;
			if (path === '') return;
			PathSelectService.ShowInFileManager(path).catch(msgAlert.error);
		});
		const unRename = Events.On(FEEvent.Rename, () => {
			const path = fState.selectedItem.id;
			if (path === '') return;
			modal.open(RenameFormModal, 'Rename');
		});
		const unDelete = Events.On(FEEvent.Delete, () => {
			const path = fState.selectedItem.id;
			if (path === '') return;
			// TODO open a modal to hold when need delete think twice
			FileAccessService.Delete(path)
				.then(() => msgAlert.success('Delete Success'))
				.catch(msgAlert.error);
		});
		return () => {
			unUpdateSubs();
			unAfterRename();
			unAfterReload();
			unAfterDeletedNode();
			unNewCollection();
			unNewFile();
			unNewFolder();
			unReload();
			unCollapseAll();
			unReveal();
			unRename();
			unDelete();
		};
	});

	const topBarBtns = [
		{
			label: 'add-file',
			icon: IconFilePlusCorner,
			handle: () => Events.Emit(MenuEvent.NewFile),
		},
		{
			label: 'add-folder',
			icon: IconFolderPlus,
			handle: () => Events.Emit(FEEvent.NewFolder),
		},
		{
			label: 'reload',
			icon: IconRotateCw,
			handle: () => Events.Emit(FEEvent.Reload),
		},
		{
			label: 'collapse-all',
			icon: IconCopyMinus,
			handle: () => Events.Emit(FEEvent.CollapseAll),
		},
		{
			label: 'runner-board',
			icon: IconPlay,
			handle: () => Events.Emit(MenuEvent.ToggleRunBoard),
		},
	];
</script>

<div class="w-full h-full min-h-0 flex flex-col">
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
	<div class="grow min-h-0 overflow-auto scrollbar-none hover:scrollbar-thin">
		<TreeMenu
			bind:items={fState.root.sub}
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
		FileExploreService.AddRoot(rootPaths).catch(msgAlert.error);
	}}
/>
{#snippet NewFolderModal()}
	{@render NewNodeModalWrapper(NewFolderForm)}
{/snippet}
{#snippet NewFileModal()}
	{@render NewNodeModalWrapper(NewFileForm)}
{/snippet}
{#snippet NewNodeModalWrapper(snippet: Snippet<[parent: string]>)}
	{const fsPath = fState.selectedItem.id}
	{const subs = fState.root.sub}
	{const defaultPath = fsPath !== '' ? fsPath : subs && subs.length > 0 ? subs.at(-1)!.id : ''}
	{#if defaultPath === ''}
		<div class="p-4 flex justify-center items-center text-warning">
			need select in file explorer
		</div>
	{:else}
		{const mode = fState.selectedItem.additional ?? NodeMode.DirMode}
		{const parent = ModeContains(mode, NodeMode.DirMode) ? defaultPath : dir(defaultPath)}
		{@render snippet(parent)}
	{/if}
{/snippet}
{#snippet NewFolderForm(parent: string)}
	{let folderName = $state('')}
	<form
		class="p-4 min-w-[33vw]"
		onsubmit={(e) => {
			e.preventDefault();
			if (!folderName) return;
			FileExploreService.AddDirNode(parent, folderName)
				.then(() => msgAlert.success(`Mkdir ${folderName} Success`))
				.catch(msgAlert.error);
			modal.close();
		}}
	>
		<div class="flex gap-1 items-center text-base-content/80">
			<span>In:</span>
			<IconFolder class="size-4" />
			{parent}
		</div>
		{#if folderName !== ''}
			<div class="flex gap-1 items-center text-base-content/80">
				<span>Create:</span>
				<IconFolder class="size-4" />
				{folderName}
			</div>
		{/if}
		<fieldset class="fieldset">
			<label class="label" for="folder-name">Folder Name</label>
			<input
				type="text"
				id="folder-name"
				class="input validator text-base-content/80 outline-none w-full"
				bind:value={folderName}
				oninvalid={(e) => e.preventDefault()}
				required
			/>
			<p class="validator-hint">folder name required</p>
		</fieldset>
		<div class="mt-4 flex gap-1 justify-end items-center">
			<div>
				<button
					type="button"
					class="btn btn-ghost font-normal hover:bg-base-content/30"
					onclick={() => modal.close()}
				>
					Cancel
				</button>
				<button type="submit" class="btn btn-info font-normal"> Create </button>
			</div>
		</div>
	</form>
{/snippet}
{#snippet NewFileForm(parent: string)}
	{let fileName = $state('')}

	{const modes = {
		RequestFile: NodeMode.FileMode | NodeMode.Request,
		Raw: NodeMode.FileMode,
	}}
	{let mode = $state(modes.RequestFile)}
	{let ext = $derived(ModeContains(mode, NodeMode.ConfigFile | NodeMode.Request) ? '.toml' : '')}
	{let reqType = $state(NodeMode.HttpReq)}
	{let methodType = $state(NodeMode.HttpGet)}

	<div class="flex grow">
		<select bind:value={mode} class="appearance-none w-full select select-sm outline-none">
			{#each Object.entries(modes) as [name, m]}
				<option value={m}>{name}</option>
			{/each}
		</select>
		{#if ModeContains(mode, NodeMode.Request)}
			{const ReqTypes = { HTTP: NodeMode.HttpReq }}
			<select bind:value={reqType} class="appearance-none select select-sm outline-none">
				{#each Object.entries(ReqTypes) as [label, type]}
					<option value={type}>{label}</option>
				{/each}
			</select>
			{#if ModeContains(reqType, NodeMode.HttpReq)}
				{const HttpMethods = {
					GET: NodeMode.HttpGet,
					POST: NodeMode.HttpPost,
					PUT: NodeMode.HttpPut,
					DELETE: NodeMode.HttpDelete,
					HEAD: NodeMode.HttpHead,
					PATCH: NodeMode.HttpPatch,
					OPTIONS: NodeMode.HttpOptions,
					CONNECT: NodeMode.HttpConnect,
					TRACE: NodeMode.HttpTrace,
				}}
				<select bind:value={methodType} class="appearance-none select select-sm outline-none">
					{#each Object.entries(HttpMethods) as [label, type]}
						<option value={type}>{label}</option>
					{/each}
				</select>
			{/if}
		{/if}
	</div>
	<form
		class="p-2 min-w-[33vw] flex flex-col gap-1"
		onsubmit={(e) => {
			e.preventDefault();
			if (!fileName) return;
			let m = mode;
			if (ModeContains(m, NodeMode.Request)) m |= reqType;
			if (ModeContains(m, NodeMode.HttpReq)) m |= methodType;
			FileExploreService.AddFileNode(parent, fileName + ext, m)
				.then(() => msgAlert.success(`Create ${fileName} Success`))
				.catch(msgAlert.error);
			modal.close();
		}}
	>
		<div class="flex gap-1 items-center text-base-content/80">
			<span>In:</span>
			<IconFolder class="size-4" />
			{parent}
		</div>
		{#if fileName !== ''}
			<div class="flex gap-1 items-center text-base-content/80">
				<span>Create:</span>
				<IconFile class="size-4" />
				{fileName}{ext}
			</div>
		{/if}
		<fieldset class="fieldset">
			<label class="label" for="folder-name">Name</label>
			<input
				type="text"
				id="folder-name"
				class="input validator text-base-content/80 outline-none w-full"
				bind:value={fileName}
				oninvalid={(e) => e.preventDefault()}
				required
			/>
			<p class="validator-hint">required</p>
		</fieldset>
		<div class="flex gap-1 justify-end items-center">
			<div>
				<button
					type="button"
					class="btn btn-ghost font-normal hover:bg-base-content/30"
					onclick={() => modal.close()}
				>
					Cancel
				</button>
				<button type="submit" class="btn btn-info font-normal"> Create </button>
			</div>
		</div>
	</form>
{/snippet}
{#snippet RenameFormModal()}
	{const oldPath = fState.selectedItem.id}
	{let fileName = $state(basename(oldPath))}
	<form
		class="p-4 min-w-[33vw]"
		onsubmit={(e) => {
			e.preventDefault();
			if (!fileName) return;
			FileAccessService.Rename(oldPath, fileName)
				.then(() => msgAlert.success('Rename Success'))
				.catch(msgAlert.error);
			modal.close();
		}}
	>
		<fieldset class="fieldset">
			<label class="label" for="name"> Rename </label>
			<input
				type="text"
				id="name"
				class="input validator text-base-content outline-none w-full"
				bind:value={fileName}
				oninvalid={(e) => e.preventDefault()}
				required
			/>
			<p class="validator-hint">required</p>
		</fieldset>
		<div class="mt-4 flex gap-1 justify-end items-center">
			<div>
				<button
					type="button"
					class="btn btn-ghost font-normal hover:bg-base-content/30"
					onclick={() => modal.close()}
				>
					Cancel
				</button>
				<button type="submit" class="btn btn-info font-normal"> Rename </button>
			</div>
		</div>
	</form>
{/snippet}
{#snippet itemContextMenu()}
	<ul
		class="menu p-0 bg-base-100 border border-base-content/30 [&_li>*]:rounded-none"
		oncontextmenu={(e) => e.preventDefault()}
	>
		{const Labels = {
			NewFile: 'New File',
			NewFolder: 'New Folder',
			Rename: 'Rename',
			Reveal: 'Reveal In File Manager',
			Delete: 'Delete',
		} as const}
		{const Fields = {
			[Labels.NewFile]: { on: () => Events.Emit(MenuEvent.NewFile), icon: IconFilePlusCorner },
			[Labels.NewFolder]: { on: () => Events.Emit(FEEvent.NewFolder), icon: IconFolderPlus },
			[Labels.Rename]: { on: () => Events.Emit(FEEvent.Rename), icon: IconSquarePen },
			[Labels.Reveal]: { on: () => Events.Emit(FEEvent.Reveal), icon: IconFolderOpenDot },
			[Labels.Delete]: { on: () => Events.Emit(FEEvent.Delete), icon: IconTrash },
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
					<FieldIcon class="size-4" />
					<span>{label}</span>
				</div>
			</button>
		{/snippet}
		{#if fState.selectedItem.type === 'group'}
			<li>{@render btn('New File')}</li>
			<li>{@render btn('New Folder')}</li>
		{/if}
		<li>{@render btn('Rename')}</li>
		<li>{@render btn('Reveal In File Manager')}</li>
		<div class="divider p-0 m-0 h-1"></div>
		<li class="text-error [&>button]:hover:bg-error/10">{@render btn('Delete')}</li>
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
			<IconFolderCog class="size-4" strokeWidth={1} nonScalingStroke />
		{:else}
			<IconFolder class="size-4" strokeWidth={1} nonScalingStroke />
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
			<IconFile class="size-4" strokeWidth={1} nonScalingStroke />
		{/if}
	{/if}
{/snippet}
