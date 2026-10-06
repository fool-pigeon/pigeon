<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	import { json, jsonParseLinter } from '@codemirror/lang-json';
	import CMTextArea from './CMTextArea.svelte';
	import { EditorState } from '@codemirror/state';
	import { linter } from '@codemirror/lint';

	type Prop = {
		value: string;
		readonly?: boolean;
	};
	let { value = $bindable(), readonly = false }: Prop = $props();
	// TODO a powerful syntax highlight config for json
</script>

<CMTextArea
	bind:value
	extensions={[json(), linter(jsonParseLinter()), EditorState.readOnly.of(readonly)]}
/>
