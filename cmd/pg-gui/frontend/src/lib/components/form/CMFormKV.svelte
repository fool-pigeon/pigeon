<script lang="ts">
	// Icons
	import IconListPlus from '@lucide/svelte/icons/list-plus';
	import IconTrash from '@lucide/svelte/icons/trash';

	import type { KV } from '@type/form';
	import equal from 'fast-deep-equal';
	import { onMount, untrack } from 'svelte';
	import CMInput from './CMInput.svelte';
	type kv = {
		k: string;
		v: string;
	};
	type Prop = {
		value: KV;
		kHeader?: string;
		vHeader?: string;
		kReadOnly?: boolean;
	};
	let {
		value = $bindable(),
		kHeader = $bindable('key'),
		vHeader = $bindable('value'),
		kReadOnly = $bindable(false),
	}: Prop = $props();
	let kvs = $state<kv[]>([]);
	let isEditing = $state(false);

	onMount(() => (kvs = marshal(value)));
	function marshal(value: KV): kv[] {
		if (!value) return [];
		return Object.entries(value).map(([k, v]) => ({ k, v: v ?? '' }));
	}
	function unmarshal(kvs: kv[]): KV {
		if (kvs.length <= 0) return null;
		const value: KV = {};
		kvs.forEach(({ k, v }) => {
			value[k] = v;
		});
		return value;
	}
	// kvs->value
	$effect(() => {
		const t = $state.snapshot(kvs);
		if (isEditing) {
			return;
		}
		if (untrack(() => equal(t, marshal(value)))) {
			return;
		}
		value = unmarshal(t);
	});
	// value->kvs
	$effect(() => {
		const v = value;
		if (untrack(() => equal(kvs, marshal(v)))) {
			return;
		}
		kvs = marshal(v);
	});
</script>

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
				<td>
					<input
						class="w-full"
						bind:value={kv.k}
						placeholder={kHeader}
						readonly={kReadOnly}
						onfocus={() => (isEditing = true)}
						onblur={() => (isEditing = false)}
					/>
				</td>
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
