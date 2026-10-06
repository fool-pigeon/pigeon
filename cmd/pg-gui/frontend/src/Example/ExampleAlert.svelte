<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	import {
		type hPosition,
		msgAlert,
		type vPosition,
	} from '@lib/components/message/AlertNotify.svelte';
	import Select from './Select.svelte';

	let h = $state<hPosition>('center');
	let v = $state<vPosition>('top');
	$effect(() => {
		msgAlert.setHPos(h);
		msgAlert.setVPos(v);
	});
	const msg = 'this is a msg';
</script>

<div class="join join-vertical">
	<button class="btn btn-info" onclick={() => msgAlert.info(msg)}>info</button>
	<button class="btn btn-success" onclick={() => msgAlert.success(msg)}>success</button>
	<button class="btn btn-warning" onclick={() => msgAlert.warn(msg)}>warn</button>
	<button class="btn btn-error" onclick={() => msgAlert.error(msg)}>error</button>
	<Select
		label="h"
		bind:value={h}
		options={[{ value: 'center' }, { value: 'left' }, { value: 'right' }]}
	/>
	<Select
		label="v"
		bind:value={v}
		options={[{ value: 'center' }, { value: 'top' }, { value: 'bottom' }]}
	/>
</div>
