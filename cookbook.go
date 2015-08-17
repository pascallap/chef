package chef

import "fmt"
import "os"
import "path/filepath"
import "crypto/md5"
import "encoding/json"
import "encoding/base64"
import "encoding/hex"
import "io"

const filechunk = 8192    // we settle for 8KB

// CookbookService  is the service for interacting with chef server cookbooks endpoint
type CookbookService struct {
	client *Client
}

// CookbookItem represents a object of cookbook file data
type CookbookItem struct {
	Url         string `json:"url"`
	Path        string `json:"path"`
	Name        string `json:"name"`
	Checksum    string `json:"checksum"`
	Specificity string `json:"specificity"`
}

// CookbookItem represents a object of cookbook file data
type CookbookItemPut struct {
	Path        string `json:"path"`
	Name        string `json:"name"`
	Checksum    string `json:"checksum"`
	Specificity string `json:"specificity"`
}

// CookbookListResult is the summary info returned by chef-api when listing
// http://docs.opscode.com/api_chef_server.html#cookbooks
type CookbookListResult map[string]CookbookVersions

// CookbookVersions is the data container returned from the chef server when listing all cookbooks
type CookbookVersions struct {
	Url      string            `json:"url"`
	Versions []CookbookVersion `json:"versions"`
}

// CookbookVersion is the data for a specific cookbook version
type CookbookVersion struct {
	Url     string `json:"url"`
	Version string `json:"version"`
}

// CookbookMeta represents a Golang version of cookbook metadata
type CookbookMeta struct {
	Name            string                 `json:"cookbook_name"`
	Version         string                 `json:"version"`
	Description     string                 `json:"description"`
	LongDescription string                 `json:"long_description"`
	Maintainer      string                 `json:"maintainer"`
	MaintainerEmail string                 `json:"maintainer_email"`
	License         string                 `json:"license"`
	Platforms       map[string]string      `json:"platforms"`
	Depends         map[string]string      `json:"dependencies"`
	Reccomends      map[string]string      `json:"recommendations"`
	Suggests        map[string]string      `json:"suggestions"`
	Conflicts       map[string]string      `json:"conflicting"`
	Provides        map[string]string      `json:"providing"`
	Replaces        map[string]string      `json:"replacing"`
	Attributes      map[string]interface{} `json:"attributes"` // this has a format as well that could be typed, but blargh https://github.com/lob/chef/blob/master/cookbooks/apache2/metadata.json
	Groupings       map[string]interface{} `json:"groupings"`  // never actually seen this used.. looks like it should be map[string]map[string]string, but not sure http://docs.opscode.com/essentials_cookbook_metadata.html
	Recipes         map[string]string      `json:"recipes"`
}

// CookbookMeta represents a Golang version of cookbook metadata
type CookbookMetaPut struct {
	Name            string                 `json:"cookbook_name"`
	Version         string                 `json:"version"`
	Description     string                 `json:"description"`
	LongDescription string                 `json:"long_description"`
	Maintainer      string                 `json:"maintainer"`
	MaintainerEmail string                 `json:"maintainer_email"`
	License         string                 `json:"license"`
	Platforms       map[string]string      `json:"platforms"`
	Depends         map[string]string      `json:"dependencies"`
	Reccomends      map[string]string      `json:"recommendations"`
	Suggests        map[string]string      `json:"suggestions"`
	Conflicts       map[string]string      `json:"conflicting"`
	Provides        map[string]string      `json:"providing"`
	Replaces        map[string]string      `json:"replacing"`
	Attributes      map[string]interface{} `json:"attributes"` // this has a format as well that could be typed, but blargh https://github.com/lob/chef/blob/master/cookbooks/apache2/metadata.json
	Groupings       map[string]interface{} `json:"groupings"`  // never actually seen this used.. looks like it should be map[string]map[string]string, but not sure http://docs.opscode.com/essentials_cookbook_metadata.html
	Recipes         map[string]string      `json:"recipes"`
}


// Cookbook represents the native Go version of the deserialized api cookbook
type Cookbook struct {
	CookbookName string         `json:"cookbook_name"`
	Name         string         `json:"name"`
	Version      string         `json:"version"`
	ChefType     string         `json:"chef_type"`
	Frozen       bool           `json:"frozen?"`
	JsonClass    string         `json:"json_class"`
	Files        []CookbookItem `json:"files"`
	Templates    []CookbookItem `json:"templates"`
	Attributes   []CookbookItem `json:"attributes"`
	Recipes      []CookbookItem `json:"recipes"`
	Definitions  []CookbookItem `json:"definitions"`
	Libraries    []CookbookItem `json:"libraries"`
	Providers    []CookbookItem `json:"providers"`
	Resources    []CookbookItem `json:"resources"`
	RootFiles    []CookbookItem `json:"root_files"`
	Metadata     CookbookMeta   `json:"metadata"`
}

