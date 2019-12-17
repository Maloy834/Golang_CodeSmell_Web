package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)
type User struct {
	Id    int     `json:"id"`
	StPath  string  `json:"stPath"`
	MtPath string  `json:"mtPath"`
	SumPath string  `json:"sumPath"`
	ListPath string  `json:"listPath"`
}

func uploadFile(w http.ResponseWriter, r *http.Request) {

	//fmt.Println("File Upload Endpoint Hit")
	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	// Create a temporary file within our temp-images directory that follows
	// a particular naming pattern
	tempFile, err := ioutil.TempFile("temp-images", handler.Filename)
	if err != nil {
		fmt.Println(err)
	}
	//println(tempFile.Name())
	defer tempFile.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}
	// write this byte array to our temporary file
	tempFile.Write(fileBytes)

    pr:= analyse(tempFile.Name())
    w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pr)
	// return that we have successfully uploaded our file!
	//fmt.Fprintf(w, "Successfully Uploaded File\n")
}


func renderError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(message))
}
func analyse(projectName string ) User {
	files, err := Unzip(projectName, "output-folder")
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Println("Unzipped:\n" + strings.Join(files, "\n"))
	tr:=ReadAllfiles(files,projectName)
	os.RemoveAll("output-folder")
   return tr
}
func Unzip(src string, dest string) ([]string, error) {

	var filenames []string

	readfile, err := zip.OpenReader(src)
	if err != nil {
		println("debug")
		return filenames, err
	}
	defer readfile.Close()
	//println("debug")
	for _, f := range readfile.File {
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

func ReadAllfiles(file []string,projectName string) User {
	var all_struct []Struct
	var all_method [] Method
	for i := 0; i < len(file); i++ {
		//if os.Stat(file[i])
		f, err := os.Stat(file[i])

		if err != nil {
			log.Fatal(err)
		}
		if f.IsDir() == false && filepath.Ext(file[i]) == ".go" {
			structs, methods := parseFile(file[i])
			/*for _,st:= range structs{
				all_struct=append(all_struct,st)
			}*/
			//all_struct=append(all_struct,structs)
			//println(len(structs))
			for _, mlist := range methods {
				for i, c := range structs {
					//structs[i].addAllMethods(methods)
					structs[i].Totalmethods = methods
					if mlist.PkgName == c.PkgName && mlist.StructName == c.StructName {
						structs[i].addUsedMethod(mlist)
						//structs[i].Totalmethods = methods
					}
				}
			}
			for _, st := range structs {

				all_struct = append(all_struct, st)
			}
			for _, mt := range methods {
				all_method = append(all_method, mt)
			}


		}

	}
	for i, st := range all_struct {
		st.NDC = calculateNDC(st)
		st.NP = calculateNP(st)
		st.ATFD = calculateATFD(st)
		st.TCC = calculateTCC(st)
		st.WMC = calculateWMC(st.Methods)
		st.ATFD = calculateATFD(st)
		st.TCC = calculateTCC(st)
		st.LCOM= calculateLCOM(st,st.Methods)
		st.GodStruct = checkGodStruct(st)
		//st.DataStruct=checkDataStruct(st)
		all_struct[i] = st
	}

	for j,m := range all_method{
		m.calculateCallerMethod(all_method)
		//println(m.FuncName,"  ",m.CM,"" ,m.CC)
		m.FDP=calculateFDP(all_struct,m)
		m.FeatureEnvy=checkFeatureEnvy(m)
		m.ShortGunSurgery=checkShortGunSurgery(m)
		m.BrainMethod=checkBrainMethod(m)
		m.LongParameter=checkLongParameter(m)
		all_method[j]=m
	}

	structFile:=analyseStruct(all_struct,projectName)
	methodFile:=analyseMethods(all_method,projectName)
	summarryFile:=writeCodeSmellSumarry(all_struct,all_method,projectName)
	smellListFile:=writeCodeSmellList(all_struct,all_method,projectName)

	stFileName:=filepath.Base(structFile.Name())
	println(stFileName)
	mtFileName:=filepath.Base(methodFile.Name())
	sumFileName:=filepath.Base(summarryFile.Name())
    smellFileName:=filepath.Base(smellListFile.Name())
	user := User{
		Id: 1,
		StPath:  stFileName,
		MtPath: mtFileName,
		SumPath: sumFileName,
		ListPath: smellFileName,
	}
	return user
	//checkMethodExtract(methods, file[i],structs)
}
func setupRoutes() {

	router := mux.NewRouter()
	//http.Handle("/", router)
	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	methods:= handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"})
	origins:= handlers.AllowedOrigins([]string{"*"})
	router.HandleFunc("/upload", uploadFile).Methods("GET", "POST", "PUT", "HEAD", "OPTIONS")
	router.HandleFunc("/check", check).Methods("GET", "POST", "PUT", "HEAD", "OPTIONS")

	//fs := http.FileServer(http.Dir("temp-images"))
	router.PathPrefix("/files/").Handler(
		http.StripPrefix("/files/", http.FileServer(http.Dir("temp-images/"))))

	//http.Handle("/files/", http.StripPrefix("/files", fs))
	//http.Handle("/files/", http.StripPrefix("/files", fs))
	//router.HandleFunc("/download", downloadFile).Methods("GET", "POST", "PUT", "HEAD", "OPTIONS")
	//http.HandleFunc("/upload", uploadFile)
	http.ListenAndServe(":8080", handlers.CORS(headers,methods,origins)(router))
	//http.ListenAndServe(":8082",nil)
}
func check(w http.ResponseWriter, r *http.Request){
	/*w.Header().Set("Content-Disposition", "attachment; filename=Wiki.png")
	w.Header().Set("Content-Type", w.Request.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", w.Request.Header.Get("Content-Length"))*/
}
func main() {
	 fmt.Println("Server Start Now on port: 8080")
	 setupRoutes()

}
