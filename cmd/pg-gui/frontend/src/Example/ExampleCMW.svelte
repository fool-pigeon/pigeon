<script lang="ts">
	import CMWrapper from '@lib/components/form/CMWrapper.svelte';
	import { basicSetup, EditorView } from 'codemirror';

	let value = $state('hellowrold');
</script>

<button class="btn btn-info btn-soft" onclick={() => (value = 'helloworld')}>reset-value</button>
<div class="text-2xl">actually: {value}</div>
<CMWrapper
	bind:value
	viewInitFunc={() => {
		return new EditorView({
			extensions: [basicSetup],
		});
	}}
/>
