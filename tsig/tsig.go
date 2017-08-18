/* tsig.go
Timberland Signature project

*/

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
	//"reflect"
	"regexp"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/sheets/v4"
)

/* the token will need to be passed in. */
type CliArgs struct {
	fullname  string
	job_title string
	address   string
	csz       string
	telephone string
	website   string
	facebook  string
	linkedin  string
	twitter   string
}

type SocialMedia struct {
	LinkedIn_URL   string
	Website_URL    string
	Twitter_Handle string
	Facebook       string
}

type ExtraContactInfo struct {
	Address_12     string
	City_State_Zip string
	Job_Title      string
	Telephone      string
}

var _args CliArgs

func main() {

	ctx := context.Background()

	b, err := ioutil.ReadFile("client-tsig.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/tsig.googleapis.com.json
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope, admin.AdminDirectoryUserReadonlyScope, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)

	var _defaults CliArgs
	sheetsSrv, sheetsErr := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets Client %v", sheetsErr)
	}

	spreadsheetId := "1IeQWuSbJBiqful-0-hxriPJsWvja7Nlk1B1gtVijRdc"
	readRange := "A2:I"
	resp, err := sheetsSrv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet. %v", err)
	}
	if len(resp.Values) > 0 {
		for _, row := range resp.Values {
			// grab default data
			_defaults.fullname = fmt.Sprintf("%s",row[0])
			_defaults.job_title = fmt.Sprintf("%s",row[1])
			_defaults.address = fmt.Sprintf("%s",row[2])
			_defaults.csz = fmt.Sprintf("%s",row[3])
			_defaults.telephone = fmt.Sprintf("%s",row[4])
			_defaults.website = fmt.Sprintf("%s",row[5])
			_defaults.facebook = fmt.Sprintf("%s",row[6])
			_defaults.linkedin = fmt.Sprintf("%s",row[7])
			_defaults.twitter = fmt.Sprintf("%s",row[8])
		}
	} else {
		fmt.Print("No data found.")
	}

	srv, err := admin.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve directory Client %v", err)
	}

	r, err := srv.Users.List().Customer("my_customer").Projection("full").
		MaxResults(300).OrderBy("email").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve users in domain.", err)
	}

	if len(r.Users) == 0 {
		fmt.Print("No users found.\n")
	} else {
		for _, u := range r.Users {
			//match, _ := regexp.MatchString("dhernandez", u.PrimaryEmail)
			//if match == true {
			if u.CustomSchemas != nil {
				sig := dumbStr()
				_args.fullname = u.Name.FullName

				// Get the Contact Info for the signature
				contactstr := u.CustomSchemas["Extra_Contact_Info"]
				var contact ExtraContactInfo
				contactErr := json.Unmarshal(contactstr, &contact)
				if contactErr != nil {
					panic(contactErr)
				}
				_args.job_title = contact.Job_Title
				_args.address = contact.Address_12
				_args.csz = contact.City_State_Zip
				_args.telephone = contact.Telephone

				// Get the Social Media Info for the signature
				socialstr := u.CustomSchemas["Social_Media"]
				var social SocialMedia
				socialErr := json.Unmarshal(socialstr, &social)
				if socialErr != nil {
					fmt.Printf("%v", socialErr)
				}
				_args.website = social.Website_URL
				_args.facebook = social.Facebook
				_args.linkedin = social.LinkedIn_URL
				_args.twitter = social.Twitter_Handle
				//fmt.Println("CustomSchemas.Social_Media", social)
				fmt.Println("")
				//fmt.Printf("%s (%s)\n", u.PrimaryEmail, u.Name.FullName)

				if _args.fullname == "" {
					_args.fullname = _defaults.fullname
				}
				if _args.job_title == "" {
					_args.job_title = _defaults.job_title
				}
				if _args.address == "" {
					_args.address = _defaults.address
				}
				if _args.csz == "" {
					_args.csz = _defaults.csz
				}
				if _args.telephone == "" {
					_args.telephone = _defaults.telephone
				}
				if _args.website == "" {
					_args.website = _defaults.website
				}
				if _args.facebook == "" {
					_args.facebook = _defaults.facebook
				}
				if _args.linkedin == "" {
					_args.linkedin = _defaults.linkedin
				}
				if _args.twitter == "" {
					_args.twitter = _defaults.twitter
				}

				var subs map[string]string
				subs = make(map[string]string)
				populate(&subs)

				for k := range subs {
					r, err := regexp.Compile(k)
					if err != nil {
						log.Fatal("Something broke: %s", err)
					}
					if subs[k] != "" {
						sig = r.ReplaceAllString(sig, subs[k])
					}
				}
				fmt.Println(sig)
			}
		}
	}
}

