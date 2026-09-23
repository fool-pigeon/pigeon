<script lang="ts" module>
	export type OpenedFS = {
		fs: FSItem;
		saved: boolean;
	};
	class EditorState {
		// state of opened files
		ofs = $state<OpenedFS[]>([]);
		// current fs
		current = $state<OpenedFS>();

		reload() {
			GetOpenedState()
				.then((val) => {
					this.ofs = [];
					for (const v of val ?? []) {
						const fs = fileExplorer.find(v.path);
						if (!fs) {
							Close(v.path);
							continue;
						}
						this.ofs.push({ fs, saved: v.saved });
					}
				})
				.catch(msgAlert.error);
		}
		close(index: number) {
			const of = this.ofs[index];
			Close(of.fs.id)
				.then(() => {
					if (activedFS && activedFS.id === of.fs.id) {
						this.selectCurrent(index > 0 ? this.ofs[index - 1] : this.ofs[index + 1]);
					}
					this.ofs.splice(index, 1);
				})
				.catch(msgAlert.error);
		}
		includesPath(path: string): boolean {
			return this.ofs.some((v) => v.fs.id === path);
		}
		find(path: string): OpenedFS | undefined {
			return this.ofs.filter((v) => v.fs.id === path).at(0);
		}
		selectCurrent(of: OpenedFS | undefined) {
			this.current = of;
			if (!of) {
				fileExplorer.setActivedFS(undefined);
				return;
			}
			fileExplorer.expandToItem(of.fs);
			fileExplorer.setSelectedFS(of.fs);
		}
	}
	const eState = new EditorState();
	let isRawFileMode = $state(false);
	let isLayoutRow = $state(true);
	let activedFS = $derived(fileExplorer.getActivedFS());
	let routes = $derived(fileExplorer.findRoutes(activedFS?.id));
</script>

<script lang="ts">
	import {
		fileExplorer,
		fsItemDisplay as fsDisplay,
		type FSItem,
	} from '../navigation/FileExplorer.svelte';
	import ScrollFlex from '../base/ScrollFlex.svelte';
	import FileView from './FileView.svelte';
	import {
		Close,
		CloseAll,
		GetOpenedState,
		Open,
		Save,
		SaveAll,
	} from '@bindings/service/fileservice';
	// Icons
	import IconCircle from '@lucide/svelte/icons/circle';
	import IconX from '@lucide/svelte/icons/x';
	import IconFileTypeCorner from '@lucide/svelte/icons/file-type-corner';
	import IconForm from '@lucide/svelte/icons/form';
	import IconLayoutCol from '@lucide/svelte/icons/columns-2';
	import IconLayoutRow from '@lucide/svelte/icons/rows-2';
	import IconEllipsis from '@lucide/svelte/icons/ellipsis';
	import IconChevronRight from '@lucide/svelte/icons/chevron-right';
	import IconSaveAll from '@lucide/svelte/icons/save-all';
	import IconTriangleAlert from '@lucide/svelte/icons/triangle-alert';
	import { onDestroy, onMount, untrack } from 'svelte';
	import { msgAlert } from '../message/AlertNotify.svelte';
	import { modal } from '../base/Modal.svelte';
	import { Events } from '@wailsio/runtime';
	import { MenuEvent } from '@bindings/gui';
	import { DataEvent } from '@lib/enum/events';
	onMount(() => {
		Events.Once(DataEvent.FileExplorerLoaded, () => eState.reload());
	});
	onDestroy(() => CloseAll());
	$effect(() => {
		if (!activedFS || activedFS.type === 'group') {
			return;
		}
		if (untrack(() => eState.includesPath(activedFS.id))) {
			const of = eState.find(activedFS.id);
			if (of && eState.current !== of) {
				eState.selectCurrent(of);
			}
			return;
		}
		Open(activedFS.id)
			.then(() => {
				eState.ofs.push({ fs: activedFS, saved: true });
				eState.selectCurrent({ fs: activedFS, saved: true });
			})
			.catch(msgAlert.error);
	});
</script>

