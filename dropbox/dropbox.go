package dropbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const TimeFormat = time.RFC1123Z

type OAuthToken struct {
	Key, Secret string
}

type token interface {
	key() string
	secret() string
}

func buildAuthString(consumerToken AppToken, tok token) string {
	var buf bytes.Buffer
	buf.WriteString(`OAuth oauth_version="1.0", oauth_signature_method="PLAINTEXT"`)
	fmt.Fprintf(&buf, `, oauth_consumer_key="%s"`, url.QueryEscape(consumerToken.Key))
	fmt.Fprintf(&buf, `, oauth_timestamp="%v"`, time.Now().Unix())
	sigend := ""
	if tok != nil {
		sigend = url.QueryEscape(tok.secret())
		fmt.Fprintf(&buf, `, oauth_token="%s"`, url.QueryEscape(tok.key()))
	}
	fmt.Fprintf(&buf, `, oauth_signature="%s&%s"`, url.QueryEscape(consumerToken.Secret), sigend)
	return buf.String()
}

type Error struct {
	Code    int
	Message string
}

func (e Error) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

func doRequest(r *http.Request, consumerTok AppToken, accessTok token) (*FileReader, error) {
	r.Header.Set("Authorization", buildAuthString(consumerTok, accessTok))
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		var info struct {
			Error string `json:"error"`
		}
		return nil, Error{
			Code:    resp.StatusCode,
			Message: info.Error}
	}
	return newFileReader(resp), nil
}

func apiURL(path string) url.URL {
	return url.URL{
		Scheme: "https",
		Host:   apiHost,
		Path:   "/" + apiVersion + path}
}

func GetAuthorizeURL(requestToken RequestToken, callback *url.URL) *url.URL {
	params := url.Values{"oauth_token": {requestToken.Key}}
	if callback != nil {
		params.Add("oauth_callback", callback.String())
	}
	return &url.URL{
		Scheme:   "https",
		Host:     webHost,
		Path:     "/" + apiVersion + "/oauth/authorize",
		RawQuery: params.Encode()}
}

type AppToken OAuthToken
type RequestToken OAuthToken
type AccessToken OAuthToken

func (at AccessToken) key() string {
	return at.Key
}

func (at AccessToken) secret() string {
	return at.Secret
}

func (rt RequestToken) key() string {
	return rt.Key
}

func (rt RequestToken) secret() string {
	return rt.Secret
}

func postForToken(u url.URL, appToken AppToken, accessToken token) (OAuthToken, error) {
	r, e := http.NewRequest("POST", u.String(), nil)
	if e != nil {
		return OAuthToken{}, e
	}
	rc, e := doRequest(r, appToken, accessToken)
	if e != nil {
		return OAuthToken{}, e
	}
	defer rc.Close()
	var buf bytes.Buffer
	buf.ReadFrom(rc)
	vals, e := url.ParseQuery(buf.String())
	if e != nil {
		return OAuthToken{}, e
	}
	return OAuthToken{
		Key:    vals.Get("oauth_token"),
		Secret: vals.Get("oauth_token_secret")}, nil
}

func StartAuth(appToken AppToken) (RequestToken, error) {
	u := apiURL("/oauth/request_token")
	t, e := postForToken(u, appToken, nil)
	return RequestToken(t), e
}

func FinishAuth(appToken AppToken, requestToken RequestToken) (AccessToken, error) {
	u := apiURL("/oauth/access_token")
	t, e := postForToken(u, appToken, requestToken)
	return AccessToken(t), e
}

type accessType int

const (
	AppFolder accessType = iota
	Dropbox
)

type Config struct {
	Access accessType
	Locale string
}

type Client struct {
	AppToken    AppToken
	AccessToken AccessToken
	Config      Config
}

type AccountInfo struct {
	ReferralLink string `json:"referral_link"`
	DisplayName  string `json:"display_name"`
	Uid          uint64 `json:"uid"`
	Country      string `json:"country"`
	QuotaInfo    struct {
		Shared uint64 `json:"shared"`
		Quota  uint64 `json:"quota"`
		Normal uint64 `json:"normal"`
	} `json:"quota_info"`
	Email string `json:"email"`
}

