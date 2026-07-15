type PointerDownOutsideEvent = {
	detail: {
		originalEvent: Event;
	};
};

const handledOverlayPointerDownEvents = new WeakSet<Event>();

export function markOverlayPointerDownHandled(event: PointerDownOutsideEvent) {
	handledOverlayPointerDownEvents.add(event.detail.originalEvent);
}

export function wasOverlayPointerDownHandled(event: PointerDownOutsideEvent) {
	return handledOverlayPointerDownEvents.has(event.detail.originalEvent);
}
