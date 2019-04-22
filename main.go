package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strings"
	"strconv"
	"errors"
  "log"
  "path/filepath"
	"github.com/gocarina/gocsv"
 	"github.com/otiai10/copy"
)

type Docs []struct {
	Id         string `json:"_id"`
	ProductUri string `json:"productUri"`
	Meta       struct {
		Source string `json:"source"`
		// Test string `json:"test"`
	} `json:"meta"`
	General    struct {
		Manufacturer string   `json:"manufacturer"`
		Model        string   `json:"model"`
		Year         int      `json:"year"`
		Msrp         float64  `json:"msrp"`
		Category     string   `json:"category"`
		Subcategory  string   `json:"subcategory"`
		Description  string   `json:"description"`
		Countries    []string `json:"countries"`
	} `json:"general"`
	Images []struct {
		Src      string `json:"src"`
		Desc     string `json:"desc"`
		Longdesc string `json:"longdesc"`
	} `json:"images,omitempty"`
	Videos []struct {
		Src      string `json:"src"`
		Desc     string `json:"desc"`
		Longdesc string `json:"longdesc"`
	} `json:"videos,omitempty"`
	Features    []string `json:"features,omitempty"`
	Options     []string `json:"options,omitempty"`
	Attachments []struct {
		AttachmentDescription     string `json:"attachmentDescription"`
		AttachmentLocation        string `json:"attachmentLocation"`
		AttachmentLongDescription string `json:"attachmentLongDescription"`
	} `json:"attachments,omitempty"`
	Operational         map[string]Specs `json:"operational,omitempty"`
	Other               map[string]Specs `json:"other,omitempty"`
	EngineDrivetrain    map[string]Specs `json:"engineDrivetrain,omitempty"`
	EngineAndDriveTrain map[string]Specs `json:"engineAndDriveTrain,omitempty"`
	Dimensions          map[string]Specs `json:"dimensions,omitempty"`
	Hydraulics          map[string]Specs `json:"hydraulics,omitempty"`
	Engine              map[string]Specs `json:"engine,omitempty"`
	Weights             map[string]Specs `json:"weights,omitempty"`
	Measurements        map[string]Specs `json:"measurements,omitempty"`
	Body                map[string]Specs `json:"body,omitempty"`
	Electrical          map[string]Specs `json:"electrical,omitempty"`
	Battery             map[string]Specs `json:"battery,omitempty"`
	Drivetrain          map[string]Specs `json:"drivetrain,omitempty"`
}

//nested inside Parent Spec Names
type Specs struct {
	Desc  string `json:"desc"`
	Label string `json:"label"`
}
type Image struct {
	Src      string `json:"src"`
	Desc     string `json:"desc"`
	Longdesc string `json:"longdesc"`
}

type PatchId struct {
	Data []struct {
		Id      string `json:"_id"`
		General struct {
			Manufacturer string `json:"manufacturer"`
			Model        string `json:"model"`
			Year         int    `json:"year"`
			Category     string `json:"category"`
			Subcategory  string `json:"subcategory"`
		} `json:"general"`
	} `json:"data"`
}

type PatchUpdatedAt struct {
	Data struct {
		UpdatedAt string `json:"updated_at"`
	} `json:"data"`
}

type CrsTrims struct {
	ProdType         string  `csv:"ProdType"`
	MakeId           string  `csv:"MakeId"`
	ModelId          string  `csv:"ModelId"`
	ModelYear        float64 `csv:"ModelYear"`
	ManufacturerName string  `csv:"ManufacturerName"`
	ModelName        string  `csv:"ModelName"`
	TrimId           string  `csv:"TrimId"`
	TrimName         string  `csv:"TrimName"`
	TrimPhoto        string  `csv:"TrimPhoto"`
	Msrp             float64  `csv:"MSRP"`
	DisplayName      string  `csv:"DisplayName"`
}

type CrsFeatures struct {
	TrimId        string  `csv:"TrimId"`
	PackageId     string   `csv:"PackageId"`
	AttributeId   string    `csv:"AttributeId"`
	FeatureName   string    `csv:"FeatureName"`
	AttributeName string     `csv:"AttributeName"`
	Value         string    `csv:"Value"`
}

type CrsOptions struct {
	TrimId        string  `csv:"TrimId"`
	PackageId     string   `csv:"PackageId"`
	AttributeId   string    `csv:"AttributeId"`
	FeatureName   string    `csv:"FeatureName"`
	AttributeName string     `csv:"AttributeName"`
	Value         string    `csv:"Value"`
}

type CrsSample struct {
	TrimId        string  `csv:"TrimId"`
	PackageId     string   `csv:"PackageId"`
	AttributeId   string    `csv:"AttributeId"`
	FeatureName   string    `csv:"FeatureName"`
	AttributeName string     `csv:"AttributeName"`
	Value         string    `csv:"Value"`
}

type CrsMfrs struct {
	ProdType         string  `csv:"ProdType"`
	MakeId           string  `csv:"MakeId"`
	ManufacturerName string  `csv:"ManufacturerName"`
}

type CrsLogic struct {
	TrimId    string  `csv:"TrimId"`
	PackageId string  `csv:"PackageId"`
	RuleType  string  `csv:"RuleType"`
	Logic     string  `csv:"Logic"`
}

type CrsPhotoGallery struct {
	PhotoMapId int      `csv:"photomapid"`
	TrimId     string   `csv:"trimid"`
	PackageId  string   `csv:"packageid"`
	PhotoName  string   `csv:"photoname"`
	Tags       string    `csv:"tags"`
}

