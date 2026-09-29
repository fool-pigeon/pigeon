<script lang="ts">
	import { json } from '@codemirror/lang-json';
	import CMTextArea from './CMTextArea.svelte';
	import { EditorState } from '@codemirror/state';

	type Prop = {
		value: string;
		readonly?: boolean;
	};
	let { value = $bindable(), readonly = false }: Prop = $props();
	// TODO a powerful syntax highlight config for json
</script>

<CMTextArea bind:value extensions={[json(), EditorState.readOnly.of(readonly)]} />
