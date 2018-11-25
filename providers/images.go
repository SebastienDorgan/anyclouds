package providers

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
}

//ImageManager defines image management functions a anyclouds provider must provide
type ImageManager interface {
	List(filter *ResourceFilter) ([]Image, error)
	Get(id string) (*Image, error)
}
