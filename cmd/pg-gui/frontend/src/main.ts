// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { mount } from 'svelte'
import App from './App.svelte'

if (import.meta.env.DEV) {
    const el=document.createElement('div')
    el.id='example-component-root'
    const {default:Exmaple}=await import('./Example/Example.svelte')
    mount(Exmaple,{target:el})
    document.body.insertBefore(el,document.querySelector("#app"))
}
// biome-ignore lint/style/noNonNullAssertion: false
mount(App, { target: document.getElementById('app')! })
