import { Types } from "./event_types.js";
export * from "./event_types.js";
/**
 * A table of data types for all known events.
 * Will be monkey-patched by the binding generator.
 */
export interface CustomEvents {
}
/**
 * Either a known event name or an arbitrary string.
 */
export type WailsEventName<E extends keyof CustomEvents = keyof CustomEvents> = E | (string & {});
/**
 * Union of all known system event names.
 */
type SystemEventName = {
    [K in keyof (typeof Types)]: (typeof Types)[K][keyof ((typeof Types)[K])];
} extends (infer M) ? M[keyof M] : never;
/**
 * The data type associated to a given event.
 */
export type WailsEventData<E extends WailsEventName = WailsEventName> = E extends keyof CustomEvents ? CustomEvents[E] : (E extends SystemEventName ? void : any);
/**
 * The type of handlers for a given event.
 */
export type WailsEventCallback<E extends WailsEventName = WailsEventName> = (ev: WailsEvent<E>) => void;
/**
 * Represents a system event or a custom event emitted through wails-provided facilities.
 */
export declare class WailsEvent<E extends WailsEventName = WailsEventName> {
    /**
     * The name of the event.
     */
    name: E;
    /**
     * Optional data associated with the emitted event.
     */
    data: WailsEventData<E>;
    /**
     * Name of the originating window. Omitted for application events.
     * Will be overridden if set manually.
     */
    sender?: string;
    constructor(name: E, data: WailsEventData<E>);
    constructor(name: WailsEventData<E> extends null | void ? E : never);
}
/**
 * Register a callback function to be called multiple times for a specific event.
 *
 * @param eventName - The name of the event to register the callback for.
 * @param callback - The callback function to be called when the event is triggered.
 * @param maxCallbacks - The maximum number of times the callback can be called for the event. Once the maximum number is reached, the callback will no longer be called.
 * @returns A function that, when called, will unregister the callback from the event.
 */
export declare function OnMultiple<E extends WailsEventName = WailsEventName>(eventName: E, callback: WailsEventCallback<E>, maxCallbacks: number): () => void;
/**
 * Registers a callback function to be executed when the specified event occurs.
 *
 * @param eventName - The name of the event to register the callback for.
 * @param callback - The callback function to be called when the event is triggered.
 * @returns A function that, when called, will unregister the callback from the event.
 */
export declare function On<E extends WailsEventName = WailsEventName>(eventName: E, callback: WailsEventCallback<E>): () => void;
/**
 * Registers a callback function to be executed only once for the specified event.
 *
 * @param eventName - The name of the event to register the callback for.
 * @param callback - The callback function to be called when the event is triggered.
 * @returns A function that, when called, will unregister the callback from the event.
 */
export declare function Once<E extends WailsEventName = WailsEventName>(eventName: E, callback: WailsEventCallback<E>): () => void;
/**
 * Removes event listeners for the specified event names.
 *
 * @param eventNames - The name of the events to remove listeners for.
 */
export declare function Off(...eventNames: [WailsEventName, ...WailsEventName[]]): void;
/**
 * Removes all event listeners.
 */
export declare function OffAll(): void;
/**
 * Emits an event.
 *
 * @returns A promise that will be fulfilled once the event has been emitted.  Resolves to true if the event was cancelled.
 * @param name - The name of the event to emit
 * @param data - The data that will be sent with the event
 */
export declare function Emit<E extends WailsEventName = WailsEventName>(name: E, data: WailsEventData<E>): Promise<boolean>;
export declare function Emit<E extends WailsEventName = WailsEventName>(name: WailsEventData<E> extends null | void ? E : never): Promise<boolean>;
