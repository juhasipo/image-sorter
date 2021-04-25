package database

type migration struct {
	id          MigrationId
	description string
	query       string
}

var migrations = []migration{
	{
		id:          0,
		description: "Initial Tables",
		query: `
			CREATE TABLE image (
			    id INTEGER PRIMARY KEY,
			    name TEXT,
			    file_name TEXT,
			    directory TEXT,
			    byte_size INT,
			    exif_orientation INT,
			    image_angle INT,
			    image_flip INT,
			    width INT,
			    height INT,
			    created_timestamp DATETIME,
			    modified_timestamp DATETIME,
			
			    UNIQUE (directory, file_name),
			    UNIQUE (name)
			);

			CREATE INDEX image_created_timestamp_idx ON image (created_timestamp);
			CREATE INDEX image_byte_size_idx ON image (byte_size);

			CREATE TABLE category (
			    id INTEGER PRIMARY KEY,
			    name TEXT,
			    sub_path TEXT,
			    shortcut INTEGER,
			
			    UNIQUE (name)
			);

			CREATE TABLE image_category (
			    image_id INTEGER,
			    category_id INTEGER,
			    operation INT,
			    
			    FOREIGN KEY(image_id) REFERENCES image(id) ON DELETE CASCADE,
			    FOREIGN KEY(category_id) REFERENCES category(id) ON DELETE CASCADE,
			    UNIQUE (image_id, category_id)
			);

			CREATE TABLE image_similar (
			    image_id INTEGER,
			    similar_image_id INTEGER,
			    rank INTEGER,
			    score REAL,
			    
			    FOREIGN KEY(image_id) REFERENCES image(id) ON DELETE CASCADE,
			    FOREIGN KEY(similar_image_id) REFERENCES image(id) ON DELETE CASCADE
			
			    -- Required indices is created dynamically so that INSERT can be optimized
			);
		`,
	},
	{
		id:          1,
		description: "Image Meta Data",
		query: `
			CREATE TABLE image_meta_data (
			    image_id INTEGER,
			    key TEXT,
			    value TEXT,

			    FOREIGN KEY(image_id) REFERENCES image(id) ON DELETE CASCADE
			);

			CREATE INDEX image_meta_data_idx ON image_meta_data (key, value);
			CREATE UNIQUE INDEX image_meta_data_uq ON image_meta_data (image_id, key);
		`,
	},
}
