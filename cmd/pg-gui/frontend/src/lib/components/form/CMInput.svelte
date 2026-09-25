<script lang="ts">
	import { EditorView } from 'codemirror';
	import CMWrapper from './CMWrapper.svelte';
	import { keymap, placeholder } from '@codemirror/view';
	import { defaultKeymap } from '@codemirror/commands';

	type Prop = {
		value: string;
		placeholder?: string;
	};
	let { value = $bindable(), placeholder: p }: Prop = $props();
	// TODO add cm_extension for highlight different env params
</script>

<CMWrapper
	bind:value
	viewInitFunc={() => {
		return new EditorView({
			extensions: [
				keymap.of(defaultKeymap),
				placeholder(p ?? ''),
				EditorView.theme({
					'&': { height: '100%' },
				}),
			],
		});
	}}
/>
