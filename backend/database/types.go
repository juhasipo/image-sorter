package database

type Image struct {
	Id        int64  `db:"id,omitempty"`
	Name      string `db:"name"`
	FileName  string `db:"file_name"`
	Directory string `db:"directory"`
	ByteSize  int64  `db:"byte_size"`
}

type Category struct {
	Id       int64  `db:"id,omitempty"`
	Name     string `db:"name"`
	SubPath  string `db:"sub_path"`
	Shortcut string `db:"shortcut"`
}

type ImageCategory struct {
	ImageId    int64 `db:"image_id"`
	CategoryId int64 `db:"category_id"`
	Operation  int64 `db:"operation"`
}

type CategorizedImage struct {
	Id        int64  `db:"id,omitempty"`
	Name      string `db:"name"`
	SubPath   string `db:"sub_path"`
	Shortcut  string `db:"shortcut"`
	Operation int64  `db:"operation"`
}

type ImageHash struct {
	ImageId int64 `db:"image_id"`
}
