package event

type Topic string
const(
	UI_READY = "event-ui-ready"
	NEXT_IMAGE = "event-next-image"
	PREV_IMAGE = "event-prev-image"
	CURRENT_IMAGE = "event-current-image"
	CATEGORIZE_IMAGE = "event-categorize-image"

	IMAGES_UPDATED = "event-images-updated"
	CATEGORIES_UPDATED = "event-categories-updated"
	IMAGE_CATEGORIZED = "event-image-categorized"
)