/* --- METHODS --- */
// Subs methods

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
		url.QueryEscape("tsig.googleapis.com.json")), err
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

func populate(a *map[string]string) {
	*a = map[string]string{
		"@FULLNAME":               _args.fullname,
		"@JOB_TITLE":              _args.job_title,
		"@ADDRESS_LINE_1_COMMA_2": _args.address,
		"@CITY_STATE_ZIP":         _args.csz,
		"@TELEPHONE":              _args.telephone,
		"@WEBSITE":                _args.website,
		"@FACEBOOK":               _args.facebook,
		"@LINKED_IN":              _args.linkedin,
		"@TWITTER":                _args.twitter,
	}
}

func dumbSlice() []string {
	// Sorry about this
	dumb := []string{"@FULLNAME", "@JOB_TITLE", "@ADDRESS_LINE_1_COMMA_2", "@CITY_STATE_ZIP", "@TELEPHONE", "@WEBSITE", "@FACEBOOK", "@LINKED_IN", "@TWITTER"}
	return dumb
}

func dumbStr() string {
	// Sorry about this.
	return "<div dir=\"ltr\" font-family=\"sans-serif\"><br clear=\"all\"><div><div class=\"gmail_signature\" data-smartmail=\"gmail_signature\"><div dir=\"ltr\" style=\"font-size:12.8px\"><b><font color=\"#6aa84f\" size=\"4\">@FULLNAME</font></b><br><font color=\"#666666\">@JOB_TITLE        <br>        @ADDRESS_LINE_1_COMMA_2        <br>@CITY_STATE_ZIP  </font></div><div dir=\"ltr\" style=\"font-size:12.8px\"><font color=\"#666666\">t. @TELEPHONE<br></font></div><div dir=\"ltr\" style=\"font-size:12.8px\"><font color=\"#666666\"><br></font></div><div dir=\"ltr\" style=\"font-size:12.8px\"><font color=\"#6aa84f\"><b><span style=\"font-family:Calibri;font-size:14.6667px\">Top Work Places |</span><span style=\"font-family:Calibri;font-size:14.6667px\"><i>StarTribune</i></span></b></font></div><div dir=\"ltr\" style=\"font-size:12.8px\"><p><a href=\"@WEBSITE\" target=\"_blank\"><img src=\"https://docs.google.com/uc?export=download&amp;id=0B2kRekvqies1a05SSDBITmpzMFk&amp;revid=0B2kRekvqies1dWdEYzBWZ3d4SmpaZ25JSlZHS0Rrak9tQkdvPQ\" width=\"200\" height=\"35\" alt=\"Timberland Partners\"></a></p><p><a href=\"@FACEBOOK/\" target=\"_blank\"><img src=\"https://docs.google.com/a/timberlandpartners.com/uc?id=0B2kRekvqies1WnVtMG0tZDh1azg&amp;export=download\"></a><a href=\"@LINKED_IN/\" target=\"_blank\"><img src=\"https://docs.google.com/a/timberlandpartners.com/uc?id=0B2kRekvqies1QlVWOWVjY0d2V00&amp;export=download\"></a><a href=\"@TWITTER\" target=\"_blank\"><img src=\"https://docs.google.com/a/timberlandpartners.com/uc?id=0B2kRekvqies1aVkxRmJZLU9MQ3M&amp;export=download\"></a></p></div></div></div></div></div>"
}
