package main

import (
	"database/sql"
	"fmt"
	"io"
	//"math"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

//--------------------------------------------------------------database part---------------------------------------------------------------------------------

type user struct {
	User_id string `json:"user_id"`
	Email   string `json:"email"`
}

func getUser(email string, password string) (user, error) {
	var u user
	query := "SELECT user_id, email, password FROM user WHERE email = ?"
	row := db.QueryRow(query, email)
	var hashedPassword string
	err := row.Scan(&u.User_id, &u.Email, &hashedPassword)
	if err != nil {
		return u, err
	}
	var uu user
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return uu, err
	}

	return u, nil

}

func registerUser(email string, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return err
	}
	query := "call addUser(?,?);"
	_, erro := db.Exec(query, email, string(hashedPassword))
	if erro != nil {
		return erro
	}
	return nil
}

func addfileentry(uuid string, actname string, ftype string, fsize float64,is_accepted bool ,owner string) (int,string) {
	var count int
	fmt.Println("dede", fsize)
	checkQuery := "SELECT COUNT(*) FROM metadata WHERE name_at_server = ? and owner = ?"
	err := db.QueryRow(checkQuery, uuid, owner).Scan(&count)
	if err != nil {
		fmt.Println(err)
		return 202,""
	}
	if count > 0 {
		return 202,""
	}
	quer := "insert into metadata (name_at_server,actualname,filetype,file_size,is_accepted,owner) values (?,?,?,?,?,?);"
	_, erri := db.Exec(quer, uuid, actname, ftype, fsize,is_accepted, owner)
	if erri != nil {
		fmt.Println(erri)
		return 202,""
	}
	r:= db.QueryRow("select used_storage from user where user_id = ?",owner); //mind it 
	var siz string;
	r.Scan(&siz);
	return 101,siz;
}

type File struct {
	Id         int    `json:"id"`
	Actualname string `json:"actualname"`
	Ftype      string `json:"filetype"`
	Date       string `json:"date"`
	FSize		string `json:"fsize"`
}
type Filewithuuid struct {
	Id           int    `json:"id"`
	Nameatserver string `json:"name_at_server"`
	Actualname   string `json:"actualname"`
	Ftype        string `json:"filetype"`
	FSize        string `json:"file_size"`
	Date         string `json:"date"`
}

func getfileloc(id int) (string,string, error) {
	query := `select name_at_server,filetype from metadata where id = ?`
	row, err := db.Query(query, id)
	if err != nil {
		fmt.Println("error occured when get file server name for downloading")
		return "","", err
	}
	var s1, s2 string
	for row.Next() {
		erro := row.Scan(&s1, &s2)
		if erro != nil {
			return "","", erro
		}
	}

	return s1,s2, nil

}