// Cookbook represents the native Go version of the deserialized api cookbook
type CookbookPut struct {
	CookbookName string            `json:"cookbook_name"`
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	ChefType     string            `json:"chef_type"`
	Frozen       bool              `json:"frozen?"`
	JsonClass    string            `json:"json_class"`
	Files        []CookbookItemPut `json:"files"`
	Templates    []CookbookItemPut `json:"templates"`
	Attributes   []CookbookItemPut `json:"attributes"`
	Recipes      []CookbookItemPut `json:"recipes"`
	Definitions  []CookbookItemPut `json:"definitions"`
	Libraries    []CookbookItemPut `json:"libraries"`
	Providers    []CookbookItemPut `json:"providers"`
	Resources    []CookbookItemPut `json:"resources"`
	RootFiles    []CookbookItemPut `json:"root_files"`
	Metadata     CookbookMetaPut      `json:"metadata"`
}

// String makes CookbookListResult implement the string result
func (c CookbookListResult) String() (out string) {
	for k, v := range c {
		out += fmt.Sprintf("%s => %s\n", k, v.Url)
		for _, i := range v.Versions {
			out += fmt.Sprintf(" * %s\n", i.Version)
		}
	}
	return out
}

// String makes Cookbook implement the string result
func (c Cookbook) String() (out string) {
	cc, err := json.MarshalIndent(c,"","    ")
  if err != nil {
     fmt.Println("error:", err)
   }
	return string(cc)
}

// String makes Cookbook implement the string result
func (c CookbookPut) String() (out string) {
	cc, err := json.Marshal(c)
  if err != nil {
     fmt.Println("error:", err)
   }
	return string(cc)
}

func (c Cookbook) ToCookbookPut() (cookbookPut CookbookPut) {
	cc, err := json.Marshal(c)
  if err != nil {
     fmt.Println("error:", err)
   }
   
   var acookbookPut CookbookPut
   
   err = json.Unmarshal(cc,&acookbookPut)
 	if err != nil {
 		fmt.Println("error:", err)
 	}
  
//  fmt.Println(acookbookPut)
  
	return acookbookPut
}


// versionParams assembles a querystring for the chef api's  num_versions
// This is used to restrict the number of versions returned in the reponse
func versionParams(path, numVersions string) string {
	if numVersions == "0" {
		numVersions = "all"
	}

	// need to optionally add numVersion args to the request
	if len(numVersions) > 0 {
		path = fmt.Sprintf("%s?num_versions=%s", path, numVersions)
	}
	return path
}

// Get retruns a CookbookVersion for a specific cookbook
//  GET /cookbooks/name
func (c *CookbookService) Get(name string) (data CookbookVersion, err error) {
	path := fmt.Sprintf("cookbooks/%s", name)
	err = c.client.magicRequestDecoder("GET", path, nil, &data)
	return
}

// GetAvailable returns the versions of a coookbook available on a server
func (c *CookbookService) GetAvailableVersions(name, numVersions string) (data CookbookListResult, err error) {
	path := versionParams(fmt.Sprintf("cookbooks/%s", name), numVersions)
	err = c.client.magicRequestDecoder("GET", path, nil, &data)
	return
}

// GetVersion fetches a specific version of a cookbooks data from the server api
//   GET /cookbook/foo/1.2.3
//   GET /cookbook/foo/_latest
//   Chef API docs: http://docs.opscode.com/api_chef_server.html#id5
func (c *CookbookService) GetVersion(name, version string) (data Cookbook, err error) {
	url := fmt.Sprintf("cookbooks/%s/%s", name, version)
	c.client.magicRequestDecoder("GET", url, nil, &data)
	return
}

// ListVersions lists the cookbooks available on the server limited to numVersions
//   Chef API docs: http://docs.opscode.com/api_chef_server.html#id2
func (c *CookbookService) ListAvailableVersions(numVersions string) (data CookbookListResult, err error) {
	path := versionParams("cookbooks", numVersions)
	err = c.client.magicRequestDecoder("GET", path, nil, &data)
	return
}

// List returns a CookbookListResult with the latest versions of cookbooks available on the server
func (c *CookbookService) List() (CookbookListResult, error) {
	return c.ListAvailableVersions("")
}

// DeleteVersion removes a version of a cook from a server
func (c *CookbookService) Delete(name, version string) (err error) {
	path := fmt.Sprintf("cookbooks/%s", name)
	err = c.client.magicRequestDecoder("DELETE", path, nil, nil)
	return
}

