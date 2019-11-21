/*
 * PrImBoard
 *
 * PrImBoard (Private Image Board) can be best described as an image board for all the picures and videos you have taken. You can invite users to the board and share specific images with them or your family members!
 *
 * API version: 1.0.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package swagger

type Media struct {

	Id string `json:"id,omitempty"`

	Title string `json:"title,omitempty"`

	Description string `json:"description,omitempty"`

	Comments []Comment `json:"comments,omitempty"`

	Creator string `json:"creator,omitempty"`

	Tags []string `json:"tags,omitempty"`

	Groups []MediaGroups `json:"groups,omitempty"`

	Timestamp int64 `json:"timestamp,omitempty"`

	TimestampUpload int64 `json:"timestamp_upload,omitempty"`

	Url string `json:"url,omitempty"`

	UrlThumb string `json:"url_thumb,omitempty"`

	Type_ string `json:"type,omitempty"`

	Format string `json:"format,omitempty"`
}
