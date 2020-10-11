Image Sorter
============

GTK3 application for sorting images to different folders

# Requirements

* GTK3
* libjpeg (or libjpeg-turbo8)

# Shortcuts

## Zoom

|Key | Description |
|----|-------------|
| +  | Zoom in
| -  | Zoom out
| Backspace | Zoom to fit

## Navigation

|Key | Description |
|----|-------------|
|Right | Move to next image
|CTRL + Right | Move 10 images forwards
|Left | Move to previous image
|CTRL + Left | Move 10 images backwards
|Page down | Move 100 images forwards
|Page up | Move 100 images backward
|Home | Go to the first image
|End | Go to the last image

# Categories

The `<key>` means a shortcut set for a category.

|Key | Description |
|----|-------------|
|`<key>` | Toggle category for image and move to the next image
|Shift + `<key>` | Toggle category but stay on the same image
|CTRL + `<key>` | Toggle category and remove all other categories set to the image, move to next image
|CTRL + Shift + `<key>` | Toggle category and remove all other categories set to the image, stay on the same image
|ALT + `<key>` | Show only images from that category (press F10 to showw all images again)

# Other

|Key | Description |
|----|-------------|
|ESC | Exit full screen/no distractions mode
|F8  | Cast to Chromecast
|F10 | Show all images (if selected to show only images from a category)
|F11 | Toggle full screen
|ALT + Enter | Toggle full screen
|CTRL + F11 | No distractions mode
|F12 | Find similar images


Development
===========

# Folder structure

|Folder   | Description |
|---------|-------------|
| api     | Common interfaces and types
| assets  | Images, icons etc.
| backend | Business logic implementation
| common  | Common implementations used by UI and backend
| duplo   | Optiomized implementation of Duplo library
| script  | Build scripts
| testassets | Test asset files (e.g. JPEG images) used in unit tests
| ui      | Graphical user interface implmentation
