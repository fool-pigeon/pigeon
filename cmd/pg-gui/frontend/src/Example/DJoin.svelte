<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	import DynamicJoin, { NewPanel } from '@lib/components/base/DynamicJoin.svelte';
	import { untrack } from 'svelte';

	function randomInt(n: number): number {
		return Math.floor(Math.random() * n);
	}

	let cubCount = $state(2);
	let panels = $state(Array.from({ length: 2 }, () => NewPanel(cubr)));
	let direction = $state(true);
	let minSize = $state(0);
	$effect(() => {
		if (!cubCount || cubCount < 0) {
			return;
		}
		untrack(() => (panels = Array.from({ length: cubCount }, () => NewPanel(cubr))));
	});
</script>

{#snippet cubr()}
	<div
		class="w-full"
		style:background="rgb({randomInt(255)}, {randomInt(255)}, {randomInt(255)});"
	></div>
{/snippet}

<div class="w-full h-2/3">
	<div class="join">
		<div class="join">
			<input
				type="text"
				min="0"
				placeholder="panel count"
				class="input input-info"
				bind:value={cubCount}
			/>
			<button class="btn" onclick={() => (cubCount = Math.max(0, --cubCount))}>-</button>
			<button class="btn" onclick={() => cubCount++}>+</button>
		</div>
		<div class="join">
			<input type="text" class="input" bind:value={minSize} />
			<select class="select" bind:value={direction}>
				<option selected value="horizontal">H</option>
				<option value="vertical">V</option>
			</select>
		</div>
	</div>
	<DynamicJoin bind:panels bind:isFlexRow={direction} />
</div>
