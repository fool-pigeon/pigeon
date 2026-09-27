<script lang="ts">
	import { basicSetup, EditorView } from 'codemirror';
	import CMWrapper from './CMWrapper.svelte';

	type Prop = {
		value: string;
	};
	let { value = $bindable() }: Prop = $props();
</script>

<CMWrapper
	bind:value
	viewInitFunc={() => {
		return new EditorView({
			extensions: [
				basicSetup,
				EditorView.theme({
					'&': { width: '100%', height: '100%' },
				}),
			],
		});
	}}
/>
