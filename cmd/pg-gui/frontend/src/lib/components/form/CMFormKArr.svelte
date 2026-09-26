<script lang="ts">
	// Icons
	import IconListPlus from '@lucide/svelte/icons/list-plus';
	import IconTrash from '@lucide/svelte/icons/trash';

	import type { KArr } from '@type/form';
	import equal from 'fast-deep-equal';
	import { onMount, untrack } from 'svelte';
	import CMInput from './CMInput.svelte';
	type kv = {
		k: string;
		v: string;
	};
	type Prop = {
		value: KArr;
		kHeader?: string;
		vHeader?: string;
	};
	let {
		value = $bindable(),
		kHeader = $bindable('key'),
		vHeader = $bindable('key'),
	}: Prop = $props();
	let kvs = $state<kv[]>([]);
	onMount(() => (kvs = marshal(value)));
	function marshal(value: KArr): kv[] {
		if (!value) return [];
		return Object.entries(value).flatMap(([k, arr]) => (arr ?? ['']).map((v) => ({ k, v })));
	}
	function unmarshal(kvs: kv[]): KArr {
		if (kvs.length <= 0) return null;
		const value: KArr = {};
		kvs.forEach(({ k, v }) => {
			value[k] ??= [];
			value[k].push(v);
		});
		return value;
	}
	// kvs->value
	$effect(() => {
		const t = $state.snapshot(kvs);
		if (untrack(() => equal(t, marshal(value)))) {
			return;
		}
		value = unmarshal(t);
	});
	// value->kvs
	$effect(() => {
		const v = value;
		if (untrack(() => equal(v, unmarshal(kvs)))) {
			return;
		}
		kvs = marshal(v);
	});
</script>

<!-- TEST viewer -->
<!-- <div class="text-success w-full break-all">{JSON.stringify(value, null, 2)}</div>
<div class="text-info w-full break-all">{JSON.stringify(kvs, null, 2)}</div> -->

<table class="w-full table-fixed border border-base-content/10">
	<thead class="w-full">
		<tr class="[&>th]:border [&>th]:border-base-content/10">
			<th class="w-1/3">{kHeader}</th>
			<th class="grow">{vHeader}</th>
			<th class="p-0 w-8 has-[button:hover]:bg-info">
				<button
					title="add"
					class="w-full flex items-center justify-center cursor-pointer text-base-content"
					onclick={() => kvs.push({ k: '', v: '' })}
				>
					<IconListPlus class="size-5" />
				</button>
			</th>
		</tr>
	</thead>
	<tbody class="w-full">
		{#each kvs as kv, i}
			<tr class="[&>td]:border [&>td]:border-base-content/10">
				<td><input class="w-full" placeholder={kHeader} bind:value={kv.k} /></td>
				<td><CMInput bind:value={kv.v} placeholder={vHeader} /></td>
				<td class="has-[button:hover]:bg-error">
					<button
						title="delete"
						class="w-full h-full flex items-center justify-center cursor-pointer text-base-content/80 hover:text-base-content"
						onclick={() => kvs.splice(i, 1)}
					>
						<IconTrash class="size-5" />
					</button>
				</td>
			</tr>
		{/each}
	</tbody>
</table>
