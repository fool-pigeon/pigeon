<script lang="ts">
	import TestTubeDiagonal from '@lucide/svelte/icons/test-tube-diagonal';
	import ExampleDJoin from './ExampleDJoin.svelte';

	let showName = $state<string>('');
</script>

<div class="absolute bottom-2 w-full">
	<div class="drawer">
		<input id="my-drawer-1" type="checkbox" class="drawer-toggle" />
		<div class="drawer-content">
			<div class="w-full flex justify-center">
				<label for="my-drawer-1" class="btn">
					<TestTubeDiagonal />
				</label>
			</div>
		</div>
		<div class="drawer-side">
			<label for="my-drawer-1" aria-label="close sidebar" class="drawer-overlay"></label>
			<div class="join h-full">
				<ul class="menu bg-base-200 min-h-full p-4">
					{@render item('Example')}
					{@render item('DJoin')}
				</ul>
				<div class="min-w-md bg-base-300">
					{#if showName === 'DJoin'}
						<ExampleDJoin />
					{:else}
						<div class="w-20 h-40 bg-info">Hello World</div>
					{/if}
				</div>
			</div>
		</div>
	</div>
</div>

{#snippet item(name: string)}
	<li>
		<button
			onclick={() => {
				showName = name;
			}}>{name}</button
		>
	</li>
{/snippet}
