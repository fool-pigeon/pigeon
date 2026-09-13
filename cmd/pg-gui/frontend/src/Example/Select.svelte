<script lang="ts" generics="T">
	type option<T> = { value: T; label?: string };
	let {
		label = 'select',
		value = $bindable<T>(),
		options = $bindable<option<T>[]>([] as option<T>[])
	} = $props();
</script>

<div class="flex items-end">
	<div>
		<span class="text-accent whitespace-nowrap">{label}:</span>
	</div>
	<select class="select select-info h-6" bind:value>
		{#each options as option, i}
			<option selected={i === 0} value={option.value}>
				{#if option.label}
					{option.label}
				{:else}
					{option.value}
				{/if}
			</option>
		{/each}
	</select>
</div>