func getfiles(owner_id string,accepted bool) ([]File, error) {
	rows, err := db.Query("SELECT id, actualname, filetype,file_size, date FROM metadata WHERE owner = ? and is_accepted = ?;", owner_id, accepted)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []File
	for rows.Next() {
		var file File
		if err := rows.Scan(&file.Id, &file.Actualname, &file.Ftype, &file.FSize,&file.Date); err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return files, nil
}

func getfile(file_id string) (Filewithuuid, error) {
	row := db.QueryRow("SELECT id,name_at_server, actualname, filetype,file_size, date FROM metadata WHERE id = ?;", file_id)
	var file Filewithuuid
	if err := row.Scan(&file.Id, &file.Nameatserver, &file.Actualname, &file.Ftype,&file.FSize, &file.Date); err != nil {
		return file, err
	} else {
		return file, nil
	}

}

func getuseridformail(fmail string) (string, error) {
	row := db.QueryRow("select user_id from user where email = ?", fmail)
	var id string
	row.Scan(&id)
	if id == "" {
		return "", fmt.Errorf("user not found")
	}
	return id, nil
}

// ----------------------------------------------------------user section------------------------------------------------------------------------------
func signup(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")
	err := registerUser(email, password)
	if err != nil {
		c.JSON(400, gin.H{"message": "error registering user"})
		return
	}
	c.JSON(200, gin.H{"message": "user registered successfully"})
}

func authenticate(c *gin.Context) {
	tokenString := c.PostForm("token")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			fmt.Println("1")
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("your_secret_key"), nil
	})

	if err != nil || !token.Valid {
		fmt.Println("1",err)
		fmt.Println(tokenString)
		c.JSON(400, gin.H{"message": "invalid token"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		fmt.Println("2",err);
		c.JSON(400, gin.H{"message": "invalid token"})
		return
	}

	userID := claims["user_id"].(string)
	email := claims["email"].(string)

	c.JSON(200, gin.H{"user_id": userID, "email": email})
}
func signin(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")
	u, err := getUser(email, password)
	if err != nil {
		fmt.Println(err)
		c.JSON(401, gin.H{"message": "invalid email or password"})
		return
	}

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": u.User_id,
		"email":   u.Email,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString([]byte("your_secret_key"))
	if err != nil {
		c.JSON(500, gin.H{"message": "could not create token"})
		return
	}

	c.JSON(200, gin.H{"user": u, "token": tokenString})
	//c.JSON(200, gin.H{"user": u})
}

//------------------------------------------------------------share section-------------------------------------------------------------------------------------

func sharefile(c *gin.Context) {
	friend_mail := c.PostForm("fmail")
	file_id := c.PostForm("file_id")
	fi, err := getfile(file_id)
	if err != nil {
		fmt.Println("err",err)
		c.JSON(400, gin.H{"message": "file not found"})
		return
	}
	id, err := getuseridformail(friend_mail)
	if err != nil {
		fmt.Println("err",err)
		c.JSON(400, gin.H{"message": "friend email not found"})
		return
	}
	fsize, err := strconv.ParseFloat(fi.FSize, 64)
	if err != nil {
		c.JSON(500, gin.H{"message": "error converting file size"})
		return
	}
	result,siz := addfileentry(fi.Nameatserver, fi.Actualname, fi.Ftype, fsize, false,id)
	if result == 202 {
		c.JSON(500, gin.H{"message": "error sharing file"})
		return
	}

	c.JSON(200, gin.H{"message": "file shared successfully","updated_file_size":siz})
}

// ------------------------------------------------------------download section----------------------------------------------------------------------------




func getsharedfiles(c *gin.Context) {
	//filelist,err:=os.ReadDir("./userfiles/");
	owner := c.PostForm("user_id")
	results, errr := getfiles(owner,false)
	if errr != nil {
		fmt.Println("podang")
		c.JSON(400, gin.H{"message": "error retrieving files from database"})
		return
	}
	r:= db.QueryRow("select used_storage from user where user_id = ?",owner);
	var siz string;
	r.Scan(&siz);
	c.JSON(200, gin.H{"files": results,"size":siz})
	// if(err!=nil){
	// 	fmt.Println("cannot read userfiles");
	// 	c.JSON(400,gin.H{"message":"files not found"});
	// }
	// var fnames [] string;
	// for _,name:= range filelist{
	// 	fnames = append(fnames, name.Name());
	// }
	// c.JSON(200,gin.H{"files":fnames});
}

func get_file(c *gin.Context) {
	//filelist,err:=os.ReadDir("./userfiles/");
	owner := c.PostForm("user_id")
	results, errr := getfiles(owner,true)
	if errr != nil {
		fmt.Println("podang")
		c.JSON(400, gin.H{"message": "error retrieving files from database"})
		return
	}
	r:= db.QueryRow("select used_storage from user where user_id = ?",owner);
	var siz string;
	r.Scan(&siz);
	c.JSON(200, gin.H{"files": results,"size":siz})
	// if(err!=nil){
	// 	fmt.Println("cannot read userfiles");
	// 	c.JSON(400,gin.H{"message":"files not found"});
	// }
	// var fnames [] string;
	// for _,name:= range filelist{
	// 	fnames = append(fnames, name.Name());
	// }
	// c.JSON(200,gin.H{"files":fnames});
}

func download_file(c *gin.Context) {
	fileid := c.PostForm("fileid")
	id, err := strconv.Atoi(fileid)
	if err != nil {
		fmt.Println("file id is not an integer", err)
		c.JSON(300, gin.H{"message": "file id is not an integer"})
	}
	filename,filetypee, err := getfileloc(id)
	if err != nil {
		fmt.Println("file is not in server or already deleted", err)
		c.JSON(400, gin.H{"message": "file is not in server or already deleted"})
	}
	filepath := "./userfiles/" + filename
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Println("cannot open the file for download")
		c.JSON(400, gin.H{"message": "cnnot retrive the file"})
	}
	defer file.Close()
	stats, err2 := file.Stat()
	if err2 != nil {
		fmt.Println("cannot open the file for download")
		c.JSON(400, gin.H{"message": "cnnot retrive the file"})
	}
	c.Writer.Header().Set("Content-Disposition", "attachment; filename="+filename+"."+filetypee)
	c.Writer.Header().Set("Content-Type", "application/octet-stream")
	c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", stats.Size()))

	_, _ = io.Copy(c.Writer, file)
}

