<script lang="ts" module>
	export const WrapperMeta = {
		external: Annotation.define<boolean>(),
	};
</script>

<script lang="ts">
	import { Annotation, StateEffect, Transaction } from '@codemirror/state';
	import { EditorView } from '@codemirror/view';
	import { onMount } from 'svelte';

	/**
	 * this is a wrapper component of codemirror
	 * bind:value
	 * use viewInitFunc to inject a codemirror view
	 */
	type Prop = {
		value: string;
		viewInitFunc: () => EditorView;
		enableExternalHistory?: boolean;
	};

	let {
		value = $bindable(),
		viewInitFunc,
		enableExternalHistory = $bindable(false),
	}: Prop = $props();
	let view = $state<EditorView>();
	let el: HTMLDivElement;

	onMount(() => {
		view = viewInitFunc();
		// inject state listener to update state.doc to value
		view.dispatch({
			effects: [
				StateEffect.appendConfig.of(
					EditorView.updateListener.of((u) => {
						if (u.docChanged) {
							value = u.state.doc.toString();
						}
					}),
				),
			],
		});
		el.appendChild(view.dom);

		return () => {
			view?.destroy();
		};
	});
	$effect(() => {
		if (!view) {
			return;
		}
		if (view.state.doc.toString() !== value) {
			setDocExternally(value);
		}
	});
	function setDocExternally(text: string) {
		view?.dispatch({
			changes: { from: 0, to: view.state.doc.length, insert: text },
			annotations: [
				WrapperMeta.external.of(true),
				Transaction.addToHistory.of(enableExternalHistory),
			],
		});
	}
</script>

<div class="w-full h-full" bind:this={el}></div>
