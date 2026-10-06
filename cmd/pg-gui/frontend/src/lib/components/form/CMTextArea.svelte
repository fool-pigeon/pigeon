<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

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
	import { daisyColorTheme, daisyThemeExtension } from './cmConfig';

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
				// prevent drag selection
				EditorView.domEventHandlers({
					dragstart: (event) => {
						event.preventDefault();
						return true;
					},
				}),
				EditorView.theme({
					'&': { width: '100%', height: '100%' },
				}),
				daisyThemeExtension,
			],
		});
	}}
/>