//-------------------------------------------------------------upload section start----------------------------------------------------------------

func init_upload(c *gin.Context) {
	uuid := c.PostForm("uuid")
	_, er := os.Stat("./uploads")
	if er != nil {
		os.Mkdir("./uploads", 0755)
	}
	err := os.Mkdir("./uploads/"+uuid, 0755)

	if err != nil {
		fmt.Println("cannot create folder to store the packets")
		c.JSON(400, gin.H{"message": "cannot initiate the upload"})
	}
	fmt.Println("folder created")
	c.JSON(200, gin.H{"message": "initiate successfull"})
}

func handleUpload(c *gin.Context) {
	ind := c.PostForm("packet_index")
	uuid := c.PostForm("uuid")
	file, err := c.FormFile("file")
	if err != nil {
		fmt.Println("got errer when recieving file")
		c.JSON(400, gin.H{"message": "cannot get the packet"})
	}
	errr := c.SaveUploadedFile(file, "./uploads/"+uuid+"/"+ind)
	if errr != nil {
		c.JSON(400, gin.H{"message": "some error occured when save packet no: " + ind})
	}
	c.JSON(200, gin.H{"message": "packet recieved succesfully"})

}

func completeUpload(c *gin.Context) {
	uuid := c.PostForm("uuid")
	namei := c.PostForm("filename")
	ext := path.Ext(namei)
	actualname := strings.TrimSuffix(namei, ext)
	filetype := ext[1:]
	ownerid := c.PostForm("user_id")
	file_list, err1 := os.ReadDir("./uploads/" + uuid)
	if err1 != nil {
		fmt.Println("folder doesnt exist")
		c.JSON(400, gin.H{"message": "folder doesnt exist"})
	}
	var file_name_list []string
	for _, file := range file_list {
		if file.IsDir() {
			fmt.Println("cannot recieve the file name")
		} else {
			file_name_list = append(file_name_list, file.Name())
		}
	}
	sort.Slice(file_name_list, func(i, j int) bool {
		indi, _ := strconv.Atoi(file_name_list[i])
		indj, _ := strconv.Atoi(file_name_list[j])
		return indi < indj

	})
	os.MkdirAll("./userfiles/", 0755)
	objfile, err2 := os.Create("./userfiles/" + uuid)
	if err2 != nil {
		fmt.Println("cannot create objfile", err2)
		c.JSON(400, gin.H{"message": "error when create obj file"})
	}
	defer objfile.Close()

	for _, f_name := range file_name_list {
		packet, err3 := os.Open("./uploads/" + uuid + "/" + f_name)
		if err3 != nil {
			fmt.Println("cannot open packets to write")
			c.JSON(400, gin.H{"message": "error when create obj file"})
		}
		defer packet.Close()
		packet.Seek(0, 0)
		_, err4 := objfile.ReadFrom(packet)
		if err4 != nil {
			fmt.Println("cannot write")
			c.JSON(400, gin.H{"message": "error when write"})
		}
		packet.Close()
	}
	stats, err := objfile.Stat()
	if err != nil {
		fmt.Println("cannot get file info", err)
		c.JSON(400, gin.H{"message": "error when getting file info"})
		return
	}
	fmt.Println(stats.Size())
	fileSizeInGB := float64(stats.Size())/(1024*1024)
	fmt.Println(fileSizeInGB)
	objfile.Close()
	err5 := os.RemoveAll("./uploads/" + uuid)
	if err5 != nil {
		fmt.Println("error delete when a file")
		c.JSON(400, gin.H{"message": "error when create obj file"})

	}
	fmt.Println("file deleted successfully", uuid)

	result,siz := addfileentry(uuid, actualname, filetype, fileSizeInGB,true, ownerid)
	if result == 202 {
		fmt.Println("there is an error when inserting the entry")
	} else {
		fmt.Println("success")
	}

	c.JSON(200, gin.H{"message": "fileuploaded succesfully","size":siz})

}

