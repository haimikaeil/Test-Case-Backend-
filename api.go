package main

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
    "io"
    "os"
    "path/filepath"
    "strings"
    "fmt"
)

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    
        tokenString := r.Header.Get("Authorization")
        if len(tokenString) == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Missing Authorization Header"))
			return
		}
		tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
        // fmt.Println(tokenString)
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if jwt.GetSigningMethod("HS256") != token.Method {
                return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
            }

            return []byte("secret"), nil
        })

        fmt.Println(token)
        fmt.Println(err)

        if token != nil && err == nil {
            log.Print("token verified")
        }else{
            w.WriteHeader(http.StatusUnauthorized)
            w.Write([]byte("Invalid Token"))
            return
        }
        next.ServeHTTP(w, r)
    })
}

func loginUser(w http.ResponseWriter, r *http.Request) {
    var users Users
	var Data []Users
	var response Response
    
    err := r.ParseMultipartForm(4096)
    if err != nil {
        panic(err)
    }
    
    Username := r.FormValue("Username")
    Password := r.FormValue("Password")

	db := connect()
	defer db.Close()

	log.Print(r)

	row, err := db.Query("Select * from user where Username=? and Password=md5(?)", Username, Password)
	if err != nil {
		log.Print(err)
	}

    var found bool
    found = false

	for row.Next(){
        found =true
        if err := row.Scan(&users.ID, &users.Username, &users.Password, &users.Nama_lengkap, &users.Foto); err != nil {
            log.Fatal(err.Error())
    
        } else{
            Data= append(Data, users)
        }
    }
    

    if(found){
        response.Status = 1
        response.Message = "Success"
        response.Data = Data
        sign := jwt.New(jwt.GetSigningMethod("HS256"))
	    token, err := sign.SignedString([]byte("secret"))
        response.Token = token

        if(err != nil){
            log.Fatal(err.Error())
        }
    }else{
        response.Status = 1
        response.Message = "User not found"
    }

	
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(response)

}

func getUserAll(w http.ResponseWriter, r *http.Request) {
	var users Users
	var arr_user []Users
	var response Response

	db := connect()
	defer db.Close()

	rows, err := db.Query("Select * from user")
	if err != nil {
		log.Print(err)
	}

	for rows.Next() {
		if err := rows.Scan(&users.ID, &users.Username, &users.Password, &users.Nama_lengkap, &users.Foto); err != nil {
			log.Fatal(err.Error())

		} else {
			arr_user = append(arr_user, users)
		}
	}

	response.Status = 1
	response.Message = "Success"
	response.Data = arr_user

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func registerUser(w http.ResponseWriter, r *http.Request) {
  var response Response
  
	db := connect()
	defer db.Close()

	err := r.ParseMultipartForm(4096)
	if err != nil {
		panic(err)
	}

	Username 		:= r.FormValue("Username")
	Password 		:= r.FormValue("Password")
	Nama_lengkap 	:= r.FormValue("Nama_lengkap")
	uploadedFile, handler, err := r.FormFile("Foto")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer uploadedFile.Close()

	dir, err := os.Getwd()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filename := handler.Filename
	fileLocation := filepath.Join(dir, "upload", filename)
	targetFile, err := os.OpenFile(fileLocation, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer targetFile.Close()

	if _, err := io.Copy(targetFile, uploadedFile); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("INSERT INTO user (Username, Password, Nama_lengkap, Foto) values (?, md5(?),?,?)",
		Username,
		Password,
		Nama_lengkap,
		filename,
	)

	if err != nil {
		log.Print(err)
	}

	response.Status = 1
	response.Message = "Success Add"
	log.Print("Insert data to database")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func updateUser(w http.ResponseWriter, r *http.Request) {
	var response Response

	db := connect()
	defer db.Close()

	err := r.ParseMultipartForm(4096)
	if err != nil {
		panic(err)
	}

	ID 				:= r.FormValue("ID")
	Username 		:= r.FormValue("Username")
	Password 		:= r.FormValue("Password")
	Nama_lengkap 	:= r.FormValue("Nama_lengkap")
	file, handler, err	:= r.FormFile("Foto")

	f, err := os.OpenFile(handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
    _, _ = io.Copy(f, file)

	_, err = db.Exec("UPDATE user set Username = ?, Password = md5(?), Nama_lengkap = ?, Foto = ? where ID = ?",
		Username,
		Password,
		Nama_lengkap,
		handler.Filename,
		ID,
	)

	if err != nil {
		log.Print(err)
	}

	response.Status = 1
	response.Message = "Success Update Data"
	log.Print("Update data to database")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	var response Response

	db := connect()
	defer db.Close()

	err := r.ParseMultipartForm(4096)
	if err != nil {
		panic(err)
	}

	ID := r.FormValue("ID")

	_, err = db.Exec("DELETE from user where ID = ?",
		ID,
	)

	if err != nil {
		log.Print(err)
	}

	response.Status = 1
	response.Message = "Success Delete Data"
	log.Print("Delete data to database")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}