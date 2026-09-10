import { mount } from 'svelte'
import App from './App.svelte'

// biome-ignore lint/style/noNonNullAssertion: false
mount(App, { target: document.getElementById('app')! })
