package webdav

import "encoding/xml"

// FileInfo — информация о файле или папке. Единый формат для локальных и удалённых файлов
type FileInfo struct {
	Path     string
	Name     string
	Size     int64
	Modified string
	IsDir    bool
}

// MultiStatus — корень XML-ответа PROPFIND. Содержит список <response>
type MultiStatus struct {
	XMLName   xml.Name   `xml:"DAV: multistatus"`
	Responses []Response `xml:"DAV: response"`
}

// Response — один файл/папка в XML-ответе
type Response struct {
	Href     string   `xml:"DAV: href"`
	PropStat PropStat `xml:"DAV: propstat"`
}

// PropStat — контейнер для свойств файла
type PropStat struct {
	Prop Prop `xml:"DAV: prop"`
}

// Prop — свойства файла, которые мы запрашиваем у сервера
type Prop struct {
	DisplayName      string       `xml:"DAV: displayname"`
	GetContentLength int64        `xml:"DAV: getcontentlength"`
	GetLastModified  string       `xml:"DAV: getlastmodified"`
	ResourceType     ResourceType `xml:"DAV: resourcetype"`
}

// ResourceType — определяет, файл это или папка, если Collection != nil — это папка
type ResourceType struct {
	Collection *string `xml:"DAV: collection"`
}

// ToFileInfo превращает Response (который пришёл из XML) в наш удобный FileInfo
func (r *Response) ToFileInfo() FileInfo {
	return FileInfo{
		Path:     r.Href,
		Name:     r.PropStat.Prop.DisplayName,
		Size:     r.PropStat.Prop.GetContentLength,
		Modified: r.PropStat.Prop.GetLastModified,
		IsDir:    r.PropStat.Prop.ResourceType.Collection != nil,
	}
}