<div class="w-full h-full flex flex-col">
	{#if eState.ofs.length <= 0}
		<!-- TODO: user tips in editor -->
		<div class="w-full h-full flex justify-center items-center text-2xl">
			<div>Some tips like shortcuts and description here</div>
		</div>
	{:else}
		<div class="select-none">
			<!-- top bar nav -->
			<div class="flex gap-1 justify-between">
				<ScrollFlex>
					{#snippet content()}
						{#each eState.ofs as of, i (of.fs.id)}
							<div
								role="button"
								tabindex="0"
								class="p-1 flex gap-0.5 items-center cursor-pointer
									text-base-content/60 hover:text-primary hover:bg-base-300
									{activedFS && activedFS.id === of.fs.id ? 'bg-primary/30' : ''}"
								onkeydown={() => {}}
								onclick={() => eState.selectCurrent(of)}
							>
								{@render fsDisplay(of.fs)}
								{#snippet UnsavedChange()}
									<div class="p-2 flex flex-col">
										<div class="p-1 flex items-center">
											<IconTriangleAlert class="size-8 text-warning" />
											<span class="text-2xl">Wait a moment..</span>
										</div>
										<div>You have unsaved changes in this file.</div>
										<div class="mt-4 flex justify-between">
											<button
												class="rounded-none btn btn-error p-2"
												onclick={() => {
													eState.close(i);
													modal.close();
												}}
											>
												Not Save
											</button>
											<div class="flex gap-1">
												<button
													class="rounded-none btn btn-ghost p-2 hover:bg-base-content/30"
													onclick={() => {
														modal.close();
													}}
												>
													Cancel
												</button>
												<button
													class="rounded-none btn btn-info p-2 hover:btn-success"
													onclick={() => {
														Events.Emit(MenuEvent.Save);
													}}
												>
													Save
												</button>
											</div>
										</div>
									</div>
								{/snippet}
								<button
									class="p-0 size-5 btn btn-ghost btn-sm text-base-content/30 hover:btn-error group"
									onclick={(e) => {
										e.stopPropagation();
										if (of.saved) {
											eState.close(i);
										} else {
											modal.open(UnsavedChange, 'Unsaved Change');
										}
									}}
								>
									{#if of.saved}
										<IconX class="size-5" />
									{:else}
										<IconCircle class="p-1 size-5 group-hover:hidden fill-current border-none" />
										<IconX class="size-5 hidden group-hover:inline" />
									{/if}
								</button>
							</div>
						{/each}
					{/snippet}
				</ScrollFlex>
				<!-- top function btns -->
				<!-- TODO: complete functions for top btns -->
				<div class="flex border border-primary/40">
					<button
						title={'Save all'}
						class="size-8 p-0 flex justify-center items-center cursor-pointer hover:bg-base-content/10"
						onclick={() => {
							SaveAll()
								.then(() => {
									eState.reload();
									msgAlert.success('Saved all');
								})
								.catch(msgAlert.error);
						}}
					>
						<IconSaveAll class="size-5" />
					</button>
					<label
						title={isRawFileMode ? 'To form mode' : 'To file mode'}
						class="size-8 swap swap-rotate hover:bg-base-content/10"
					>
						<input type="checkbox" bind:checked={isRawFileMode} />
						<div class="swap-on">
							<IconForm class="size-5" />
						</div>
						<div class="swap-off">
							<IconFileTypeCorner class="size-5" />
						</div>
					</label>
					<label
						title={isLayoutRow ? 'To column layout' : 'To row layout'}
						class="size-8 swap swap-rotate hover:bg-base-content/10"
					>
						<input type="checkbox" bind:checked={isLayoutRow} />
						<div class="swap-on">
							<IconLayoutRow class="size-5" />
						</div>
						<div class="swap-off">
							<IconLayoutCol class="size-5" />
						</div>
					</label>
					<button
						title={'more'}
						class="size-8 p-0 flex justify-center items-center cursor-pointer hover:bg-base-content/10"
						onclick={() => {}}
					>
						<IconEllipsis class="size-5" />
					</button>
				</div>
			</div>
			<!-- breadcrumb -->
			<div class="pt-1 pl-1 border-b border-primary/50">
				<ScrollFlex>
					{#snippet content()}
						{#each routes as r, i (r.id)}
							<div class="flex items-baseline">
								<button
									class="hover:underline hover:cursor-pointer"
									onclick={() => {
										fileExplorer.expandToItem(r);
										fileExplorer.setSelectedFS(r);
									}}
								>
									{@render fsDisplay(r)}
								</button>
								{#if i < routes.length - 1}
									<IconChevronRight class="size-4 text-base-content/70" />
								{/if}
							</div>
						{/each}
					{/snippet}
				</ScrollFlex>
			</div>
		</div>
		<!-- content -->
		<!-- TODO: fill content of editor -->
		{#if !eState.current}{:else}
			<div class="grow">
				{#key eState.current.fs.id}
					{#if isRawFileMode}
						<FileView bind:openedFS={eState.current} />
					{:else}
						<div>Form Mode</div>
						<div>{JSON.stringify(eState.ofs, null, 2)}</div>
						<div>{JSON.stringify(eState.current, null, 2)}</div>
					{/if}
				{/key}
			</div>
		{/if}
	{/if}
</div>
