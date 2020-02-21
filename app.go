package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"github.com/justinas/alice"
)

type App struct {
	Router *mux.Router
	Middlewares *Middleware
	config *Env
}

type shortenReq struct {
	URL string `json:"url" validate:"required"`
	ExpirationInMinutes int64 `json:"expiration_in_minutes" validate:"min=0"`
}

type shortlinkResp struct {
	Shortlink string `json:"shortlink"`
}

//Initialize app
func (a *App) Initialize(e *Env) {
	//set log formatter
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	a.config = e
	a.Router = mux.NewRouter()
	a.Middlewares = &Middleware{}
	a.initializeRoutes()
}

func (a *App) initializeRoutes() {

	//a.Router.HandleFunc("/api/shorten",a.createShortLink).Methods("POST")
	//a.Router.HandleFunc("/api/info",a.getShortlinkInfo).Methods("GET")
	//a.Router.HandleFunc("/{shortlink:[a-zA-Z0-9]{1,11}}",a.redirect).Methods("GET")
	m := alice.New(a.Middlewares.LoggingHandler,a.Middlewares.RecoverHandler)
	a.Router.Handle("/api/shorten",m.ThenFunc(a.createShortLink)).Methods("POST")
	a.Router.Handle("/api/info",m.ThenFunc(a.getShortlinkInfo)).Methods("GET")
	a.Router.Handle("/{shortlink:[a-zA-Z0-9]{1,11}}",m.ThenFunc(a.redirect)).Methods("GET")

}

func (a *App) createShortLink(w http.ResponseWriter,r *http.Request) {
	var req shortenReq
	log.Print("createShortLink...")
	if err := json.NewDecoder(r.Body).Decode(&req);err != nil{
		respondWithError(w,StatusError{
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("parse parameters failed %v",r.Body),
		})
		return
	}
	log.Print(req)
	//http.HandlerFunc()
	validate := validator.New()
	if err := validate.Struct(req);err != nil{
		respondWithError(w,StatusError{
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("validate parameters failed %v",req),
		})
		return
	}
	defer r.Body.Close()

	s,err := a.config.S.Shorten(req.URL,req.ExpirationInMinutes)
	if err != nil{
		respondWithError(w,err)
		return
	}
	respondWithJSON(w,http.StatusCreated,shortlinkResp{Shortlink:s})
}

func (a *App) getShortlinkInfo(w http.ResponseWriter,r *http.Request) {
	vals := r.URL.Query()
	s := vals.Get("shortlink")

	info,err := a.config.S.ShortlinkInfo(s)
	if err != nil{
		respondWithError(w,err)
	}
	respondWithJSON(w,http.StatusOK,info)
}

func (a *App) redirect(w http.ResponseWriter,r *http.Request) {
	vars := mux.Vars(r)

	url,err := a.config.S.Unshorten(vars["shortlink"])
	if err != nil{
		respondWithError(w,err)
	}

	http.Redirect(w,r,url,http.StatusTemporaryRedirect)
}

//listen and serve
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr,a.Router))
	
}

func respondWithError(w http.ResponseWriter,err error)  {
	switch e:=err.(type) {
	case Error:
		log.Printf("HTTP %d - %s",e.Status(),e)
		respondWithJSON(w,e.Status(),e.Error())
	default:
		respondWithJSON(w,http.StatusInternalServerError,http.StatusText(http.StatusInternalServerError))
	}
	
}

func respondWithJSON(w http.ResponseWriter,code int,payload interface{})  {
	resp,_ := json.Marshal(payload)

	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(code)
	w.Write(resp)
}