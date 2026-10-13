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
	import { EditorState, type Extension } from '@codemirror/state';
	import { daisyColorTheme, daisyThemeExtension } from './cmConfig';
	import { settings } from '@comp/context/AppSettings.svelte';

	/**
	 * @type value: the bindabled string value of area
	 * @type extensions: extensions of codemirror,default []
	 */
	type Prop = {
		value: string;
		extensions?: Extension;
		readonly?: boolean;
	};
	let { value = $bindable(), extensions = [], readonly = false }: Prop = $props();
</script>

{#key settings.interface.editor.preventWrapLine}
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
					EditorState.readOnly.of(readonly),
					settings.interface.editor.preventWrapLine ? [] : EditorView.lineWrapping,
					EditorView.theme({
						'&': { width: '100%', height: '100%' },
					}),
					daisyThemeExtension,
				],
			});
		}}
	/>
{/key}
