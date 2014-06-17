package torch

import (
	"database/sql"
	"log"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"
)

type Parser struct {
	Store          *SessionStore
	DB             *sql.DB
	GetDatabase    func(session *Session) (*sql.DB, error)
	PublicPatterns []*regexp.Regexp
}

// Wraps a function expecting a Request to make it work with httpResponseWriter, http.Request
func (parser *Parser) WrapReturn(handler func(*Request)) func(w http.ResponseWriter, r *http.Request) *Request {
	return func(w http.ResponseWriter, r *http.Request) *Request {
		log.Printf("Begin Request %s\n", r.RequestURI)
		d, _ := httputil.DumpRequest(r, false)
		log.Println(string(d))
		defer log.Printf("End Request\n")

		requestTorch, err := parser.ParseRequest(w, r)
		if err != nil {
			log.Fatal(err)
			w.Write([]byte("An error occurred"))
			return nil
		}
		if requestTorch.Session.User == nil {
			log.Printf("PUBLIC: Check Path %s", r.URL.Path)
			for _, p := range parser.PublicPatterns {
				log.Println(p.String())
				if p.MatchString(r.URL.Path) {
					log.Printf("PUBLIC: Matched Public Path %s", p.String())
					handler(requestTorch)
					return requestTorch
				}
			}

			log.Println("PUBLIC: No Public Pathes Matched")
			if strings.HasSuffix(r.URL.Path, ".html") {
				requestTorch.Session.LoginTarget = &r.URL.Path
			}
			requestTorch.Redirect("/login")
		} else {
			handler(requestTorch)
		}
		return requestTorch
	}
}

func (parser *Parser) Wrap(handler func(*Request)) func(w http.ResponseWriter, r *http.Request) {
	f := parser.WrapReturn(handler)

	return func(w http.ResponseWriter, r *http.Request) {
		_ = f(w, r)
		return
	}
}

// WrapSplit checks the method of a request, and uses the handlers passed in order GET, POST, PUT, DELETE.
// To skip a method, pass nil (Or don't specify) Will return 404
// The order is an obsuciry I'm not proud of... probably should be a map?
// Ideally the methods should be registered separately, but that requires taking over more of the default
// functionality in httpRequestHandler which is not the plan at this stage. - These are Helpers, not a framework.
// (Who am I kidding, I love building frameworks)
func (parser *Parser) WrapSplit(handlers ...func(*Request)) func(w http.ResponseWriter, r *http.Request) {
	methods := []string{"GET", "POST", "PUT", "DELETE"}
	return parser.Wrap(func(request *Request) {
		for i, m := range methods {
			if request.Method == m {
				if len(handlers) > i && handlers[i] != nil {
					handlers[i](request)
				} else {
					//TODO: 404
				}
				return
			}
		}
	})
}

// ParseRequest is a utility usually used internally to give a Request object to a standard http request
// Exported for better flexibility
func (parser *Parser) ParseRequest(w http.ResponseWriter, r *http.Request) (*Request, error) {
	request := Request{
		writer: w,
		raw:    r,
		Method: r.Method,
	}

	sessCookie, err := r.Cookie("gsd_session")
	if err != nil {
		request.NewSession(parser.Store)
	} else {
		sess, err := parser.Store.GetSession(sessCookie.Value)
		if err != nil {
			request.NewSession(parser.Store)
		} else {
			request.Session = sess
		}
	}

	db, err := parser.GetDatabase(request.Session)
	if err != nil {
		return nil, err
	}
	request.db = db
	return &request, nil
}