type Delta struct {
	Reset   bool            `json:"reset"`
	Cursor  string          `json:"cursor"`
	HasMore bool            `json:"has_more"`
	Entries [][]interface{} `json:"entries"`
}

type FileMetadata struct {
	Size        string         `json:"size"`
	Rev         string         `json:"rev"`
	ThumbExists bool           `json:"thumb_exists"`
	Bytes       int64          `json:"bytes"`
	Modified    string         `json:"modified"`
	Path        string         `json:"path"`
	IsDir       bool           `json:"is_dir"`
	Icon        string         `json:"icon"`
	Root        string         `json:"root"`
	MimeType    string         `json:"mime_type"`
	Revision    int64          `json:"revision"`
	Hash        *string        `json:"hash"`
	Contents    []FileMetadata `json:"contents"`
}

func (md *FileMetadata) ModTime() time.Time {
	t, _ := time.Parse(TimeFormat, md.Modified)
	return t
}

type Link struct {
	URL     string `json:"url"`
	Expires string `json:"expires"`
}

func (s *Client) doGet(u url.URL) (*FileReader, error) {
	r, e := http.NewRequest("GET", u.String(), nil)
	if e != nil {
		return nil, e
	}
	return doRequest(r, s.AppToken, s.AccessToken)
}

func (s *Client) getForJson(u url.URL, jdata interface{}) error {
	buf, err := s.doGet(u)
	if err != nil {
		return err
	}
	defer buf.Close()
	return json.NewDecoder(buf).Decode(jdata)
}

func (s *Client) postForJson(u url.URL, jdata interface{}) error {
	r, e := http.NewRequest("POST", u.String(), nil)
	if e != nil {
		return e
	}
	rc, e := doRequest(r, s.AppToken, s.AccessToken)
	if e != nil {
		return e
	}
	defer rc.Close()
	return json.NewDecoder(rc).Decode(jdata)
}

const (
	apiVersion  = "1"
	apiHost     = "api.dropbox.com"
	contentHost = "api-content.dropbox.com"
	webHost     = "www.dropbox.com"
)

func (s *Client) GetDelta() (*Delta, error) {
	u := apiURL("/delta")
	u.RawQuery = s.Config.localeQuery()
	var delta Delta
	if e := s.postForJson(u, &delta); e != nil {
		return nil, e
	}
	return &delta, nil
}

func (s *Client) GetAccountInfo() (*AccountInfo, error) {
	u := apiURL("/account/info")
	u.RawQuery = s.Config.localeQuery()
	var info AccountInfo
	if e := s.getForJson(u, &info); e != nil {
		return nil, e
	}
	return &info, nil
}

func (s *Client) root() string {
	if s.Config.Access == Dropbox {
		return "dropbox"
	}
	return "sandbox"
}

func (s *Client) GetMetadata(path string, list bool) (*FileMetadata, error) {
	u := apiURL("/metadata/" + s.root() + path)
	v := url.Values{"list": {strconv.FormatBool(list)}}
	u.RawQuery = s.Config.setLocale(v).Encode()

	var md FileMetadata
	if e := s.getForJson(u, &md); e != nil {
		return nil, e
	}
	return &md, nil
}

func contentURL(path string) url.URL {
	return url.URL{
		Scheme: "https",
		Host:   contentHost,
		Path:   "/" + apiVersion + path}
}

type ThumbSize string

const (
	ThumbSmall  ThumbSize = "small"
	ThumbMedium ThumbSize = "medium"
	ThumbLarge  ThumbSize = "large"
	ThumbL      ThumbSize = "l"
	ThumbXL     ThumbSize = "xl"
)

func (s *Client) GetThumb(path string, size ThumbSize) (*FileReader, error) {
	u := contentURL("/thumbnails/" + s.root() + path)
	if size != "" {
		u.RawQuery = url.Values{"size": {string(size)}}.Encode()
	}
	rc, e := s.doGet(u)
	return rc, e
}

