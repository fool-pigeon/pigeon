<script lang="ts">
	import TreeMenu, {
		type IndentGuideType,
		NewTreeGroup,
		NewTreeLeaf,
		type TreeItem,
	} from '@lib/components/base/TreeMenu.svelte';
	import File from '@lucide/svelte/icons/file';
	import Folder from '@lucide/svelte/icons/folder';
	import FolderOpen from '@lucide/svelte/icons/folder-open';
	import Input from './Input.svelte';
	import Select from './Select.svelte';

	let items = $state([
		NewTreeLeaf('1', 'file', {
			snippet: fileItem,
			on: {
				onclick: () => (msg = 'file:click'),
				onauxclick: () => (msg = 'file:auxclick'),
				oncontextmenu: (e) => {
					e.preventDefault();
					msg = 'file:contextmenu';
				},
			},
		}),
		NewTreeGroup('2', 'folder', {
			isOpen: true,
			snippet: folderItem,
			next: [
				NewTreeLeaf('20', 'f0', { snippet: fileItem }),
				NewTreeGroup('f2', 'folder2', {
					isOpen: true,
					next: [NewTreeLeaf('21', 'f1', { snippet: fileItem }), NewTreeLeaf('22', 'f2')],
				}),
			],
			on: {
				ontoggle: () => {
					msg = 'folder:toggle';
				},
				oncontextmenu: () => {
					msg = 'folder:contextmenu';
				},
			},
		}),
		NewTreeLeaf('3', 'music'),
	]);
	let indentSize = $state(5);
	let indentGuide = $state<IndentGuideType>('always');
	let itemClass = $state('border border-dashed border-transparent hover:border-info/20');
	let itemStyle = $state('');
	let msg = $state('');
</script>

<div class="join max-w-xs join-vertical w-full h-full">
	<Input label="item-class" bind:value={itemClass} />
	<Input label="item-style" bind:value={itemStyle} />
	<Input label="indent" bind:value={indentSize} addButton={true} />
	<Select
		label="guide"
		bind:value={indentGuide}
		options={[{ value: 'always' }, { value: 'hover' }, { value: 'none' }]}
	/>
	<div>msg:<span class="text-success">{msg}</span></div>
	<div class="w-md h-auto overflow-auto">
		<TreeMenu bind:items bind:indentSize bind:indentGuide />
	</div>
</div>

{#snippet fileItem(item: TreeItem)}
	<div class="flex gap-1 items-center"><File class="w-4" />{item.name}</div>
{/snippet}
{#snippet folderItem(item: TreeItem)}
	<div class="flex items-center gap-1">
		{#if item.isOpen}
			<FolderOpen class="w-4" />
		{:else}
			<Folder class="w-4" />
		{/if}
		{item.name}
	</div>
{/snippet}
