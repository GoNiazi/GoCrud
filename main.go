package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/schema"

	validation "github.com/go-ozzo/ozzo-validation"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("Crud123"))
var tpl *template.Template

// User Struct
type User struct {
	ID       int
	Username string
	Email    string
	Password string
}

// Product struct
type Product struct {
	ID     int
	Name   string
	Price  int
	Resale int
}

func init() {
	tpl = template.Must(template.ParseGlob("templates/*"))
}

func dbConn() (db *sql.DB) {
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := ""
	dbName := "crud"
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	return db

}

//Validatelogin func
func (u User) ValidateLogin() error {
	return validation.ValidateStruct(&u,
		// Street cannot be empty, and the length must between 5 and 50
		validation.Field(&u.Email, validation.Required, validation.Length(5, 100)),
		// City cannot be empty, and the length must between 5 and 50
		validation.Field(&u.Password, validation.Required, validation.Length(5, 50)),
		// State cannot be empty, and must be a string consisting of two letters in upper case

		// State cannot be empty, and must be a string consisting of five digits
	)
}

//ValidateRegister func
func (u User) ValidateRegister() error {
	return validation.ValidateStruct(&u,
		// Street cannot be empty, and the length must between 5 and 50
		validation.Field(&u.Email, validation.Required, validation.Length(5, 100)),
		// City cannot be empty, and the length must between 5 and 50
		validation.Field(&u.Password, validation.Required, validation.Length(5, 50)),
		// State cannot be empty, and must be a string consisting of two letters in upper case
		validation.Field(&u.Username, validation.Required, validation.Length(5, 50)),
	)
}

