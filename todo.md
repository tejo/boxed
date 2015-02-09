# high pri

* use email as key, not user id
- put api tokens and default user in config
- repository to parse extra data (title and data)
- article slug bucket

# low pri

- change router


{
  "reset":true,
  "cursor":"AAHskXVmJSRVG_bgh4Oq2VPQqG79nWM26Tl8jSmRyiGM3M0EGLzi_df861bAJ6q5xzkbQD7WhMTh8cEA2s1GUlDg79gRBPQJ6D5vJvah2sZrBg",
  "has_more":false,
  "entries":[
    ["/proj1",{"revision":3,"rev":"30ec9923d","thumb_exists":false,"bytes":0,"modified":"Wed, 20 Mar 2013 05:58:43 +0000","path":"/proj1","is_dir":true,"icon":"folder_app","root":"app_folder","size":"0 bytes"}],
    ["/proj1/current",{"revision":6,"rev":"60ec9923d","thumb_exists":false,"bytes":0,"modified":"Wed, 20 Mar 2013 05:58:48 +0000","path":"/proj1/current","is_dir":true,"icon":"folder_app","root":"app_folder","size":"0 bytes"}],
    ["/proj1/backup",{"revision":9,"rev":"90ec9923d","thumb_exists":false,"bytes":0,"modified":"Wed, 20 Mar 2013 05:58:59 +0000","path":"/proj1/backup","is_dir":true,"icon":"folder_app","root":"app_folder","size":"0 bytes"}],
    ["/proj1/current/test.txt",{"revision":12,"rev":"c0ec9923d","thumb_exists":false,"bytes":34,"modified":"Wed, 20 Mar 2013 06:10:51 +0000","client_mtime":"Wed, 20 Mar 2013 06:10:46 +0000","path":"/proj1/current/test.txt","is_dir":false,"icon":"page_white_text","root":"app_folder","mime_type":"text/plain","size":"34 bytes"}]
  ]
}

type Cursor struct {
  Reset bool `json:"reset"`
  Cursor string `json:"cursor"`
  HasMore bool `json:"has_more"`
  Entries [][]interface{} `json:"entries"`
}


{
    "delta": {
        "users": [
            12345678,
            23456789
        ]
    }
}

type DeltaPayLoad struct {
  Delta struct {
    Users []int `json:"users"`
  } `json:"delta"`
}
