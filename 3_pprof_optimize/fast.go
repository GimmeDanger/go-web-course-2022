package hw3

import (
	"github.com/mailru/easyjson"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	// "log"
)

type UserBrowser struct {
	Browsers []string `json:"browsers"`
	Company  string   `json:"company"`
	Country  string   `json:"country"`
	Email    string   `json:"email"`
	Job      string   `json:"job"`
	Name     string   `json:"name"`
	Phone    string   `json:"phone"`
}

func browserChecker(browser string, seenBrowsers map[string]bool, isAndroid *bool, isMSIE *bool) {
	insertCandidate := false
	if strings.Contains(browser, "Android") {
		*isAndroid = true
		insertCandidate = true
	}
	if strings.Contains(browser, "MSIE") {
		*isMSIE = true
		insertCandidate = true
	}		
	if insertCandidate {
		if _, ok := seenBrowsers[browser]; !ok {
			seenBrowsers[browser] = true
		}
	}
}

func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	r := regexp.MustCompile("@")
	seenBrowsers := make(map[string]bool, 0)
	foundUsers := ""

	lines := strings.Split(string(fileContents), "\n")

	for i, line := range lines {
		user := UserBrowser{}
		err := easyjson.Unmarshal([]byte(line), &user)
		if err != nil {
			panic(err)
		}

		isAndroid := false
		isMSIE := false
		for _, browser := range user.Browsers {
			browserChecker(browser, seenBrowsers, &isAndroid, &isMSIE)
		}
		if !(isAndroid && isMSIE) {
			continue
		}

		email := r.ReplaceAllString(user.Email, " [at] ")
		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user.Name, email)
	}

	fmt.Fprintln(out, "found users:\n"+foundUsers)
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}
