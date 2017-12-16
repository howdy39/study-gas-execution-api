package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/script/v1"
)

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("script-go-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func main() {
	ctx := context.Background()

	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/script-go-quickstart.json
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/drive")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	scriptId := "Mn_YoQoNj_iufS59FmWsY-JgYYRqhh78z"
	client := getClient(ctx, config)

	// Generate a service object.
	srv, err := script.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve script Client %v", err)
	}

	// Create an execution request object.
	req := script.ExecutionRequest{Function: "getFoldersUnderRoot"}

	// Make the API request.
	resp, err := srv.Scripts.Run(scriptId, &req).Do()
	if err != nil {
		// The API encountered a problem before the script started executing.
		log.Fatalf("Unable to execute Apps Script function. %v", err)
	}

	if resp.Error != nil {
		fmt.Printf("%s", resp.Error)
		// The API executed, but the script returned an error.

		// Extract the first (and only) set of error details and cast as a map.
		// The values of this map are the script's 'errorMessage' and
		// 'errorType', and an array of stack trace elements (which also need to
		// be cast as maps).
		//error := resp.Error.Details[0].(map[string]interface{})
		//fmt.Printf("Script error message: %s\n", error["errorMessage"]);

		//if (error["scriptStackTraceElements"] != nil) {
		//	// There may not be a stacktrace if the script didn't start executing.
		//	fmt.Printf("Script error stacktrace:\n")
		//	for _, trace := range error["scriptStackTraceElements"].([]interface{}) {
		//		t := trace.(map[string]interface{})
		//		fmt.Printf("\t%s: %d\n", t["function"], int(t["lineNumber"].(float64)))
		//	}
		//}
	} else {
		// The result provided by the API needs to be cast into the correct type,
		// based upon what types the Apps Script function returns. Here, the
		// function returns an Apps Script Object with String keys and values, so
		// must be cast into a map (folderSet).
		//r := resp.Response.(map[string]interface{})
		json, _ := resp.Response.MarshalJSON()
		fmt.Printf("%s", json)
		//folderSet := r["result"].(map[string]interface{})
		//if len(folderSet) == 0 {
		//	fmt.Printf("No folders returned!\n")
		//} else {
		//	fmt.Printf("Folders under your root folder:\n")
		//	for id, folder := range folderSet {
		//		fmt.Printf("\t%s (%s)\n", folder, id)
		//	}
		//}
	}

}
