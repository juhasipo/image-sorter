package library

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"vincit.fi/image-sorter/api/apitype"
)

var (
	sutService *Service
)

func initializeSutService() *Service {
	sut = initializeSut()

	return NewImageService(nil, sut, statusStore)
}

func TestService_calculateNewIndexAndWrapNegative(t *testing.T) {
	a := assert.New(t)

	sutService = initializeSutService()

	t.Run("Calculate new index in range: 0", func(t *testing.T) {
		a.Equal(0, sutService.calculateNewIndexAndWrapNegative(0, 10))
	})
	t.Run("Calculate new index in range: positive", func(t *testing.T) {
		a.Equal(1, sutService.calculateNewIndexAndWrapNegative(1, 10))
		a.Equal(2, sutService.calculateNewIndexAndWrapNegative(2, 10))
		a.Equal(3, sutService.calculateNewIndexAndWrapNegative(3, 10))
		a.Equal(9, sutService.calculateNewIndexAndWrapNegative(9, 10))
	})
	t.Run("Calculate new index in range: clamp positive", func(t *testing.T) {
		a.Equal(9, sutService.calculateNewIndexAndWrapNegative(10, 10))
		a.Equal(9, sutService.calculateNewIndexAndWrapNegative(11, 10))
		a.Equal(9, sutService.calculateNewIndexAndWrapNegative(12, 10))
	})
	t.Run("Calculate new index in range: wrap negative", func(t *testing.T) {
		a.Equal(9, sutService.calculateNewIndexAndWrapNegative(-1, 10))
		a.Equal(8, sutService.calculateNewIndexAndWrapNegative(-2, 10))
		a.Equal(7, sutService.calculateNewIndexAndWrapNegative(-3, 10))
		a.Equal(6, sutService.calculateNewIndexAndWrapNegative(-4, 10))
		a.Equal(5, sutService.calculateNewIndexAndWrapNegative(-5, 10))
		a.Equal(4, sutService.calculateNewIndexAndWrapNegative(-6, 10))
		a.Equal(3, sutService.calculateNewIndexAndWrapNegative(-7, 10))
		a.Equal(2, sutService.calculateNewIndexAndWrapNegative(-8, 10))
		a.Equal(1, sutService.calculateNewIndexAndWrapNegative(-9, 10))
		a.Equal(0, sutService.calculateNewIndexAndWrapNegative(-10, 10))
	})
	t.Run("Calculate new index in range: clamp negative when wrapping the second time", func(t *testing.T) {
		a.Equal(0, sutService.calculateNewIndexAndWrapNegative(-11, 10))
		a.Equal(0, sutService.calculateNewIndexAndWrapNegative(-12, 10))
	})

}

func TestService_calculateIndexOffsetAndClamp(t *testing.T) {
	a := assert.New(t)

	sutService = initializeSutService()

	t.Run("Calculate new index in range: 0", func(t *testing.T) {
		a.Equal(0, sutService.calculateIndexOffsetAndClamp(0, 0, 10))
	})
	t.Run("Calculate new index in range: positive from 0/10", func(t *testing.T) {
		a.Equal(1, sutService.calculateIndexOffsetAndClamp(0, 1, 10))
		a.Equal(2, sutService.calculateIndexOffsetAndClamp(0, 2, 10))
		a.Equal(3, sutService.calculateIndexOffsetAndClamp(0, 3, 10))
		a.Equal(9, sutService.calculateIndexOffsetAndClamp(0, 9, 10))
		a.Equal(9, sutService.calculateIndexOffsetAndClamp(0, 10, 10))
	})
	t.Run("Calculate new index in range: positive from 5/10", func(t *testing.T) {
		a.Equal(6, sutService.calculateIndexOffsetAndClamp(5, 1, 10))
		a.Equal(7, sutService.calculateIndexOffsetAndClamp(5, 2, 10))
		a.Equal(8, sutService.calculateIndexOffsetAndClamp(5, 3, 10))
		a.Equal(9, sutService.calculateIndexOffsetAndClamp(5, 4, 10))
		a.Equal(9, sutService.calculateIndexOffsetAndClamp(5, 5, 10))
		a.Equal(9, sutService.calculateIndexOffsetAndClamp(5, 9, 10))
	})
	t.Run("Calculate new index in range: wrap negative", func(t *testing.T) {
		a.Equal(0, sutService.calculateIndexOffsetAndClamp(0, -1, 10))
		a.Equal(0, sutService.calculateIndexOffsetAndClamp(0, -2, 10))
	})

}

func TestService_moveToImage(t *testing.T) {
	a := assert.New(t)

	sutService = initializeSutService()

	t.Run("Find by ID no category", func(tt *testing.T) {
		imageFiles := []*apitype.ImageFile{
			apitype.NewImageFileWithId(1, "/tmp", "foo1"),
			apitype.NewImageFileWithId(2, "/tmp", "foo2"),
			apitype.NewImageFileWithId(3, "/tmp", "foo3"),
			apitype.NewImageFileWithId(4, "/tmp", "foo4"),
			apitype.NewImageFileWithId(5, "/tmp", "foo5"),
		}
		sutService.AddImageFiles(imageFiles)

		tt.Run("Find by ID", func(t *testing.T) {
			a.Equal(0, sutService.findImageIndex(1, apitype.NoCategory))
			a.Equal(1, sutService.findImageIndex(2, apitype.NoCategory))
			a.Equal(2, sutService.findImageIndex(3, apitype.NoCategory))
			a.Equal(3, sutService.findImageIndex(4, apitype.NoCategory))
			a.Equal(4, sutService.findImageIndex(5, apitype.NoCategory))
		})
		tt.Run("Find by ID: not found", func(t *testing.T) {
			a.Equal(0, sutService.findImageIndex(10, apitype.NoCategory))
		})
	})

	t.Run("Find by ID in category", func(tt *testing.T) {
		imageFiles := []*apitype.ImageFile{
			apitype.NewImageFileWithId(1, "/tmp", "foo1"),
			apitype.NewImageFileWithId(2, "/tmp", "foo2"),
			apitype.NewImageFileWithId(3, "/tmp", "foo3"),
			apitype.NewImageFileWithId(4, "/tmp", "foo4"),
			apitype.NewImageFileWithId(5, "/tmp", "foo5"),
		}
		sutService.AddImageFiles(imageFiles)
		category1, _ := categoryStore.AddCategory(apitype.NewCategory("category1", "cat1", "C"))
		category2, _ := categoryStore.AddCategory(apitype.NewCategory("category2", "cat2", "D"))

		imageCategoryStore.CategorizeImage(2, category1.Id(), apitype.CATEGORIZE)
		imageCategoryStore.CategorizeImage(4, category1.Id(), apitype.CATEGORIZE)
		imageCategoryStore.CategorizeImage(5, category1.Id(), apitype.CATEGORIZE)

		imageCategoryStore.CategorizeImage(1, category2.Id(), apitype.CATEGORIZE)
		imageCategoryStore.CategorizeImage(2, category2.Id(), apitype.CATEGORIZE)

		tt.Run("Find by ID", func(t *testing.T) {
			a.Equal(0, sutService.findImageIndex(2, category1.Id()))
			a.Equal(1, sutService.findImageIndex(4, category1.Id()))
			a.Equal(2, sutService.findImageIndex(5, category1.Id()))
		})
		tt.Run("Find by ID: not found", func(t *testing.T) {
			a.Equal(0, sutService.findImageIndex(1, category1.Id()))
		})
	})

}