//Auth func
func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session")
		_, ok := session.Values["email"]
		_, myok := session.Values["password"]
		if !ok && !myok {
			http.Redirect(w, r, "/login", 302)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	if r.Method == "POST" {
		log.Println("hello im here in login")
		r.ParseForm()
		email := r.FormValue("email")
		password := r.FormValue("password")
		//
		user := User{}
		decoder := schema.NewDecoder()
		err := decoder.Decode(&user, r.PostForm)
		if err != nil {
			log.Println(err)
		}

		user = User{Email: user.Email, Password: user.Password}
		err = user.ValidateLogin()
		//	log.Println(err)
		if err != nil {
			//	log.Println("heolli i  m in err")
			tpl.ExecuteTemplate(w, "login.html", err)
			return
		}

		//session
		session, _ := store.Get(r, "session")
		session.Values["email"] = email
		session.Values["password"] = password
		session.Save(r, w)
		log.Println("session")

		err = db.QueryRow("SELECT email , password FROM gocrud WHERE email=?", email).Scan(&user.Email, &user.Password)
		if err != nil {
			http.Redirect(w, r, "/login", 302)
		}

		if email == user.Email && password == user.Password {
			log.Println(user.Email, password)
			http.Redirect(w, r, "/index", 302)
			log.Println("hello im here in login")
		} else {
			http.Redirect(w, r, "/login", 302)
		}
		log.Println("hello im here in login")
	}
	log.Println("hello i m login")
	tpl.ExecuteTemplate(w, "login.html", nil)
}
func addproduct(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	if r.Method == "POST" {
		r.ParseForm()
		name := r.FormValue("name")
		price := r.FormValue("price")
		resale := r.FormValue("resale")
		rec, err := db.Prepare("INSERT INTO product(name,price,resale) VALUES(?,?,?)")
		if err != nil {
			http.Redirect(w, r, "/addproduct", 302)
		}
		_, execerr := rec.Exec(name, price, resale)
		if execerr != nil {
			http.Redirect(w, r, "/addproduct", 302)
		}
		log.Println("Values store")
		defer db.Close()
		http.Redirect(w, r, "/index", http.StatusFound)
	}

	tpl.ExecuteTemplate(w, "addproduct.html", nil)

}
func index(w http.ResponseWriter, r *http.Request) {
	log.Println("im here")
	db := dbConn()
	rec, err := db.Query("SELECT * FROM product ORDER BY id DESC")
	if err != nil {
		panic(err.Error())
	}
	pro := []Product{}
	for rec.Next() {
		product := Product{}
		proerr := rec.Scan(&product.ID, &product.Name, &product.Price, &product.Resale)
		if proerr != nil {
			panic(proerr.Error())
		}
		pro = append(pro, product)

	}
	tpl.ExecuteTemplate(w, "index.html", pro)
	defer db.Close()
}
func show(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	id := mux.Vars(r)["id"]
	rec := db.QueryRow("SELECT * FROM product WHERE id=?", id)
	product := Product{}
	err := rec.Scan(&product.ID, &product.Name, &product.Price, &product.Resale)
	if err != nil {
		panic(err.Error())
	}
	tpl.ExecuteTemplate(w, "show.html", product)
	defer db.Close()

}
func edit(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	id := mux.Vars(r)["id"]
	rec := db.QueryRow("SELECT * FROM product Where id=?", id)
	product := Product{}
	err := rec.Scan(&product.ID, &product.Name, &product.Price, &product.Resale)
	if err != nil {
		panic(err.Error())
	}
	tpl.ExecuteTemplate(w, "edit.html", product)
	defer db.Close()
}
func update(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	if r.Method == "POST" {
		r.ParseForm()
		name := r.FormValue("name")
		price := r.FormValue("price")
		resale := r.FormValue("resale")
		id := r.FormValue("productid")
		updform, err := db.Prepare("UPDATE product SET  name=?, price=?, resale=? WHERE id=? ")
		if err != nil {
			panic(err.Error())
		}
		updform.Exec(name, price, resale, id)
		log.Println("VALUES UPDATED")

	}

	defer db.Close()
	http.Redirect(w, r, "/index", 302)
}
func delete(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	id := mux.Vars(r)["id"]
	delrec, err := db.Prepare("DELETE FROM product WHERE id=?")
	if err != nil {
		panic(err.Error())
	}
	delrec.Exec(id)
	defer db.Close()
	http.Redirect(w, r, "/index", 302)
}
func register(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	if r.Method == "POST" {
		r.ParseForm()
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		//validation
		user := User{}
		decoder := schema.NewDecoder()
		decoerr := decoder.Decode(&user, r.PostForm)
		if decoerr != nil {
			log.Println(decoerr)
		}
		user = User{Username: user.Username, Email: user.Email, Password: user.Password}
		decoerr = user.ValidateRegister()
		if decoerr != nil {
			tpl.ExecuteTemplate(w, "register.html", decoerr)
			return
		}
		rec, err := db.Prepare("INSERT INTO gocrud(username,email,password) VALUES(?,?,?)")
		if err != nil {
			http.Redirect(w, r, "/register", 302)
		}
		_, executionerr := rec.Exec(username, email, password)
		if executionerr != nil {
			http.Redirect(w, r, "/register", 302)
		}
		defer db.Close()

		http.Redirect(w, r, "/login", http.StatusFound)
	}
	log.Println("REGISTER PAGE")
	tpl.ExecuteTemplate(w, "register.html", nil)
}
func logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	session.Values["email"] = ""
	session.Values["password"] = ""
	session.Options.MaxAge = -1

	err := session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/login", 302)

}
func main() {
	r := mux.NewRouter()

	r.HandleFunc("/login", login).Methods("POST")
	r.HandleFunc("/login", login).Methods("GET")
	r.HandleFunc("/index", Auth(index)).Methods("GET")
	r.HandleFunc("/index", Auth(index)).Methods("POST")
	r.HandleFunc("/register", register).Methods("POST")
	r.HandleFunc("/register", register).Methods("GET")
	r.HandleFunc("/addproduct", Auth(addproduct)).Methods("GET")
	r.HandleFunc("/addproduct", addproduct).Methods("POST")
	r.HandleFunc("/show/{id:[0-9]+}", Auth(show)).Methods("GET")
	r.HandleFunc("/edit/{id:[0-9]+}", Auth(edit)).Methods("GET")
	r.HandleFunc("/update", update).Methods("POST")
	r.HandleFunc("/delete/{id:[0-9]+}", delete).Methods("GET")
	r.HandleFunc("/logout", logout).Methods("GET")
	http.Handle("/", r)
	http.ListenAndServe(":8000", nil)

}
