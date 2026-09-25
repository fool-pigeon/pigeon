<script lang="ts">
	import { basicSetup, EditorView } from 'codemirror';
	import CMWrapper from './CMWrapper.svelte';
	import { json } from '@codemirror/lang-json';

	type Prop = {
		value: string;
	};
	let { value = $bindable() }: Prop = $props();
	// TODO a powerful syntax highlight config for json
</script>

<CMWrapper
	bind:value
	viewInitFunc={() => {
		return new EditorView({
			extensions: [
				basicSetup,
				json(),
				EditorView.theme({
					'&': { width: '100%', height: '100%' },
				}),
			],
		});
	}}
/>
