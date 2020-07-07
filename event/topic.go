package event

type Topic string
const(
	UI_READY Topic = "event-ui-ready"
	NEXT_IMAGE Topic = "event-next-image"
	PREV_IMAGE Topic = "event-prev-image"
	JUMP_NEXT_IMAGE Topic = "event-jump-next-image"
	JUMP_PREV_IMAGE Topic = "event-jump-prev-image"
	SIMILAR_IMAGE Topic = "event-similar-image"
	CURRENT_IMAGE Topic = "event-current-image"
	JUMP_TO_IMAGE Topic = "event-jump-to-image"
	CATEGORIZE_IMAGE Topic = "event-categorize-image"
	PERSIST_CATEGORIES Topic = "event-persis-categories"
	GENERATE_HASHES Topic = "event-generate-hashes"
	STOP_HASHES Topic = "event-stop-hashes"

	IMAGE_CHANGED Topic = "event-image-changed"

	IMAGES_UPDATED Topic = "event-images-updated"
	CATEGORIES_UPDATED Topic = "event-categories-updated"
	IMAGE_CATEGORIZED Topic = "event-image-categorized"
	UPDATE_HASH_STATUS Topic = "event-update-hash-status"

	CAST_FIND_DEVICES Topic = "event-cast-find-devices"
	CAST_DEVICE_FOUND Topic = "event-cast-device-found"
	CAST_SELECT_DEVICE Topic = "event-cast-select-device"
	CAST_READY Topic = "event-cast-ready"
	CAST_STOP Topic = "event-cast-stop"
)

