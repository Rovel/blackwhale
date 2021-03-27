package handlers

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/gorilla/sessions"
	"github.com/joaopandolfi/blackwhale/configurations"
	"github.com/joaopandolfi/blackwhale/handlers/conjson"
	"github.com/joaopandolfi/blackwhale/handlers/conjson/transform"
	"github.com/joaopandolfi/blackwhale/utils"
)

// --- Responses ---

// Regexp definitions
var keyMatchRegex = regexp.MustCompile(`\"(\w+)\":`)
var wordBarrierRegex = regexp.MustCompile(`([a-z_0-9])([A-Z])`)

// marshaler
var marshaler func(v interface{}) ([]byte, error) = json.Marshal

// ActiveSnakeCase default json encoder
func ActiveSnakeCase() {
	marshaler = func(v interface{}) ([]byte, error) {
		marshaler := conjson.NewMarshaler(v, transform.ConventionalKeys())
		return json.MarshalIndent(marshaler, "", " ")
	}
}

// SnakeCaseDecoder json
func SnakeCaseDecoder(r io.Reader) conjson.Decoder {
	return conjson.NewDecoder(json.NewDecoder(r), transform.ConventionalKeys(), transform.ValidIdentifierKeys())
}

// header -
func header(w http.ResponseWriter) {
	w.Header().Add("Access-Control-Allow-Origin", configurations.Configuration.CORS)
	w.Header().Add("Content-Type", "application/json")
}

// responseError - Private function to make response
func responseError(w http.ResponseWriter, message string) {
	b, _ := json.Marshal(map[string]string{"message": message})
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(b)
}

// restResponseError - Private function to response in mode RES error
func restResponseError(w http.ResponseWriter, message string) {
	b, _ := json.Marshal(map[string]interface{}{"success": false, "message": message})
	w.Write(b)
}

// RESTResponse - Make default REST API response
func RESTResponse(w http.ResponseWriter, resp interface{}) {
	Response(w, map[string]interface{}{"success": true, "data": resp}, http.StatusOK)
}

// RESTResponseWithStatus - Make default REST API response with statuscode
func RESTResponseWithStatus(w http.ResponseWriter, resp interface{}, status int) {
	Response(w, map[string]interface{}{"success": true, "data": resp}, status)
}

// Response - Make default generic response
func Response(w http.ResponseWriter, resp interface{}, status int) {
	// set Header
	header(w)
	w.WriteHeader(status)
	b, err := marshaler(resp)

	if err == nil {
		// Responde
		w.Write(b)
	} else {
		utils.Error("Error on convert response to JSON", err)
		ResponseError(w, "Error on convert response to JSON")
	}
}

// ResponseError - Make default generic response
func ResponseError(w http.ResponseWriter, resp interface{}) {
	// set Header
	header(w)
	b, _ := marshaler(resp)
	responseError(w, string(b))
}

// RESTResponseError - Make REST API default response
func RESTResponseError(w http.ResponseWriter, resp interface{}) {
	// set Header
	header(w)
	b, _ := marshaler(resp)
	restResponseError(w, string(b))
}

// Redirect - Redirect page
func Redirect(r *http.Request, w http.ResponseWriter, url string) {
	http.Redirect(w, r, url, http.StatusSeeOther)
}

// --- Parameters ---

// GetVars - Return url vars
// @example /api/{key}/send
// @vars = {"key":data}
func GetVars(r *http.Request) map[string]string {
	return mux.Vars(r)
}

// GetHeader - Return Header value stored on passed key
func GetHeader(r *http.Request, key string) string {
	return r.Header.Get(key)
}

// InjectHeader - Inject data on header request
func InjectHeader(r *http.Request, key, val string) {
	r.Header.Add(key, val)
}

// GetQueryes - Return queryes values
// @example /api?key=data
func GetQueryes(r *http.Request) url.Values {
	return r.URL.Query()
}

// GetBody - Return byte body data
func GetBody(r *http.Request) ([]byte, error) {
	return ioutil.ReadAll(r.Body)
}

// GetForm - Return parsed form data
func GetForm(r *http.Request) (form url.Values, err error) {
	err = r.ParseForm()
	form = r.Form
	return
}

// DecodeForm - Decoded parsed form data on interface
func DecodeForm(dst interface{}, src map[string][]string) error {
	decoder := schema.NewDecoder()
	return decoder.Decode(dst, src)
}

// GetSession returns stored Session
// @global
// Login session keys: `logged`, `username`, `institution`, `level`, `token`
func GetSession(r *http.Request) (*sessions.Session, error) {
	return configurations.Configuration.Session.Store.Get(r, configurations.Configuration.Session.Name)
}

// GetNamedSession - Return data sored on specific session
func GetNamedSession(r *http.Request, name string) (*sessions.Session, error) {
	return configurations.Configuration.Session.Store.Get(r, name)
}

// ExtractToken - Extract Jwt Token
func ExtractToken(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	//normally Authorization the_token_xxx
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}
