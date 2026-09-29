<!-- @component
a text editor of codemirror basicSetup.
the value will be bindabled.

@example
```svelte
<CMTextArea bind:value extensions={...} />
```
-->
<script lang="ts">
	import { basicSetup, EditorView } from 'codemirror';
	import CMWrapper from './CMWrapper.svelte';
	import type { Extension } from '@codemirror/state';

	/**
	 * @type value: the bindabled string value of area
	 * @type extensions: extensions of codemirror,default []
	 */
	type Prop = {
		value: string;
		extensions?: Extension;
	};
	let { value = $bindable(), extensions = [] }: Prop = $props();
</script>

<CMWrapper
	bind:value
	viewInitFunc={() => {
		return new EditorView({
			extensions: [
				basicSetup,
				extensions,
				EditorView.theme({
					'&': { width: '100%', height: '100%' },
				}),
			],
		});
	}}
/>