// Download a version of a cookbook from a server
func (c *CookbookService) Download(name, version, destination string) (err error) {
	url := fmt.Sprintf("cookbooks/%s/%s", name, version)
	var data Cookbook
	err = c.client.magicRequestDecoder("GET", url, nil, &data)

	basedir := filepath.Join(destination, name + "-" + version)

  //if _, err := os.Stat(basedir); err == nil {
  //  fmt.Println("There is already a cookbook in the tmp folder for the cookbook: " + name)
  //  fmt.Println("    Delete the directory " + basedir + " and retry!")
  //  os.Exit(1)
  //}

	c.DownloadCookbookItems(data.Attributes,basedir)
	c.DownloadCookbookItems(data.Recipes,basedir)
	c.DownloadCookbookItems(data.Providers,basedir)
	c.DownloadCookbookItems(data.Definitions,basedir)
	c.DownloadCookbookItems(data.Libraries,basedir)
	c.DownloadCookbookItems(data.Files,basedir)
	c.DownloadCookbookItems(data.Templates,basedir)
	c.DownloadCookbookItems(data.RootFiles,basedir)
	c.DownloadCookbookItems(data.Resources,basedir)
	return
}

func (c *CookbookService) DownloadCookbookItems(object []CookbookItem, destination string) (err error){
	for _,cbitems := range object{
		err = os.MkdirAll(filepath.Dir(filepath.Join(destination, cbitems.Path)),0777)
		if err != nil {
			return err
		}
		err := c.client.Download(cbitems.Url, filepath.Join(destination, cbitems.Path))
		if err != nil {
			return err
		}
	}
	return
}

// Upload a Cookbook
//   Chef API docs: https://docs.chef.io/api_chef_server.html#id30
// We should add implementation for chefignore...
// Lazy geneartion of json descriptor of the cookbook ... Use a preprepared one...
func (c *CookbookService) Upload(name, version, source string, cookbookDestriptionJson CookbookPut) (err error) {
  fileList := []string{}
  err = filepath.Walk(filepath.Join(source,name), func(path string, f os.FileInfo, err error) error {
      if !f.IsDir(){
        fileList = append(fileList, path)
      }
      return nil
  })

  fileChecksums := make(map[string]string)
  var checksums []string
  for _, file := range fileList {
      //fmt.Println(file)
      checksum,err := ComputeMd5(file)
      if err != nil{
       return err 
      }
      fileChecksums[checksum] = file
      checksums = append(checksums,checksum)
  }

	// post the new sums/files to the sandbox
	postResp, err := c.client.Sandboxes.Post(checksums)
	if err != nil {
		fmt.Println("error making request: ", err)
		os.Exit(1)
	}

  for respChecksumID, item := range postResp.Checksums{
    if item.Upload {
      iofile,err := os.Open(fileChecksums[respChecksumID])
      defer iofile.Close()

      hexChecksum,_:= hex.DecodeString(respChecksumID)
      checksum64 := base64.StdEncoding.EncodeToString(hexChecksum)

      req, err := c.client.NewRequest("PUT", item.Url, iofile)
      if err !=nil {
        fmt.Println(err)
        return err
      }

      req.Header.Set("content-type", "application/x-binary")
      req.Header.Set("content-md5", checksum64)
      req.Header.Set("accept", "application/json")

      _, err = c.client.Do(req, nil)
      iofile.Close()
      if err !=nil {
        fmt.Println("Error on push cookbookitems: " + err.Error())
        return err
      }
    }    
  }

	box , err := c.client.Sandboxes.Put(postResp.ID)
	if err != nil {
		fmt.Println("Error commiting sandbox: ", err.Error())
    fmt.Println(box)
		os.Exit(1)
	}
  fmt.Println("Sandbox has been commited: ")
  
  c.Put(name,version,cookbookDestriptionJson)
  
  return

}

func ComputeCookbookItemFromFile(filePath, checksum, basePath string)(cookbookitem CookbookItem, err error){
  cookbookitem.Checksum = checksum
  cookbookitem.Name = filepath.Base(filePath)
  if basePath != ""{
    relPath,err := filepath.Rel(basePath, filePath)
    if err != nil{
      fmt.Println("Oups")
      return cookbookitem, err
    }
    cookbookitem.Path = relPath
  }else
  {
    cookbookitem.Path = filePath
  }
  return cookbookitem ,err
}


func ComputeMd5(filePath string) (string, error) {
  var result []byte
  file, err := os.Open(filePath)
  if err != nil {
    return fmt.Sprintf("%x",result), err
  }
  defer file.Close()
 
  hash := md5.New()
  if _, err := io.Copy(hash, file); err != nil {
    return fmt.Sprintf("%x",result), err
  }
 
  return fmt.Sprintf("%x",hash.Sum(result)), nil
}


// ListVersions lists the cookbooks available on the server limited to numVersions
//   Chef API docs: http://docs.opscode.com/api_chef_server.html#id2
func (c *CookbookService) Put(name, version string, cookbook CookbookPut) (err error) {
  url := fmt.Sprintf("cookbooks/%s/%s", name, version)
  
  fmt.Println("************ Upload "+name)
  
  body, err :=  JSONReader(cookbook)
	if err != nil {
    fmt.Println("Json Reading err" + err.Error())
		return
	}

	err = c.client.magicRequestDecoder("PUT", url, body, nil)
	if err != nil {
    fmt.Println("Error on magic decoder" + err.Error())
		return
	}
	return
}

