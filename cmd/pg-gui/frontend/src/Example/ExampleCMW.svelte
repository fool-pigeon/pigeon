<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	import CMInput from '@lib/components/form/CMInput.svelte';
	import CMJson from '@lib/components/form/CMJson.svelte';

	let value = $state('{"hello":"world"}');
</script>

<button class="btn btn-info btn-soft" onclick={() => (value = 'helloworld')}>reset-value</button>
<div class="text-2xl">actually: {value}</div>
<div class="w-1/2">
	<CMInput bind:value placeholder={'input'} />
</div>
<div class="w=1/2">
	<CMJson bind:value />
</div>
