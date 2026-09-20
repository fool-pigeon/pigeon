<script lang="ts">
	import { Icon } from '@lucide/svelte';
	import { fileExplorer, fsItemDisplay, type FSItem } from '../navigation/FileExplorer.svelte';
	import IconX from '@lucide/svelte/icons/x';
	import ScrollFlex from '../base/ScrollFlex.svelte';
	let topItems = $state<FSItem[]>([]);
	$effect(() => {
		const activedFS = fileExplorer.getActivedFS();
		if (!activedFS || activedFS.type === 'group') {
			return;
		}
		if (topItems.some((fs) => fs.id === activedFS.id)) {
			return;
		}
		topItems.push(activedFS);
	});
	let scrollDiv = $state<HTMLDivElement>();
</script>

<div class="w-full h-full flex flex-col">
	<!-- top bar -->
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
						onclick={() => fileExplorer.setActivedFS(item)}
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
		<div class="flex whitespace-nowrap">
			<button class="border">file-view</button>
			<button class="border">change-direction</button>
			<button class="border">more</button>
		</div>
	</div>
	<!-- breadcrumb -->
	<div>breadcumb</div>
	<!-- content -->
	<div class="overflow-auto">content</div>
	<div>{JSON.stringify(topItems, null, 2)}</div>
	<div>{JSON.stringify(fileExplorer.getActivedFS(), null, 2)}</div>
</div>
