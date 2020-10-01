package main

type Users struct {
	ID        string `form:"ID" json:"ID"`
	Username string `form:"Username" json:"Username"`
	Password  string `form:"Password" json:"Password"`
	Nama_lengkap  string `form:"Nama_lengkap" json:"Nama_lengkap"`
	Foto  string `form:"Foto" json:"Foto"`

}

type Response struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    []Users
	Token	string
}