func (s *Client) AddFile(path string, contents io.Reader, size int64) (*FileMetadata, error) {
	return s.putFile(path, contents, size, url.Values{"overwrite": {"false"}})
}

func (s *Client) UpdateFile(path string, contents io.Reader, size int64, parentRev string) (*FileMetadata, error) {
	return s.putFile(path, contents, size, url.Values{"parent_rev": {parentRev}})
}

func (s *Client) ForceFile(path string, contents io.Reader, size int64) (*FileMetadata, error) {
	return s.putFile(path, contents, size, url.Values{"overwrite": {"true"}})
}

func (s *Client) putFile(path string, contents io.Reader, size int64, vals url.Values) (*FileMetadata, error) {
	u := contentURL("/files_put/" + s.root() + path)
	if vals == nil {
		vals = make(url.Values)
	}
	u.RawQuery = s.Config.setLocale(vals).Encode()
	r, e := http.NewRequest("PUT", u.String(), contents)
	if e != nil {
		return nil, e
	}
	r.ContentLength = size
	buf, err := doRequest(r, s.AppToken, s.AccessToken)
	if err != nil {
		return nil, err
	}
	var md FileMetadata
	dec := json.NewDecoder(buf)
	if e := dec.Decode(&md); e != nil {
		return nil, e
	}
	return &md, nil
}

type FileReader struct {
	io.ReadCloser
	// -1 if unknown.
	Size        int64
	ContentType string
}

func newFileReader(r *http.Response) *FileReader {
	return &FileReader{
		r.Body,
		r.ContentLength,
		r.Header.Get("Content-Type")}
}

func (s *Client) GetFile(path string) (*FileReader, error) {
	return s.doGet(contentURL("/files/" + s.root() + path))
}

func (s *Client) GetLink(path string) (*Link, error) {
	u := apiURL("/shares/" + s.root() + path)
	u.RawQuery = s.Config.localeQuery()
	var link Link
	if e := s.postForJson(u, &link); e != nil {
		return nil, e
	}
	return &link, nil
}

func (s *Client) GetMedia(path string) (*Link, error) {
	u := apiURL("/media/" + s.root() + path)
	u.RawQuery = s.Config.localeQuery()
	var link Link
	if e := s.postForJson(u, &link); e != nil {
		return nil, e
	}
	return &link, nil
}

func (c *Config) localeQuery() string {
	return c.setLocale(url.Values{}).Encode()
}

func (c *Config) setLocale(v url.Values) url.Values {
	if c.Locale != "" {
		v.Set("locale", c.Locale)
	}
	return v
}

func (s *Client) fileOp(op string, vals url.Values) (*FileMetadata, error) {
	u := apiURL("/fileops/" + op)
	vals.Set("root", s.root())
	u.RawQuery = s.Config.setLocale(vals).Encode()
	var md FileMetadata
	if e := s.postForJson(u, &md); e != nil {
		return nil, e
	}
	return &md, nil
}

func (s *Client) Move(from, to string) (*FileMetadata, error) {
	return s.fileOp("move", url.Values{"from_path": {from}, "to_path": {to}})
}

func (s *Client) Copy(from, to string) (*FileMetadata, error) {
	return s.fileOp("copy", url.Values{"from_path": {from}, "to_path": {to}})
}

func (s *Client) CreateDir(path string) (*FileMetadata, error) {
	return s.fileOp("create_folder", url.Values{"path": {path}})
}

func (s *Client) Delete(path string) (*FileMetadata, error) {
	return s.fileOp("delete", url.Values{"path": {path}})
}

func (c *Client) Search(path, query string, limit int) ([]FileMetadata, error) {
	u := apiURL("/search/" + c.root() + path)
	v := url.Values{"query": {query}}
	if limit > 0 {
		v.Set("limit", strconv.Itoa(limit))
	}
	u.RawQuery = c.Config.setLocale(v).Encode()
	var md []FileMetadata
	if e := c.getForJson(u, &md); e != nil {
		return nil, e
	}
	return md, nil
}
