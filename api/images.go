package api

import (
	"time"
)

//Image defines Image type
type Image struct {
	//unique identifier of the image
	ID string
	//name of the image
	Name string
	//minimum disk requirements to run the image (in GB)
	MinDisk int
	//minimum RAM requirements to run the image (in MB)
	MinRAM int
	//Creation date of the image
	CreatedAt time.Time
	//last update time
	UpdatedAt time.Time
}

//ImageManager defines image management functions a anyclouds provider must provide
type ImageManager interface {
	List() ([]Image, *ListImageError)
	Get(id string) (*Image, *GetImageError)
}

//ListImageError list image error type
type ListImageError struct {
	ErrorStack
}

//NewListImageError create a new ListImageError
func NewListImageError(cause error) *ListImageError {
	if cause == nil {
		return nil
	}
	return &ListImageError{*NewErrorStack(cause, "error listing images")}
}

//GetImageError get image error type
type GetImageError struct {
	ErrorStack
}

//NewGetImageError create a new GetImageError
func NewGetImageError(cause error, imageID string) *GetImageError {
	if cause == nil {
		return nil
	}
	return &GetImageError{*NewErrorStack(cause, "error getting image", imageID)}
}