type CrsPackages struct {
	TrimId       string `csv:"TrimId"`
	PackageId    string `csv:"PackageId"`
	PackageCode  string `csv:"PackageCode"`
	PackageTitle string `csv:"PackageTitle"`
	Msrp         string `csv:"Msrp"`
}
type CrsSpecs struct {
	TrimId       string `csv:"TrimId"`
	PackageId    string `csv:"PackageId,omitempty"`
	PackageTitle    string `csv:"PackageTitle,omitempty"`
	AttributeId  string `csv:"AttributeId,omitempty"`
	FeatureName  string `csv:"FeatureName,omitempty"`
	AttributeName string `csv:"AttributeName,omitempty"`
	Value  string `csv:"Value,omitempty"`
}

type CrsSpecParentNames struct {
	SpecParentName string `csv:"SpecParentName"`
}

type CrsCategories struct {
	TrimId       string `csv:"TrimId"`
  Value        string `csv:"Value"`
	ModelName     string  `csv:"ModelName"`
  ProdType     string  `csv:"ProdType"`
  MappedCategory string `csv:"mappedCategory"`
}

type ManufacturersUpdated struct{
	OEM_Name string `csv:"OEM_Name"`
	ManufacturerName string `csv:"ManufacturerName"`

}

// Token ...
type Token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

func createFile(path string) {
	// detect if file exists
	var _, err = os.Stat(path)

	// create file if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		if err != nil {
			panic(err)
		}
		defer file.Close()
	}
}

func deleteFile(path string) {
	// delete file
	os.Remove(path)
}

var url string

func getAPI() string{
	fmt.Print("Enter api discordia or igneous: ")
	var api string
	fmt.Scanln(&api)

	fmt.Print("Enter api type stage or prod: ")
	var apiType string
	fmt.Scanln(&apiType)

	if (api == "igneous"){
		if apiType == "stage"{
			url = "https://api.stage.cwsplatform.com/specs"
		}else if apiType == "prod"{
			url = "https://api.prod.cwsplatform.com/specs"
		}
	}else if api == "discordia"{
			if apiType == "stage"{
				url = "http://127.0.0.1:5000/v1/model/"
			}else if apiType == "prod"{
				url = "https://discordia.blackbook.tilabs.tech/v1/model/"
			}
		}else{
			err :=errors.New("Please enter valid input")
			if err != nil {
				log.Fatal(err)
			}
		}
		return url
	}

	func ObtainNebulousToken() string {
		// url := os.Getenv("NEB_TOKEN_ENDPOINT")
		//
		// clientID := os.Getenv("NEB_CLIENT_ID")
		// clientSecret := os.Getenv("NEB_CLIENT_SECRET")

		url := "https://apis.traderonline.com/vLatest/token"
		fmt.Println("token endpoint url is",url)
		clientID := "blackbook"
		clientSecret := "#ZoDDn08ZlQjncScAfs1?3URoX8JPDZN"
	 fmt.Println("clientsecret is",clientSecret)

		nebulousToken := ""

		payload := strings.NewReader("------WebKitFormBoundary7MA4YWxkTrZu0gW\r\nContent-Disposition: form-data; name=\"client_id\"\r\n\r\n" + clientID + "\r\n------WebKitFormBoundary7MA4YWxkTrZu0gW\r\nContent-Disposition: form-data; name=\"client_secret\"\r\n\r\n" + clientSecret + "\r\n------WebKitFormBoundary7MA4YWxkTrZu0gW\r\nContent-Disposition: form-data; name=\"grant_type\"\r\n\r\nclient_credentials\r\n------WebKitFormBoundary7MA4YWxkTrZu0gW--")
		req, err := http.NewRequest("POST", url, payload)
		if err != nil {
			log.Println("An issue occured while creating the new request to obtain a nebulous token, the reported error was: " + err.Error())
			return nebulousToken
		}
		req.Header.Add("content-type", "multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW")
		req.Header.Add("Cache-Control", "no-cache")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("An issue arose while attempting to capture the response from nebulous to obtain a token, the reported error was: " + err.Error())
			res.Body.Close()
			return nebulousToken
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println("An issue occurred during the reading of the response body, the reported error was: " + err.Error())
			res.Body.Close()
			return nebulousToken
		}
		defer res.Body.Close()

		var token Token

		err = json.Unmarshal(body, &token)
		if err != nil {
			log.Println("Unable to marshal the nebulous token response into token struct, the reported error was: " + err.Error())
			return nebulousToken
		}

		nebulousToken = token.AccessToken
		fmt.Println(nebulousToken)
		fmt.Println("In obtain token function")
		return nebulousToken
	}

