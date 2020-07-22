package common

import "log"

type ImageOperation interface {
	Apply(*Handle) error
}

type ImageOperationGroup struct {
	handle     *Handle
	operations []ImageOperation
}

func ImageOperationGroupNew(handle *Handle, operations []ImageOperation) *ImageOperationGroup {
	return &ImageOperationGroup{
		handle:     handle,
		operations: operations,
	}
}

func (s *ImageOperationGroup) Apply() error {
	for _, operation := range s.operations {
		log.Printf("Applying: '%s'", operation)
		err := operation.Apply(s.handle)
		if err != nil {
			return err
		}
	}
	return nil
}

type fileOperation struct {
	dstPath string
	dstFile string
}

type ImageCopy struct {
	fileOperation

	ImageOperation
}

func ImageCopyNew(targetDir string, targetFile string) ImageOperation {
	return &ImageCopy{
		fileOperation: fileOperation{
			dstPath: targetDir,
			dstFile: targetFile,
		},
		ImageOperation: nil,
	}
}
func (s *ImageCopy) Apply(handle *Handle) error {
	log.Printf("Copy %s", handle.GetPath())
	return CopyFile(handle.GetDir(), handle.GetFile(), s.dstPath, s.dstFile)
}
func (s *ImageCopy) String() string {
	return "Copy to " + s.dstPath + " " + s.dstFile
}

type ImageMove struct {
	fileOperation

	ImageOperation
}

func (s *ImageMove) Apply(handle *Handle) error {
	log.Printf("Move %s", handle.GetPath())
	return nil
}
func (s *ImageMove) String() string {
	return "Move to " + s.dstPath + " " + s.dstFile
}

type ImageRemove struct {
	ImageOperation
}

func ImageRemoveNew() ImageOperation {
	return &ImageRemove{}
}
func (s *ImageRemove) Apply(handle *Handle) error {
	log.Printf("Remove %s", handle.GetPath())
	return RemoveFile(handle.path)
}
func (s *ImageRemove) String() string {
	return "Remove"
}
