package event

type Topic string

const (
	// General
	UPDATE_PROCESS_STATUS Topic = "update-process-status"
	UI_READY Topic = "ui-ready"

	// Image related
	IMAGE_REQUEST             Topic = "image-request"
	IMAGE_REQUEST_NEXT        Topic = "image-request-next"
	IMAGE_REQUEST_PREV        Topic = "image-request-prev"
	IMAGE_REQUEST_NEXT_OFFSET Topic = "image-request-next-offset"
	IMAGE_REQUEST_PREV_OFFSET Topic = "image-request-prev-offset"
	IMAGE_REQUEST_CURRENT     Topic = "image-request-current"
	IMAGE_REQUEST_SIMILAR     Topic = "image-request-similar"
	IMAGE_CHANGED             Topic = "image-changed"
	IMAGE_UPDATE              Topic = "image-update"
	IMAGE_LIST_SIZE_CHANGED   Topic = "image-list-size-changed"

	// Categorization
	CATEGORIZE_IMAGE      Topic = "categorize-image"
	CATEGORY_PERSIST_ALL  Topic = "category-persist-all"
	CATEGORIES_UPDATED    Topic = "categories-updated"
	CATEGORY_IMAGE_UPDATE Topic = "category-image-update"
	CATEGORIES_SAVE       Topic = "categories-save"
	CATEGORIES_SAVE_DEFAULT       Topic = "categories-save-default"

	// Similar image search
	SIMILAR_REQUEST_SEARCH Topic = "similar-request-serach"
	SIMILAR_REQUEST_STOP   Topic = "similar-request-stop"

	// Chrome Cast
	CAST_DEVICE_SEARCH       Topic = "cast-device-search"
	CAST_DEVICE_FOUND        Topic = "cast-device-found"
	CAST_DEVICE_SELECT       Topic = "cast-device-select"
	CAST_READY               Topic = "cast-ready"
	CAST_DEVICES_SEARCH_DONE Topic = "cast-devices-search-done"
)
