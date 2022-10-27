package main

import (
	"context"
	"day-7/connection"
	"day-7/middleware"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	route := mux.NewRouter()
	connection.DatabaseConnect()

	route.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public/"))))
	route.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	route.HandleFunc("/", homePage).Methods("GET")
	route.HandleFunc("/contact", contact).Methods("GET")
	route.HandleFunc("/add-project", blogPage).Methods("GET")
	route.HandleFunc("/project-detail/{id}", blogDetail).Methods("GET")
	route.HandleFunc("/send-data-add", middleware.UploadFile(sendDataAdd)).Methods("POST")
	route.HandleFunc("/delete-project/{id}", deleteProject).Methods("GET")
	route.HandleFunc("/edit-project/{id}", editProject).Methods("GET")
	route.HandleFunc("/update-project/{id}", middleware.UploadFile(updateProject)).Methods("POST")
	route.HandleFunc("/form-register", formRegister).Methods("GET")
	route.HandleFunc("/register", register).Methods("POST")
	route.HandleFunc("/form-login", formLogin).Methods("GET")
	route.HandleFunc("/login", login).Methods("POST")
	route.HandleFunc("/logout", logout).Methods("GET")

	fmt.Println("Server running on port 8000")
	http.ListenAndServe("localhost:8000", route)

}

type MetaData struct {
	UserId    int
	IsLogin   bool
	UserName  string
	FlashData string
}

var Data = MetaData{}

type Project struct {
	Id           int
	ProjectName  string
	StartDate    time.Time
	EndDate      time.Time
	sFormat      string
	enFormat     string
	Duration     string
	Description  string
	Technologies []string
	Image        string
	IsLogin      bool
}

type User struct {
	Id       int
	Name     string
	Email    string
	Password string
}

var dataProject = []Project{}

func homePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("views/index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Name"].(string)
		Data.UserId = session.Values["ID"].(int)
	}

	fm := session.Flashes("message")

	var flashes []string
	if len(fm) > 0 {
		session.Save(r, w)
		for _, fl := range fm {
			flashes = append(flashes, fl.(string))
		}
	}

	Data.FlashData = strings.Join(flashes, "")
	if Data.IsLogin != true {
		row, _ := connection.Conn.Query(context.Background(), "SELECT id, project_name, start_date, end_date, duration, description, technologies, image FROM tb_projects")

		var result []Project

		for row.Next() {
			var each = Project{}

			var err = row.Scan(&each.Id, &each.ProjectName, &each.StartDate, &each.EndDate, &each.Duration, &each.Description, &each.Technologies, &each.Image)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			result = append(result, each)

		}

		var response = map[string]interface{}{
			"Data":     Data,
			"Projects": result,
		}

		w.WriteHeader(http.StatusOK)
		tmpl.Execute(w, response)
	} else {
		row, _ := connection.Conn.Query(context.Background(), "SELECT id, project_name, start_date, end_date, duration, description, technologies, image FROM tb_projects WHERE tb_projects.author_id=$1 ORDER BY id DESC", Data.UserId)

		var result []Project

		for row.Next() {
			var each = Project{}

			var err = row.Scan(&each.Id, &each.ProjectName, &each.StartDate, &each.EndDate, &each.Duration, &each.Description, &each.Technologies, &each.Image)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			result = append(result, each)

		}

		var response = map[string]interface{}{
			"Data":     Data,
			"Projects": result,
		}

		w.WriteHeader(http.StatusOK)
		tmpl.Execute(w, response)
	}

}

func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("views/contact.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, nil)
}

func blogPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("views/add-project.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, nil)
}

func blogDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	var tmpl, err = template.ParseFiles("views/detail-project.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	var BlogDetail = Project{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT id, project_name, start_date, end_date, description, technologies, image FROM tb_projects WHERE id=$1", id).Scan(
		&BlogDetail.Id, &BlogDetail.ProjectName, &BlogDetail.StartDate, &BlogDetail.EndDate, &BlogDetail.Description, &BlogDetail.Technologies, &BlogDetail.Image,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	data := map[string]interface{}{
		"Project": BlogDetail,
	}

	fmt.Println(data)

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, data)
}

