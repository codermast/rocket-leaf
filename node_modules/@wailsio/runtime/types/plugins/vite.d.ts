import type { Plugin } from 'vite';
/**
 * A plugin that extends the wails runtime with locally generated code
 * to provide support for typed custom events.
 * With the plugin installed, vite will fail to build the project
 * unless wails bindings have been generated first.
 *
 * @param {string} [bindingsRoot] - The root import path for generated bindings
 */
export default function WailsTypedEvents(bindingsRoot: string): Plugin;