func rejectfile(c *gin.Context){
	id:= c.PostForm(("file_id"))
	query := "DELETE FROM metadata WHERE id = ?"
	_, err := db.Exec(query, id)
	if err != nil {
		fmt.Println("error updating file acceptance", err)
		c.JSON(500, gin.H{"message": "error updating file acceptance"})
		return
	}
	c.JSON(200, gin.H{"message": "file accepted successfully"})
}

func acceptfile(c *gin.Context){
	id:= c.PostForm(("file_id"))
	query := "UPDATE metadata SET is_accepted = true WHERE id = ?"
	_, err := db.Exec(query, id)
	if err != nil {
		fmt.Println("error updating file acceptance", err)
		c.JSON(500, gin.H{"message": "error updating file acceptance"})
		return
	}
	c.JSON(200, gin.H{"message": "file accepted successfully"})
}

func deletefile(c *gin.Context) {
	id := c.PostForm("file_id")
	file, err := getfile(id)
	if err != nil {
		fmt.Println("file not found")
		c.JSON(500, gin.H{"message": "error deleting file"})
		return

	}

	query := "DELETE FROM metadata WHERE id = ?"
	_, erri := db.Exec(query, id)
	if erri != nil {
		fmt.Println("file not found",erri)
		c.JSON(500, gin.H{"message": "error deleting file"})
		return
	}
	var count int
	checkQuery := "SELECT COUNT(*) FROM metadata WHERE name_at_server = ?"
	err = db.QueryRow(checkQuery, file.Nameatserver).Scan(&count)
	if err != nil {
		fmt.Println("file not found",err)
		c.JSON(500, gin.H{"message": "error checking file count"})
		return
	}
	if count > 0 {
		c.JSON(200, gin.H{"message": "file deleted successfully"})
		return
	}
	fpath := "./userfiles/" + file.Nameatserver
	err3 := os.Remove(fpath)
	if err3 != nil {
		fmt.Println("file not found",err3)
		c.JSON(500, gin.H{"message": "error deleting file"})
		return
	}
	c.JSON(200, gin.H{"message": "file deleted successfully"})
}

// Health check endpoint for Docker
func healthCheck(c *gin.Context) {
	// Check database connection
	if err := db.Ping(); err != nil {
		c.JSON(500, gin.H{"status": "unhealthy", "database": "disconnected"})
		return
	}
	c.JSON(200, gin.H{"status": "healthy", "database": "connected", "timestamp": time.Now().Unix()})
}

// Get database connection string from environment variables
func getDatabaseConnectionString() string {
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "3306"
	}
	
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "root"
	}
	
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "vithu"
	}
	
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "vcloud"
	}
	
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPassword, dbHost, dbPort, dbName)
}

// -------------------------------------------------------------------------------------end upload section --------------------------------------------------------
var db *sql.DB

func main() {
	r := gin.Default()
	de := "hi.exe"
	parts := strings.Split(de, ".")
	fmt.Println(parts)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://chat.shancloudservice.com", "https://v-cloud-mlag.vercel.app", "http://localhost:3000", "http://192.168.130.41:3000", "http://10.10.21.213:3000","http://192.168.94.41:3000","http://192.168.131.140:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	var err error

	// Get database connection string from environment variables
	connectionString := getDatabaseConnectionString()
	db, err = sql.Open("mysql", connectionString)
	if err != nil {
		//log.Fatalf("Error opening database: %v", err)
		fmt.Println("cant connect to database:", err)
		return
	}

	if err := db.Ping(); err != nil {
		fmt.Println("cant ping database:", err)
		//log.Fatalf("Error connecting to database: %v", err)
		return
	}
	
	fmt.Println("Successfully connected to database!")
	
	// Health check endpoint
	r.GET("/health", healthCheck)
	
	// API endpoints
	r.POST("/init-upload", init_upload)
	r.POST("/upload", handleUpload)
	r.POST("/complete-upload", completeUpload)
	r.POST("/get-files", get_file)
	r.POST("/download", download_file)
	r.POST("/signup", signup)
	r.POST("/signin", signin)
	r.POST("/authenticate", authenticate)
	r.POST("/share", sharefile)
	r.POST("/deletefile", deletefile)
	r.POST("/accept",acceptfile)
	r.POST("/reject",rejectfile)
	r.POST("/get-shared-files",getsharedfiles)
	r.Run("0.0.0.0:8080")

}
