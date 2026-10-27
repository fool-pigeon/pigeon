<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	// Icons
	import type { KArr, KV } from '@type/form';

	import { onMount, untrack } from 'svelte';
	import CMInput from './CMInput.svelte';
	import equal from 'fast-deep-equal';
	import { regex } from '@lib/utils/regex';
	type UrlPart = {
		uri: string;
		query: KArr;
	};
	type Prop = { path?: KV } & UrlPart;
	let { uri = $bindable(), query = $bindable(), path = $bindable() }: Prop = $props();
	let url = $state('');
	onMount(() => (url = marshal({ uri, query })));
	function marshal({ uri, query }: UrlPart): string {
		let u = uri;
		if (query && Object.keys(query).length > 0) {
			u += '?';
			u += Object.entries(query)
				.flatMap(([k, arr]) => (arr ?? ['']).map((v) => `${k}=${v}`))
				.join('&');
		}
		return u;
	}
	function unmarshal(u: string): UrlPart {
		const splitQuery = u.indexOf('?');
		const uri = splitQuery < 0 ? u : u.slice(0, splitQuery);
		if (splitQuery < 0) return { uri: uri, query: null };

		const qlist = u
			.slice(splitQuery + 1)
			.split('&')
			.filter((v) => v.length > 0);
		if (qlist.length <= 0) return { uri: uri, query: null };

		const query: KArr = {};
		for (const kv of qlist) {
			const i = kv.indexOf('=');
			const k = i < 0 ? kv : kv.slice(0, i);
			const v = i < 0 ? '' : kv.slice(i + 1);
			query[k] ??= [];
			query[k].push(v);
		}
		return { uri: uri, query };
	}
	function equalURL(u1: string, u2: string): boolean {
		const i1 = u1.indexOf('?');
		const i2 = u2.indexOf('?');
		const uri1 = i1 === -1 ? u1 : u1.slice(0, i1);
		const uri2 = i2 === -1 ? u2 : u2.slice(0, i2);
		const query1 = i1 === -1 ? '' : u1.slice(i1 + 1);
		const query2 = i2 === -1 ? '' : u2.slice(i2 + 1);
		return uri1 === uri2 && query1 === query2;
	}
	// url->uri,query
	$effect(() => {
		const u = url;
		if (untrack(() => equalURL(u, marshal({ uri: uri, query })))) return;
		untrack(() => {
			const data = unmarshal(u);
			uri = data.uri;
			query = data.query;
			// refresh path params
			if (path !== undefined) {
				const pathParams = Array.from(regex.urlPathParams(u));
				if (pathParams.length <= 0) {
					path = null;
				} else {
					const newPath: KV = {};
					for (const p of pathParams) {
						newPath[p] = path?.[p] ?? '';
					}
					path = newPath;
				}
			}
		});
	});
	// uri,query->url
	$effect(() => {
		const p: UrlPart = { uri: uri, query: $state.snapshot(query) };
		if (untrack(() => equal(p, unmarshal(url)))) return;
		untrack(() => (url = marshal(p)));
	});
</script>

<CMInput bind:value={url} placeholder={'URL'} />
