<script lang="ts">
	import CMFormKArr from '@comp/form/CMFormKArr.svelte';
	import CMFormKV from '@comp/form/CMFormKV.svelte';
	import type { KV, KArr } from '@type/form';

	let karr = $state<KArr>({ a: ['1', '2'] });
	let kv = $state<KV>({ a: '1', b: '1' });
	let fator = $state(false);
</script>

<button class="btn btn-info" onclick={() => (fator = !fator)}>reload</button>
{#key fator}
	<div class="text-info">{JSON.stringify(karr, null, 2)}</div>
	<div class="w-1/3">
		<CMFormKArr bind:value={karr} kHeader="Key" vHeader="Value" />
	</div>
	<div class="text-info">{JSON.stringify(kv, null, 2)}</div>
	<div class="w-1/3">
		<CMFormKV bind:value={kv} kHeader="KV_K" vHeader="KV_V" />
	</div>
{/key}
