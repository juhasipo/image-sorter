package event

type Topic string

const (
	// General
	ProcessStatusUpdated Topic = "process-status-updated"
	UiReady              Topic = "ui-ready"
	DirectoryChanged     Topic = "directory-changed"

	// Image related
	ImageRequest           Topic = "image-request"
	ImageRequestAtIndex    Topic = "image-request-at-index"
	ImageRequestNext       Topic = "image-request-next"
	ImageRequestPrev       Topic = "image-request-prev"
	ImageRequestNextOffset Topic = "image-request-next-offset"
	ImageRequestPrevOffset Topic = "image-request-prev-offset"
	ImageRequestCurrent    Topic = "image-request-current"
	ImageRequestSimilar    Topic = "image-request-similar"
	ImageShowOnly          Topic = "image-show-only"
	ImageShowAll           Topic = "image-show-all"
	ImageChanged           Topic = "image-changed"
	ImageListUpdated       Topic = "image-list-updated"
	ImageCurrentUpdated    Topic = "image-current-updated"
	ImageListSizeChanged   Topic = "image-list-size-changed"

	// Categorization
	CategorizeImage       Topic = "categorize-image"
	CategoryPersistAll    Topic = "category-persist-all"
	CategoriesUpdated     Topic = "categories-updated"
	CategoryImageUpdate   Topic = "category-image-update"
	CategoriesSave        Topic = "categories-save"
	CategoriesSaveDefault Topic = "categories-save-default"
	CategoriesShowOnly    Topic = "categories-show-only"

	// Similar image search
	SimilarRequestSearch Topic = "similar-request-search"
	SimilarRequestStop   Topic = "similar-request-stop"
	SimilarSetShowImages Topic = "similar-set-show-images"

	// Chrome Cast
	CastDeviceSearch      Topic = "cast-device-search"
	CastDeviceFound       Topic = "cast-device-found"
	CastDeviceSelect      Topic = "cast-device-select"
	CastReady             Topic = "cast-ready"
	CastDevicesSearchDone Topic = "cast-devices-search-done"
)