func sendDataAdd(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	projectName := r.PostForm.Get("project-name")
	startDate := r.PostForm.Get("start-date")
	endDate := r.PostForm.Get("end-date")
	var duration string
	description := r.PostForm.Get("desc-project")
	var techno []string
	techno = r.Form["techno"]

	dataContex := r.Context().Value("dataFile")
	image := dataContex.(string)

	FormatDate := "2006-01-02"
	startDateParse, _ := time.Parse(FormatDate, startDate)
	endDateParse, _ := time.Parse(FormatDate, endDate)

	hour := 1
	day := hour * 24
	week := hour * 24 * 7
	month := hour * 24 * 30
	year := hour * 24 * 365

	differHour := endDateParse.Sub(startDateParse).Hours()
	var differHours int = int(differHour)
	days := differHours / day
	weeks := differHours / week
	months := differHours / month
	years := differHours / year

	if differHours < week {
		duration = strconv.Itoa(int(days)) + " Days"
	} else if differHours < month {
		duration = strconv.Itoa(int(weeks)) + " Weeks"
	} else if differHours < year {
		duration = strconv.Itoa(int(months)) + " Months"
	} else if differHours > year {
		duration = strconv.Itoa(int(years)) + " Years"
	}

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_projects(author_id, project_name, start_date, end_date, duration, description, technologies, image) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", Data.UserId, projectName, startDate, endDate, duration, description, techno, image)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func deleteProject(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	fmt.Println(id)

	_, err := connection.Conn.Exec(context.Background(), "DELETE FROM tb_projects WHERE id=$1", id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func editProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset-utf=8")

	tmpl, err := template.ParseFiles("views/edit-project.html")

	if tmpl == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Message : " + err.Error()))
	}

	id, _ := strconv.Atoi((mux.Vars(r)["id"]))

	var update = Project{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT id, project_name, start_date, end_date, description, technologies, image FROM tb_projects WHERE id=$1", id).Scan(
		&update.Id, &update.ProjectName, &update.StartDate, &update.EndDate, &update.Description, &update.Technologies, &update.Image,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	update.sFormat = update.StartDate.Format("2006-01-02")
	update.enFormat = update.EndDate.Format("2006-01-02")
	fmt.Println(update.sFormat)

	response := map[string]interface{}{
		"ProjectData": update,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, response)
}

func updateProject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	id, _ := strconv.Atoi((mux.Vars(r)["id"]))

	projectName := r.PostForm.Get("project-name")
	startDate := r.PostForm.Get("start-date")
	endDate := r.PostForm.Get("end-date")
	var duration string
	description := r.PostForm.Get("desc-project")
	var techno []string
	techno = r.Form["techno"]

	dataContex := r.Context().Value("dataFile")
	image := dataContex.(string)

	FormatDate := "2006-01-02"
	startDateParse, _ := time.Parse(FormatDate, startDate)
	endDateParse, _ := time.Parse(FormatDate, endDate)

	hour := 1
	day := hour * 24
	week := hour * 24 * 7
	month := hour * 24 * 30
	year := hour * 24 * 365

	differHour := endDateParse.Sub(startDateParse).Hours()
	var differHours int = int(differHour)
	days := differHours / day
	weeks := differHours / week
	months := differHours / month
	years := differHours / year

	if differHours < week {
		duration = strconv.Itoa(int(days)) + " Days"
	} else if differHours < month {
		duration = strconv.Itoa(int(weeks)) + " Weeks"
	} else if differHours < year {
		duration = strconv.Itoa(int(months)) + " Months"
	} else if differHours > year {
		duration = strconv.Itoa(int(years)) + " Years"
	}

	_, error := connection.Conn.Exec(context.Background(), "UPDATE tb_projects SET project_name=$1, start_date=$2, end_date=$3, duration=$4, description=$5, technologies=$6, image=$7 WHERE id=$8", projectName, startDate, endDate, duration, description, techno, image, id)
	if error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + error.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func formRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var tmpl, err = template.ParseFiles("views/form-register.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, Data)
}

func register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var name = r.PostForm.Get("inputName")
	var email = r.PostForm.Get("inputEmail")
	var password = r.PostForm.Get("inputPassword")

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_users(name, email, password) VALUES($1, $2, $3)", name, email, passwordHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/form-login", http.StatusMovedPermanently)
}

func formLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content_Type", "text/html; charset=utf-8")

	tmpl, err := template.ParseFiles("views/form-login.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	fm := session.Flashes("Message")

	var flashes []string
	if len(fm) > 0 {
		session.Save(r, w)
		for _, fl := range fm {
			flashes = append(flashes, fl.(string))
		}
	}
	Data.FlashData = strings.Join(flashes, "")
	fmt.Println(Data.FlashData)

	response := map[string]interface{}{
		"DataSession": Data,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, response)
}

func login(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {

		log.Fatal(err)
	}

	email := r.PostForm.Get("inputEmail")
	password := r.PostForm.Get("inputPassword")

	user := User{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT * FROM tb_users WHERE email=$1", email).Scan(&user.Id, &user.Name, &user.Email, &user.Password)

	if err != nil {
		store := sessions.NewCookieStore([]byte("SESSION_KEY"))
		session, _ := store.Get(r, "SESSION_KEY")
		session.AddFlash("Email belum terdaftar!", "Message")
		session.Save(r, w)

		http.Redirect(w, r, "/form-login", http.StatusMovedPermanently)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) //convert pass for login
	if err != nil {
		store := sessions.NewCookieStore([]byte("SESSION_KEY"))
		session, _ := store.Get(r, "SESSION_KEY")
		session.AddFlash("Password Salah!", "Message")
		session.Save(r, w)

		http.Redirect(w, r, "/form-login", http.StatusMovedPermanently)
	}

	store := sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	session.Values["IsLogin"] = true
	session.Values["Name"] = user.Name
	session.Values["Email"] = user.Email // baru tak tambahin
	session.Values["ID"] = user.Id
	session.Options.MaxAge = 10800 //lama waktu login 3 hour

	session.AddFlash("Successfully Login!", "message")
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func logout(w http.ResponseWriter, r *http.Request) {
	store := sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")
	session.Options.MaxAge = -1
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
