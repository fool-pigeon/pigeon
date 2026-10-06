<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts" generics="T">
	let { value = $bindable<T>(), label = 'input', addButton = false } = $props();
</script>

<div class="flex items-end">
	<div>
		<span class="text-accent whitespace-nowrap">{label}:</span>
	</div>
	<input type="text" class="input input-info h-7" bind:value />
	{#if addButton}
		<button class="btn btn-outline btn-info size-7" onclick={() => value--}>-</button>
		<button class="btn btn-outline btn-info size-7" onclick={() => value++}>+</button>
	{/if}
</div>
