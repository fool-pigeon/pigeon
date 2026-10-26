<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	// Icons
	import IconKeyRound from '@lucide/svelte/icons/key-round';
	import IconEye from '@lucide/svelte/icons/eye';
	import IconEyeClosed from '@lucide/svelte/icons/eye-closed';

	type Prop = { pwd: string; placeholder?: string };
	let { pwd = $bindable(), placeholder }: Prop = $props();
</script>

<div class="w-full flex">
	{let isShow = $state(false)}
	<IconKeyRound class="shrink-0 size-4" />
	<input
		class="grow input"
		type={isShow ? 'text' : 'password'}
		placeholder={placeholder ?? 'password'}
		bind:value={pwd}
	/>
	<label class="swap shrink-0">
		<input type="checkbox" bind:checked={isShow} />
		<IconEye class="swap-on size-4" />
		<IconEyeClosed class="swap-off size-4" />
	</label>
</div>
