package event

type Topic string
const(
	UI_READY Topic = "event-ui-ready"
	NEXT_IMAGE Topic = "event-next-image"
	PREV_IMAGE Topic = "event-prev-image"
	CURRENT_IMAGE Topic = "event-current-image"
	CATEGORIZE_IMAGE Topic = "event-categorize-image"

	IMAGES_UPDATED Topic = "event-images-updated"
	CATEGORIES_UPDATED Topic = "event-categories-updated"
	IMAGE_CATEGORIZED Topic = "event-image-categorized"
)

