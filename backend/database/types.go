package database

import "vincit.fi/image-sorter/api/apitype"

type Image struct {
	Id        apitype.HandleId `db:"id,omitempty"`
	Name      string           `db:"name"`
	FileName  string           `db:"file_name"`
	Directory string           `db:"directory"`
	ByteSize  int64            `db:"byte_size"`
}

type Category struct {
	Id       apitype.CategoryId `db:"id,omitempty"`
	Name     string             `db:"name"`
	SubPath  string             `db:"sub_path"`
	Shortcut string             `db:"shortcut"`
}

type ImageCategory struct {
	ImageId    apitype.HandleId   `db:"image_id"`
	CategoryId apitype.CategoryId `db:"category_id"`
	Operation  int64              `db:"operation"`
}

type CategorizedImage struct {
	ImageId    apitype.HandleId   `db:"image_id"`
	CategoryId apitype.CategoryId `db:"category_id"`
	Name       string             `db:"name"`
	SubPath    string             `db:"sub_path"`
	Shortcut   string             `db:"shortcut"`
	Operation  int64              `db:"operation"`
}

type ImageHash struct {
	ImageId int64 `db:"image_id"`
}

type Count struct {
	Count int `db:"c"`
}
