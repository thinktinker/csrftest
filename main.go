package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"lenslocked.com/controllers"
	"lenslocked.com/middleware"
	"lenslocked.com/models"
	"lenslocked.com/rand"
)

func notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<h1>The page you were looking for could not be found. =(</h1>")

}

func main() {

	// Use: go run *.go --help to view the instruction
	// Use: go build . && ./lenslocked.com -prod to run for production
	// Use: go build . && ./lenslocked.com to run in development
	boolPtr := flag.Bool("prod", false, "Provide this flag in production to ensure that a config file is provided before the application starts.")
	flag.Parse()

	// Setup the connection string to the database "lenslocked_dev"
	cfg := LoadConfig(*boolPtr)
	dbCfg := cfg.Database

	// Connect to the database using the above connection string
	services, err := models.NewServices(
		models.WithGorm(dbCfg.Dialect(), dbCfg.Connection()),
		models.WithLogMode(!cfg.IsProd()),
		models.WithUser(cfg.Pepper, cfg.HMACKey),
		models.WithGallery(),
		models.WithImage(),
	)

	// Print a panic statement if the database cannot be connected
	must(err)

	// TO FIX: Close and Automigrate
	defer services.Close()
	services.AutoMigrate()

	r := mux.NewRouter() //instantiate a variable r which stores the gorilla mux router
	usersC := controllers.NewUsers(services.User)
	galleriesC := controllers.NewGalleries(services.Gallery, services.Image, r) //Update: pass the mux router to NewGalleries controller to create named routes
	staticC := controllers.NewStatic()

	// CSRF middleware
	bytes, err := rand.Bytes(32)
	must(err)
	csrfMw := csrf.Protect(bytes, csrf.Secure(cfg.IsProd()))

	// Testing the RequireUser middleware
	// Instantiate the middleware
	// By passing the UseMW to requireUserMW, we know that when requireUserMW is run UserMW is already run
	userMW := middleware.User{UserService: services.User}
	requireUserMW := middleware.RequireUser{User: userMW}

	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/contact", staticC.Contact).Methods("GET")
	r.Handle("/about", staticC.About).Methods("GET")
	r.HandleFunc("/signup", usersC.New).Methods("GET")     //this handler for /signups manages e GET method
	r.HandleFunc("/signup", usersC.Create).Methods("POST") //this handler for /signups manages e POST method
	r.Handle("/login", usersC.LoginView).Methods("GET")    //this handles for /login manages e GET method
	r.HandleFunc("/login", usersC.Login).Methods("POST")   //this handler for /login manages e POST method

	userLogout := requireUserMW.ApplyFn(usersC.Logout)
	r.HandleFunc("/logout", userLogout).Methods("POST") //this handler for /login manages e POST method
	// r.HandleFunc("/cookietest", usersC.CookieTest).Methods("GET")

	// When galleryNew is invoked, it would apply galleriesC.New to be processed
	galleryNew := requireUserMW.Apply(galleriesC.New)
	galleryCreate := requireUserMW.ApplyFn(galleriesC.Create)
	galleryEdit := requireUserMW.ApplyFn(galleriesC.Edit)
	galleryUpdate := requireUserMW.ApplyFn(galleriesC.Update)
	galleryDelete := requireUserMW.ApplyFn(galleriesC.Delete)
	galleryIndex := requireUserMW.ApplyFn(galleriesC.Index)
	galleryImageUpload := requireUserMW.ApplyFn(galleriesC.ImageUpload)
	galleryImageDelete := requireUserMW.ApplyFn(galleriesC.ImageDelete)

	// galleryRoutes
	r.HandleFunc("/galleries", galleryIndex).Methods("GET")
	r.Handle("/galleries/new", galleryNew).Methods("GET")
	r.HandleFunc("/galleries", galleryCreate).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/edit", galleryEdit).Methods("GET").Name(controllers.EditGallery)
	r.HandleFunc("/galleries/{id:[0-9]+}/update", galleryUpdate).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/delete", galleryDelete).Methods("POST")

	r.HandleFunc("/galleries/{id:[0-9]+}/images", galleryImageUpload).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/images/{filename}/delete", galleryImageDelete).Methods("POST")

	r.HandleFunc("/galleries/{id:[0-9]+}", galleriesC.Show).Methods("GET").Name(controllers.ShowGallery) // ShowGallery is a named route to construct the requests to a gallery with an id

	// //Assets
	assetHandler := http.FileServer(http.Dir("./assets/"))
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", assetHandler))

	// // Image routes
	imageHandler := http.FileServer(http.Dir("./images/"))
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", imageHandler))

	r.NotFoundHandler = http.HandlerFunc(notFound) //special property to handle notfound errors

	fmt.Printf("Starting the server on: %d ... \n", cfg.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), csrfMw(userMW.Apply(r)))

	//add r to ensure gorilla mux handles the routing process
	// by adding userMW.Apply to r (route), i.e. http.ListenAndServe(":3000", userMW.Apply(r))
	// the routing middleware will actually run when a request comes in
	// 	1. it will check the cookie
	// 	2. see if there's a user
	// 	3. set the context if there is
	// 	4. and then it'll call the routing code to decide where it's supposed to go
	// 	5. this will apply to every single route
	// Also, by adding csrf middleware to r (route), it ensures that the routes with POST methods
	// must be validated with a csrf token

	// Since all the routes have already applied the 1st pass of checking the cookie
	// to ensure that a valid remmeber token and its hashed token belongs to a valid user
	// the subsequent use of requireUserMW only need to check the context of the user
	// before allowing the user to access the gallery page (and display the gallery button)

}

// must take care of any error handling
func must(err error) {
	if err != nil {
		panic(err)
	}
}
