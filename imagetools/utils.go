package imagetools

func ScaleToFit(sourceWidth int, sourceHeight int, targetWidth int, targetHeight int) (int, int) {
	ratio := float32(sourceWidth) / float32(sourceHeight)
	newWidth := int(float32(targetHeight) * ratio)
	newHeight := targetHeight

	if newWidth > targetWidth {
		newWidth = targetWidth
		newHeight = int(float32(targetWidth) / ratio)
	}
	return newWidth, newHeight
}
