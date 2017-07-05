/* tsig.go
Timberland Signature project

*/

package main

import (
	"flag"
	"fmt"
	"log"
	//"os"
	"regexp"
)

/* the token will need to be passed in. Also, use a verbose flag */
type CliArgs struct {
	verbose bool
	token   string
	fullname string
	job_title string
	address string
	csz string
	telephone string
	website string
	facebook string
	linkedin string
	twitter string
}

var args CliArgs

func main() {
	args.process()

	sig := dumbStr()

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

	//log.Println(r.MatchString("peach"))
}

/* --- METHODS --- */
// Subs methods
func populate(a *map[string]string) {
	*a = map[string]string{
		"@FULLNAME" : args.fullname,
		"@JOB_TITLE" : args.job_title,
		"@ADDRESS_LINE_1_COMMA_2" : args.address,
		"@CITY_STATE_ZIP" : args.csz,
		"@TELEPHONE" : args.telephone,
		"@WEBSITE" : args.website,
		"@FACEBOOK" : args.facebook,
		"@LINKED_IN" : args.linkedin,
		"@TWITTER" : args.twitter,
	}
}

// CliArgs methods
func (a *CliArgs) process() {
	flag.BoolVar(&a.verbose, "v", false, "turn on verbose messages")
	flag.StringVar(&a.token, "token", "", "api process token")
	flag.StringVar(&a.fullname, "name", "", "full name appearing in the signature")
	flag.StringVar(&a.job_title, "job", "", "job title appearing in the signature")
	flag.StringVar(&a.address, "address", "", "address lines 1/2, comma separated")
	flag.StringVar(&a.csz, "city", "", "really the city, state and zip")
	flag.StringVar(&a.telephone, "telephone", "", "phone number for the signature")
	flag.StringVar(&a.website, "website", "", "web site for the signature")
	flag.StringVar(&a.facebook, "facebook", "", "facebook for the signature")
	flag.StringVar(&a.linkedin, "linkedin", "", "linkedin for the signature")
	flag.StringVar(&a.twitter, "twitter", "", "twitter for the signature")

	flag.Parse()
}

func (a *CliArgs) String() string {
	return fmt.Sprintf("(%v, %v)", a.verbose, a.token)
}

func dumbSlice() []string {
	// Sorry about this
	dumb := []string{"@FULLNAME", "@JOB_TITLE", "@ADDRESS_LINE_1_COMMA_2", "@CITY_STATE_ZIP", "@TELEPHONE", "@WEBSITE", "@FACEBOOK", "@LINKED_IN", "@TWITTER"}
	return dumb
}

func dumbStr() string {
	// Sorry about this.
	return "<div dir=\"ltr\"><br clear=\"all\"><div><div class=\"gmail_signature\" data-smartmail=\"gmail_signature\"><div dir=\"ltr\" style=\"font-size:12.8px\"><b><font color=\"#6aa84f\" size=\"4\">@FULLNAME</font></b><br><font color=\"#666666\">@JOB_TITLE        <br>        @ADDRESS_LINE_1_COMMA_2        <br>@CITY_STATE_ZIP  </font></div><div dir=\"ltr\" style=\"font-size:12.8px\"><font color=\"#666666\">t. @TELEPHONE<br></font></div><div dir=\"ltr\" style=\"font-size:12.8px\"><font color=\"#666666\"><br></font></div><div dir=\"ltr\" style=\"font-size:12.8px\"><font color=\"#6aa84f\"><b><span style=\"font-family:Calibri;font-size:14.6667px\">Top Work Places |</span><span style=\"font-family:Calibri;font-size:14.6667px\"><i>StarTribune</i></span></b></font></div><div dir=\"ltr\" style=\"font-size:12.8px\"><p><a href=\"@WEBSITE\" target=\"_blank\"><img src=\"https://docs.google.com/uc?export=download&amp;id=0B2kRekvqies1a05SSDBITmpzMFk&amp;revid=0B2kRekvqies1dWdEYzBWZ3d4SmpaZ25JSlZHS0Rrak9tQkdvPQ\" width=\"200\" height=\"35\" alt=\"Timberland Partners\"></a></p><p><a href=\"@FACEBOOK/\" target=\"_blank\"><img src=\"https://docs.google.com/a/timberlandpartners.com/uc?id=0B2kRekvqies1WnVtMG0tZDh1azg&amp;export=download\"></a><a href=\"@LINKED_IN/\" target=\"_blank\"><img src=\"https://docs.google.com/a/timberlandpartners.com/uc?id=0B2kRekvqies1QlVWOWVjY0d2V00&amp;export=download\"></a><a href=\"@TWITTER\" target=\"_blank\"><img src=\"https://docs.google.com/a/timberlandpartners.com/uc?id=0B2kRekvqies1aVkxRmJZLU9MQ3M&amp;export=download\"></a></p></div></div></div></div></div>"
}
