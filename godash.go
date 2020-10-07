package main

import (
	"encoding/json"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/evorts/godash/pkg"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

var (
	templates map[string]*template.Template
	api       *pkg.API
)

type Request struct {
	w     http.ResponseWriter
	r     *http.Request
	token string
	crypt *pkg.Crypt
}

type RenderData struct {
	PageAttributes map[string]string
	Groups         []pkg.Group
	Errors         map[string]string
	Csrf           string
	Forms          map[string]string
	LoggedIn       bool
}

func (req Request) isMethodGet() bool {
	return strings.ToUpper(req.r.Method) == "GET"
}

func (req Request) parseForm() {
	_ = req.r.ParseForm()
}

func (req Request) get(field string) []string {
	return req.r.Form[field]
}

func (req Request) renderContent(value interface{}) {
	_, _ = fmt.Fprintln(req.w, value)
}

func (req Request) renderJson(value interface{}) {
	if v, err := json.Marshal(value); err == nil {
		_, _ = fmt.Fprintln(req.w, string(v))
		return
	}
	_, _ = fmt.Fprintln(req.w, "{}")
}

func (req Request) prepare() Request {
	req.token = req.crypt.CryptWithSalt(time.Now().String())
	return req
}

func (req Request) isLoggedIn() bool {
	user := getAPI().Session().Get(req.r.Context(), "user")
	if user == nil || user == "" {
		return false
	}
	return true
}

func (req Request) render(w http.ResponseWriter, templateName string, data RenderData) error {
	tmpl := getTemplate(templateName)
	if tmpl == nil {
		return fmt.Errorf("the template %s does not exist", templateName)
	}
	if data.PageAttributes == nil {
		data.PageAttributes = make(map[string]string, 0)
	}
	data.PageAttributes["LogoUrl"] = getAPI().Config().GetConfig().App.Logo.Url
	data.PageAttributes["LogoAlt"] = getAPI().Config().GetConfig().App.Logo.Alt
	data.PageAttributes["ContactEmail"] = getAPI().Config().GetConfig().App.Contact.Email
	data.PageAttributes["ContactPhone1"] = getAPI().Config().GetConfig().App.Contact.Phone[0]
	data.PageAttributes["ContactPhone2"] = getAPI().Config().GetConfig().App.Contact.Phone[1]
	data.PageAttributes["ContactAddress"] = getAPI().Config().GetConfig().App.Contact.Address
	data.Csrf = req.token
	data.LoggedIn = req.isLoggedIn()
	// Render the template 'name' with data
	err := tmpl.ExecuteTemplate(w, templateName, data)
	if err != nil {
		getAPI().Logger().Log("template_err", err.Error())
		return err
	}
	return nil
}

func getAPI() *pkg.API {
	return api
}

func getTemplate(name string) *template.Template {
	if v, ok := templates[name]; ok {
		return v
	}
	return nil
}

func loadTemplates(dir string) {
	if templates == nil {
		templates = make(map[string]*template.Template, 0)
	}
	layouts, err := filepath.Glob(fmt.Sprintf("%s/layouts/*.html", dir))
	fmt.Println(fmt.Sprintf("parsing views template at: %s/layouts/*.html", dir))
	if err != nil {
		log.Fatal(err)
	}
	views, err2 := filepath.Glob(fmt.Sprintf("%s/views/*.html", dir))
	fmt.Println(fmt.Sprintf("parsing views template at: %s/views/*.html", dir))
	if err2 != nil {
		log.Fatal(err2)
	}
	for _, view := range views {
		files := append(layouts, view)
		templates[filepath.Base(view)] = template.Must(template.ParseFiles(files...))
	}
}

func main() {
	config := pkg.NewConfig("config.yml")
	err := config.Initiate()
	if err != nil {
		log.Fatal("error reading configuration")
	}
	sm := scs.New()
	sm.Lifetime = time.Second * config.GetConfig().App.SessionExpiration
	sm.IdleTimeout = 30 * time.Minute
	sm.Cookie.Name = "dashid"
	sm.Cookie.Domain = config.GetConfig().App.CookieDomain
	sm.Cookie.HttpOnly = true
	sm.Cookie.Persist = true
	sm.Cookie.SameSite = http.SameSiteStrictMode
	sm.Cookie.Secure = config.GetConfig().App.CookieSecure == 1
	sm.Cookie.Path = "/"

	api = pkg.NewAPI(config, sm, pkg.NewLogging(), pkg.NewCrypt(config.GetConfig().App.Salt), pkg.NewCrypt(""))
	loadTemplates(api.Config().GetConfig().App.TemplateDirectory)
	o := http.NewServeMux()
	// serving assets
	fs := http.FileServer(http.Dir(getAPI().Config().GetConfig().App.AssetDirectory))
	o.Handle("/assets/", http.StripPrefix("/assets", fs))
	// serving pages
	o.HandleFunc("/ping", ping)
	o.HandleFunc("/", dashboardHandler)
	o.HandleFunc("/login", loginHandler)
	o.HandleFunc("/logout", logoutHandler)
	o.HandleFunc("/not-found", notFoundHandler)
	api.Logger().Log("started", "Dashboard app started.")
	if err := http.ListenAndServe(fmt.Sprintf(":%d", api.Config().GetConfig().App.Port), sm.LoadAndSave(o)); err != nil {
		log.Fatal(err)
	}
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	getAPI().Logger().Log("dashboard_handler", "request received")
	req := Request{w: w, r: r}
	// check if the request already authenticated
	rp := r.URL.Path
	if rp != "/" && rp != "/login" && rp != "/logout" && !strings.HasPrefix(rp, "/assets/") {
		http.Redirect(w, r, "/not-found", http.StatusPermanentRedirect)
		return
	}
	if !req.isLoggedIn() {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	// render dashboard page
	if err := req.render(w, "dashboard.html", RenderData{
		PageAttributes: map[string]string{
			"Title": "Dashboard Page",
		},
		Groups: getAPI().Config().GetConfig().Groups,
	}); err != nil {
		getAPI().Logger().Log("dashboard_handler", err.Error())
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	getAPI().Logger().Log("login_handler", "request received")
	req := Request{w: w, r: r, crypt: getAPI().Crypt()}.prepare()
	if req.isLoggedIn() {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	renderData := RenderData{
		PageAttributes: map[string]string{
			"Title": "Login Page",
		},
	}
	if req.isMethodGet() {
		// render login page
		getAPI().Session().Put(r.Context(), "token", req.token)
		if err := req.render(w, "login.html", renderData); err != nil {
			getAPI().Logger().Log("login_handler", err.Error())
		}
		return
	}
	req.parseForm()
	// validate form
	user := req.get("username")
	pass := req.get("password")
	remember := req.get("remember")
	csrf := req.get("csrf")
	validationErrors := make(map[string]string, 0)
	if len(user) < 1 || strings.TrimSpace(user[0]) == "" {
		validationErrors["username"] = "Please fill up your username correctly"
	}
	if len(pass) < 1 || strings.TrimSpace(pass[0]) == "" {
		validationErrors["password"] = "Please fill up your password correctly"
	}
	if len(csrf) < 1 || strings.TrimSpace(csrf[0]) == "" {
		validationErrors["global"] = "Invalid request session"
	}
	if len(validationErrors) > 0 {
		renderData.Errors = validationErrors
		if err := req.render(w, "login.html", renderData); err != nil {
			getAPI().Logger().Log("login_handler", err.Error())
		}
		return
	}
	// csrf check
	sessionCsrf := getAPI().Session().Get(r.Context(), "token")
	if sessionCsrf == nil || csrf[0] != sessionCsrf.(string) {
		validationErrors["global"] = "Invalid request session"
		renderData.Errors = validationErrors
		if err := req.render(w, "login.html", renderData); err != nil {
			getAPI().Logger().Log("login_handler", err.Error())
		}
		return
	}
	// ensure the user and password are correct
	var userFound *pkg.User
	for _, u := range getAPI().Config().GetConfig().Users {
		if strings.ToLower(u.Username) == strings.ToLower(user[0]) {
			userFound = &u
			break
		}
	}
	if userFound == nil {
		validationErrors["global"] = "User not found. Please ensure you input it correctly."
		renderData.Errors = validationErrors
		if err := req.render(w, "login.html", renderData); err != nil {
			getAPI().Logger().Log("login_handler", err.Error())
		}
		return
	}
	passCrypt := getAPI().Hash().Crypt(pass[0])
	if strings.ToLower(passCrypt) != strings.ToLower(userFound.Password) {
		validationErrors["global"] = "Invalid authentication"
		renderData.Errors = validationErrors
		if err := req.render(w, "login.html", renderData); err != nil {
			getAPI().Logger().Log("login_handler", err.Error())
		}
		return
	}
	cookieExpiration := 3 * 24 * time.Hour
	if len(remember) > 0 && len(remember[0]) > 0 {
		getAPI().Session().Lifetime = cookieExpiration
	}
	getAPI().Session().Put(r.Context(), "user", user[0])
	if err := getAPI().Session().RenewToken(r.Context()); err != nil {
		validationErrors["global"] = "Failed to process"
		renderData.Errors = validationErrors
		if err := req.render(w, "login.html", renderData); err != nil {
			getAPI().Logger().Log("login_handler", err.Error())
		}
		return
	}
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	req := Request{w: w, r: r}
	if !req.isLoggedIn() {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	if err := getAPI().Session().Destroy(r.Context()); err != nil {
		_ = getAPI().Session().RenewToken(r.Context())
	}
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	getAPI().Logger().Log("404_handler", "request received")
	req := Request{w: w, r: r}
	if err := req.render(w, "404.html", RenderData{
		PageAttributes: map[string]string{
			"Title": "Nothing Found",
		},
	}); err != nil {
		getAPI().Logger().Log("404_handler", err.Error())
	}
}

func ping(w http.ResponseWriter, r *http.Request) {
	req := Request{w: w, r: r}
	if !req.isMethodGet() {
		req.renderContent("NOK")
		return
	}
	req.renderContent("OK")
}
