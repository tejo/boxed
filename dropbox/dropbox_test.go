package dropbox_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tejo/boxed/dropbox"
)

func Test_ParseDelta(t *testing.T) {
	a := assert.New(t)

	var d *dropbox.Delta
	json.Unmarshal([]byte(delta), &d)

	dropbox.ParseDelta(d)

	a.Contains(d.Updated, "/published/bar.md")
	a.Contains(d.Deleted, "/published/boxed super blog engine.md")
	a.NotContains(d.Updated, "/published/foo/")
}

var delta string = `
{
	"has_more": false, 
	"cursor": "asdfasdf", 
	"entries": [
		["/published/boxed super blog engine.md", null],
		["/published/bar.md", 
		 {"rev": "4b30b9d2f1", 
		 "thumb_exists": false, 
		 "path": "/published/boxed super blog engine.md", 
		 "is_dir": false, 
		 "client_mtime": 
		 "Fri, 06 Feb 2015 16:43:10 +0000", 
		 "icon": "page_white_text", 
		 "bytes": 90, 
		 "modified": "Fri, 20 Feb 2015 07:31:07 +0000", 
		 "size": "90 bytes", 
		 "root": "app_folder", 
		 "mime_type": "application/octet-stream", 
		 "revision": 75}],
		["/published/foo/", 
		 {"rev": "4b30b9d2f1", 
		 "thumb_exists": false, 
		 "path": "/published/boxed super blog engine.md", 
		 "is_dir": true, 
		 "client_mtime": 
		 "Fri, 06 Feb 2015 16:43:10 +0000", 
		 "icon": "page_white_text", 
		 "bytes": 90, 
		 "modified": "Fri, 20 Feb 2015 07:31:07 +0000", 
		 "size": "90 bytes", 
		 "root": "app_folder", 
		 "mime_type": "application/octet-stream", 
		 "revision": 75}]
	], 
	"reset": false}
`
