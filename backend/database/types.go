package database

import (
	"time"
	"vincit.fi/image-sorter/api/apitype"
)

type MigrationId uint32

type Migration struct {
	Id MigrationId `db:"id"`
}

type Image struct {
	Id              apitype.ImageId `db:"id,omitempty"`
	Name            string          `db:"name"`
	FileName        string          `db:"file_name"`
	Directory       string          `db:"directory"`
	ByteSize        int64           `db:"byte_size"`
	ExifOrientation uint8           `db:"exif_orientation"`
	ImageAngle      int             `db:"image_angle"`
	ImageFlip       bool            `db:"image_flip"`
	CreatedTime     time.Time       `db:"created_timestamp"`
	Width           uint32          `db:"width"`
	Height          uint32          `db:"height"`
	ModifiedTime    time.Time       `db:"modified_timestamp"`
}

type ImageMetaData struct {
	ImageId apitype.ImageId `db:"image_id"`
	Key     string          `db:"key"`
	Value   string          `db:"value"`
}

type Category struct {
	Id       apitype.CategoryId `db:"id,omitempty"`
	Name     string             `db:"name"`
	SubPath  string             `db:"sub_path"`
	Shortcut string             `db:"shortcut"`
}

type ImageCategory struct {
	ImageId    apitype.ImageId    `db:"image_id"`
	CategoryId apitype.CategoryId `db:"category_id"`
	Operation  int64              `db:"operation"`
}

type CategorizedImage struct {
	ImageId    apitype.ImageId    `db:"image_id"`
	CategoryId apitype.CategoryId `db:"category_id"`
	Name       string             `db:"name"`
	SubPath    string             `db:"sub_path"`
	Shortcut   string             `db:"shortcut"`
	Operation  int64              `db:"operation"`
}

type ImageSimilar struct {
	ImageId        apitype.ImageId `db:"image_id"`
	SimilarImageId apitype.ImageId `db:"similar_image_id"`
	Rank           int             `db:"rank"`
	Score          float64         `db:"score"`
}

type Count struct {
	Count int `db:"c"`
}

type Timestamp struct {
	Timestamp time.Time `db:"t"`
}