func postJson() {
	nebulousToken:=ObtainNebulousToken()
	docs := getDocs("out.json")
	deleteFile("nonExistingManufacturers.csv")
	fmt.Println("deleted nonExistingManufacturers.csv")

	createFile("nonExistingManufacturers.csv")
	fmt.Println("created nonExistingManufacturers.csv")

	deleteFile("Report.csv")
	fmt.Println("deleted report.csv")

  createFile("Report.csv")
	fmt.Println("created report.csv")

	var link map[string]int
	link = make(map[string]int)
	f, err := os.OpenFile("nonExistingManufacturers.csv", os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	writer := csv.NewWriter(f)

	f1, err := os.OpenFile("Report.csv", os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	defer f1.Close()
	writer1 := csv.NewWriter(f1)

	for s := 0; s < len(docs); s++ {
		fmt.Println(s)
		statCode := ""
		bodyString := ""
		// idStr, _ := getSpecId(docs,s)
		mJ, _ := json.Marshal(docs[s])

		fmt.Println("Token for model in post json", nebulousToken)

		// This check if the model is already existing in discordia.
		checkStatus:=checkIfRecordExists(docs,s,nebulousToken)
		fmt.Println("Token print loop in post json",s)

		// If model does not exist then it posts it and repsonse is added in the report file. Otherwise skips.
   if (checkStatus==false){
		statCode,bodyString=postRecord(mJ,nebulousToken)
		// fmt.Println(idStr)
		// if idStr != "" {
		// 		statCode, bodyString = patchRecord(idStr, mJ)
		// } else {
		// 	statCode, bodyString = postRecord(mJ)
		// }
		// if idStr == ""{
		// 	statCode, bodyString = postRecord(mJ)
		// }
		// fmt.Println(statCode)
		// fmt.Println(bodyString)
		// if bodyString== strings.Replace("{'error':true,'msg':{'general.manufacturer':['Manufacturer was not found.']}}","'","\"",-1) {
		// 	var csvData []string
		// 	csvData = append(csvData, statCode)
		// 	csvData = append(csvData, docs[s].General.Manufacturer)
		// 	csvData = append(csvData, bodyString)
		// 	writer.Write(csvData)
		// 	writer.Flush()
		// }
		if statCode != "200 OK" && statCode != "201 Created" && statCode != "" {
				if docs[s].Id == "" {
					docs[s].Id = strconv.Itoa(s)
				}
				var csvData []string
				csvData = append(csvData, statCode)
				csvData = append(csvData, docs[s].General.Manufacturer)
				csvData = append(csvData, bodyString)
				writer.Write(csvData)
				writer.Flush()
			}
		link = countStatus(statCode, link)
	}else{
		statCode="500 Duplicate"
		bodyString=docs[s].General.Model+ " already exists in discordia. So skipped."
		var csvData []string
		csvData = append(csvData, docs[s].General.Manufacturer,docs[s].General.Model)
		csvData = append(csvData, bodyString)
		writer.Write(csvData)
		writer.Flush()
		link = countStatus(statCode, link)
	}
}
	var csvData1 []string
	csvData1 = append(csvData1, "Report")
	writer1.Write(csvData1)
	writer1.Flush()
	for key, value := range link {
		var csvData []string
		csvData = append(csvData, key)
		csvData = append(csvData, strconv.Itoa(value))
		writer1.Write(csvData)
		writer1.Flush()
	}
}


func countStatus(statCode string, link map[string]int) map[string]int {
	if len(link) == 0 {
		link[statCode] = 1
	} else {
		if _, ok := link[statCode]; ok {
			link[statCode] = link[statCode] + 1
		} else {
			link[statCode] = 1
		}
	}
	return link
}

func checkIfRecordExists(docs Docs,s int,nebulousToken string) bool{
	type GetError struct {
		Error  string `json:"error"`
	}
	showModel:="https://discordia.blackbook.tilabs.tech/v1/models"
	man:=docs[s].General.Manufacturer
	cat:=docs[s].General.Category
	subcat:=docs[s].General.Subcategory
	year:=docs[s].General.Year
	model:=docs[s].General.Model
	showModelUrl:=showModel+"?manufacturer="+man+"&category="+cat+"&subcategory="+subcat+"&year="+strconv.Itoa(year)+"&model="+model
	showModelUrl=strings.Replace(showModelUrl, " ", "%20", -1)

	// try to get a record if exists
	getReq, getErr := http.NewRequest("GET", showModelUrl,nil)
	if getErr != nil {
		panic(getErr)
	}
	getReq.Header.Add("Authorization", nebulousToken)
	getResp, getErr := http.DefaultClient.Do(getReq)
	if getErr != nil {
		fmt.Println(getErr)
		fmt.Println("error","could not get record")
	}
	body, err := ioutil.ReadAll(getResp.Body)
	if err != nil {
		panic(getErr)
	}

	var er GetError
	json.Unmarshal(body,&er)
	if(er.Error==""){
		return true
	}else{
		return false
	}

}

func postRecord(mJ []byte,nebulousToken string) (string, string) {
	fmt.Println("Im in post record function")
	statCode := ""
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(mJ))
	if err != nil {
		panic(err)
	}
	fmt.Println(url)
	fmt.Println(nebulousToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", nebulousToken)
	// client := &http.Client{}
	// resp, err := client.Do(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		fmt.Println("In if err loop")
		return "error","no patch"
	}
	fmt.Println("Response Status from Discordia:", resp.Status)
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	statCode = resp.Status
	fmt.Println(statCode)
	return statCode, bodyString
}

func getDocs(path string) Docs{
	raw, err := os.Open(path)
	defer raw.Close()

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	byteRaw, _ := ioutil.ReadAll(raw)

	var d Docs
	json.Unmarshal(byteRaw, &d)
	return d
}

// func getSpecId(docs Docs, s int) (string, string) {
// 	manufacturer := docs[s].General.Manufacturer
// 	modelName := docs[s].General.Model
// 	year := docs[s].General.Year
// 	category := docs[s].General.Category
// 	idStr := ""
// 	updatedAt := ""
// 	urlString := url + "?manufacturer=" + manufacturer + "&model=" + modelName + "&year=" + strconv.Itoa(year) + "&category=" + category
// 	urlStringFinal := strings.Replace(urlString, " ", "%20", -1)
// 	resp, err := http.Get(urlStringFinal)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	if resp != nil {
// 		defer resp.Body.Close()
// 		body, err := ioutil.ReadAll(resp.Body)
// 		if err != nil {
// 			panic(err)
// 		}
// 		var p PatchId
// 		err = json.Unmarshal(body, &p)
// 		if err != nil {
// 			panic(err)
// 		}
// 		if len(p.Data) != 0 {
// 			idStr = p.Data[0].Id
// 			urlGetUpdatedAt := url + "/" + idStr
// 			resp, err = http.Get(urlGetUpdatedAt)
// 			if err != nil {
// 				panic(err)
// 			}
// 			if resp != nil {
// 				defer resp.Body.Close()
// 				body, err := ioutil.ReadAll(resp.Body)
// 				if err != nil {
// 					panic(err)
// 				}
// 				var p PatchUpdatedAt
// 				err = json.Unmarshal(body, &p)
// 				if err != nil {
// 					panic(err)
// 				}
// 				updatedAt = p.Data.UpdatedAt
// 			}
// 		}
// 	}
// 	return idStr, updatedAt
// }


func in_array(val string, array []string) (exists bool) {
    exists = false

    for _, v := range array {
        if val == v {
            exists = true
            return
        }
    }

    return
}

//get general and trims from file to struct
func getCtFromTrimsFile() []CrsTrims{
	fmt.Println("getCtFromTrimsFile");
	ct := []CrsTrims{}

	trimsFile, err := os.Open("Data_2019_03_01/PS_Trims.csv")
	if err != nil {
		fmt.Println(err)
	}
	defer trimsFile.Close()
	reader := csv.NewReader(trimsFile)
	reader.FieldsPerRecord = -1

	if err := gocsv.UnmarshalFile(trimsFile, &ct); err != nil {
		panic(err)
	}
	return ct
}

//get features from file to struct
func getCfFromFeaturesFile() []CrsFeatures{
	fmt.Println("getCfFromFeaturesFile");
	cf := []CrsFeatures{}

	featuresFile, err := os.Open("Data_2019_03_01/PS_Features.csv")
	if err != nil {
		fmt.Println(err)
	}
	defer featuresFile.Close()
	reader := csv.NewReader(featuresFile)
	reader.FieldsPerRecord = -1

	if err := gocsv.UnmarshalFile(featuresFile, &cf); err != nil {
		panic(err)
	}
	return cf
}

//get packages from file to struct
func getCpFromPackagesFile() []CrsPackages{
	cp := []CrsPackages{}

	packagesFile, err := os.Open("Data_2019_03_01/pkgs.csv")
	if err != nil {
		fmt.Println(err)
	}
	defer packagesFile.Close()
	reader := csv.NewReader(packagesFile)
	reader.FieldsPerRecord = -1

	if err := gocsv.UnmarshalFile(packagesFile, &cp); err != nil {
		panic(err)
	}
	return cp
}

//get general data from file to struct
func getCsdFromSampleDataFile() []CrsSample{
	fmt.Println("getCsdFromSampleDataFile");
	csd := []CrsSample{}

	sampleFile, err := os.Open("Data_2019_03_01/PS_SampleData.csv")
	if err != nil {
		fmt.Println(err)
	}

	defer sampleFile.Close()
	reader := csv.NewReader(sampleFile)
	reader.FieldsPerRecord = -1

	if err := gocsv.UnmarshalFile(sampleFile, &csd); err != nil {
		panic(err)
	}
	return csd
}

//get options from file to struct
func getCoFromOptionsFile() []CrsOptions{
	fmt.Println("getCoFromOptionsFile");
	co := []CrsOptions{}

	optionsFile, err := os.Open("Data_2019_03_01/PS_Options.csv")
	if err != nil {
		fmt.Println(err)
	}
	defer optionsFile.Close()
	reader := csv.NewReader(optionsFile)
	reader.FieldsPerRecord = -1

	if err := gocsv.UnmarshalFile(optionsFile, &co); err != nil {
		panic(err)
	}
	return co
}

//get photos from photogallery file
func getCpgFromPhotoGalleryFile() []CrsPhotoGallery{
	fmt.Println("getCpgFromPhotoGalleryFile");
	cpg := []CrsPhotoGallery{}

	PhotoGalleryFile, err := os.Open("Data_2019_03_01/photogallery.csv")
	if err != nil {
		fmt.Println(err)
	}
	defer PhotoGalleryFile.Close()
	reader := csv.NewReader(PhotoGalleryFile)
	reader.FieldsPerRecord = -1

	if err := gocsv.UnmarshalFile(PhotoGalleryFile, &cpg); err != nil {
		panic(err)
	}
	return cpg
}

//get specs from file to struct
func getCsFromSpecsFile() []CrsSpecs{
	fmt.Println("getCsFromSpecsFile");
	cs := []CrsSpecs{}

	specsFile, err := os.Open("Data_2019_03_01/PS_Specs_withpkgs.csv")
	if err != nil {
		fmt.Println(err)
	}
	defer specsFile.Close()
	reader := csv.NewReader(specsFile)
  reader.FieldsPerRecord = -1

	if err := gocsv.UnmarshalFile(specsFile, &cs); err != nil {
		panic(err)
	}
	// fmt.Println(cs)
	return cs
}

//get categories from file to struct
func getCaFromCategoryMappingFile() []CrsCategories{
	fmt.Println("getCaFromCategoriesAvailableFile");
	ca := []CrsCategories{}

	CategoryMappingFile, err := os.Open("Data_2019_03_01/CategoryMapping.csv")
 	if err != nil {
 		fmt.Println(err)
 	}
	defer CategoryMappingFile.Close()
	reader := csv.NewReader(CategoryMappingFile)
	reader.FieldsPerRecord = -1

	if err := gocsv.UnmarshalFile(CategoryMappingFile, &ca); err != nil {
		panic(err)
	}
	return ca
}

// func getUpdatedManufacturersFile() []ManufacturersUpdated{
//
// 	mu := []ManufacturersUpdated{}
//
// 	ManufacturersUpdatedFile, err := os.Open("Data/Oems_Updated_to_2018.csv")
//  	if err != nil {
//  		fmt.Println(err)
//  	}
// 	defer ManufacturersUpdatedFile.Close()
// 	reader := csv.NewReader(ManufacturersUpdatedFile)
// 	reader.FieldsPerRecord = -1
//
// 	if err = gocsv.UnmarshalFile(ManufacturersUpdatedFile, &mu); err != nil {
// 		panic(err)
// 	}
// 	return mu
// }

func getMappedCategory(cat string, trimId string) string{
	newCat:=""
  // fmt.Println(cat)
  ca :=getCaFromCategoryMappingFile()
  for i:=0;i<len(ca);i++{
    if ca[i].Value==cat && ca[i].TrimId==trimId {
      fmt.Println("I am inside mapping category loop")
      fmt.Println(ca[i].Value)
			fmt.Println(ca[i].TrimId)
      fmt.Println(ca[i].MappedCategory)
      newCat = ca[i].MappedCategory
    }}
	return newCat
}

func visit(files *[]string) filepath.WalkFunc {
    return func(path string, info os.FileInfo, err error) error {
        if err != nil {
            log.Fatal(err)
        }
        *files = append(*files, path)
        return nil
    }
}
func charRemover(oldString string, oldChar string, newChar string) string {
	if strings.ContainsAny(oldString, oldChar) {
		oldString = strings.Replace(oldString, oldChar, newChar, -1)
	}

	return oldString
}
// func checkIfOemUpdated(oem string, manuf string)bool{
// 	fmt.Println("checking if oem is updated");
// 	mu :=getUpdatedManufacturersFile()
// 	flag:=false
// 	for i:=0;i<len(mu);i++{
// 		if strings.EqualFold(oem,mu[i].OEM_Name) && strings.EqualFold(manuf,mu[i].ManufacturerName){
// 			flag= true
// 			fmt.Println("json oem:"+oem)
// 			fmt.Println("json manuf:"+manuf)
// 			fmt.Println(flag)
// 			break
// 		}
// 	}
// 	fmt.Println(flag);
// 	return flag
// }

func buildJson(){
	fmt.Println("im here")
	ct:=getCtFromTrimsFile()
	cf:=getCfFromFeaturesFile()
	csd:=getCsdFromSampleDataFile()
	co :=getCoFromOptionsFile()
	cs :=getCsFromSpecsFile()
	cpg :=getCpgFromPhotoGalleryFile()
	d := make(Docs, len(ct), len(ct) )
	// flag := false
	// loop thrugh each trim (model) and build json
	for t := 0; t < len(ct); t++ {
		// flag:= checkIfOemUpdated(ct[t].ManufacturerName, ct[t].ModelName)
		// if flag==false{
			trimId := ct[t].TrimId
			fmt.Println(t)
			fmt.Println("trim id is" , trimId)
			fmt.Println("\n")
			d[t].Meta.Source = "CRS"
			// d[t].Meta.Test = "Test Powersports-sneha-2019-03-07"
			d[t].General.Manufacturer = ct[t].ManufacturerName
			d[t].General.Model = ct[t].ModelName + " "+ct[t].TrimName
			d[t].General.Year = int(math.Round(ct[t].ModelYear))
			d[t].General.Msrp = ct[t].Msrp
			// building general
			fmt.Println("general build");
			for sd := 0; sd < len(csd); sd++ {
				mappedCategory:=""
				if csd[sd].TrimId == trimId {
					if csd[sd].FeatureName == "Identifiers" {
						if csd[sd].AttributeName == "Generic Type (Primary)" {
							 mappedCategory = getMappedCategory(csd[sd].Value, csd[sd].TrimId)
							d[t].General.Category=mappedCategory
							d[t].General.Description = "Description: " + d[t].General.Manufacturer+ " - " +mappedCategory
						}
						if csd[sd].AttributeName == "Manufacturer Country" {
						 	d[t].General.Countries = append(d[t].General.Countries,csd[sd].Value)
					 	} else{
							countries:= []string{"US","CA"}
							d[t].General.Countries = countries
						}
						if csd[sd].AttributeName == "Generic Type 2" {
							d[t].General.Subcategory = csd[sd].Value
						}else {
							d[t].General.Subcategory = d[t].General.Category
						}
						//extracting images
						folder_name:= ""
						image_name := ""
						// if csd[sd].AttributeName == "Photo Name"&& flag == true{
						// 	image_name = csd[sd].Value
					 	// 	folder_name = "800x400"
					 	// 	imgLinkAWS := "https://s3.amazonaws.com/cws-cdn-east/crs-ps-images/CRS+Datafeed+tractor+images+2018-11-29/"+folder_name+"/"+image_name
					 	// 	img := Image{}
					 	// 	img.Src=imgLinkAWS
					 	// 	img.Src=strings.Replace(imgLinkAWS," ","%20",-1)
					 	// 	img.Src = strings.Replace(img.Src," ","%20",-1)
					 	// 	d[t].Images=append(d[t].Images,img)
					 	// }
					  //}

					if csd[sd].AttributeName == "Photo Name" {
							image_name = csd[sd].Value
							folder_name = "800x400"
							imgLinkAWS := "https://s3.amazonaws.com/cws-cdn-east/crs-ps-images/"+folder_name+"/"+image_name
	 						img := Image{}
	 						img.Src=imgLinkAWS
							img.Src=strings.Replace(imgLinkAWS," ","%20",-1)
	 					  img.Src = strings.Replace(img.Src," ","%20",-1)
	 					  d[t].Images=append(d[t].Images,img)
					 }
					 if csd[sd].AttributeName == "Photo Name (Floorplan)"{
					 	 image_name = csd[sd].Value
					 	 folder_name = "Floorplan800"
					 	 imgLinkAWS := "https://s3.amazonaws.com/cws-cdn-east/crs-ps-images/"+folder_name+"/"+image_name
					 	 img := Image{}
					 	 img.Src=imgLinkAWS
					 	 img.Src=strings.Replace(imgLinkAWS," ","%20",-1)
					 	 img.Src = strings.Replace(img.Src," ","%20",-1)
					 	 d[t].Images=append(d[t].Images,img)
					 }
					}
				}
			}

			//extracting images from photogallery file
			fmt.Println("gallery build");
			for pg :=0; pg< len(cpg); pg++{
				if cpg[pg].TrimId == trimId {
					fmt.Println(cpg[pg].PhotoName)
					image_name := cpg[pg].PhotoName
					folder_name := "gallery"
					imgLinkAWS := "https://s3.amazonaws.com/cws-cdn-east/crs-ps-images/"+folder_name+"/"+image_name
				 img := Image{}
				 img.Src=imgLinkAWS
				 img.Src=strings.Replace(imgLinkAWS," ","%20",-1)
				img.Src = strings.Replace(img.Src," ","%20",-1)
				d[t].Images=append(d[t].Images,img)
				}
			}

			fmt.Println("specs build");
			//building specs
		   for s := 0; s < len(cs); s++ {
				 package_name:=""
				  if cs[s].TrimId == trimId {
						engineParentNames := []string{"Engine", "Carburetion","Transmission"}
						measurementsParentNames := []string{"Dimensions","Awning","Cargo Area Dimensions","Exterior Cargo Deck","Measurements"}
						bodyParentNames := []string{"Wheels","Tires","Brakes","Seat","Construction","Seat Specifications","Front Suspension","Rear Suspension"}
						operationalParentNames := []string{"Capacities","Performance","Holding Tanks","Propane Tank(s)","Air Conditioning","Water Heater Tank","Cargo Area Auxiliary Gas Tank","Performance"}
						weightsParentNames := []string{"Weight","Rear Hitch"}
				   	if in_array(cs[s].FeatureName,engineParentNames){
						 	// fmt.Println("in Engine "+  cs[s].FeatureName)
							if (cs[s].AttributeName != "NA"){
							Labelv := cs[s].AttributeName
							package_name=cs[s].PackageTitle
						  name := strings.Title(Labelv)
							firstL:=strings.ToLower(string(name[0]))
							remL:=name[1:]
						 	name = strings.Replace(firstL+remL, " ","", -1)
							name=replaceSpecialCharacters(name)
						  Descv:= cs[s].Value
						 	 s := Specs{}
							 s.Label=Labelv
							 s.Desc=Descv
							 if d[t].Engine == nil {
								 d[t].Engine=make(map[string]Specs)
								 d[t].Engine[name]=s
						 	}else{
								_, ok := 	d[t].Engine[name]
								if ok{
								 	name = name+" - "+package_name
									name = strings.Replace(name,"."," ",-1)

									name=replaceSpecialCharacters(name)
								  s.Label=Labelv 	+ " - "+package_name
									s.Label = strings.Replace(s.Label,"."," ",-1)
								 	d[t].Engine[name]=s
								}else{
								 	d[t].Engine[name]=s
								}
							}
						}
		      }else if in_array(cs[s].FeatureName,measurementsParentNames){
						// fmt.Println("in Dimensions")
						if (cs[s].AttributeName != "NA"){
							Labelv := cs[s].AttributeName
							package_name:=cs[s].PackageTitle
						  name := strings.Title(Labelv)
							firstL:=strings.ToLower(string(name[0]))
							remL:=name[1:]
						 	name = strings.Replace(firstL+remL, " ","", -1)
							name=replaceSpecialCharacters(name)
							Descv:= cs[s].Value
						 	 s := Specs{}
							 s.Label=Labelv
							 s.Desc=Descv
							 if d[t].Measurements == nil {
								 d[t].Measurements=make(map[string]Specs)
								 d[t].Measurements[name]=s
						 	}else{
								_, ok := 	d[t].Measurements[name]
								if ok{
								 	name = name+" - "+package_name
									name = strings.Replace(name,"."," ",-1)
									name=replaceSpecialCharacters(name)
									s.Label=Labelv 	+ " - "+package_name
									s.Label = strings.Replace(s.Label,"."," ",-1)
								 	d[t].Measurements[name]=s
								}else{
							 	d[t].Measurements[name]=s
							}
						 }
						}
					}else if in_array(cs[s].FeatureName,bodyParentNames){
					// fmt.Println("in Body")
						if (cs[s].AttributeName != "NA"){
							Labelv := cs[s].AttributeName
							package_name:=cs[s].PackageTitle
						  name := strings.Title(Labelv)
							firstL:=strings.ToLower(string(name[0]))
							remL:=name[1:]
						 	name = strings.Replace(firstL+remL, " ","", -1)
							name=replaceSpecialCharacters(name)
							Descv:= cs[s].Value
						 	 s := Specs{}
							 s.Label=Labelv
							 s.Desc=Descv
							 if d[t].Body == nil {
								 d[t].Body=make(map[string]Specs)
								 d[t].Body[name]=s
						 	}else{
								_, ok := 	d[t].Body[name]
								if ok{
								 	name = name+" - "+package_name
									name = strings.Replace(name,"."," ",-1)
									name=replaceSpecialCharacters(name)
									s.Label=Labelv 	+ " - "+package_name
									s.Label = strings.Replace(s.Label,"."," ",-1)
								 	d[t].Body[name]=s
								}else{
							 	d[t].Body[name]=s
							 }
							}
						}
					}else if in_array(cs[s].FeatureName,operationalParentNames){
							// fmt.Println("in Operational")
							if (cs[s].AttributeName != "NA"){
								Labelv := cs[s].AttributeName
								package_name:=cs[s].PackageTitle
								name := strings.Title(Labelv)
								firstL:=strings.ToLower(string(name[0]))
								remL:=name[1:]
								name = strings.Replace(firstL+remL, " ","", -1)
								name=replaceSpecialCharacters(name)
								Descv:= cs[s].Value
								 s := Specs{}
								 s.Label=Labelv
								 s.Desc=Descv
								 if d[t].Operational == nil {
									 d[t].Operational=make(map[string]Specs)
									 d[t].Operational[name]=s
								}else{
									_, ok := 	d[t].Operational[name]
									if ok{
									 	name = name+" - "+package_name
										name = strings.Replace(name,"."," ",-1)
									  name=replaceSpecialCharacters(name)
										s.Label=Labelv 	+ " - "+package_name
										s.Label = strings.Replace(s.Label,"."," ",-1)
									 	d[t].Operational[name]=s
									}else{
									d[t].Operational[name]=s
								}
							}
						}
					}else if in_array(cs[s].FeatureName,weightsParentNames){
							// fmt.Println("in Weights")
							if (cs[s].AttributeName != "NA"){
								Labelv := cs[s].AttributeName
								package_name:=cs[s].PackageTitle
								name := strings.Title(Labelv)
								firstL:=strings.ToLower(string(name[0]))
								remL:=name[1:]
								name = strings.Replace(firstL+remL, " ","", -1)
								name=replaceSpecialCharacters(name)
								Descv:= cs[s].Value
								 s := Specs{}
								 s.Label=Labelv
								 s.Desc=Descv
								 if d[t].Weights == nil {
									 d[t].Weights=make(map[string]Specs)
									 d[t].Weights[name]=s
								}else{
									_, ok := 	d[t].Weights[name]
									if ok{
									 	name = name+"-"+package_name
										name = strings.Replace(name,"."," ",-1)
										name=replaceSpecialCharacters(name)
										s.Label=Labelv 	+ " - "+package_name
										s.Label = strings.Replace(s.Label,"."," ",-1)
									 	d[t].Weights[name]=s
									}else{
									d[t].Weights[name]=s
								}
							}
						}
					}else if  cs[s].FeatureName == "Hydraulics"{
							// fmt.Println("in Hydraulics")
							if (cs[s].AttributeName != "NA"){
		 					Labelv := cs[s].AttributeName
							package_name:=cs[s].PackageTitle
		 				  name := strings.Title(Labelv)
		 					firstL:=strings.ToLower(string(name[0]))
		 					remL:=name[1:]
		 				 	name = strings.Replace(firstL+remL, " ","", -1)
							name=replaceSpecialCharacters(name)
		 					Descv:= cs[s].Value
		 				 	 s := Specs{}
		 					 s.Label=Labelv
		 					 s.Desc=Descv
		 					 if d[t].Hydraulics == nil {
		 						 d[t].Hydraulics=make(map[string]Specs)
		 						 d[t].Hydraulics[name]=s
		 				 	}else{
								_, ok := 	d[t].Hydraulics[name]
								if ok{
									name = name+" - "+package_name
									name = strings.Replace(name,"."," ",-1)
									name=replaceSpecialCharacters(name)
									s.Label=Labelv 	+ " - "+package_name
									s.Label = strings.Replace(s.Label,"."," ",-1)
									d[t].Hydraulics[name]=s
								}else{
		 					 	d[t].Hydraulics[name]=s
		 					}
						 }
		 				}
					}	else if  cs[s].FeatureName == "Electrical"{
							// fmt.Println("in Electrical")
							if (cs[s].AttributeName != "NA"){
		 					Labelv := cs[s].AttributeName
							package_name:=cs[s].PackageTitle
		 				  name := strings.Title(Labelv)
		 					firstL:=strings.ToLower(string(name[0]))
		 					remL:=name[1:]
		 				 	name = strings.Replace(firstL+remL, " ","", -1)
							name=replaceSpecialCharacters(name)
		 					Descv:= cs[s].Value
		 				 	 s := Specs{}
		 					 s.Label=Labelv
		 					 s.Desc=Descv
		 					 if d[t].Electrical == nil {
		 						 d[t].Electrical=make(map[string]Specs)
		 						 d[t].Electrical[name]=s
		 				 	}else{
								_, ok := 	d[t].Electrical[name]
								if ok{
									name = name+" - "+package_name
									name = strings.Replace(name,"."," ",-1)
									name=replaceSpecialCharacters(name)
									s.Label=Labelv 	+ " - "+package_name
										s.Label = strings.Replace(s.Label,"."," ",-1)
									d[t].Electrical[name]=s
								}else{
		 					 	d[t].Electrical[name]=s
		 					}
						 }
		 				}
					}else{
						// fmt.Println("in Other")
							if (cs[s].AttributeName != "NA"){
		 					Labelv := cs[s].AttributeName
							package_name:=cs[s].PackageTitle
		 				  name := strings.Title(Labelv)
		 					firstL:=strings.ToLower(string(name[0]))
		 					remL:=name[1:]
		 				 	name = strings.Replace(firstL+remL, " ","", -1)
							name=replaceSpecialCharacters(name)
		 					Descv:= cs[s].Value
		 				 	 s := Specs{}
		 					 s.Label=Labelv
		 					 s.Desc=Descv
		 					 if d[t].Other == nil {
		 						 d[t].Other=make(map[string]Specs)
		 						 d[t].Other[name]=s
		 				 	}else{
								_, ok := 	d[t].Other[name]
								if ok{
									name = name+" - "+package_name
									name = strings.Replace(name,"."," ",-1)
									name=replaceSpecialCharacters(name)
									s.Label=Labelv 	+ " - "+package_name
									s.Label = strings.Replace(s.Label,"."," ",-1)
									d[t].Other[name]=s
								}else{
		 					 	d[t].Other[name]=s
		 					}
						}
						}
		 				}
			     }
			}

			fmt.Println("features build");
			//extracting features
			for f := 0; f < len(cf); f++ {
				 if cf[f].TrimId == trimId {
					 	if !in_array(cf[f].FeatureName,d[t].Features) {

			        d[t].Features = append(d[t].Features,cf[f].FeatureName)
						}
				}
			}

			fmt.Println("options build");
			//extracting options
			for o := 0; o < len(co); o++ {
				 if co[o].TrimId == trimId {
					if !in_array(co[o].FeatureName,d[t].Options) {
			        d[t].Options = append(d[t].Options,co[o].FeatureName)
						}
				}
			}
			out, _ := json.Marshal(d)
			err := ioutil.WriteFile("./out.json", out, 0644)
			if err != nil {
				fmt.Println(err)
		  }
	//	}
	}
}

func moveImagesBasedOnManuf(){
		var files []string
		err := copy.Copy("images_after_delete", "images_after_delete_Copy")
		if err != nil {
				panic(err)
		}
		err = filepath.Walk("images_after_delete_Copy", visit(&files))
		if err != nil {
				panic(err)
		}
		for _, file := range files {
			os.Mkdir("sorted_images", 0777)
			if (!strings.Contains(file,"ColorSwatches")){
				splitPathArr:=strings.Split(file,"/")
				if(len(splitPathArr) ==3){
						splitFilenameArr:=strings.Split(splitPathArr[2],"_")
						if (len(splitFilenameArr) >=4){
								p:="sorted_images/"+splitFilenameArr[1]
								if _, err := os.Stat(p); os.IsNotExist(err) {
								    os.Mkdir(p, 0777)
								}
								p1:="sorted_images/"+splitFilenameArr[1]+"/"+splitFilenameArr[2]
								if _, err := os.Stat(p1); os.IsNotExist(err) {
										os.Mkdir(p1, 0777)
								}
								fName:=""
								if(len(splitFilenameArr) ==4){
										fName=splitFilenameArr[3]
								}else{
									for i:=3;i<len(splitFilenameArr);i++{
										if(fName==""){
											fName=splitFilenameArr[i]
										}else{
											fName=fName+"_"+splitFilenameArr[i]
										}
									}
								}
			      		err := os.Rename(file,p1+"/"+fName)
								if err != nil {
										panic(err)
								}
						}else if(len(splitFilenameArr) ==3){
							p:="sorted_images/"+splitFilenameArr[1]
							if _, err := os.Stat(p); os.IsNotExist(err) {
									os.Mkdir(p, 0777)
							}
							err := os.Rename(file,p+"/"+splitFilenameArr[2])
							if err != nil {
									panic(err)
							}
						}
				}
			}else{
					splitPathArr:=strings.Split(file,"/")
					if (len(splitPathArr)>3){
						oem:=splitPathArr[2]
						fileName:=splitPathArr[3]
						p:="sorted_images/"+oem
						if _, err := os.Stat(p); os.IsNotExist(err) {
								os.Mkdir(p, 0777)
						}
						err := os.Rename(file,p+"/"+fileName)
						if err != nil {
								panic(err)
						}
					}

			}
		}
		os.RemoveAll("images_after_delete_Copy")
}

func getOnly2018Images(){
	err := copy.Copy("all_images", "images_after_delete")
	if err != nil {
			panic(err)
	}
	var files []string
	err = filepath.Walk("images_after_delete", visit(&files))
	if err != nil {
			panic(err)
	}
	for _, file := range files {
		if(strings.Contains(file,"2018") || strings.Contains(file,"ColorSwatches") ){
		}else{
			os.Remove(file)
		}
	}
}

func replaceSpecialCharacters(name string) string{
	name = charRemover(name, "(", "")
	name = charRemover(name, ")", "")
	name = charRemover(name, "-", "")
	name = charRemover(name, "_", "")
	name = charRemover(name, ".", "")
	name = charRemover(name, "@", "")
	name = charRemover(name, "&", "And")
	name = charRemover(name, ",", "")
	name = charRemover(name, "/", "")
	name = charRemover(name, " ", "")
	return name
}
func getPackageName(package_id string) string{
		cp :=getCpFromPackagesFile()
		package_name:=""
		//extracting Package Title
		for p := 0; p < len(cp); p++ {
			 if cp[p].PackageId == package_id {
					package_name = cp[p].PackageTitle
			}
		}
		return package_name
}
func main() {
	getAPI()
	buildJson()
	postJson()
}
