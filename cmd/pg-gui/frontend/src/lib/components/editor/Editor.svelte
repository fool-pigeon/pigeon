<script lang="ts">
	import { fileExplorer, fsItemDisplay, type FSItem } from '../navigation/FileExplorer.svelte';
	import IconX from '@lucide/svelte/icons/x';
	import IconFileTypeCorner from '@lucide/svelte/icons/file-type-corner';
	import IconForm from '@lucide/svelte/icons/form';
	import IconLayoutCol from '@lucide/svelte/icons/columns-2';
	import IconLayoutRow from '@lucide/svelte/icons/rows-2';
	import IconEllipsis from '@lucide/svelte/icons/ellipsis';
	import IconChevronRight from '@lucide/svelte/icons/chevron-right';
	import IconSaveAll from '@lucide/svelte/icons/save-all';
	import ScrollFlex from '../base/ScrollFlex.svelte';
	import { ReadFile } from '@bindings/service/fileservice';
	import FileView from './FileView.svelte';
	type RawFile = {
		// original file content
		initValue: string;
		// current file content
		value: string;
		// meta of file
		fs: FSItem;
	};
	let topItems = $state<FSItem[]>([]);
	let rawFile = $state('');
	let isRawFileMode = $state(false);
	let isLayoutRow = $state(true);
	let routes = $derived(fileExplorer.findRoutes(fileExplorer.getActivedFS()?.id));
	let activedFS = $derived(fileExplorer.getActivedFS());
	$effect(() => {
		if (!activedFS || activedFS.type === 'group') {
			return;
		}
		ReadFile(activedFS.id)
			.then((val) => (rawFile = val))
			.catch(console.error);
		if (topItems.some((fs) => fs.id === activedFS.id)) {
			return;
		}
		topItems.push(activedFS);
	});
</script>

<div class="w-full h-full flex flex-col">
	{#if !activedFS || activedFS.type === 'group'}
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
						{#each topItems as item, i (item.id)}
							<div
								role="button"
								tabindex="0"
								class="p-1 flex gap-0.5 items-center cursor-pointer
							text-base-content/60 hover:text-primary hover:bg-base-300
							{fileExplorer.getActivedFS()?.id === item.id ? 'bg-primary/30' : ''}"
								onkeydown={() => {}}
								onclick={() => {
									fileExplorer.setSelectedFS(item);
									fileExplorer.expandToItem(item);
								}}
							>
								{@render fsItemDisplay(item)}
								<button
									class="p-1 size-5 btn btn-ghost btn-sm hover:btn-error"
									onclick={(e) => {
										e.stopPropagation();
										if (fileExplorer.getActivedFS()?.id === item.id) {
											fileExplorer.setActivedFS(i > 0 ? topItems[i - 1] : topItems[i + 1]);
										}
										topItems.splice(i, 1);
									}}
								>
									<IconX class="size-5" />
								</button>
							</div>
						{/each}
					{/snippet}
				</ScrollFlex>
				<!-- top function btns -->
				<!-- TODO: complete functions for top btns -->
				<div class="flex border border-primary/40">
					<button
						title={'save'}
						class="size-8 p-0 flex justify-center items-center hover:bg-base-content/10"
						onclick={() => {}}
					>
						<IconSaveAll class="size-5" />
					</button>
					<label
						title={isRawFileMode ? 'To form mode' : 'To file mode'}
						class="size-8 swap swap-rotate hover:bg-base-content/10"
					>
						<input type="checkbox" bind:checked={isRawFileMode} />
						<div class="swap-on">
							<IconForm class="size=5" />
						</div>
						<div class="swap-off">
							<IconFileTypeCorner class="size=5" />
						</div>
					</label>
					<label
						title={isLayoutRow ? 'To column layout' : 'To row layout'}
						class="size-8 swap swap-rotate hover:bg-base-content/10"
					>
						<input type="checkbox" bind:checked={isLayoutRow} />
						<div class="swap-on">
							<IconLayoutRow class="size=5" />
						</div>
						<div class="swap-off">
							<IconLayoutCol class="size=5" />
						</div>
					</label>
					<button
						title={'more'}
						class="size-8 p-0 flex justify-center items-center hover:bg-base-content/10"
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
										fileExplorer.setSelectedFS(r);
										fileExplorer.expandToItem(r);
									}}
								>
									{@render fsItemDisplay(r)}
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
		<div class="grow">
			{#if isRawFileMode}
				<FileView bind:value={rawFile} />
			{:else}
				<div>Form Mode</div>
				<div>{JSON.stringify(topItems, null, 2)}</div>
				<div>{JSON.stringify(fileExplorer.getActivedFS(), null, 2)}</div>
			{/if}
		</div>
	{/if}
</div>
